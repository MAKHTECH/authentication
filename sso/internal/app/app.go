package app

import (
	"log/slog"
	"strconv"

	grpcapp "sso/sso/internal/app/gprc"
	"sso/sso/internal/app/telegram_callback_auth"
	"sso/sso/internal/config"
	"sso/sso/internal/lib/kafka"
	"sso/sso/internal/lib/ratelimiter"
	"sso/sso/internal/services/auth"
	"sso/sso/internal/services/user"
	"sso/sso/internal/storage/postgres"
	redis "sso/sso/internal/storage/redis"
)

type App struct {
	GRPCSrv                *grpcapp.App
	TelegramCallbackServer *telegram_callback_auth.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	rStorage := redis.InitRedis(cfg.Redis.DB, cfg.Redis.Host, strconv.Itoa(cfg.Redis.Port))

	log.Info("Initializing PostgreSQL storage",
		"host", cfg.Postgres.Host,
		"port", cfg.Postgres.Port,
		"database", cfg.Postgres.DBName,
	)
	storage, err := postgres.New(cfg.Postgres)
	if err != nil {
		panic(err)
	}

	rateLimiter := ratelimiter.NewRateLimiter(rStorage.Client)
	kafkaProducer, err := kafka.NewKafkaProducer(cfg.Kafka.Brokers)
	if err != nil {
		panic(err)
	}

	authService := auth.New(
		log, cfg, kafkaProducer,
		storage, storage, storage, rStorage,
	)

	userService := user.New(
		log, storage, rStorage, cfg,
	)

	grpcApp := grpcapp.New(log, cfg, authService, userService, rateLimiter, kafkaProducer)
	telegramApp := telegram_callback_auth.New(log, cfg, authService)

	return &App{
		GRPCSrv:                grpcApp,
		TelegramCallbackServer: telegramApp,
	}
}
