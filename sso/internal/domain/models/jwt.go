package models

import (
	ssov1 "sso/protos/gen/go/sso"
	"time"
)

type RefreshSession struct {
	RefreshToken string        `json:"refreshToken"`
	UserId       string        `json:"userId"`
	Ua           string        `json:"ua"`
	Ip           string        `json:"ip"`
	Fingerprint  string        `json:"fingerprint"`
	ExpiresIn    time.Duration `json:"expiresIn"`
	CreatedAt    time.Time     `json:"createdAt"`
}

type AccessTokenData struct {
	Username string
	Email    *string // может быть nil для Telegram авторизации
	PhotoURL string
	Role     ssov1.Role
	UserID   int64
	AppID    int32
	Balance  int64 // баланс в копейках
	Exp      int64
}

//claims["sub"] = user.ID
//claims["app_id"] = user.AppID
//claims["username"] = user.Username
//claims["exp"] = time.Now().Add(time.Minute * duration).Unix()
