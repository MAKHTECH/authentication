package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// RateLimiter - структура для хранения лимитеров и redis клиента
type RateLimiter struct {
	RedisClient *redis.Client
	UberLimiter ratelimit.Limiter
	Attempts    map[string]int
	MaxAttempts int
	BlockTime   time.Duration
}

// NewRateLimiter - конструктор для RateLimiter
func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		RedisClient: client,
		UberLimiter: ratelimit.New(3, ratelimit.Per(time.Minute)), // 3 попыток в минуту
		Attempts:    make(map[string]int),
		MaxAttempts: 10, // max 5 attempt
		BlockTime:   time.Hour,
	}
}

// CheckAndIncrementAttempts - проверка и увеличение счетчика попыток
func (r *RateLimiter) CheckAndIncrementAttempts(ctx context.Context, clientIP string) error {
	key := fmt.Sprintf("auth_attempts:%s", clientIP)

	// Получаем текущие попытки из Redis
	attempts, err := r.RedisClient.Get(ctx, key).Int()
	if errors.Is(err, redis.Nil) {
		attempts = 0
	} else if err != nil {
		return status.Errorf(codes.Internal, "failed to get attempts: %v", err)
	}

	attempts++
	if attempts >= r.MaxAttempts {
		// Блокируем пользователя
		if err := r.BlockUser(ctx, clientIP); err != nil {
			return err
		}
		return status.Errorf(codes.PermissionDenied, "too many failed attempts")
	}

	// Сохраняем обновленное количество попыток с TTL
	err = r.RedisClient.Set(ctx, key, attempts, time.Minute*10).Err()
	if err != nil {
		return status.Errorf(codes.Internal, "failed to set attempts: %v", err)
	}

	return nil
}

// BlockUser - блокирует пользователя в Redis
func (r *RateLimiter) BlockUser(ctx context.Context, clientIP string) error {
	blockKey := fmt.Sprintf("blocked_user:%s", clientIP)
	return r.RedisClient.Set(ctx, blockKey, "blocked", r.BlockTime).Err()
}

// IsBlocked - проверяет, заблокирован ли пользователь
func (r *RateLimiter) IsBlocked(ctx context.Context, clientIP string) bool {
	blockKey := fmt.Sprintf("blocked_user:%s", clientIP)
	_, err := r.RedisClient.Get(ctx, blockKey).Result()
	return !errors.Is(err, redis.Nil)
}

// ResetAttempts - сбрасывает счетчик попыток
func (r *RateLimiter) ResetAttempts(ctx context.Context, clientIP string) error {
	key := fmt.Sprintf("auth_attempts:%s", clientIP)
	return r.RedisClient.Del(ctx, key).Err()
}
