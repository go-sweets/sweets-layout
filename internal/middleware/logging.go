package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	// LogRequestBody whether to log request body
	LogRequestBody bool
	// LogResponseBody whether to log response body
	LogResponseBody bool
	// SkipPaths that should not be logged
	SkipPaths map[string]bool
	// MaxBodySize maximum body size to log
	MaxBodySize int
}

// NewLoggingConfig creates a new logging configuration
func NewLoggingConfig() *LoggingConfig {
	return &LoggingConfig{
		LogRequestBody:  false, // Disable by default for security
		LogResponseBody: false, // Disable by default for performance
		SkipPaths: map[string]bool{
			"/ping":    true,
			"/health":  true,
			"/metrics": true,
		},
		MaxBodySize: 1024, // 1KB max
	}
}

// HertzLoggingMiddleware returns a Hertz logging middleware
func HertzLoggingMiddleware(config *LoggingConfig) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())

		// Skip logging for configured paths
		if config.SkipPaths[path] {
			c.Next(ctx)
			return
		}

		// Prepare request log data
		requestData := map[string]interface{}{
			"method":     string(c.Request.Method()),
			"path":       path,
			"query":      string(c.Request.URI().QueryString()),
			"user_agent": string(c.Request.Header.UserAgent()),
			"ip":         c.ClientIP(),
			"timestamp":  start.Format(time.RFC3339),
		}

		// Add authentication info if available
		if authType := GetAuthType(c); authType != "" {
			requestData["auth_type"] = authType
			if authType == "api_key" {
				requestData["api_key"] = GetAPIKey(c)
			}
		}

		// Log request body if enabled and reasonable size
		if config.LogRequestBody {
			body := c.Request.Body()
			if len(body) > 0 && len(body) <= config.MaxBodySize {
				requestData["request_body"] = string(body)
			}
		}

		// Log request
		hlog.CtxInfof(ctx, "HTTP Request: %+v", requestData)

		// Process request
		c.Next(ctx)

		// Calculate duration
		duration := time.Since(start)

		// Prepare response log data
		responseData := map[string]interface{}{
			"method":      string(c.Request.Method()),
			"path":        path,
			"status_code": c.Response.StatusCode(),
			"duration_ms": float64(duration.Nanoseconds()) / 1e6,
			"timestamp":   time.Now().Format(time.RFC3339),
		}

		// Log response body if enabled and reasonable size
		if config.LogResponseBody {
			body := c.Response.Body()
			if len(body) > 0 && len(body) <= config.MaxBodySize {
				responseData["response_body"] = string(body)
			}
		}

		// Log response with appropriate level based on status code
		statusCode := c.Response.StatusCode()
		if statusCode >= 500 {
			hlog.CtxErrorf(ctx, "HTTP Response (Error): %+v", responseData)
		} else if statusCode >= 400 {
			hlog.CtxWarnf(ctx, "HTTP Response (Client Error): %+v", responseData)
		} else {
			hlog.CtxInfof(ctx, "HTTP Response (Success): %+v", responseData)
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Response.Header.Set("X-Request-ID", requestID)

		// Add request ID to logger context
		hlog.CtxInfof(ctx, "Request started with ID: %s", requestID)

		c.Next(ctx)
	}
}

// generateRequestID generates a simple request ID
// In production, use a proper UUID library
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
