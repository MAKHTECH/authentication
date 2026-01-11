package main

import (
	"log/slog"
	"os"
	"os/signal"
	"sso/sso/internal/app"
	"sso/sso/internal/config"
	"sso/sso/internal/domain/custom_models"
	"sso/sso/internal/lib/logger/handlers/slogpretty"
	"sso/sso/pkg/directories"
	"sso/sso/pkg/migrations"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting application", slog.Any("config", cfg))

	// Применяем миграции при старте приложения
	migrationsPath := directories.FindDirectoryName("migrations")
	err := migrations.ApplyMigrations(
		migrations.PostgresConfig{
			Host:     cfg.Postgres.Host,
			Port:     cfg.Postgres.Port,
			User:     cfg.Postgres.User,
			Password: cfg.Postgres.Password,
			DBName:   cfg.Postgres.DBName,
			SSLMode:  cfg.Postgres.SSLMode,
		},
		migrationsPath,
		"migrations",
	)
	if err != nil {
		log.Error("failed to apply migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	application := app.New(log, cfg)
	go application.GRPCSrv.MustRun()
	go application.TelegramCallbackServer.MustRun()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping application", slog.String("signal", sign.String()))

	application.GRPCSrv.Stop()
	log.Info("application stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}))
	case envProd:

		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	// Создание файла для логов
	file, err := os.OpenFile(
		directories.FindDirectoryName(
			"logger",
		)+
			"\\logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666,
	)
	if err != nil {
		panic(err)
	}
	//defer file.Close()
	customWriter := &custom_models.CustomFileWriter{File: file}
	handler := opts.NewPrettyHandler(os.Stdout, customWriter)

	return slog.New(handler)
}
