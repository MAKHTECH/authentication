package models

import (
	ssov1 "sso/protos/gen/go/sso"
	"time"
)

// Transaction модель транзакции баланса
// Использует ssov1.TransactionType напрямую из proto
type Transaction struct {
	ID             string                // UUID транзакции
	UserID         int64                 // ID пользователя
	AppID          int32                 // ID приложения
	ReservationID  *string               // ID резервирования (для RESERVE это собственный ID, для COMMIT/CANCEL - ссылка на RESERVE)
	Type           ssov1.TransactionType // Тип транзакции из proto
	Status         string                // Статус транзакции: pending, success, failed
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
func (t *Transaction) IsReservation() bool {
	return t.Type == ssov1.TransactionType_TRANSACTION_TYPE_RESERVE
}

// IsActiveReservation проверяет, активно ли резервирование (нет commit/cancel)
// Эта проверка должна выполняться на уровне storage, здесь только проверка типа
func (t *Transaction) IsActiveReservation() bool {
	return t.IsReservation() && !t.IsExpired()
}

// IsExpired проверяет, истекло ли резервирование
func (t *Transaction) IsExpired() bool {
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

type RedisTransaction struct {
	IdempotentKey string
	Status        ssov1.TransactionStatus
	OperationType ssov1.TransactionType
	UserID        int64
	Amount        int64 // СУММА В КОПЕЙКАХ!!
	CreatedAt     time.Time
}

func (r *RedisTransaction) Key() string {
	return "transaction:" + r.IdempotentKey
}

func (r *RedisTransaction) GetTypeString() string {
	switch r.OperationType {
	case ssov1.TransactionType_TRANSACTION_TYPE_DEPOSIT:
		return "deposit"
	case ssov1.TransactionType_TRANSACTION_TYPE_RESERVE:
		return "withdraw"
	case ssov1.TransactionType_TRANSACTION_TYPE_COMMIT:
		return "commit"
	case ssov1.TransactionType_TRANSACTION_TYPE_CANCEL:
		return "cancel"
	case ssov1.TransactionType_TRANSACTION_TYPE_REFUND:
		return "refund"
	case ssov1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL:
		return "withdrawal"
	}

	return "unknown"
}

func (r *RedisTransaction) GetStatusString() string {
	switch r.Status {
	case ssov1.TransactionStatus_TRANSACTION_PENDING:
		return "pending"
	case ssov1.TransactionStatus_TRANSACTION_SUCCESS:
		return "success"
	case ssov1.TransactionStatus_TRANSACTION_FAILED:
		return "failed"
	}

	return "unknown"
}
