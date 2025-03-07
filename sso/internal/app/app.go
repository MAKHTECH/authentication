package app

import (
	"log/slog"
	grpcapp "sso/sso/internal/app/gprc"
	"sso/sso/internal/config"
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

	authService := auth.New(log, cfg, storage, storage, storage, rStorage)
	userService := user.New(log, storage, rStorage, cfg)

	grpcApp := grpcapp.New(log, cfg, authService, userService)

	return &App{
		GRPCSrv: grpcApp,
	}
}
