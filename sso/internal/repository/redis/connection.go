package redis

import (
	"context"
	"sso/sso/internal/config"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// Repository представляет Redis репозиторий
type Repository struct {
	Client *redis.Client
	cfg    *config.Config
}

// New создает новое подключение к Redis
func New(cfg *config.Config) *Repository {
	addr := cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   cfg.Redis.DB,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return &Repository{Client: rdb, cfg: cfg}
}

// Close закрывает соединение с Redis
func (r *Repository) Close() error {
	return r.Client.Close()
}
