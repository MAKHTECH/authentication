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
	return nil, status.Error(codes.Unimplemented, "method GetTransactions not implemented")
}

// ----- For Service Role -----

func (s *ServerAPI) CommitReserve(ctx context.Context, req *ssov1.CommitReserveRequest) (*ssov1.CommitReserveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CommitReserve not implemented")
}

func (s *ServerAPI) CancelReserve(ctx context.Context, req *ssov1.CancelReserveRequest) (*ssov1.CancelReserveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CancelReserve not implemented")
}

// TODO: когда добавим платежку
func (s *ServerAPI) Deposit(ctx context.Context, req *ssov1.DepositRequest) (*ssov1.DepositResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Deposit not implemented")
}
