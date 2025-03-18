package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// dumpRateLimitsToRedis logs the number of requests per endpoint per IP
func (rl *RateLimiter) dumpRateLimitsToRedis() {
	ticker := time.NewTicker(10 * time.Second) // Adjust frequency as needed
	defer ticker.Stop()

	fmt.Println("Running ticker for Dumping rate limits to Redis...")
	for range ticker.C {
		hashSet := make(map[string]string)
		rl.rateLimits.Range(func(key, value interface{}) bool {
			hashSet[fmt.Sprintf("%s", key)] = fmt.Sprintf("%d", value.(int))

			return true
		})
		rl.redisClient.SaveToRedisHash(context.Background(), rl.config.RedisHashName, hashSet, 15*time.Minute)

	}
}

// logClients periodically prints stored rate limiters for debugging
func (rl *RateLimiter) logClients() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		rl.clients.Range(func(key, value interface{}) bool {
			client := value.(*RateLimiterClient)
			fmt.Printf("Client Key: %s, Rate: %v\n", key, client.limiter)
			return true
		})
	}
}

// cleanupOldClients removes stale clients periodically
func (rl *RateLimiter) cleanupOldClients() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	fmt.Println("Running ticker for cleaning up old clients...")
	for range ticker.C {
		rl.rateLimitsMu.Lock()
		now := time.Now()
		rl.clients.Range(func(key, value interface{}) bool {
			client := value.(*RateLimiterClient)
			if now.Sub(client.lastSeen) > 5*time.Second {
				fmt.Printf("Removing old client: %s\n", key)
				rl.clients.Delete(key)
			}
			return true
		})
		rl.rateLimitsMu.Unlock()
	}
}

func (rl *RateLimiter) periodicRateLimitCleanup() {
	ticker := time.NewTicker(rl.config.TimeFrame)
	defer ticker.Stop()
	fmt.Println("Running ticker for periodic cleanup...")
	for range ticker.C {
		rl.rateLimitsMu.Lock()
		rl.clients.Clear()
		rl.rateLimitsMu.Unlock()
	}
}

// trackExceededIP tracks unique IPs exceeding the rate limit for an endpoint
func (rl *RateLimiter) trackExceededIP(ip, endpoint string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	ipSet, _ := rl.exceedingIPs.LoadOrStore(endpoint, &sync.Map{})
	ipSet.(*sync.Map).Store(ip, struct{}{})
}

// monitorExceededLimits checks if any endpoint has more than 10 unique IPs exceeding their limit
func (rl *RateLimiter) monitorExceededLimits() {
	ticker := time.NewTicker(rl.config.TimeFrame)
	defer ticker.Stop()
	fmt.Println("Running ticker for monitoring exceeded limits...")
	for range ticker.C {
		rl.exceedingIPs.Range(func(endpoint, ipMap interface{}) bool {
			count := 0
			ipMap.(*sync.Map).Range(func(key, _ interface{}) bool {
				count++
				return true
			})
			if count >= rl.config.IPThreshold {
				fmt.Printf("⚠️ High traffic detected! More than %d unique IPs exceeded rate limits for endpoint: %s\n", rl.config.IPThreshold, endpoint)
				a, ok := rl.rateLimits.Load(endpoint)
				if !ok {
					return true
				}
				currentLimit, ok := a.(int)
				if !ok {
					return true
				}
				if currentLimit < rl.config.MaxRateLimit {
					newLimit := currentLimit + rl.config.IncreaseFactor
					rl.rateLimits.Store(endpoint, newLimit)
				}
			}
			return true
		})
	}
}
