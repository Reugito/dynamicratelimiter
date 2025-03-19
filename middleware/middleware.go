package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/Reugito/dynamicratelimiter/config"
	"github.com/Reugito/dynamicratelimiter/internal/storage"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type rateLimiter struct {
	config      config.RateLimitConfig
	redisClient *storage.RedisService
	clients     sync.Map // Tracks request counts per IP + endpoint

	rateLimits sync.Map

	mu           sync.Mutex
	rateLimitsMu sync.Mutex // Lock for rateLimits

	once sync.Once
	// requestStats sync.Map // Tracks request counts per IP + endpoint:  count of requests  souly for monitoring purpose no responsibility in ratelimiter
	exceedingIPs sync.Map // Tracks unique IPs exceeding limits per endpoint
}

type RateLimiterClient struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Middleware applies rate limiting using Redis (if enabled) or in-memory limiters
func (rl *rateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip, err := getNetworkIP(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Failed to find IP address"})
			return
		}

		endpoint := c.FullPath()
		clientKey := getClientKey(ip, endpoint)

		// In-memory rate limiting
		rl.mu.Lock()
		clientIface, exists := rl.clients.Load(clientKey)
		var client *RateLimiterClient
		if !exists {
			rateLimit, exists := rl.rateLimits.Load(endpoint)
			if !exists {
				rateLimit = rl.config.DefaultLimit
				rl.rateLimits.Store(endpoint, rateLimit)
			}
			client = &RateLimiterClient{
				limiter:  rate.NewLimiter(rate.Limit(rateLimit.(int)), 1),
				lastSeen: time.Now(),
			}
			rl.clients.Store(clientKey, client)
		} else {
			client = clientIface.(*RateLimiterClient)
		}
		client.lastSeen = time.Now()
		rl.mu.Unlock()

		// Use Wait for precise rate limits, else fallback to Allow
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := client.limiter.Wait(ctx); err != nil {
			// Track request count
			if rl.config.EnableDynamicRateLimiting {
				// rl.incrementRequestCount(clientKey)
				rl.trackExceededIP(ip, endpoint)
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Next()
	}
}
