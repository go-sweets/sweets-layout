package server

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	hertz_server "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	app_config "github.com/go-sweets/sweets-layout/internal/config"
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/application/handlers"
	"github.com/go-sweets/sweets-layout/internal/middleware"
	"github.com/go-sweets/sweets-layout/internal/service"
)

func NewHertzServer(c *app_config.Config, helloService *service.HelloService) *hertz_server.Hertz {
	addr := fmt.Sprintf(":%d", c.Server.HTTPPort)

	// Performance optimized configuration
	opts := []config.Option{
		hertz_server.WithHostPorts(addr),
		// Network optimization
		hertz_server.WithReadTimeout(c.Server.ReadTimeout),
		hertz_server.WithWriteTimeout(c.Server.WriteTimeout),
		hertz_server.WithIdleTimeout(c.Server.IdleTimeout),
		hertz_server.WithMaxRequestBodySize(64 * 1024 * 1024), // 64MB
		hertz_server.WithKeepAliveTimeout(time.Minute * 2),
		// Enable HTTP/2
		hertz_server.WithH2C(true),
		// Stream configuration
		hertz_server.WithStreamBody(true),
		// Disable default recovery to use custom error handling
		hertz_server.WithDisableDefaultDate(true),
	}

	// Production optimizations
	if c.Server.Mode == "prod" {
		opts = append(opts,
			// Enable connection pooling optimizations
			hertz_server.WithDisablePrintRoute(true),
			// Note: Hertz already includes netpoll optimizations by default
		)
	}

	// Development mode optimizations
	if c.Server.Mode == "dev" && c.Server.Debug {
		opts = append(opts,
			hertz_server.WithDisableDefaultContentType(false),
		)
	}

	h := hertz_server.New(opts...)

	// Initialize middleware configurations
	authConfig := middleware.NewAuthConfig()
	loggingConfig := middleware.NewLoggingConfig()
	errorConfig := middleware.NewErrorHandlingConfig()

	// Register middleware chain (order matters!)
	h.Use(middleware.RequestIDMiddleware())                     // 1. Request ID first
	h.Use(middleware.HertzLoggingMiddleware(loggingConfig))     // 2. Logging
	h.Use(middleware.HertzErrorHandlingMiddleware(errorConfig)) // 3. Error handling
	h.Use(middleware.ValidationErrorMiddleware())               // 4. Validation errors
	h.Use(middleware.HertzAuthMiddleware(authConfig))           // 5. Authentication

	// Performance middleware
	// TODO: Implement compression middleware for Hertz
	// if c.FeatureFlags.EnableCompression {
	// 	h.Use(middleware.CompressionMiddleware())
	// }

	// TODO: Implement rate limiting middleware for Hertz
	// if c.FeatureFlags.EnableRateLimiting {
	// 	h.Use(middleware.RateLimitMiddleware(c.Security.RateLimitRequests, c.Security.RateLimitDuration))
	// }

	// Add CORS headers with configuration support
	if c.Security.EnableCORS {
		corsOrigins := c.Security.CORSOrigins
		h.Use(func(ctx context.Context, c *app.RequestContext) {
			c.Header("Access-Control-Allow-Origin", corsOrigins)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			c.Next(ctx)
		})
	}

	// Health check endpoint (no auth required)
	h.GET("/ping", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, utils.H{"message": "pong", "timestamp": time.Now().Unix()})
	})

	// Health check with detailed system info
	serverMode := c.Server.Mode
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		health := utils.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
			"mode":      serverMode,
		}
		c.JSON(consts.StatusOK, health)
	})

	// Hello API endpoint (auth required)
	h.GET("/v1/hello", func(ctx context.Context, c *app.RequestContext) {
		response, err := helloService.SayHelloHTTP(ctx, &handlers.HelloRequest{Id: 1})
		if err != nil {
			c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
			return
		}
		c.JSON(consts.StatusOK, response)
	})

	return h
}

// Note: Hertz already includes netpoll optimizations by default
// Additional transport optimizations can be added here if needed
