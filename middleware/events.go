package middleware

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// dumpRateLimitsToRedis logs the number of requests per endpoint per IP
func (rl *rateLimiter) dumpRateLimitsToRedis() {
	ticker := time.NewTicker(10 * time.Second) // Adjust frequency as needed
	defer ticker.Stop()

	fmt.Println("Running ticker for Dumping rate limits to Redis...")
	for range ticker.C {
		hashSet := make(map[string]string)
		rl.rateLimits.Range(func(key, value interface{}) bool {
			hashSet[fmt.Sprintf("%s", key)] = fmt.Sprintf("%d", value.(int))

			return true
		})
		rl.redisClient.SaveToRedisHash(context.Background(), rl.config.Redis.RateLimitKey, hashSet, 0)

	}
}

// logClients periodically prints stored rate limiters for debugging
func (rl *rateLimiter) logClients() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		rl.clients.Range(func(key, value interface{}) bool {
			client := value.(*rateLimiterClient)
			fmt.Printf("Client Key: %s, Rate: %v\n", key, client.limiter)
			return true
		})
	}
}

// cleanupOldClients removes stale clients periodically
func (rl *rateLimiter) cleanupOldClients() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	fmt.Println("Running ticker for cleaning up old clients...")
	for range ticker.C {
		rl.clientsMu.Lock()
		now := time.Now()
		rl.clients.Range(func(key, value interface{}) bool {
			client := value.(*rateLimiterClient)
			if now.Sub(client.lastSeen) > 5*time.Second {
				rl.clients.Delete(key)
			}
			return true
		})
		rl.clientsMu.Unlock()
	}
}

func (rl *rateLimiter) periodicRateLimitCleanup() {
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	timeUntilMidnight := time.Until(nextMidnight)

	// Schedule first run at midnight
	time.AfterFunc(timeUntilMidnight, func() {
		for {
			fmt.Println("🔄 Running periodic cleanup at midnight...")
			rl.rateLimits.Clear()
			if rl.config.Redis.EnableRedis {
				rl.loadRateLimitsFromRedis()
			}

			// Sleep for 24 hours before next cleanup
			time.Sleep(24 * time.Hour)
		}
	})
}

func (rl *rateLimiter) clearClients() {
	rl.clientsMu.Lock()
	rl.clients.Clear()
	rl.clientsMu.Unlock()
}

// trackExceededIP tracks unique IPs exceeding the rate limit for an endpoint
func (rl *rateLimiter) trackExceededIP(ip, endpoint string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	ipSet, _ := rl.exceedingIPs.LoadOrStore(endpoint, &sync.Map{})
	ipSet.(*sync.Map).Store(ip, struct{}{})
}

// monitorExceededLimits checks if any endpoint has more than 10 unique IPs exceeding their limit
func (rl *rateLimiter) monitorExceededLimits() {
	ticker := time.NewTicker(rl.config.RateLimits.MonitoringTimeFrame)
	defer ticker.Stop()
	fmt.Println("Running ticker for monitoring exceeded limits...")
	for range ticker.C {
		rl.exceedingIPs.Range(func(endpoint, ipMap interface{}) bool {
			count := 0
			exceedingIPsArr := []string{}

			ipMap.(*sync.Map).Range(func(key, _ interface{}) bool {
				count++
				exceedingIPsArr = append(exceedingIPsArr, key.(string))
				return true
			})
			if count >= rl.config.RateLimits.IPExceedThreshold {
				fmt.Printf("⚠️ High traffic detected! More than %d unique IPs exceeded rate limits for endpoint: %s\n", rl.config.RateLimits.IPExceedThreshold, endpoint)
				a, ok := rl.rateLimits.Load(endpoint)
				if !ok {
					return true
				}
				currentLimit, ok := a.(int)
				if !ok {
					return true
				}
				if currentLimit < rl.config.RateLimits.GlobalMaxRequestsPerSec {
					newLimit := currentLimit + rl.config.RateLimits.IncreaseFactor
					rl.rateLimits.Store(endpoint, newLimit)
					if rl.config.Redis.EnableRedis {
						go rl.logExceedingIPsInRedis(endpoint.(string), currentLimit, newLimit, exceedingIPsArr)
					}
					rl.clearClients()
				}

			}
			return true
		})
		rl.exceedingIPs.Clear()
	}
}

func (rl *rateLimiter) logExceedingIPsInRedis(endpoint string, rateLimit, newLimit int, exceedingIPs []string) {
	data := map[string]string{
		"rate_limit_log": endpoint,
		"timestamp":      time.Now().Format(time.RFC3339),
		"previous_limit": fmt.Sprintf("%d", rateLimit),
		"new_limit":      fmt.Sprintf("%d", newLimit),
		"exceeding_ips":  strings.Join(exceedingIPs, ","),
	}
	rl.redisClient.PushToList(context.Background(), "ratelimit_log_"+endpoint, data, 0)
}
