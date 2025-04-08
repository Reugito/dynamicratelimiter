package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RateLimitMetricsHandler exposes request stats for monitoring
func (rl *rateLimiter) RateLimitMetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := make(map[string]int)
		rl.rateLimits.Range(func(key, value interface{}) bool {
			stats[key.(string)] = value.(int)
			return true
		})

		c.JSON(http.StatusOK, stats)
	}
}

func (rl *rateLimiter) DefaultRequestsPerSec() gin.HandlerFunc {
	return func(c *gin.Context) {
		rl.rateLimits.Clear()
		rl.clients.Clear()
		c.JSON(http.StatusOK, gin.H{"default_requests_per_sec": rl.config.RateLimits.DefaultRequestsPerSec})
	}
}
