package user_jwt

import (
	"fmt"
	"github.com/o1egl/paseto"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/config"
	"sso/sso/internal/domain/models"
	"time"
)

// GenerateAccessToken генерирует PASETO access токен для пользователя.
func GenerateAccessToken(user *models.User, duration time.Duration, secretKey string) (string, error) {
	v2 := paseto.NewV2()
	claims := map[string]interface{}{
		"sub":      user.ID,
		"app_id":   user.AppID,
		"username": user.Username,
		"role":     int32(user.Role),
		"exp":      time.Now().Add(duration).Unix(),
	}
	token, err := v2.Encrypt([]byte(secretKey), claims, nil)
	if err != nil {
		return "", fmt.Errorf("не удалось зашифровать токен: %w", err)
	}
	return token, nil
}

// GenerateRefreshToken генерирует PASETO refresh токен для пользователя.
func GenerateRefreshToken(user *models.User, duration time.Duration, secretKey string) (string, int64, error) {
	v2 := paseto.NewV2()
	exp := time.Now().Add(duration).Unix()
	claims := map[string]interface{}{
		"sub": user.ID,
		"exp": exp,
	}
	token, err := v2.Encrypt([]byte(secretKey), claims, nil)
	if err != nil {
		return "", 0, fmt.Errorf("не удалось зашифровать токен: %w", err)
	}
	return token, exp, nil
}

// CreateTokenPair создает пару access и refresh токенов с использованием PASETO.
func CreateTokenPair(user *models.User, config *config.Config) (*models.TokenPair, int64, error) {
	// Создаем access токен
	accessToken, err := GenerateAccessToken(user, config.Jwt.AccessTokenTTL, config.Secret)
	if err != nil {
		return nil, 0, err
	}

	// Создаем refresh токен
	refreshToken, exp, err := GenerateRefreshToken(user, config.Jwt.RefreshTokenTTL, config.Secret)
	if err != nil {
		return nil, 0, err
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, exp, nil
}

// ParseToken разбирает и проверяет PASETO токен, возвращая AccessTokenData.
func ParseToken(tokenString string, isAccessToken bool, secretKey string) (*models.AccessTokenData, error) {
	v2 := paseto.NewV2()
	var payload map[string]interface{}
	err := v2.Decrypt(tokenString, []byte(secretKey), &payload, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось расшифровать токен: %w", err)
	}

	// Проверка времени истечения
	exp, ok := payload["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("отсутствует или неверный exp claim")
	}
	if time.Now().Unix() > int64(exp) {
		return nil, fmt.Errorf("токен истек")
	}

	// Извлечение sub
	sub, ok := payload["sub"].(float64)
	if !ok {
		return nil, fmt.Errorf("отсутствует или неверный sub claim")
	}

	data := &models.AccessTokenData{
		UserID: int64(sub),
		Exp:    int64(exp),
	}

	// Дополнительные проверки для access токена
	if isAccessToken {
		username, ok := payload["username"].(string)
		if !ok {
			return nil, fmt.Errorf("отсутствует или неверный username claim")
		}
		appID, ok := payload["app_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("отсутствует или неверный app_id claim")
		}
		role, ok := payload["role"].(float64)
		if !ok {
			return nil, fmt.Errorf("отсутствует или неверный role claim")
		}
		data.Username = username
		data.AppID = int32(appID)
		data.Role = ssov1.Role(int32(role))
	}

	return data, nil
}
