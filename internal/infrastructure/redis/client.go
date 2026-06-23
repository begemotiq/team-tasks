package redis

import (
	goredis "github.com/redis/go-redis/v9"

	"task-service/internal/config"
)

func NewClient(cfg config.RedisConfig) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
}
