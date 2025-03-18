package config

import (
	"time"
)

type RateLimitConfig struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	RedisHashName string // provide the redis hash name for storing the rate limit

	EnableRedis bool // true if you want to store the endpoint and ratelimit configs in redis

	MaxRateLimitRange int

	DefaultLimit            int
	TimeFrame               time.Duration
	IPThreshold             int
	IncreaseFactor          int
	DecreaseFactor          int
	EnableDynamicMonitoring bool
}
