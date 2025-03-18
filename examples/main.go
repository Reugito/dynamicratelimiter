package main

import (
	"github.com/Reugito/dynamicratelimiter/config"
	"github.com/Reugito/dynamicratelimiter/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Initialize rate limiter
	limiter := middleware.NewRateLimiter(config.RateLimitConfig{
		RedisHost:     "127.0.0.1",
		RedisPort:     "6379",
		RedisPassword: "",

		EnableRedis:             false,
		RedisHashName:           "request_stats",
		EnableDynamicMonitoring: false,
		DefaultLimit:            3,
		IPThreshold:             3,
	})

	// Apply middleware
	r.Use(limiter.Middleware())

	r.GET("/example", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})

	// Run API server
	r.Run(":8080")
}
