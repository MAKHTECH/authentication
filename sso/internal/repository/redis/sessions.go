package redis

import (
	"context"
	"fmt"
	"time"

	"sso/sso/internal/domain/models"
)

// SaveRefreshSession сохраняет refresh сессию
func (r *Repository) SaveRefreshSession(ctx context.Context, rs *models.RefreshSession, refreshTTL time.Duration) error {
	key := "user:" + rs.UserId + ":" + rs.Fingerprint
	response := r.Client.HMSet(ctx, key, map[string]interface{}{
		"refreshToken": rs.RefreshToken,
		"userId":       rs.UserId,
		"ua":           rs.Ua,
		"ip":           rs.Ip,
		"fingerprint":  rs.Fingerprint,
		"expiresIn":    rs.ExpiresIn,
		"createdAt":    rs.CreatedAt,
	})
	if response.Err() != nil {
		return response.Err()
	}

	// Установка TTL для ключа в Redis
	if err := r.Client.Expire(ctx, key, refreshTTL).Err(); err != nil {
		return err
	}

	return nil
}

// GetRefreshSession получает refresh сессию по fingerprint
func (r *Repository) GetRefreshSession(ctx context.Context, fingerprint string) (*models.RefreshSession, error) {
	keyPattern := "user:*:" + fingerprint
	keys, err := r.Client.Keys(ctx, keyPattern).Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("refresh session not found")
	}

	key := keys[0]
	result, err := r.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	refreshSession := &models.RefreshSession{
		ExpiresIn:    0,
		Ua:           result["ua"],
		Fingerprint:  result["fingerprint"],
		RefreshToken: result["refreshToken"],
		CreatedAt:    time.Time{},
		UserId:       result["userId"],
		Ip:           result["ip"],
	}

	expiresIn, err := time.ParseDuration(result["expiresIn"] + "s")
	if err != nil {
		return nil, fmt.Errorf("invalid expiresIn format: %v", err)
	}
	refreshSession.ExpiresIn = expiresIn

	createdAt, err := time.Parse(time.RFC3339Nano, result["createdAt"])
	if err != nil {
		return nil, fmt.Errorf("invalid createdAt format: %v", err)
	}
	refreshSession.CreatedAt = createdAt

	return refreshSession, nil
}

// GetRefreshSessionsByUserId возвращает все сессии пользователя по его ID
func (r *Repository) GetRefreshSessionsByUserId(ctx context.Context, userID string) ([]*models.RefreshSession, error) {
	// Формируем паттерн для поиска ключей: user:<user_id>:*
	keyPattern := fmt.Sprintf("user:%s:*", userID)

	// Получаем все ключи, соответствующие паттерну
	keys, err := r.Client.Keys(ctx, keyPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys for user %s: %w", userID, err)
	}

	// Если ключей нет, возвращаем пустой список без ошибки
	if len(keys) == 0 {
		return []*models.RefreshSession{}, nil
	}

	// Собираем все сессии
	var refreshSessions []*models.RefreshSession
	for _, key := range keys {
		session, err := r.getRefreshSessionFromKey(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get session for key %s: %w", key, err)
		}
		refreshSessions = append(refreshSessions, session)
	}

	return refreshSessions, nil
}

// DeleteRefreshSession удаляет refresh сессию
func (r *Repository) DeleteRefreshSession(ctx context.Context, fingerprint, id string) error {
	key := "user:" + id + ":" + fingerprint
	if err := r.Client.Del(ctx, key).Err(); err != nil {
		return err
	}

	return nil
}

// getRefreshSessionFromKey извлекает данные сессии по конкретному ключу
func (r *Repository) getRefreshSessionFromKey(ctx context.Context, key string) (*models.RefreshSession, error) {
	// Получаем все поля хэша по ключу
	result, err := r.Client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get hash data for key %s: %w", key, err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no data found for key %s", key)
	}

	// Создаём объект сессии
	session := &models.RefreshSession{
		UserId:       result["userId"],
		Fingerprint:  result["fingerprint"],
		RefreshToken: result["refreshToken"],
		Ua:           result["ua"],
		Ip:           result["ip"],
	}

	// Парсим expiresIn
	if expiresInStr, ok := result["expiresIn"]; ok {
		expiresIn, err := time.ParseDuration(expiresInStr + "s")
		if err != nil {
			return nil, fmt.Errorf("invalid expiresIn format for key %s: %w", key, err)
		}
		session.ExpiresIn = expiresIn
	}

	// Парсим createdAt
	if createdAtStr, ok := result["createdAt"]; ok {
		createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("invalid createdAt format for key %s: %w", key, err)
		}
		session.CreatedAt = createdAt
	}

	return session, nil
}
