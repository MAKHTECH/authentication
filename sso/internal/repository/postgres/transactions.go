package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"sso/sso/internal/domain/models"
	"sso/sso/internal/repository"

	ssov1 "sso/protos/gen/go/sso"
)

// Reserve создает резервирование в одной DB-транзакции
// 1. UPDATE users SET reserve_balance WHERE (balance - reserve_balance) >= amount
// 2. INSERT transactions с idempotency_key
// При unique_violation по idempotency_key - возвращает существующую транзакцию
func (r *Repository) Reserve(
	ctx context.Context,
	userID int64,
	appID int32,
	amount int64,
	idempotencyKey string,
	description string,
	expiresAt time.Time,
) (*models.Transaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. UPDATE users: увеличиваем reserve_balance
	// Условие: (balance - reserve_balance) >= amount
	var balanceBefore, reservedBefore int64
	err = tx.QueryRowContext(ctx, `
		UPDATE users 
		SET reserve_balance = reserve_balance + $1, updated_at = NOW()
		WHERE id = $2 AND (balance - reserve_balance) >= $1
		RETURNING balance, reserve_balance - $1
	`, amount, userID).Scan(&balanceBefore, &reservedBefore)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Либо пользователь не найден, либо недостаточно средств
			// Проверим существование пользователя
			var exists bool
			r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
			if !exists {
				return nil, repository.ErrUserNotFound
			}
			return nil, repository.ErrInsufficientFunds
		}
		return nil, err
	}

	reservedAfter := reservedBefore + amount

	// 2. INSERT transaction с idempotency_key
	var transaction models.Transaction
	var txExpiresAt sql.NullTime
	var typeStr string

	err = tx.QueryRowContext(ctx, `
		INSERT INTO transactions (
			user_id, app_id, type, amount, 
			balance_before, balance_after, 
			reserved_before, reserved_after,
			description, idempotency_key, expires_at, status
		) VALUES ($1, $2, 'reserve', $3, $4, $4, $5, $6, $7, $8, $9, 'pending')
		RETURNING id, user_id, app_id, type, amount, 
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, expires_at, status, created_at
	`, userID, appID, amount, balanceBefore, reservedBefore, reservedAfter,
		description, idempotencyKey, expiresAt).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.AppID,
		&typeStr,
		&transaction.Amount,
		&transaction.BalanceBefore,
		&transaction.BalanceAfter,
		&transaction.ReservedBefore,
		&transaction.ReservedAfter,
		&transaction.Description,
		&transaction.IdempotencyKey,
		&txExpiresAt,
		&transaction.Status,
		&transaction.CreatedAt,
	)

	if err != nil {
		// Проверяем unique_violation по idempotency_key
		if isUniqueViolation(err, "idempotency_key") {
			// Откатываем текущую транзакцию и возвращаем существующую
			tx.Rollback()
			return r.GetTransactionByIdempotencyKey(ctx, idempotencyKey)
		}
		return nil, err
	}

	if txExpiresAt.Valid {
		transaction.ExpiresAt = &txExpiresAt.Time
	}

	// Конвертация type из строки в proto enum
	transaction.Type = stringToTransactionType(typeStr)

	// Устанавливаем reservation_id = собственный ID для reserve транзакции
	_, err = tx.ExecContext(ctx, `
		UPDATE transactions SET reservation_id = id WHERE id = $1
	`, transaction.ID)
	if err != nil {
		return nil, err
	}
	transaction.ReservationID = &transaction.ID

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &transaction, nil
}

