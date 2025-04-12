package rstorage

import (
	"context"
	"github.com/redis/go-redis/v9"
)

type RClient struct {
	Client *redis.Client
}

func InitRedis(db int, host, port string) *RClient {
	// создаем экземпляр клиента Redis
	var addr string = host + ":" + port
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	return &RClient{Client: rdb}
}
