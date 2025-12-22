package models

type AuthUser struct {
	Email    string
	Username string
	Password string
	AppID    int32
}

// TelegramAuthUser данные авторизации через Telegram
type TelegramAuthUser struct {
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	PhotoURL   string
	AuthDate   int64
	Hash       string
	AppID      int32
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