// GetTransactionByIdempotencyKey получает транзакцию по idempotency_key
func (r *Repository) GetTransactionByIdempotencyKey(ctx context.Context, idempotencyKey string) (*models.Transaction, error) {
	var transaction models.Transaction
	var txExpiresAt sql.NullTime
	var reservationID sql.NullString
	var typeStr string
	var statusStr string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, reservation_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, expires_at, status, created_at
		FROM transactions
		WHERE idempotency_key = $1
	`, idempotencyKey).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.AppID,
		&reservationID,
		&typeStr,
		&transaction.Amount,
		&transaction.BalanceBefore,
		&transaction.BalanceAfter,
		&transaction.ReservedBefore,
		&transaction.ReservedAfter,
		&transaction.Description,
		&transaction.IdempotencyKey,
		&txExpiresAt,
		&statusStr,
		&transaction.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrReservationNotFound
		}
		return nil, err
	}

	if reservationID.Valid {
		transaction.ReservationID = &reservationID.String
	}
	if txExpiresAt.Valid {
		transaction.ExpiresAt = &txExpiresAt.Time
	}

	// Конвертация type из строки в proto enum
	transaction.Type = stringToTransactionType(typeStr)

	return &transaction, nil
}

// isUniqueViolation проверяет, является ли ошибка нарушением уникальности
func isUniqueViolation(err error, constraint string) bool {
	// PostgreSQL unique violation code: 23505
	errStr := err.Error()
	return strings.Contains(errStr, "23505") ||
		(strings.Contains(errStr, "unique") && strings.Contains(errStr, constraint))
}

// stringToTransactionType конвертирует строку в proto enum
func stringToTransactionType(s string) ssov1.TransactionType {
	switch s {
	case "deposit":
		return ssov1.TransactionType_TRANSACTION_TYPE_DEPOSIT
	case "reserve":
		return ssov1.TransactionType_TRANSACTION_TYPE_RESERVE
	case "commit":
		return ssov1.TransactionType_TRANSACTION_TYPE_COMMIT
	case "cancel":
		return ssov1.TransactionType_TRANSACTION_TYPE_CANCEL
	case "refund":
		return ssov1.TransactionType_TRANSACTION_TYPE_REFUND
	case "withdrawal":
		return ssov1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL
	default:
		return ssov1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
	}
}

// GetReservationByID получает резервирование по ID
func (r *Repository) GetReservationByID(ctx context.Context, reservationID string) (*models.Transaction, error) {
	var transaction models.Transaction
	var txExpiresAt sql.NullTime
	var reservationIDNull sql.NullString
	var typeStr string
	var statusStr string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, reservation_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, expires_at, status, created_at
		FROM transactions
		WHERE id = $1 AND type = 'reserve'
	`, reservationID).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.AppID,
		&reservationIDNull,
		&typeStr,
		&transaction.Amount,
		&transaction.BalanceBefore,
		&transaction.BalanceAfter,
		&transaction.ReservedBefore,
		&transaction.ReservedAfter,
		&transaction.Description,
		&transaction.IdempotencyKey,
		&txExpiresAt,
		&statusStr,
		&transaction.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrReservationNotFound
		}
		return nil, err
	}

	if reservationIDNull.Valid {
		transaction.ReservationID = &reservationIDNull.String
	}
	if txExpiresAt.Valid {
		transaction.ExpiresAt = &txExpiresAt.Time
	}

	transaction.Type = stringToTransactionType(typeStr)
	transaction.Status = statusStr

	return &transaction, nil
}

