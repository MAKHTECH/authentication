package gprc_user

import (
	"context"
	"errors"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/services/user"
	"sso/sso/pkg/utils"

	validation "github.com/go-ozzo/ozzo-validation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
