package app

import (
	"log/slog"

	"sso/sso/internal/app/cron"
	grpcapp "sso/sso/internal/app/gprc"
	"sso/sso/internal/app/telegram_callback_auth"
	"sso/sso/internal/config"
	"sso/sso/internal/lib/kafka"
	"sso/sso/internal/lib/ratelimiter"
	"sso/sso/internal/repository/postgres"
	"sso/sso/internal/repository/redis"
	"sso/sso/internal/services/auth"
	"sso/sso/internal/services/transactions"
	"sso/sso/internal/services/user"
)

type App struct {
	GRPCSrv                   *grpcapp.App
	TelegramCallbackServer    *telegram_callback_auth.App
	ExpiredReservationsWorker *cron.ExpiredReservationsWorker
}

func New(log *slog.Logger, cfg *config.Config) *App {
	redisRepo := redis.New(cfg)

	log.Info("Initializing PostgreSQL storage",
		"host", cfg.Postgres.Host,
		"port", cfg.Postgres.Port,
		"database", cfg.Postgres.DBName,
	)
	postgresRepo, err := postgres.New(cfg.Postgres)
	if err != nil {
		panic(err)
	}

	rateLimiter := ratelimiter.NewRateLimiter(redisRepo.Client)
	kafkaProducer, err := kafka.NewKafkaProducer(cfg.Kafka.Brokers)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, cfg, kafkaProducer, postgresRepo, postgresRepo, postgresRepo, redisRepo)
	userService := user.New(log, postgresRepo, redisRepo, cfg)
	transactionsService := transactions.New(log, cfg, redisRepo, postgresRepo)

	grpcApp := grpcapp.New(log, cfg, kafkaProducer, rateLimiter, authService, userService, transactionsService)
	telegramApp := telegram_callback_auth.New(log, cfg, authService)

	// Cron worker для отмены истёкших резервирований
	expiredWorker := cron.NewExpiredReservationsWorker(
		log,
		postgresRepo,
		cfg.Cron.ExpiredReservationsInterval,
		cfg.Cron.ExpiredReservationsBatchSize,
	)

	return &App{
		GRPCSrv:                   grpcApp,
		TelegramCallbackServer:    telegramApp,
		ExpiredReservationsWorker: expiredWorker,
	}
}
