package telegram_callback_auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sso/sso/internal/config"
	telegram_http_callback "sso/sso/internal/http"
	"sso/sso/internal/services/auth"
)

type App struct {
	log    *slog.Logger
	server *http.Server
	cfg    *config.Config
}

func New(log *slog.Logger, cfg *config.Config, service auth.TelegramService) *App {
	// регистрация сервиса
	// регистрация хэндлеров
	handlers := telegram_http_callback.New(log, cfg, service)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback/telegram/auth", handlers.TelegramCallbackHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Telegram.CallbackPort),
		Handler: mux,
	}

	return &App{
		log:    log,
		cfg:    cfg,
		server: server,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "telegram_callback_auth.Run"

	a.log.With(
		slog.String("op", op),
		slog.Int("port", a.cfg.Telegram.CallbackPort),
	)

	a.log.Info("HTTP server is running", slog.String("addr", a.server.Addr), slog.String("op", op))

	if err := a.server.ListenAndServe(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "telegram_callback_auth.Stop"
	if err := a.server.Shutdown(context.Background()); err != nil {
		a.log.Error("failed to stop HTTP server", slog.String("error", err.Error()), slog.String("op", op))
	}
}
