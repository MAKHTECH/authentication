package telegram_http_callback

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sso/sso/internal/config"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/lib/logger/sl"
	"sso/sso/internal/services/auth"

	telegramloginwidget "github.com/LipsarHQ/go-telegram-login-widget"
)

type Server interface {
	TelegramCallbackHandler(writer http.ResponseWriter, req *http.Request)
}

type server struct {
	log     *slog.Logger
	config  *config.Config
	service auth.TelegramService
}

func New(logger *slog.Logger, cfg *config.Config, service auth.TelegramService) Server {
	return &server{
		log:     logger,
		config:  cfg,
		service: service,
	}
}

type User struct {
	Id        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type payload struct {
	User User `json:"user"`
}

// test data
/*
https://auth.makhkets.ru/?id=5285375327&first_name=Makhkets%20%F0%9F%97%BD&username=Makhkets&photo_url=https://t.me/i/userpic/320/ZjWaWT3rxUckT5cSElSrVZaJDWrI8ArA1Ovbpoxv3UJBVxlxx75BU1C9GpMBrlBA.jpg&auth_date=1766425258&hash=106e836c6513ad26dfdbcb82471fc77182c8def2d5b9754f6d79701a6acc9384
https://localhost:8099/callback/telegram/auth?id=5285375327&first_name=Makhkets%20%F0%9F%97%BD&username=Makhkets&photo_url=https://t.me/i/userpic/320/ZjWaWT3rxUckT5cSElSrVZaJDWrI8ArA1Ovbpoxv3UJBVxlxx75BU1C9GpMBrlBA.jpg&auth_date=1766425258&hash=106e836c6513ad26dfdbcb82471fc77182c8def2d5b9754f6d79701a6acc9384
*/

func (s *server) TelegramCallbackHandler(w http.ResponseWriter, r *http.Request) {
	modelAuthorizationData, err := telegramloginwidget.NewFromURI(r.URL.String())
	if err != nil {
		slog.Error("Failed to parse Telegram login data", sl.Err(err))
		return
	}

	// validate hash
	if err = modelAuthorizationData.Check(s.config.Telegram.Token); err != nil {
		s.log.Error("Invalid hash", sl.Err(err))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Получаем данные для контекста
	clientIP := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		clientIP = xff
	}
	userAgent := r.Header.Get("User-Agent")
	fingerprint := r.Header.Get("X-Fingerprint")
	if fingerprint == "" {
		fingerprint = "telegram-widget"
	}

	// Создаём контекст с необходимыми значениями
	ctx := context.WithValue(context.Background(), "fingerprint", fingerprint)
	ctx = context.WithValue(ctx, "ip", clientIP)
	ctx = context.WithValue(ctx, "user-agent", userAgent)

	tokenPair, err := s.service.LoginTelegram(ctx, models.TelegramAuthUser{
		TelegramID: modelAuthorizationData.ID,
		FirstName:  modelAuthorizationData.FirstName,
		LastName:   modelAuthorizationData.LastName,
		Username:   modelAuthorizationData.Username,
		PhotoURL:   modelAuthorizationData.PhotoURL,
		AuthDate:   modelAuthorizationData.AuthDate,
		AppID:      1,
	})
	if err != nil {
		s.log.Error("Failed to login via Telegram", sl.Err(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Логируем успешный ответ
	s.log.Info("Sending token pair response", slog.Any("tokenPair", tokenPair))

	responseData, err := json.Marshal(tokenPair)
	if err != nil {
		s.log.Error("Failed to marshal response", sl.Err(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.log.Info("Response JSON", slog.String("json", string(responseData)))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(responseData)
	if err != nil {
		s.log.Error("Failed to write response", sl.Err(err))
		return
	}

	s.log.Info("Response sent successfully")
}
