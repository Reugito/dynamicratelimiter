package config

import "time"

type RateLimitConfig struct {
	Redis                   RedisConfig       // Redis-related settings
	RateLimits              RateLimitSettings // Rate-limiting rules
	EnableAdaptiveRateLimit bool              // If true, dynamically adjusts limits based on traffic
}

type RedisConfig struct {
	EnableRedis   bool   // If true, Redis will be used for rate limiting
	Host          string // Redis server hostname or IP address
	Port          string // Redis server port
	Password      string // Password for Redis authentication
	DatabaseIndex int    // Redis database index
	RateLimitKey  string // Redis key used for storing rate limit data
}

type RateLimitSettings struct {
	GlobalMaxRequestsPerSec int           // Maximum allowed requests per second globally
	DefaultRequestsPerSec   int           // Default rate limit if no specific limit is set
	MonitoringTimeFrame     time.Duration // Time frame for periodically monitoring excessive requests
	IPExceedThreshold       int           // Number of unique IPs exceeding limit before increasing the rate
	IncreaseFactor          int           // How much to increase the limit dynamically (new Limit = current limit  + IncreaseFactor)
}
