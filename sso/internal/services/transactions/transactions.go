package transactions

import (
	"context"
	"errors"
	"log/slog"
	"time"

	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/config"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/lib/logger/sl"
	"sso/sso/internal/repository"
)

type TransactionsManagement interface {
	Reserve(ctx context.Context, userID int64, appID int32, amount int64, idempotentKey string, description string, ttl time.Duration) (*models.Transaction, error)
}

type Transactions struct {
	log                   *slog.Logger
	cfg                   *config.Config
	transactionRepository repository.RTransactionRepository
	dbRepository          repository.TransactionRepository
}

var (
	InternalError      = errors.New("internal error")
	AlreadyInProgress  = errors.New("transaction already in progress")
	InsufficientFunds  = errors.New("insufficient funds")
	ReservationExpired = errors.New("reservation expired")
)

func New(
	log *slog.Logger,
	cfg *config.Config,
	rTransactionRepository repository.RTransactionRepository,
	dbRepository repository.TransactionRepository,
) TransactionsManagement {
	return &Transactions{
		log:                   log,
		cfg:                   cfg,
		transactionRepository: rTransactionRepository,
		dbRepository:          dbRepository,
	}
}

// Reserve - резервирование средств с идемпотентностью
//
// Алгоритм:
// 1. Проверка Redis:
//   - success → вернуть сохранённый response
//   - processing → вернуть AlreadyInProgress
//   - нет ключа → идём дальше
//
// 2. Сохраняем в Redis со статусом processing
// 3. Выполняем reserve в одной DB-транзакции:
//   - UPDATE users WHERE (balance - reserve_balance) >= amount
//   - INSERT transactions с idempotency_key
//
// 4. При unique_violation:
//   - читаем существующую транзакцию
//   - проверяем expires_at
//   - возвращаем ТОТ ЖЕ успешный ответ
//
// 5. При успехе: Redis → success
// 6. При ошибке: Redis → failed (короткий TTL)
func (t *Transactions) Reserve(
	ctx context.Context,
	userID int64,
	appID int32,
	amount int64,
	idempotentKey string,
	description string,
	ttl time.Duration,
) (*models.Transaction, error) {
	const op = "transactions.Reserve"

	log := t.log.With(
		slog.String("op", op),
		slog.String("key", idempotentKey),
		slog.Int64("user_id", userID),
		slog.Int64("amount", amount),
	)

	// 1. Проверка Redis
	existing, err := t.transactionRepository.GetIdempotentKey(ctx, idempotentKey)
	if err != nil {
		log.Error("failed to check idempotent key in redis", sl.Err(err))
		return nil, InternalError
	}

	if existing != nil {
		switch existing.Status {
		case ssov1.TransactionStatus_TRANSACTION_SUCCESS:
			// Уже успешно выполнено - возвращаем из БД
			log.Info("found successful transaction in redis, returning existing")
			return t.getExistingTransaction(ctx, idempotentKey, log)

		case ssov1.TransactionStatus_TRANSACTION_PENDING:
			// В процессе выполнения
			log.Info("transaction already in progress")
			return nil, AlreadyInProgress

		case ssov1.TransactionStatus_TRANSACTION_FAILED:
			// Ранее failed - можно попробовать снова, удаляем ключ
			log.Info("previous attempt failed, retrying")
			_ = t.transactionRepository.DeleteIdempotentKey(ctx, idempotentKey)
		}
	}

	// 2. Сохраняем в Redis со статусом processing (pending)
	redisTransaction := &models.RedisTransaction{
		IdempotentKey: idempotentKey,
		Status:        ssov1.TransactionStatus_TRANSACTION_PENDING,
		OperationType: ssov1.TransactionType_TRANSACTION_TYPE_RESERVE,
		UserID:        userID,
		Amount:        amount,
		CreatedAt:     time.Now(),
	}

	if err := t.transactionRepository.SaveIdempotentKey(ctx, redisTransaction); err != nil {
		log.Error("failed to save idempotent key to redis", sl.Err(err))
		return nil, InternalError
	}

	// 3. Выполняем reserve в БД
	expiresAt := time.Now().Add(ttl)
	transaction, err := t.dbRepository.Reserve(ctx, userID, appID, amount, idempotentKey, description, expiresAt)

	if err != nil {
		// Обработка ошибок
		if errors.Is(err, repository.ErrInsufficientFunds) {
			log.Warn("insufficient funds")
			t.setFailed(ctx, idempotentKey, log)
			return nil, InsufficientFunds
		}

		if errors.Is(err, repository.ErrUserNotFound) {
			log.Warn("user not found")
			t.setFailed(ctx, idempotentKey, log)
			return nil, repository.ErrUserNotFound
		}

		log.Error("failed to reserve in db", sl.Err(err))
		t.setFailed(ctx, idempotentKey, log)
		return nil, InternalError
	}

	// 4. Проверяем expires ТОЛЬКО для existing транзакции (из unique_violation)
	// Если транзакция создана давно (> 1 сек назад) - это existing
	// НЕ отменяем здесь - это задача cron worker
	isExisting := time.Since(transaction.CreatedAt) > time.Second
	if isExisting && transaction.ExpiresAt != nil && time.Now().After(*transaction.ExpiresAt) {
		log.Warn("existing reservation expired", slog.Time("expires_at", *transaction.ExpiresAt))
		// Удаляем PROCESSING из Redis, чтобы можно было повторить с новым ключом
		_ = t.transactionRepository.DeleteIdempotentKey(ctx, idempotentKey)
		return nil, ReservationExpired
	}

	// 5. Резервирование успешно создано, статус остаётся PENDING
	// SUCCESS будет установлен только после commit, FAILED - после cancel или истечения срока

	log.Info("reserve successful",
		slog.String("transaction_id", transaction.ID),
		slog.Int64("amount", transaction.Amount),
	)

	return transaction, nil
}

// getExistingTransaction получает существующую транзакцию из БД и проверяет expires
func (t *Transactions) getExistingTransaction(ctx context.Context, idempotentKey string, log *slog.Logger) (*models.Transaction, error) {
	transaction, err := t.dbRepository.GetTransactionByIdempotencyKey(ctx, idempotentKey)
	if err != nil {
		log.Error("failed to get existing transaction", sl.Err(err))
		return nil, InternalError
	}

	// Проверяем expires - НЕ отменяем, это задача cron worker
	if transaction.ExpiresAt != nil && time.Now().After(*transaction.ExpiresAt) {
		log.Warn("existing reservation expired")
		return nil, ReservationExpired
	}

	return transaction, nil
}

// setFailed устанавливает статус failed в Redis
func (t *Transactions) setFailed(ctx context.Context, idempotentKey string, log *slog.Logger) {
	if err := t.transactionRepository.SetIdempotentKeyStatus(ctx, idempotentKey, ssov1.TransactionStatus_TRANSACTION_FAILED); err != nil {
		log.Error("failed to set redis status to failed", sl.Err(err))
	}
}
