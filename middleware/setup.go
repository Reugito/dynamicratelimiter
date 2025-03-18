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

	if cfg.EnableRedis {
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
	redisClient := storage.NewRedisService(rl.config.RedisHost, rl.config.RedisPort, rl.config.RedisPassword, rl.config.RedisDB)
	rl.redisClient = redisClient
	if redisClient != nil {
		fmt.Println("✅✅ Loading rate limits from Redis...")
		if rl.config.RedisHashName == "" {
			rl.config.RedisHashName = "ratelimits"
			fmt.Println("⚠️ No Redis hash name provided, using default:", rl.config.RedisHashName)
		}
		rl.loadRateLimitsFromRedis()
		rl.redisClient.CreateRedisHash(context.Background(), rl.config.RedisHashName)
	} else {
		fmt.Println("❌ Failed to connect to Redis, falling back to in-memory rate limiter...")
		rl.config.EnableRedis = false
	}
}

func setDefaultConfigValues(rl *rateLimiter) {
	if rl.config.DefaultLimit == 0 {
		rl.config.DefaultLimit = 5
		fmt.Println("⚠️ No default rate limit provided, using default:", rl.config.DefaultLimit)
	}

	if rl.config.EnableDynamicRateLimiting {
		if rl.config.MaxRateLimit == 0 {
			rl.config.MaxRateLimit = 15
			fmt.Println("⚠️ No max rate limit provided, using default:", rl.config.MaxRateLimit)
		}
		if rl.config.TimeFrame == 0 {
			rl.config.TimeFrame = 1 * time.Minute
			fmt.Println("⚠️ No time frame provided, using default:", rl.config.TimeFrame)
		}
		if rl.config.IPThreshold == 0 {
			rl.config.IPThreshold = 2
			fmt.Println("⚠️ No IP threshold provided, using default:", rl.config.IPThreshold)
		}
		if rl.config.IncreaseFactor == 0 {
			rl.config.IncreaseFactor = 1
			fmt.Println("⚠️ No increase factor provided, using default:", rl.config.IncreaseFactor)
		}
	}
}

func initializeMonitoring(rl *rateLimiter) {
	go rl.cleanupOldClients()

	if rl.config.EnableRedis {
		go rl.dumpRateLimitsToRedis() // Monitor request counts
	}
	if rl.config.EnableDynamicRateLimiting {
		go rl.monitorExceededLimits() // Monitor IP exceed limits
	}
}
