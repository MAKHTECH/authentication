package app

import (
	"log/slog"
	grpcapp "sso/sso/internal/app/gprc"
	"sso/sso/internal/config"
	"sso/sso/internal/lib/kafka"
	"sso/sso/internal/lib/ratelimiter"
	"sso/sso/internal/services/auth"
	"sso/sso/internal/services/user"
	redis "sso/sso/internal/storage/redis"
	"sso/sso/internal/storage/sqlite"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	rStorage := redis.InitRedis(cfg.Redis.DB, cfg.Redis.Addr, cfg.Redis.Password)
	log.Info("Initializing SQLite storage", "path", cfg.StoragePath)
	storage, err := sqlite.New(cfg.StoragePath)

	if err != nil {
		panic(err)
	}

	rateLimiter := ratelimiter.NewRateLimiter(rStorage.Client)
	kafkaProducer, err := kafka.NewKafkaProducer(cfg.Kafka.Brokers, "")
	if err != nil {
		panic(err)
	}
	defer kafkaProducer.Close()

	authService := auth.New(
		log, cfg, kafkaProducer,
		storage, storage, storage, rStorage,
	)

	userService := user.New(
		log, storage, rStorage, cfg,
	)

	grpcApp := grpcapp.New(log, cfg, authService, userService, rateLimiter, kafkaProducer)

	return &App{
		GRPCSrv: grpcApp,
	}
}
