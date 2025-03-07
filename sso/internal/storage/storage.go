package storage

import "errors"

var (
	ErrUserExists     = errors.New("user already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrAppNotFound    = errors.New("app not found")
	ErrUserRoleExists = errors.New("user role already exists or (user, app) not found")
)
