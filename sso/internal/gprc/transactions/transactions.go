package grpc_transactions

import (
	"context"
	"errors"
	"time"

	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/services/transactions"
	"sso/sso/internal/services/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServerAPI struct {
	ssov1.TransactionsServer
	UserManagement         user.UserManagement
	TransactionsManagement transactions.TransactionsManagement
}

func Register(
	gRPC *grpc.Server,
	UserManagement user.UserManagement,
	transactionsManagements transactions.TransactionsManagement) {
	ssov1.RegisterTransactionsServer(gRPC, &ServerAPI{
		UserManagement:         UserManagement,
		TransactionsManagement: transactionsManagements,
	})
}

func (s *ServerAPI) GetBalance(ctx context.Context, req *ssov1.GetBalanceRequest) (*ssov1.GetBalanceResponse, error) {
	balance, reservedBalance, availableBalance, err := s.UserManagement.GetBalance(ctx, int(ctx.Value("data").(*models.AccessTokenData).UserID))
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.GetBalanceResponse{
		Balance:          balance,
		ReservedBalance:  reservedBalance,
		AvailableBalance: availableBalance,
	}, nil
}

func (s *ServerAPI) Reserve(ctx context.Context, req *ssov1.ReserveRequest) (*ssov1.ReserveResponse, error) {
	userData := ctx.Value("data").(*models.AccessTokenData)

	// Используем TTL по умолчанию 15 минут для резервирования
	const defaultReserveTTL = 15 * time.Minute

	transaction, err := s.TransactionsManagement.Reserve(
		ctx,
		userData.UserID,
		req.GetAppId(),
		req.GetAmount(),
		req.GetIdempotencyKey(),
		req.GetDescription(),
		defaultReserveTTL,
	)

	if err != nil {
		switch {
		case errors.Is(err, transactions.InsufficientFunds):
			return &ssov1.ReserveResponse{
				Status:       ssov1.TransactionStatus_TRANSACTION_FAILED,
				ErrorMessage: "insufficient funds",
			}, nil
		case errors.Is(err, transactions.AlreadyInProgress):
			return &ssov1.ReserveResponse{
				Status:       ssov1.TransactionStatus_TRANSACTION_PENDING,
				ErrorMessage: "transaction already in progress",
			}, nil
		case errors.Is(err, transactions.ReservationExpired):
			return &ssov1.ReserveResponse{
				Status:       ssov1.TransactionStatus_TRANSACTION_FAILED,
				ErrorMessage: "reservation expired",
			}, nil
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &ssov1.ReserveResponse{
		Status:           ssov1.TransactionStatus_TRANSACTION_PENDING,
		ReservationId:    transaction.ID,
		ReservedAmount:   transaction.Amount,
		RemainingBalance: transaction.BalanceAfter - transaction.ReservedAfter,
	}, nil
}

func (s *ServerAPI) GetTransactions(ctx context.Context, req *ssov1.GetTransactionsRequest) (*ssov1.GetTransactionsResponse, error) {
	userData := ctx.Value("data").(*models.AccessTokenData)

	const maxRecords = 10

	from := int(req.GetFrom())
	to := int(req.GetTo())

	// Валидация from: не может быть отрицательным
	if from < 0 {
		return nil, status.Error(codes.InvalidArgument, "from must be non-negative")
	}

	// Валидация to: должен быть больше from
	if to <= from {
		return nil, status.Error(codes.InvalidArgument, "to must be greater than from")
	}

	// Ограничение: максимум 10 записей за раз
	requestedCount := to - from
	if requestedCount > maxRecords {
		return nil, status.Errorf(codes.InvalidArgument, "requested range exceeds maximum of %d records", maxRecords)
	}

	limit := requestedCount
	offset := from

	txList, totalCount, err := s.TransactionsManagement.GetTransactions(ctx, userData.UserID, limit, offset)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get transactions")
	}

	// Если from >= totalCount, возвращаем пустой список (не ошибка, просто нет данных)
	// Это уже обрабатывается репозиторием - он вернёт пустой slice

	// Вычисляем фактический диапазон
	actualTo := int32(from) + int32(len(txList))

	// Конвертируем модели в proto-сообщения
	protoTransactions := make([]*ssov1.Transaction, 0, len(txList))
	for _, tx := range txList {
		protoTx := &ssov1.Transaction{
			Id:            tx.ID,
			Type:          tx.Type,
			Amount:        tx.Amount,
			BalanceAfter:  tx.BalanceAfter,
			ReservedAfter: tx.ReservedAfter,
			Description:   tx.Description,
			CreatedAt:     tx.CreatedAt.Unix(),
		}

		if tx.ReservationID != nil {
			protoTx.ReservationId = *tx.ReservationID
		}

		protoTransactions = append(protoTransactions, protoTx)
	}

	return &ssov1.GetTransactionsResponse{
		Transactions: protoTransactions,
		TotalCount:   totalCount,
		From:         int32(from),
		To:           actualTo,
	}, nil
}

func (s *ServerAPI) CommitReserve(ctx context.Context, req *ssov1.CommitReserveRequest) (*ssov1.CommitReserveResponse, error) {
	reservationID := req.GetReservationId()
	if reservationID == "" {
		return nil, status.Error(codes.InvalidArgument, "reservation_id is required")
	}

	// Генерируем idempotency key для commit на основе reservation_id и app_id
	commitIdempotencyKey := "commit:" + reservationID

	transaction, err := s.TransactionsManagement.Commit(ctx, reservationID, commitIdempotencyKey)
	if err != nil {
		switch {
		case errors.Is(err, transactions.ReservationNotFound):
			return &ssov1.CommitReserveResponse{
				Success:      false,
				ErrorMessage: "reservation not found",
			}, nil
		case errors.Is(err, transactions.AlreadyInProgress):
			return &ssov1.CommitReserveResponse{
				Success:      false,
				ErrorMessage: "commit already in progress",
			}, nil
		case errors.Is(err, transactions.ReservationExpired):
			return &ssov1.CommitReserveResponse{
				Success:      false,
				ErrorMessage: "reservation expired or closed",
			}, nil
		case errors.Is(err, transactions.InvalidTransactionType):
			return &ssov1.CommitReserveResponse{
				Success:      false,
				ErrorMessage: "invalid transaction type",
			}, nil
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &ssov1.CommitReserveResponse{
		Success:         true,
		CommittedAmount: transaction.Amount,
		NewBalance:      transaction.BalanceAfter,
	}, nil
}

func (s *ServerAPI) CancelReserve(ctx context.Context, req *ssov1.CancelReserveRequest) (*ssov1.CancelReserveResponse, error) {
	reservationID := req.GetReservationId()
	if reservationID == "" {
		return nil, status.Error(codes.InvalidArgument, "reservation_id is required")
	}

	// Генерируем idempotency key для cancel на основе reservation_id
	cancelIdempotencyKey := "cancel:" + reservationID

	transaction, err := s.TransactionsManagement.Cancel(ctx, reservationID, cancelIdempotencyKey)
	if err != nil {
		switch {
		case errors.Is(err, transactions.ReservationNotFound):
			return &ssov1.CancelReserveResponse{
				Success:      false,
				ErrorMessage: "reservation not found",
			}, nil
		case errors.Is(err, transactions.AlreadyInProgress):
			return &ssov1.CancelReserveResponse{
				Success:      false,
				ErrorMessage: "cancel already in progress",
			}, nil
		case errors.Is(err, transactions.AlreadyCommitted):
			return &ssov1.CancelReserveResponse{
				Success:      false,
				ErrorMessage: "reservation already committed",
			}, nil
		case errors.Is(err, transactions.InvalidTransactionType):
			return &ssov1.CancelReserveResponse{
				Success:      false,
				ErrorMessage: "invalid transaction type",
			}, nil
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	// Для cancel: available_balance = balance - reserve_balance
	// После cancel reserve_balance уменьшается, значит available увеличивается
	newAvailableBalance := transaction.BalanceAfter - transaction.ReservedAfter

	return &ssov1.CancelReserveResponse{
		Success:        true,
		ReleasedAmount: transaction.Amount,
		NewBalance:     newAvailableBalance,
	}, nil
}

// TODO: когда добавим платежку
func (s *ServerAPI) Deposit(ctx context.Context, req *ssov1.DepositRequest) (*ssov1.DepositResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Deposit not implemented")
}
