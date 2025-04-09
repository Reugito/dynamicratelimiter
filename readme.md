📌 Dynamic Rate Limiter for Gin Framework
=========================================

A **highly configurable, endpoint-based rate limiter** for the **Gin framework** in Golang. Supports **Redis or in-memory storage**, **dynamic scaling based on API usage**, and **fine-grained user control**.

📚 Table of Contents
--------------------

*   Features
    
*   Installation
    
*   Configuration
    
*   Usage
    
*   Dynamic Monitoring
    
*   API Integration
    
*   Future Enhancements
    
*   License
    

🎯 **Features**  
-----------  

1. **Per-Endpoint Rate Limiting** - Limits requests based on IP & endpoint.  
2. **Dynamic Scaling** - Automatically increases or decreases rate limits based on API usage.  
3. **User-Controlled Monitoring** - Enable or disable dynamic scaling via configuration.  
4. **Configurable Storage** - Supports **Redis** or **in-memory map** storage.  
5. **Customizable Thresholds** - Adjust rate limits based on API traffic patterns.  
6. **Efficient & Scalable** - Optimized for high performance with **Gin**.  
7. **Analysis & Tracking** - Users can check `ratelimit_log_*(endpoint path)` in Redis to monitor how many times the rate limit increased and identify which IPs are responsible for exceeding limits.  

   ```json
   {
       "exceeding_ips": "xxx.xxx.xx.xx, yyy.yyy.yy.yy",
       "new_limit": "6",
       "previous_limit": "5",
       "rate_limit_log": "/abc/efg",
       "timestamp": "2025-04-02T17:33:06+05:30"
   }
   ```

8. **🔴 Persistent Rate Limits** - Redis is used solely for storing rate limits persistently. If the system restarts, the rate limits will be restored from Redis, ensuring continuity. Users have full access to manage these limits.


⚙️ Installation
---------------

### **1️⃣ Install Package**

Download the package using Go modules.
```sh
go get github.com/Reugito/dynamicratelimiter
```

### **2️⃣ Import into Your Project**

Import the package to use it within your Gin-based API.
```go
import (
    "github.com/Reugito/dynamicratelimiter"
)
```

🔧 Configuration
----------------
1. **This configuration enables dynamic rate limiting with Redis and monitoring. The rate limiter adjusts limits based on API consumption patterns**
   
3. **Define rate-limiting settings using a structured configuration file. The settings allow you to specify the default rate limit, time frame, thresholds for dynamic scaling, and whether to enable monitoring.**

## Setup
The `SetupRateLimiter` function initializes and configures the rate limiter using environment variables.

## Configuration Details

- **Redis Configuration**
  - `EnableRedis`: Enables or disables Redis integration.
  - `Host`: Redis server hostname.
  - `Port`: Redis server port.
  - `Password`: Redis authentication password.
  - `DatabaseIndex`: Redis database index.
  - `RateLimitKey`: Key for storing rate limiter configurations in Redis.

- **Rate Limit Settings**
  - `GlobalMaxRequestsPerSec`: Maximum allowed requests per second globally. i.e dynamic ratlimiting will not breach this limit ± DefaultRequestsPerSec.
  - `DefaultRequestsPerSec`: Default rate limit per second.
  - `MonitoringTimeFrame`: Time frame (in seconds) for monitoring API consumption.
  - `IncreaseFactor`: Factor by which rate limits increase when thresholds are exceeded.
  - `IPExceedThreshold`: Number of IPs to monitor for increasing the rate limit.

- **Adaptive Rate Limiting**
  - `EnableAdaptiveRateLimit`: Enables dynamic scaling of rate limits.

## Code Implementation
 **1. This configuration enables dynamic rate limiting with Redis and monitoring. The rate limiter adjusts limits based on API consumption patterns.**
