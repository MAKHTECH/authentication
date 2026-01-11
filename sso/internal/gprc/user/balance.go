package gprc_user

import (
	"context"
	"errors"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/services/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ServerAPI) GetBalance(ctx context.Context, req *ssov1.GetBalanceRequest) (*ssov1.GetBalanceResponse, error) {
	// todo обработать ошибки
	balance, reservedBalance, availableBalance, err := s.UserManagement.GetBalance(ctx, int(ctx.Value("data").(*models.AccessTokenData).UserID))
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.GetBalanceResponse{
		Balance:          balance,
		ReservedBalance:  reservedBalance,
		AvailableBalance: availableBalance,
	}, nil //
}

func (s *ServerAPI) Reserve(ctx context.Context, req *ssov1.ReserveRequest) (*ssov1.ReserveResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Reserve not implemented")
}

func (s *ServerAPI) Deposit(ctx context.Context, req *ssov1.DepositRequest) (*ssov1.DepositResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Deposit not implemented")
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
