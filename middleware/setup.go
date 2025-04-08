package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/Reugito/dynamicratelimiter/config"
	"github.com/Reugito/dynamicratelimiter/internal/storage"
)

// NewRateLimiter initializes a new rate limiter instance
func NewRateLimiter(cfg config.RateLimitConfig) *rateLimiter {
	rl := &rateLimiter{config: cfg}

	setDefaultConfigValues(rl)

	if cfg.Redis.EnableRedis {
		setupRedis(rl)
	} else {
		fmt.Println("✅✅ Using in-memory rate limiter...")
	}

	// Start cleanup and monitoring once
	rl.once.Do(func() {
		initializeMonitoring(rl)
	})

	return rl
}

func setupRedis(rl *rateLimiter) {
	fmt.Println("🚀 Using Redis as rate limiter backend...")
	redisClient := storage.NewRedisService(rl.config.Redis.Host, rl.config.Redis.Port, rl.config.Redis.Password, rl.config.Redis.DatabaseIndex)
	rl.redisClient = redisClient
	if redisClient != nil {
		fmt.Println("✅ ✅ Loading rate limits from Redis...")
		if rl.config.Redis.RateLimitKey == "" {
			rl.config.Redis.RateLimitKey = "rate_limits_config"
			fmt.Println("⚠️ No Redis hash name provided, using default:", rl.config.Redis.RateLimitKey)
		}
		rl.loadRateLimitsFromRedis()
		rl.redisClient.CreateRedisHash(context.Background(), rl.config.Redis.RateLimitKey)
	} else {
		fmt.Println("❌ Failed to connect to Redis, falling back to in-memory rate limiter...")
		rl.config.Redis.EnableRedis = false
	}
}

func setDefaultConfigValues(rl *rateLimiter) {
	if rl.config.RateLimits.DefaultRequestsPerSec == 0 {
		rl.config.RateLimits.DefaultRequestsPerSec = 5
		fmt.Println("⚠️ No default rate limit provided, using default:", rl.config.RateLimits.DefaultRequestsPerSec)
	}

	if rl.config.EnableAdaptiveRateLimit {
		if rl.config.RateLimits.GlobalMaxRequestsPerSec == 0 {
			rl.config.RateLimits.GlobalMaxRequestsPerSec = 15
			fmt.Println("⚠️ No max rate limit provided, using default:", rl.config.RateLimits.GlobalMaxRequestsPerSec)
		}
		if rl.config.RateLimits.MonitoringTimeFrame == 0 {
			rl.config.RateLimits.MonitoringTimeFrame = 1 * time.Minute
			fmt.Println("⚠️ No time frame provided, using default:", rl.config.RateLimits.MonitoringTimeFrame)
		}
		if rl.config.RateLimits.IPExceedThreshold == 0 {
			rl.config.RateLimits.IPExceedThreshold = 2
			fmt.Println("⚠️ No IP threshold provided, using default:", rl.config.RateLimits.IPExceedThreshold)
		}
		if rl.config.RateLimits.IncreaseFactor == 0 {
			rl.config.RateLimits.IncreaseFactor = 1
			fmt.Println("⚠️ No increase factor provided, using default:", rl.config.RateLimits.IncreaseFactor)
		}
	}
}

func initializeMonitoring(rl *rateLimiter) {
	go rl.cleanupOldClients()
	go rl.periodicRateLimitCleanup()

	if rl.config.Redis.EnableRedis {
		go rl.dumpRateLimitsToRedis() // Monitor request counts
	}
	if rl.config.EnableAdaptiveRateLimit {
		go rl.monitorExceededLimits() // Monitor IP exceed limits
	}
}
