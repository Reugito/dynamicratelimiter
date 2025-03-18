package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

func getClientKey(ip, endpoint string) string {
	return fmt.Sprintf("%s@%s", ip, endpoint)
}

func getNetworkIP(c *gin.Context) (string, error) {
	ip := c.ClientIP()
	if ip == "" {
		return "", errors.New("failed to get client IP")
	}
	return ip, nil
}

// trackRequest increments the request count per IP and endpoint
func (rl *RateLimiter) incrementRequestCount(clientKey string) {
	countIface, _ := rl.requestStats.LoadOrStore(clientKey, 0)
	count := countIface.(int) + 1
	rl.requestStats.Store(clientKey, count)
}

// RateLimitMetricsHandler exposes request stats for monitoring
func (rl *RateLimiter) RateLimitMetricsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := make(map[string]int)
		rl.requestStats.Range(func(key, value interface{}) bool {
			stats[key.(string)] = value.(int)
			return true
		})

		c.JSON(http.StatusOK, stats)
	}
}

func (rl *RateLimiter) loadRateLimitsFromRedis() {
	ctx := context.Background()

	// Fetch all rate limits in a single Redis request
	rateLimits, err := rl.redisClient.FetchFromRedisHash(ctx, "rate_limits")
	if err != nil {
		fmt.Println("❌ Failed to load rate limits:", err)
		return
	}

	// Create a worker pool of 10 goroutines
	const maxWorkers = 10
	workerPool := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for endpoint, limitStr := range rateLimits {
		workerPool <- struct{}{} // Block if pool is full
		wg.Add(1)

		go func(ep, lim string) {
			defer wg.Done()
			defer func() { <-workerPool }() // Free up a worker slot

			if limit, err := strconv.Atoi(lim); err == nil {
				if limit > 0 {
					rl.rateLimits.Store(ep, limit)
				}
			} else {
				fmt.Printf("⚠️ Invalid rate limit for %s: %v\n", ep, lim)
			}
		}(endpoint, limitStr)
	}

	wg.Wait()
}
