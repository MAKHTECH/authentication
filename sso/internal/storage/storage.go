package storage

import "errors"

var (
	ErrUserExists     = errors.New("user already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrAppNotFound    = errors.New("app not found")
	ErrUserRoleExists = errors.New("user role already exists or (user, app) not found")

	ErrUsernameUnique = errors.New("username must be unique")
	ErrEmailUnique    = errors.New("email must be unique")
)
