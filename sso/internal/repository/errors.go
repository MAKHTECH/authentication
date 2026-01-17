package repository

import "errors"

var (
	ErrUserExists     = errors.New("user already exists")
	ErrUserNotFound   = errors.New("user not found")
	ErrAppNotFound    = errors.New("app not found")
	ErrUserRoleExists = errors.New("user role already exists or (user, app) not found")

	ErrUsernameUnique = errors.New("username must be unique")
	ErrEmailUnique    = errors.New("email must be unique")

	IdempotentKeyNotFound = errors.New("idempotent key not found")

	// Transaction errors
	ErrInsufficientFunds      = errors.New("insufficient funds")
	ErrIdempotentKeyExists    = errors.New("idempotent key already exists")
	ErrReservationNotFound    = errors.New("reservation not found")
	ErrReservationExpired     = errors.New("reservation expired")
	ErrTransactionNotPending  = errors.New("transaction is not pending")
	ErrAlreadyCommitted       = errors.New("reservation already committed")
	ErrAlreadyCancelled       = errors.New("reservation already cancelled")
	ErrInvalidTransactionType = errors.New("invalid transaction type")
)
