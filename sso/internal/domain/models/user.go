package models

import ssov1 "sso/protos/gen/go/sso"

type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
	RoleService   Role = "service"
)

// RoleToProto конвертирует строковую роль в proto Role
func RoleToProto(role string) ssov1.Role {
	switch role {
	case string(RoleUser):
		return ssov1.Role_USER
	case string(RoleModerator):
		return ssov1.Role_MODERATOR
	case string(RoleAdmin):
		return ssov1.Role_ADMIN
	case string(RoleService):
		return ssov1.Role_SERVICE
	default:
		return ssov1.Role_USER
	}
}

// ProtoToRole конвертирует proto Role в строковую роль
func ProtoToRole(role ssov1.Role) Role {
	switch role {
	case ssov1.Role_USER:
		return RoleUser
	case ssov1.Role_MODERATOR:
		return RoleModerator
	case ssov1.Role_ADMIN:
		return RoleAdmin
	case ssov1.Role_SERVICE:
		return RoleService
	default:
		return RoleUser
	}
}

// AuthType представляет тип авторизации пользователя
type AuthType string

const (
	AuthTypeEmail    AuthType = "email"
	AuthTypeTelegram AuthType = "telegram"
)

type User struct {
	ID              int64
	Email           *string // nil для Telegram авторизации
	Username        string
	PassHash        *string  // nil для Telegram авторизации
	TelegramID      *int64   // nil для Email авторизации
	FirstName       *string  // Имя из Telegram
	LastName        *string  // Фамилия из Telegram
	PhotoURL        *string  // URL фото из Telegram
	Balance         int64    // Баланс пользователя в копейках
	ReservedBalance int64    // Замороженные средства в копейках
	AuthType        AuthType // "email" или "telegram"
	AppID           int32
	Role            ssov1.Role
}

// AvailableBalance возвращает доступный баланс в копейках (баланс минус замороженные средства)
func (u *User) AvailableBalance() int64 {
	return u.Balance - u.ReservedBalance
}

// BalanceRubles возвращает баланс в рублях (для отображения)
func (u *User) BalanceRubles() float64 {
	return CopecksToRubles(u.Balance)
}

// ReservedBalanceRubles возвращает замороженные средства в рублях
func (u *User) ReservedBalanceRubles() float64 {
	return CopecksToRubles(u.ReservedBalance)
}

// AvailableBalanceRubles возвращает доступный баланс в рублях
func (u *User) AvailableBalanceRubles() float64 {
	return CopecksToRubles(u.AvailableBalance())
}

// CanReserve проверяет, можно ли зарезервировать указанную сумму (в копейках)
func (u *User) CanReserve(amountCopecks int64) bool {
	return u.AvailableBalance() >= amountCopecks && amountCopecks > 0
}

// CopecksToRubles конвертирует копейки в рубли
func CopecksToRubles(copecks int64) float64 {
	return float64(copecks) / 100
}

// RublesToCopecks конвертирует рубли в копейки
func RublesToCopecks(rubles float64) int64 {
	return int64(rubles * 100)
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
