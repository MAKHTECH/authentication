package user_jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	ssov1 "sso/protos/gen/go/sso"
	"sso/sso/internal/config"
	"sso/sso/internal/domain/models"
	"time"
)

// GenerateAccessToken generates a new JWT access token for the user
func GenerateAccessToken(user *models.User, duration time.Duration, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"app_id":   user.AppID,
		"username": user.Username,
		"role":     int32(user.Role),
		"exp":      time.Now().Add(duration).Unix(),
	})

	accessToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return accessToken, nil
}

func GenerateRefreshToken(user *models.User, duration time.Duration, secretKey string) (string, int64, error) {
	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)
	exp := time.Now().Add(time.Minute * duration).Unix()

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	claims["exp"] = exp

	// Sign the token
	refreshToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", 0, err
	}

	return refreshToken, exp, nil
}

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

// ParseToken parses and validates a JWT token and returns AccessTokenData
func ParseToken(tokenString string, isAccessToken bool, secretKey string) (*models.AccessTokenData, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Verify token validity
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Extract required fields
	sub, ok := claims["sub"].(float64) // JWT typically stores numbers as float64
	if !ok {
		return nil, fmt.Errorf("missing or invalid subject claim")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid expiration claim")
	}

	// Create AccessTokenData
	data := &models.AccessTokenData{
		UserID: int64(sub),
		Exp:    int64(exp),
	}

	// Additional checks for access token
	if isAccessToken {
		username, ok := claims["username"].(string)
		if !ok {
			return nil, fmt.Errorf("missing or invalid username claim")
		}

		appID, ok := claims["app_id"].(float64)
		if !ok {
			return nil, fmt.Errorf("missing or invalid app_id claim")
		}

		data.Username = username
		data.AppID = int32(appID)
		data.Role = ssov1.Role(int32(claims["role"].(float64)))
	}

	return data, nil
}
