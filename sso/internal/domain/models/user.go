package models

import ssov1 "sso/protos/gen/go/sso"

type Role uint

const (
	user Role = iota
	admin
)

type User struct {
	ID       int64
	Email    string
	Username string
	PassHash string
	AppID    int32
	Role     ssov1.Role
}

type UserDTO struct {
	ID    int64
	Email string
}
