package models

import ssov1 "sso/protos/gen/go/sso"

type Role uint

const (
	user Role = iota
	admin
)

// AuthType представляет тип авторизации пользователя
type AuthType string

const (
	AuthTypeEmail    AuthType = "email"
	AuthTypeTelegram AuthType = "telegram"
)

type User struct {
	ID         int64
	Email      *string // nil для Telegram авторизации
	Username   string
	PassHash   *string  // nil для Telegram авторизации
	TelegramID *int64   // nil для Email авторизации
	FirstName  *string  // Имя из Telegram
	LastName   *string  // Фамилия из Telegram
	PhotoURL   *string  // URL фото из Telegram
	Balance    float64  // Баланс пользователя
	AuthType   AuthType // "email" или "telegram"
	AppID      int32
	Role       ssov1.Role
}

type UserDTO struct {
	ID    int64
	Email string
}

// IsTelegramUser проверяет, авторизован ли пользователь через Telegram
func (u *User) IsTelegramUser() bool {
	return u.AuthType == AuthTypeTelegram
}

// IsEmailUser проверяет, авторизован ли пользователь через Email
func (u *User) IsEmailUser() bool {
	return u.AuthType == AuthTypeEmail
}

// GetDisplayName возвращает отображаемое имя пользователя
func (u *User) GetDisplayName() string {
	if u.Username != "" {
		return u.Username
	}
	if u.FirstName != nil {
		name := *u.FirstName
		if u.LastName != nil {
			name += " " + *u.LastName
		}
		return name
	}
	if u.Email != nil {
		return *u.Email
	}
	return ""
}