// Commit подтверждает резервирование и списывает средства
// Алгоритм:
// 1. BEGIN транзакция
// 2. SELECT FOR UPDATE резервирование
// 3. Валидация (статус, тип, expires_at)
// 4. UPDATE users: reserve_balance -= amount, balance -= amount
// 5. UPDATE transactions: status = 'committed'
// 6. INSERT транзакция commit (аудит)
// 7. COMMIT
func (r *Repository) Commit(ctx context.Context, reservationID string, commitIdempotencyKey string) (*models.Transaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Блокируем резервирование FOR UPDATE
	var reservation struct {
		ID        string
		UserID    int64
		AppID     int32
		Amount    int64
		Status    string
		Type      string
		ExpiresAt sql.NullTime
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, amount, status, type, expires_at
		FROM transactions
		WHERE id = $1
		FOR UPDATE
	`, reservationID).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.AppID,
		&reservation.Amount,
		&reservation.Status,
		&reservation.Type,
		&reservation.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrReservationNotFound
		}
		return nil, err
	}

	// 2. Валидация типа - должен быть reserve
	if reservation.Type != "reserve" {
		return nil, repository.ErrInvalidTransactionType
	}

	// 3. Валидация статуса
	switch reservation.Status {
	case "committed":
		// Уже подтверждено - идемпотентный успех
		// Возвращаем существующую транзакцию commit по reservation_id
		return r.getCommitTransactionByReservationID(ctx, reservationID)
	case "expired", "cancelled":
		// Резервирование закрыто
		return nil, repository.ErrReservationExpired
	case "pending":
		// OK, продолжаем
	default:
		return nil, repository.ErrTransactionNotPending
	}

	// 4. Проверка expires_at
	if reservation.ExpiresAt.Valid && time.Now().After(reservation.ExpiresAt.Time) {
		return nil, repository.ErrReservationExpired
	}

	// 5. Финансовые изменения: списываем из reserve_balance и balance
	var balanceBefore, reservedBefore int64
	err = tx.QueryRowContext(ctx, `
		UPDATE users 
		SET reserve_balance = reserve_balance - $1,
			balance = balance - $1,
			updated_at = NOW()
		WHERE id = $2 AND reserve_balance >= $1 AND balance >= $1
		RETURNING balance + $1, reserve_balance + $1
	`, reservation.Amount, reservation.UserID).Scan(&balanceBefore, &reservedBefore)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("failed to commit: balance inconsistency")
		}
		return nil, err
	}

	balanceAfter := balanceBefore - reservation.Amount
	reservedAfter := reservedBefore - reservation.Amount

	// 6. Обновляем статус резервирования на 'committed'
	_, err = tx.ExecContext(ctx, `
		UPDATE transactions 
		SET status = 'committed', updated_at = NOW()
		WHERE id = $1
	`, reservationID)
	if err != nil {
		return nil, err
	}

	// 7. Создаём транзакцию commit (аудит)
	var commitTx models.Transaction
	var txExpiresAt sql.NullTime
	var typeStr string

	err = tx.QueryRowContext(ctx, `
		INSERT INTO transactions (
			user_id, app_id, reservation_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, status
		) VALUES ($1, $2, $3, 'commit', $4, $5, $6, $7, $8, 'commit for reservation', $9, 'success')
		RETURNING id, user_id, app_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, expires_at, status, created_at
	`, reservation.UserID, reservation.AppID, reservationID, reservation.Amount,
		balanceBefore, balanceAfter, reservedBefore, reservedAfter, commitIdempotencyKey).Scan(
		&commitTx.ID,
		&commitTx.UserID,
		&commitTx.AppID,
		&typeStr,
		&commitTx.Amount,
		&commitTx.BalanceBefore,
		&commitTx.BalanceAfter,
		&commitTx.ReservedBefore,
		&commitTx.ReservedAfter,
		&commitTx.Description,
		&commitTx.IdempotencyKey,
		&txExpiresAt,
		&commitTx.Status,
		&commitTx.CreatedAt,
	)

	if err != nil {
		// Проверяем unique_violation по idempotency_key
		if isUniqueViolation(err, "idempotency_key") {
			// Уже выполнен с этим ключом
			tx.Rollback()
			return r.GetTransactionByIdempotencyKey(ctx, commitIdempotencyKey)
		}
		return nil, err
	}

	commitTx.Type = stringToTransactionType(typeStr)
	commitTx.ReservationID = &reservationID

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &commitTx, nil
}

// getCommitTransactionByReservationID возвращает транзакцию commit по reservation_id
func (r *Repository) getCommitTransactionByReservationID(ctx context.Context, reservationID string) (*models.Transaction, error) {
	var transaction models.Transaction
	var txExpiresAt sql.NullTime
	var reservationIDNull sql.NullString
	var typeStr string
	var statusStr string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, reservation_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, expires_at, status, created_at
		FROM transactions
		WHERE reservation_id = $1 AND type = 'commit'
		LIMIT 1
	`, reservationID).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.AppID,
		&reservationIDNull,
		&typeStr,
		&transaction.Amount,
		&transaction.BalanceBefore,
		&transaction.BalanceAfter,
		&transaction.ReservedBefore,
		&transaction.ReservedAfter,
		&transaction.Description,
		&transaction.IdempotencyKey,
		&txExpiresAt,
		&statusStr,
		&transaction.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrReservationNotFound
		}
		return nil, err
	}

	if reservationIDNull.Valid {
		transaction.ReservationID = &reservationIDNull.String
	}
	if txExpiresAt.Valid {
		transaction.ExpiresAt = &txExpiresAt.Time
	}

	transaction.Type = stringToTransactionType(typeStr)
	transaction.Status = statusStr

	return &transaction, nil
}

