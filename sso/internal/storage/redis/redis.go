package rstorage

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RClient struct {
	Client *redis.Client
}

func InitRedis(db int, address, password string) *RClient {
	// создаем экземпляр клиента Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return &RClient{Client: rdb}
}
