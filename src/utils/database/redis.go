package database

import (
	"context"
	"fmt"
	"time"

	"ticket-booking-system/src/utils"

	"github.com/redis/go-redis/v9"
)

// RedisClient represents a Redis client
type RedisClient struct {
	Client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config *utils.Config) (*RedisClient, error) {
	// Create Redis options
	opts := &redis.Options{
		Addr:     config.GetRedisAddr(),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	}

	// Create Redis client
	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{Client: client}, nil
}

// Close closes the Redis connection
func (c *RedisClient) Close() error {
	return c.Client.Close()
}

// Ping tests the Redis connection
func (c *RedisClient) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// Health checks Redis health
func (c *RedisClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.Ping(ctx)
}
