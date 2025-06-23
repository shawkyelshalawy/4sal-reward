package cache

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient() *redis.Client {
	addr := getEnvOrDefault("REDIS_ADDR", "localhost:6379")
	password := os.Getenv("REDIS_PASSWORD")
	
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func HealthCheck(client *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := client.Ping(ctx).Result()
	return err
}