```go
import (
	"github.com/Reugito/dynamicratelimiter/config"
	"github.com/Reugito/dynamicratelimiter/middleware"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
	"time"
	)
func SetupRateLimiter(router *gin.Engine) {
	redisHost, redisPort := os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")
	if redisHost == "" || redisPort == "" {
		return
	}

	globalMaxRateLimitInt, err := strconv.Atoi(os.Getenv("GLOBAL_MAX_RATE_LIMIT"))
	if err != nil {
		globalMaxRateLimitInt = 20
	}

	monitoringTimeFrameInt, err := strconv.Atoi(os.Getenv("MONITORING_TIME_FRAME_IN_SECONDS"))
	if err != nil {
		monitoringTimeFrameInt = 10
	}

	ipExceedThresholdInt, err := strconv.Atoi(os.Getenv("IP_EXCEED_THRESHOLD"))
	if err != nil {
		ipExceedThresholdInt = 2
	}

	rateLimitConf := config.RateLimitConfig{
		Redis: config.RedisConfig{
			EnableRedis:   true,
			Host:          redisHost,
			Port:          redisPort,
			Password:      os.Getenv("REDIS_PASSWORD"),
			DatabaseIndex: 0,
			RateLimitKey:  "api_rate_limit_config",
		},
		RateLimits: config.RateLimitSettings{
			GlobalMaxRequestsPerSec: globalMaxRateLimitInt,
			DefaultRequestsPerSec:   5,
			MonitoringTimeFrame:     time.Duration(monitoringTimeFrameInt) * time.Second,
			IncreaseFactor:          1,
			IPExceedThreshold:       ipExceedThresholdInt,
		},
		EnableAdaptiveRateLimit: true,
	}

	rl := middleware.NewRateLimiter(rateLimitConf)

	// attach ratelimiter to middleware
	router.Use(rl.Middleware())

	router.GET("/rate-limit-conf", rl.RateLimitMetricsHandler())
}
```

**2. Dynamic Rate Limiter with Monitoring and Without Redis**
```go
rateLimitConf := config.RateLimitConfig{
	Redis: config.RedisConfig{
		EnableRedis: false,
	},
	RateLimits: config.RateLimitSettings{
		GlobalMaxRequestsPerSec: globalMaxRateLimitInt,
		DefaultRequestsPerSec:   5,
		MonitoringTimeFrame:     time.Duration(monitoringTimeFrameInt) * time.Second,
		IncreaseFactor:          1,
		IPExceedThreshold:       ipExceedThresholdInt,
	},
	EnableAdaptiveRateLimit: true,
}
```

**3. Simple Rate Limiter Without Monitoring and Without Redis**
```go
rateLimitConf := config.RateLimitConfig{
	Redis: config.RedisConfig{
		EnableRedis: false,
	},
	RateLimits: config.RateLimitSettings{
		DefaultRequestsPerSec: 5,
	},
	EnableAdaptiveRateLimit: false,
}
```

   

🚀 Usage
--------

1️⃣ Initialize Rate Limiter

Set up the rate limiter by loading the configuration, initializing Redis (if used), and applying the middleware to your Gin router.

2️⃣ Apply Middleware

Attach the rate-limiting middleware to the API routes to enforce request limits per endpoint.

3️⃣ Enable Monitoring

Optionally start the dynamic monitoring system to automatically adjust limits based on traffic.


📊 Dynamic Monitoring

The **monitoring service** dynamically adjusts rate limits **based on API usage**. If enabled, it periodically evaluates traffic patterns and updates the limits accordingly.

 **How It Works:**

- Increases rate limits if API traffic exceeds a specified threshold.
- Prevents excessive requests from overwhelming the server.
- Can be enabled or disabled via configuration.
- Uses an adaptive approach to prevent sudden spikes in traffic from causing disruptions.
- Regularly logs and reports rate limit changes for better visibility and debugging.
- Optimizes API performance by balancing request load dynamically.

> **🔴 Important:** Redis is only used to store rate limits persistently. If the system restarts, the rate limits will be reloaded from Redis, ensuring continuity. Users have full access to manage these limits.

🔌 API Integration
------------------

To integrate the rate limiter into an existing Gin API:

1.  **Import and initialize** the package.
    
2.  **Load configurations** from a file or environment variables.
    
3.  **Apply the middleware** to the Gin router.
    
4.  **Enable monitoring** if required.
    

🔮 Future Enhancements
----------------------

✅ **Admin API to Enable/Disable Monitoring**
✅ **Dashboard for Real-Time Monitoring**
✅ **Custom Limits for Specific Endpoints**
✅ **Logging & Alert System**


📝 License
----------



💡 **Contributions Welcome!** Feel free to submit **PRs or suggestions**. 🚀
