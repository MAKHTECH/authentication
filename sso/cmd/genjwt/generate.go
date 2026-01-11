package main

import (
	"fmt"
	"sso/sso/internal/config"
	user_jwt "sso/sso/internal/lib/jwt"
	"time"

	"github.com/o1egl/paseto"
)

func main() {
	cfg := config.MustLoad()

	access, err := generateAccessToken(cfg.PrivateKey)
	if err != nil {
		panic(err)
	}
	fmt.Println(access)
}

// generateAccessToken генерирует PASETO v2.public access токен для пользователя.
func generateAccessToken(privateKeyHex string) (string, error) {
	privateKey, _, err := user_jwt.GetKeyPair(privateKeyHex)
	if err != nil {
		return "", err
	}

	v2 := paseto.NewV2()
	claims := map[string]interface{}{
		"sub":       2,
		"app_id":    1,
		"username":  "makhkets",
		"photo_url": "",
		"role":      int32(2), // ADMIN = 2 (согласно enum Role в proto)
		"exp":       time.Now().Add(time.Hour * 24000).Unix(),
	}

	token, err := v2.Sign(privateKey, claims, nil)
	if err != nil {
		return "", fmt.Errorf("не удалось подписать токен: %w", err)
	}
	return token, nil
}