// Cancel отменяет резервирование и возвращает средства
// Алгоритм:
// 1. BEGIN транзакция
// 2. SELECT FOR UPDATE резервирование
// 3. Валидация (статус, тип)
// 4. UPDATE users: reserve_balance -= amount (возвращаем средства)
// 5. UPDATE transactions: status = 'cancelled'
// 6. INSERT транзакция cancel (аудит)
// 7. COMMIT
func (r *Repository) Cancel(ctx context.Context, reservationID string, cancelIdempotencyKey string) (*models.Transaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Блокируем резервирование FOR UPDATE
	var reservation struct {
		ID        string
		UserID    int64
		AppID     int32
		Amount    int64
		Status    string
		Type      string
		ExpiresAt sql.NullTime
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, amount, status, type, expires_at
		FROM transactions
		WHERE id = $1
		FOR UPDATE
	`, reservationID).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.AppID,
		&reservation.Amount,
		&reservation.Status,
		&reservation.Type,
		&reservation.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrReservationNotFound
		}
		return nil, err
	}

	// 2. Валидация типа - должен быть reserve
	if reservation.Type != "reserve" {
		return nil, repository.ErrInvalidTransactionType
	}

	// 3. Валидация статуса
	switch reservation.Status {
	case "committed":
		// Уже подтверждено - нельзя отменить
		return nil, repository.ErrAlreadyCommitted
	case "expired", "cancelled":
		// Уже отменено/истекло - идемпотентный успех
		return r.getCancelTransactionByReservationID(ctx, reservationID)
	case "pending":
		// OK, продолжаем
	default:
		return nil, repository.ErrTransactionNotPending
	}

	// 4. Возврат средств: уменьшаем reserve_balance (balance не меняется!)
	var balanceBefore, reservedBefore int64
	err = tx.QueryRowContext(ctx, `
		UPDATE users 
		SET reserve_balance = reserve_balance - $1,
			updated_at = NOW()
		WHERE id = $2 AND reserve_balance >= $1
		RETURNING balance, reserve_balance + $1
	`, reservation.Amount, reservation.UserID).Scan(&balanceBefore, &reservedBefore)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("failed to cancel: reserve_balance inconsistency")
		}
		return nil, err
	}

	reservedAfter := reservedBefore - reservation.Amount

	// 5. Обновляем статус резервирования на 'cancelled'
	_, err = tx.ExecContext(ctx, `
		UPDATE transactions 
		SET status = 'cancelled', updated_at = NOW()
		WHERE id = $1
	`, reservationID)
	if err != nil {
		return nil, err
	}

	// 6. Создаём транзакцию cancel (аудит)
	var cancelTx models.Transaction
	var txExpiresAt sql.NullTime
	var typeStr string

	err = tx.QueryRowContext(ctx, `
		INSERT INTO transactions (
			user_id, app_id, reservation_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, status
		) VALUES ($1, $2, $3, 'cancel', $4, $5, $5, $6, $7, 'cancel for reservation', $8, 'success')
		RETURNING id, user_id, app_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, expires_at, status, created_at
	`, reservation.UserID, reservation.AppID, reservationID, reservation.Amount,
		balanceBefore, reservedBefore, reservedAfter, cancelIdempotencyKey).Scan(
		&cancelTx.ID,
		&cancelTx.UserID,
		&cancelTx.AppID,
		&typeStr,
		&cancelTx.Amount,
		&cancelTx.BalanceBefore,
		&cancelTx.BalanceAfter,
		&cancelTx.ReservedBefore,
		&cancelTx.ReservedAfter,
		&cancelTx.Description,
		&cancelTx.IdempotencyKey,
		&txExpiresAt,
		&cancelTx.Status,
		&cancelTx.CreatedAt,
	)

	if err != nil {
		// Проверяем unique_violation по idempotency_key
		if isUniqueViolation(err, "idempotency_key") {
			// Уже выполнен с этим ключом
			tx.Rollback()
			return r.GetTransactionByIdempotencyKey(ctx, cancelIdempotencyKey)
		}
		return nil, err
	}

	cancelTx.Type = stringToTransactionType(typeStr)
	cancelTx.ReservationID = &reservationID

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &cancelTx, nil
}

// getCancelTransactionByReservationID возвращает транзакцию cancel по reservation_id
func (r *Repository) getCancelTransactionByReservationID(ctx context.Context, reservationID string) (*models.Transaction, error) {
	var transaction models.Transaction
	var txExpiresAt sql.NullTime
	var reservationIDNull sql.NullString
	var typeStr string
	var statusStr string

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, reservation_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, expires_at, status, created_at
		FROM transactions
		WHERE reservation_id = $1 AND type = 'cancel'
		LIMIT 1
	`, reservationID).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.AppID,
		&reservationIDNull,
		&typeStr,
		&transaction.Amount,
		&transaction.BalanceBefore,
		&transaction.BalanceAfter,
		&transaction.ReservedBefore,
		&transaction.ReservedAfter,
		&transaction.Description,
		&transaction.IdempotencyKey,
		&txExpiresAt,
		&statusStr,
		&transaction.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Нет транзакции cancel, но статус expired/cancelled - создаём пустой ответ
			return &models.Transaction{
				ReservationID: &reservationID,
				Status:        "success",
			}, nil
		}
		return nil, err
	}

	if reservationIDNull.Valid {
		transaction.ReservationID = &reservationIDNull.String
	}
	if txExpiresAt.Valid {
		transaction.ExpiresAt = &txExpiresAt.Time
	}

	transaction.Type = stringToTransactionType(typeStr)
	transaction.Status = statusStr

	return &transaction, nil
}

