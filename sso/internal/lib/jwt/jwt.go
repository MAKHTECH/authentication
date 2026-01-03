package user_jwt

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/config"
	"sso/sso/internal/domain/models"
	"time"

	"github.com/o1egl/paseto"
)

// GetKeyPair извлекает Ed25519 ключевую пару из hex-закодированного приватного ключа.
func GetKeyPair(privateKeyHex string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, nil, fmt.Errorf("не удалось декодировать приватный ключ: %w", err)
	}
	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil, nil, fmt.Errorf("приватный ключ должен быть %d байт, получено: %d", ed25519.PrivateKeySize, len(privateKeyBytes))
	}
	privateKey := ed25519.PrivateKey(privateKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return privateKey, publicKey, nil
}

// GenerateAccessToken генерирует PASETO v2.public access токен для пользователя.
func GenerateAccessToken(user *models.User, duration time.Duration, privateKeyHex string) (string, error) {
	privateKey, _, err := GetKeyPair(privateKeyHex)
	if err != nil {
		return "", err
	}

	v2 := paseto.NewV2()
	claims := map[string]interface{}{
		"sub":       user.ID,
		"app_id":    user.AppID,
		"username":  user.Username,
		"photo_url": user.PhotoURL,
		"role":      int32(user.Role),
		"exp":       time.Now().Add(duration).Unix(),
	}
	token, err := v2.Sign(privateKey, claims, nil)
	if err != nil {
		return "", fmt.Errorf("не удалось подписать токен: %w", err)
	}
	return token, nil
}

// GenerateRefreshToken генерирует PASETO v2.public refresh токен для пользователя.
func GenerateRefreshToken(user *models.User, duration time.Duration, privateKeyHex string) (string, int64, error) {
	privateKey, _, err := GetKeyPair(privateKeyHex)
	if err != nil {
		return "", 0, err
	}

	v2 := paseto.NewV2()
	exp := time.Now().Add(duration).Unix()
	claims := map[string]interface{}{
		"sub": user.ID,
		"exp": exp,
	}
	token, err := v2.Sign(privateKey, claims, nil)
	if err != nil {
		return "", 0, fmt.Errorf("не удалось подписать токен: %w", err)
	}
	return token, exp, nil
}

// CreateTokenPair создает пару access и refresh токенов с использованием PASETO v2.public.
func CreateTokenPair(user *models.User, cfg *config.Config) (*models.TokenPair, int64, error) {
	// Создаем access токен
	accessToken, err := GenerateAccessToken(user, cfg.Jwt.AccessTokenTTL, cfg.PrivateKey)
	if err != nil {
		return nil, 0, err
	}

	// Создаем refresh токен
	refreshToken, exp, err := GenerateRefreshToken(user, cfg.Jwt.RefreshTokenTTL, cfg.PrivateKey)
	if err != nil {
		return nil, 0, err
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, exp, nil
}

// ParseToken разбирает и проверяет PASETO v2.public токен, возвращая AccessTokenData.
func ParseToken(tokenString string, isAccessToken bool, privateKeyHex string) (*models.AccessTokenData, error) {
	_, publicKey, err := GetKeyPair(privateKeyHex)
	if err != nil {
		return nil, err
	}

	v2 := paseto.NewV2()
	var payload map[string]interface{}
	err = v2.Verify(tokenString, publicKey, &payload, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось верифицировать токен: %w", err)
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
