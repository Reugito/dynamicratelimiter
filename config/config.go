package config

import (
	"time"
)

type RateLimitConfig struct {
	RedisHost     string // Redis server host (e.g., "localhost")
	RedisPort     string // Redis server port (e.g., "6379")
	RedisPassword string // Password for Redis authentication (leave empty if not required)
	RedisDB       int    // Redis database index (0 by default)
	RedisHashName string // Redis hash key used for storing rate limit data

	EnableRedis bool // If true, rate limit configurations will be stored and managed in Redis

	MaxRateLimit int // Maximum allowed rate limit (requests per second) across all clients

	DefaultLimit int           // Default rate limit (requests per second) when no specific limit is set
	TimeFrame    time.Duration // Time frame for monitoring exceeding requests before adjusting the rate limit
	IPThreshold  int           // Number of unique IPs that must exceed the limit before increasing the rate limit

	IncreaseFactor            int  // Addition factor for increasing the rate limit dynamically
	EnableDynamicRateLimiting bool // If true, enables dynamic rate limiting based on traffic patterns
}
