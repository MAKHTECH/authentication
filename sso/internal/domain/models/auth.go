package models

type AuthUser struct {
	Email    string
	Username string
	Password string
	AppID    int32
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
