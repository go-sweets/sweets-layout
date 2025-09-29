package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// JWTSecret for JWT token validation
	JWTSecret string
	// APIKeys for API key validation
	APIKeys map[string]bool
	// SkipPaths that don't require authentication
	SkipPaths map[string]bool
}

// NewAuthConfig creates a new auth configuration
func NewAuthConfig() *AuthConfig {
	return &AuthConfig{
		JWTSecret: "your-secret-key", // In production, load from config
		APIKeys: map[string]bool{
			"api-key-1": true,
			"api-key-2": true,
		},
		SkipPaths: map[string]bool{
			"/ping":    true,
			"/health":  true,
			"/metrics": true,
			"/docs":    true,
			"/swagger": true,
		},
	}
}

// HertzAuthMiddleware returns a Hertz authentication middleware
func HertzAuthMiddleware(config *AuthConfig) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		path := string(c.Request.URI().Path())

		// Skip authentication for configured paths
		if config.SkipPaths[path] {
			c.Next(ctx)
			return
		}

		// Check for API key in header
		apiKey := string(c.Request.Header.Get("X-API-Key"))
		if apiKey != "" {
			if config.APIKeys[apiKey] {
				// Set user context for downstream handlers
				c.Set("auth_type", "api_key")
				c.Set("api_key", apiKey)
				hlog.CtxInfof(ctx, "API key authentication successful for path: %s", path)
				c.Next(ctx)
				return
			}
			hlog.CtxWarnf(ctx, "Invalid API key provided for path: %s", path)
		}

		// Check for Bearer token in Authorization header
		authHeader := string(c.Request.Header.Get("Authorization"))
		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if validateJWTToken(token, config.JWTSecret) {
					// Set user context for downstream handlers
					c.Set("auth_type", "jwt")
					c.Set("jwt_token", token)
					hlog.CtxInfof(ctx, "JWT authentication successful for path: %s", path)
					c.Next(ctx)
					return
				}
				hlog.CtxWarnf(ctx, "Invalid JWT token provided for path: %s", path)
			}
		}

		// Authentication failed
		hlog.CtxWarnf(ctx, "Authentication failed for path: %s", path)
		c.JSON(consts.StatusUnauthorized, utils.H{
			"error":   "Unauthorized",
			"message": "Authentication required. Provide a valid API key or JWT token.",
			"code":    consts.StatusUnauthorized,
		})
		c.Abort()
	}
}

// validateJWTToken validates a JWT token (simplified implementation)
// In production, use a proper JWT library like golang-jwt/jwt
func validateJWTToken(token, secret string) bool {
	// Simplified validation - in production, implement proper JWT validation
	// This is just for demonstration purposes
	if len(token) < 10 {
		return false
	}

	// Simulate JWT validation logic
	// In a real implementation, you would:
	// 1. Parse the JWT token
	// 2. Verify the signature using the secret
	// 3. Check expiration time
	// 4. Validate claims

	// For demo purposes, accept tokens that start with "valid_"
	return strings.HasPrefix(token, "valid_")
}

// GetAuthType returns the authentication type from context
func GetAuthType(c *app.RequestContext) string {
	if authType, exists := c.Get("auth_type"); exists {
		return authType.(string)
	}
	return ""
}

// GetAPIKey returns the API key from context
func GetAPIKey(c *app.RequestContext) string {
	if apiKey, exists := c.Get("api_key"); exists {
		return apiKey.(string)
	}
	return ""
}

// GetJWTToken returns the JWT token from context
func GetJWTToken(c *app.RequestContext) string {
	if token, exists := c.Get("jwt_token"); exists {
		return token.(string)
	}
	return ""
}
