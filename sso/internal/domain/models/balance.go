package models

import (
	"time"

	ssov1 "sso/protos/gen/go/sso"
)

// BalanceTransaction модель транзакции баланса
// Использует ssov1.TransactionType напрямую из proto
type BalanceTransaction struct {
	ID             string                // UUID транзакции
	UserID         int64                 // ID пользователя
	AppID          int32                 // ID приложения
	ReservationID  *string               // ID резервирования (для RESERVE это собственный ID, для COMMIT/CANCEL - ссылка на RESERVE)
	Type           ssov1.TransactionType // Тип транзакции из proto
	Amount         int64                 // Сумма транзакции в копейках (всегда положительная)
	BalanceBefore  int64                 // Баланс до операции в копейках
	BalanceAfter   int64                 // Баланс после операции в копейках
	ReservedBefore int64                 // Reserved баланс до операции в копейках
	ReservedAfter  int64                 // Reserved баланс после операции в копейках
	Description    string                // Описание операции
	IdempotencyKey string                // Ключ идемпотентности
	Metadata       string                // JSON с дополнительными данными
	ExpiresAt      *time.Time            // Время истечения резервирования (только для type=RESERVE)
	CreatedAt      time.Time             // Время создания
}

// IsReservation проверяет, является ли транзакция резервированием
func (t *BalanceTransaction) IsReservation() bool {
	return t.Type == ssov1.TransactionType_TRANSACTION_TYPE_RESERVE
}

// IsActiveReservation проверяет, активно ли резервирование (нет commit/cancel)
// Эта проверка должна выполняться на уровне storage, здесь только проверка типа
func (t *BalanceTransaction) IsActiveReservation() bool {
	return t.IsReservation() && !t.IsExpired()
}

// IsExpired проверяет, истекло ли резервирование
func (t *BalanceTransaction) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

// BalanceInfo информация о балансе пользователя
type BalanceInfo struct {
	UserID           int64 // ID пользователя
	Balance          int64 // Общий баланс в копейках
	ReservedBalance  int64 // Замороженные средства в копейках
	AvailableBalance int64 // Доступный баланс в копейках
}

// NewBalanceInfo создает информацию о балансе
func NewBalanceInfo(userID int64, balance, reservedBalance int64) *BalanceInfo {
	return &BalanceInfo{
		UserID:           userID,
		Balance:          balance,
		ReservedBalance:  reservedBalance,
		AvailableBalance: balance - reservedBalance,
	}
}
