package gprc_user

import (
	"context"
	"errors"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/services/user"
	"sso/sso/pkg/utils"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserManagement interface {
	AssignRole(ctx context.Context, role ssov1.Role, userID, appID int) (bool, error)
	CheckPermission(ctx context.Context, userID int, appID int) (bool, error)

	ChangePhoto(ctx context.Context, userID int, photoURL string) (bool, error)
	ChangeUsername(ctx context.Context, userID int, username string) (string, error)
	ChangeEmail(ctx context.Context, userID int, newEmail string) (string, error)
	ChangePassword(ctx context.Context, userID int, newPassword, currentPassword string) (bool, error)
}

type ServerAPI struct {
	ssov1.UserServer
	UserManagement UserManagement
}

func Register(gRPC *grpc.Server, UserManagement UserManagement) {
	ssov1.RegisterUserServer(gRPC, &ServerAPI{UserManagement: UserManagement})
}

// AssignRole assigns a role to a user
// Metadata: access_token: <token>
func (s *ServerAPI) AssignRole(ctx context.Context, req *ssov1.AssignRoleRequest) (*ssov1.AssignRoleResponse, error) {
	userData := ctx.Value("data").(*models.AccessTokenData)
	if userData.Role != ssov1.Role_ADMIN {
		return nil, status.Errorf(codes.PermissionDenied, "only admin can assign roles")
	}
	// validation
	userID := req.GetUserId()
	role := req.GetRole()

	// Валидация userID (uint32)
	if err := validation.Validate(userID,
		validation.Required, // Обязательное поле
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	//Валидация role с динамическим списком значений
	if err := validation.Validate(role,
		validation.Required,
		validation.In(utils.GetValidRoles()...), // Распаковываем срез в аргументы
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role: %v", err)
	}

	ok, err := s.UserManagement.AssignRole(ctx, role, int(userID), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, user.ErrUserRoleExists) {
			return nil, status.Errorf(codes.AlreadyExists, user.ErrUserRoleExists.Error())
		}

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.AssignRoleResponse{Success: ok}, nil
}

func (s *ServerAPI) ChangeAvatar(ctx context.Context, req *ssov1.ChangeAvatarRequest) (*ssov1.ChangeAvatarResponse, error) {
	userData := ctx.Value("data").(*models.AccessTokenData)

	// Валидация userID (uint32)
	if err := validation.Validate(
		userData.UserID, validation.Required,
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	// Валидация avatar_url - проверяем, что это валидный URL
	photo := req.GetPhotoUrl()
	if err := validation.Validate(photo,
		validation.Required,
		is.URL, // Проверяем, что это валидный URL
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid photo_url: %v", err)
	}

	// Проверяем, что по URL действительно находится изображение
	if err := utils.IsImageURL(photo); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "photo_url is not a valid image: %v", err)
	}

	success, err := s.UserManagement.ChangePhoto(ctx, int(userData.UserID), photo)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, user.ErrUserNotFound.Error())
		}
		if errors.Is(err, user.ErrInvalidPhotoURL) {
			return nil, status.Errorf(codes.InvalidArgument, user.ErrInvalidPhotoURL.Error())
		}

		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.ChangeAvatarResponse{Success: success}, nil
}

func (s *ServerAPI) ChangeUsername(ctx context.Context, req *ssov1.ChangeUsernameRequest) (*ssov1.ChangeUsernameResponse, error) {
	userData := ctx.Value("data").(*models.AccessTokenData)

	// validate username
	if err := validation.Validate(
		req.GetUsername(),
		validation.Required, is.Alphanumeric,
		validation.Length(3, 30),
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid username: %v", err)
	}

	newUsername, err := s.UserManagement.ChangeUsername(ctx, int(userData.UserID), req.GetUsername())
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		} else if errors.Is(err, user.ErrUsernameUnique) {
			return nil, status.Errorf(codes.AlreadyExists, "username already taken")
		}
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.ChangeUsernameResponse{
		Username: newUsername,
	}, nil
}

func (s *ServerAPI) ChangeEmail(ctx context.Context, req *ssov1.ChangeEmailRequest) (*ssov1.ChangeEmailResponse, error) {
	userData := ctx.Value("data").(*models.AccessTokenData)

	// validate email
	if err := validation.Validate(
		req.GetEmail(),
		validation.Required,
		is.Email,
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid email: %v", err)
	}

	// сменить имейл, проверив не занят ли он
	newEmail, err := s.UserManagement.ChangeEmail(ctx, int(userData.UserID), req.GetEmail())
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		} else if errors.Is(err, user.ErrEmailUnique) {
			return nil, status.Errorf(codes.AlreadyExists, "email already taken")
		} else if errors.Is(err, user.ErrInvalidEmail) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid email format")
		}
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.ChangeEmailResponse{
		Email: newEmail,
	}, nil
}

func (s *ServerAPI) ChangePassword(ctx context.Context, req *ssov1.ChangePasswordRequest) (*ssov1.ChangePasswordResponse, error) {
	userData := ctx.Value("data").(*models.AccessTokenData)

	// validate current password
	if err := validation.Validate(
		req.GetCurrentPassword(),
		validation.Required,
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "current password is required")
	}

	// validate new password
	if err := validation.Validate(
		req.GetNewPassword(),
		validation.Required,
		validation.Length(8, 100),
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid new password: %v", err)
	}

	success, err := s.UserManagement.ChangePassword(ctx, int(userData.UserID), req.GetNewPassword(), req.GetCurrentPassword())
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "user not found")
		} else if errors.Is(err, user.ErrWrongCurrentPassword) {
			return nil, status.Errorf(codes.InvalidArgument, "current password is incorrect")
		} else if errors.Is(err, user.ErrInvalidPassword) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid password")
		}
		return nil, status.Errorf(codes.Internal, "internal error")
	}

	return &ssov1.ChangePasswordResponse{
		Success: success,
	}, nil
}