// CancelExpiredReservation отменяет истёкшее резервирование и возвращает средства
// 1. Проверяем что резервирование существует, status='pending', expires_at < NOW()
// 2. Возвращаем reserve_balance пользователю
// 3. Обновляем статус резервирования на 'failed'
// 4. Создаём транзакцию отмены
func (r *Repository) CancelExpiredReservation(ctx context.Context, reservationID string) (*models.Transaction, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// 1. Получаем и блокируем резервирование FOR UPDATE
	var reservation struct {
		ID        string
		UserID    int64
		AppID     int32
		Amount    int64
		Status    string
		ExpiresAt sql.NullTime
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id, user_id, app_id, amount, status, expires_at
		FROM transactions
		WHERE id = $1 AND type = 'reserve'
		FOR UPDATE
	`, reservationID).Scan(
		&reservation.ID,
		&reservation.UserID,
		&reservation.AppID,
		&reservation.Amount,
		&reservation.Status,
		&reservation.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrReservationNotFound
		}
		return nil, err
	}

	// 2. Проверяем статус - должен быть pending
	if reservation.Status != "pending" {
		return nil, repository.ErrTransactionNotPending
	}

	// 3. Проверяем что expires_at истёк
	if !reservation.ExpiresAt.Valid || time.Now().Before(reservation.ExpiresAt.Time) {
		// Резервирование ещё не истекло - нельзя отменить по этой причине
		return nil, errors.New("reservation has not expired yet")
	}

	// 4. Возвращаем reserve_balance пользователю
	var balanceBefore, reservedBefore int64
	err = tx.QueryRowContext(ctx, `
		UPDATE users 
		SET reserve_balance = reserve_balance - $1, updated_at = NOW()
		WHERE id = $2 AND reserve_balance >= $1
		RETURNING balance, reserve_balance + $1
	`, reservation.Amount, reservation.UserID).Scan(&balanceBefore, &reservedBefore)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("failed to release reserve: insufficient reserve_balance")
		}
		return nil, err
	}

	reservedAfter := reservedBefore - reservation.Amount

	// 5. Обновляем статус оригинального резервирования на 'failed'
	_, err = tx.ExecContext(ctx, `
		UPDATE transactions 
		SET status = 'failed', updated_at = NOW()
		WHERE id = $1
	`, reservationID)
	if err != nil {
		return nil, err
	}

	// 6. Создаём транзакцию отмены
	var cancelTx models.Transaction
	var txExpiresAt sql.NullTime
	var typeStr string

	err = tx.QueryRowContext(ctx, `
		INSERT INTO transactions (
			user_id, app_id, reservation_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, status
		) VALUES ($1, $2, $3, 'cancel', $4, $5, $5, $6, $7, 'auto-cancelled: reservation expired', 'success')
		RETURNING id, user_id, app_id, type, amount,
			balance_before, balance_after, reserved_before, reserved_after,
			description, idempotency_key, expires_at, status, created_at
	`, reservation.UserID, reservation.AppID, reservationID, reservation.Amount,
		balanceBefore, reservedBefore, reservedAfter).Scan(
		&cancelTx.ID,
		&cancelTx.UserID,
		&cancelTx.AppID,
		&typeStr,
		&cancelTx.Amount,
		&cancelTx.BalanceBefore,
		&cancelTx.BalanceAfter,
		&cancelTx.ReservedBefore,
		&cancelTx.ReservedAfter,
		&cancelTx.Description,
		&cancelTx.IdempotencyKey,
		&txExpiresAt,
		&cancelTx.Status,
		&cancelTx.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	cancelTx.Type = stringToTransactionType(typeStr)
	cancelTx.ReservationID = &reservationID

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &cancelTx, nil
}

// GetExpiredReservations возвращает список ID истёкших резервирований
// status='pending', expires_at < NOW(), type='reserve'
func (r *Repository) GetExpiredReservations(ctx context.Context, limit int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id FROM transactions
		WHERE type = 'reserve' 
			AND status = 'pending' 
			AND expires_at < NOW()
		ORDER BY expires_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ids, nil
}
