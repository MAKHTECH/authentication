package redis

import (
	"context"
	"strconv"
	"time"

	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/domain/models"
)

func (r *Repository) SaveIdempotentKey(ctx context.Context, transaction *models.RedisTransaction) error {
	response := r.Client.HMSet(ctx, transaction.Key(), map[string]interface{}{
		"status": transaction.Status,

		"type":    transaction.OperationType,
		"amount":  transaction.Amount,
		"user_id": transaction.UserID,

		"createdAt": transaction.CreatedAt,
	})
	if response.Err() != nil {
		return response.Err()
	}

	// Установка TTL для PROCESSING статуса (5 минут)
	if err := r.Client.Expire(ctx, transaction.Key(), r.cfg.Idempotency.ProcessingTTL).Err(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetIdempotentKey(ctx context.Context, idempotentKey string) (*models.RedisTransaction, error) {
	result, err := r.Client.HGetAll(ctx, "transaction:"+idempotentKey).Result()
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, nil
	}

	// Парсинг status
	statusInt, err := strconv.Atoi(result["status"])
	if err != nil {
		return nil, err
	}

	// Парсинг type
	typeInt, err := strconv.Atoi(result["type"])
	if err != nil {
		return nil, err
	}

	// Парсинг amount
	amount, err := strconv.ParseInt(result["amount"], 10, 64)
	if err != nil {
		return nil, err
	}

	// Парсинг user_id
	userID, err := strconv.ParseInt(result["user_id"], 10, 64)
	if err != nil {
		return nil, err
	}

	// Парсинг createdAt
	createdAt, err := time.Parse(time.RFC3339Nano, result["createdAt"])
	if err != nil {
		return nil, err
	}

	transaction := &models.RedisTransaction{
		IdempotentKey: idempotentKey,
		Status:        ssov1.TransactionStatus(statusInt),
		OperationType: ssov1.TransactionType(typeInt),
		UserID:        userID,
		Amount:        amount,
		CreatedAt:     createdAt,
	}

	return transaction, nil
}

func (r *Repository) SetIdempotentKeyStatus(ctx context.Context, idempotentKey string, status ssov1.TransactionStatus) error {
	key := "transaction:" + idempotentKey

	// Обновляем статус
	if err := r.Client.HSet(ctx, key, "status", int(status)).Err(); err != nil {
		return err
	}

	// Устанавливаем TTL в зависимости от статуса
	var ttl time.Duration
	switch status {
	case ssov1.TransactionStatus_TRANSACTION_SUCCESS:
		ttl = r.cfg.Idempotency.SuccessTTL // 24h
	case ssov1.TransactionStatus_TRANSACTION_FAILED:
		ttl = r.cfg.Idempotency.FailedTTL // 5 минут
	default:
		ttl = r.cfg.Idempotency.ProcessingTTL // 5 минут
	}

	return r.Client.Expire(ctx, key, ttl).Err()
}

func (r *Repository) DeleteIdempotentKey(ctx context.Context, idempotentKey string) error {
	key := "transaction:" + idempotentKey
	return r.Client.Del(ctx, key).Err()
}
