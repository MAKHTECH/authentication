package grpc_auth

import (
	"context"
	"errors"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/services/auth"
)

type Auth interface {
	Login(ctx context.Context, user models.AuthUser) (tokenPair *models.TokenPair, err error)
	Logout(ctx context.Context, accessToken string) (bool, error)

	RefreshToken(ctx context.Context, refreshToken string) (*models.TokenPair, error)
	RegisterNewUser(ctx context.Context, user models.AuthUser) (tokenPair *models.TokenPair, err error)

	GetDevices(ctx context.Context, userID int32) ([]*models.RefreshSession, error)
}

type ServerAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &ServerAPI{auth: auth})
}

func (s *ServerAPI) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	// username validation
	if err := validation.Validate(
		req.GetUsername(), validation.Required, validation.Length(4, 100),
	); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// password validation
	if err := validation.Validate(
		req.GetPassword(), validation.Required, validation.Length(8, 100),
	); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// app_id validation
	if err := validation.Validate(
		req.GetAppId(), validation.Required, validation.Max(15),
	); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tokenPair, err := s.auth.Login(ctx, models.AuthUser{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		AppID:    req.GetAppId(),
	})

	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "Invalid Credentials")

		} else if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")

		} else if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "User already exists")
		} else if errors.Is(err, auth.ErrInvalidApp) {
			return nil, status.Error(codes.InvalidArgument, "Invalid App")
		}

		fmt.Println(err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{Tokens: &ssov1.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}}, nil
}

func (s *ServerAPI) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	// email validation
	if err := validation.Validate(
		req.GetEmail(), validation.Required, validation.Length(12, 100), is.Email,
	); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// password validation
	if err := validation.Validate(
		req.GetPassword(), validation.Required, validation.Length(8, 100),
	); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// app_id validation
	if err := validation.Validate(
		req.GetAppId(), validation.Required, validation.Max(15),
	); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// registration new user in the auth service and return user id
	tokenPair, err := s.auth.RegisterNewUser(ctx, models.AuthUser{
		Email:    req.GetEmail(),
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		AppID:    req.GetAppId(),
	})
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		} else if errors.Is(err, auth.ErrInvalidApp) {
			return nil, status.Error(codes.NotFound, "app not found")
		}
		fmt.Println(err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RegisterResponse{
		Tokens: &ssov1.TokenPair{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		},
	}, nil
}

func (s *ServerAPI) RefreshToken(ctx context.Context, req *ssov1.RefreshTokenRequest) (*ssov1.RefreshTokenResponse, error) {
	// refresh token validation
	refreshToken := req.GetRefreshToken()
	if err := validation.Validate(
		refreshToken, validation.Required, validation.Length(10, 200),
	); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	tokenPair, err := s.auth.RefreshToken(ctx, refreshToken)
	if err != nil {

		if errors.Is(err, auth.InvalidRefreshToken) {
			return nil, status.Error(codes.InvalidArgument, "Invalid refresh token")

		} else if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "User not found")

		} else if errors.Is(err, auth.ErrInvalidApp) {
			return nil, status.Error(codes.NotFound, "App not found")
		}

		fmt.Println(err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.RefreshTokenResponse{Tokens: &ssov1.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}}, nil
}

func (s *ServerAPI) Logout(ctx context.Context, req *ssov1.LogoutRequest) (*ssov1.LogoutResponse, error) {
	accessToken := req.GetAccessToken()
	if err := validation.Validate(accessToken, validation.Required); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	logout, err := s.auth.Logout(ctx, accessToken)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &ssov1.LogoutResponse{Success: logout}, nil
}

func (s *ServerAPI) GetDevices(ctx context.Context, req *ssov1.GetDevicesRequest) (*ssov1.GetDevicesResponse, error) {
	userData := ctx.Value("data").(*models.AccessTokenData)
	if userData.Role != ssov1.Role_ADMIN {
		return nil, status.Errorf(codes.PermissionDenied, "only admin can assign roles")
	}

	userID := req.GetUserId()
	if err := validation.Validate(userID, validation.Required); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	sessions, err := s.auth.GetDevices(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	var deviceList []*ssov1.Device
	for _, session := range sessions {
		deviceList = append(deviceList, &ssov1.Device{
			RefreshToken: session.RefreshToken,
			UserId:       session.UserId,
			Ua:           session.Ua,
			Ip:           session.Ip,
			Fingerprint:  session.Fingerprint,
			ExpiresIn:    int64(session.ExpiresIn),
			CreatedAt:    session.CreatedAt.Unix(),
		})
	}

	return &ssov1.GetDevicesResponse{Devices: deviceList}, nil
}
