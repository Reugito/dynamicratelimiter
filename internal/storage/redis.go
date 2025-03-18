package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService wraps the Redis client
type RedisService struct {
	client *redis.Client
}

// NewRedisService initializes a RedisService// NewRedisService initializes a RedisService
func NewRedisService(ip, port, password string, db int) *RedisService {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", ip, port),
		Password: password,
		DB:       db,
	})

	// Check Redis connection
	if err := redisClient.Ping(context.TODO()).Err(); err != nil {
		fmt.Println("❌ Failed to connect to Redis:", err)
		return nil // Fallback to in-memory rate limiting
	} else {
		fmt.Println("✅ Connected to Redis at", fmt.Sprintf("%s:%s", ip, port))
	}

	return &RedisService{
		client: redisClient,
	}
}

// Fetch data from Redis hash set// Fetch data from Redis hash set
func (r *RedisService) FetchFromRedisHash(ctx context.Context, key string) (map[string]string, error) {
	data, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *RedisService) SaveToRedisHash(ctx context.Context, key string, data map[string]string, ttl time.Duration) error {
	// Save hash data
	if err := r.client.HMSet(ctx, key, data).Err(); err != nil {
		return err
	}

	// Set expiration only if ttl > 0
	if ttl > 0 {
		if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisService) CreateRedisHash(ctx context.Context, key string) error {
	// Check if the hash set exists
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return err
	}

	// If the hash set does not exist, create an empty one
	if exists == 0 {
		if err := r.client.HSet(ctx, key, "init", "1").Err(); err != nil {
			return err
		}
	}

	return nil
}
