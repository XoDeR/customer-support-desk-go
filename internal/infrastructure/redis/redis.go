package redis

import (
	"context"

	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/config"
	"github.com/redis/go-redis/v9"
)

func New(c *config.Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{Addr: c.Redis.Address, Password: c.Redis.Password, DB: c.Redis.DB})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return client, nil
}
