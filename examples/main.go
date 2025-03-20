package main

import (
	"time"

	"github.com/Reugito/dynamicratelimiter/config"
	"github.com/Reugito/dynamicratelimiter/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Initialize rate limiter
	limiter := middleware.NewRateLimiter(config.RateLimitConfig{
		Redis: config.RedisConfig{
			EnableRedis:   true,
			Host:          "localhost",
			Port:          "6379",
			Password:      "",
			DatabaseIndex: 0,
			RateLimitKey:  "rate_limit_data",
		},
		RateLimits: config.RateLimitSettings{
			GlobalMaxRequestsPerSec: 20,
			DefaultRequestsPerSec:   5,
			MonitoringTimeFrame:     time.Minute,
			IPExceedThreshold:       5,
			IncreaseFactor:          2,
		},
		EnableAdaptiveRateLimit: true,
	},
	)

	// Apply middleware
	r.Use(limiter.Middleware())

	r.GET("/example", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})

	// Run API server
	r.Run(":8080")
}
