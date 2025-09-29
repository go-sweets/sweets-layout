package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

// KitexAuthConfig holds Kitex authentication configuration
type KitexAuthConfig struct {
	// JWTSecret for JWT token validation
	JWTSecret string
	// APIKeys for API key validation
	APIKeys map[string]bool
	// SkipMethods that don't require authentication
	SkipMethods map[string]bool
}

// NewKitexAuthConfig creates a new Kitex auth configuration
func NewKitexAuthConfig() *KitexAuthConfig {
	return &KitexAuthConfig{
		JWTSecret: "your-secret-key", // In production, load from config
		APIKeys: map[string]bool{
			"api-key-1": true,
			"api-key-2": true,
		},
		SkipMethods: map[string]bool{
			"Ping":   true,
			"Health": true,
		},
	}
}

// KitexAuthInterceptor returns a Kitex authentication interceptor
func KitexAuthInterceptor(config *KitexAuthConfig) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			// Get RPC info
			ri := rpcinfo.GetRPCInfo(ctx)
			if ri == nil {
				return fmt.Errorf("no RPC info found in context")
			}

			methodName := ri.Invocation().MethodName()

			// Skip authentication for configured methods
			if config.SkipMethods[methodName] {
				klog.CtxInfof(ctx, "Skipping auth for method: %s", methodName)
				return next(ctx, req, resp)
			}

			// Check for authentication in metadata
			authenticated := false
			authType := ""

			// Get metadata from context (implementation depends on how metadata is passed)
			if apiKey := getMetadataValue(ctx, "x-api-key"); apiKey != "" {
				if config.APIKeys[apiKey] {
					authenticated = true
					authType = "api_key"
					klog.CtxInfof(ctx, "API key authentication successful for method: %s", methodName)
				} else {
					klog.CtxWarnf(ctx, "Invalid API key provided for method: %s", methodName)
				}
			}

			// Check for Bearer token
			if !authenticated {
				if authHeader := getMetadataValue(ctx, "authorization"); authHeader != "" {
					if strings.HasPrefix(authHeader, "Bearer ") {
						token := strings.TrimPrefix(authHeader, "Bearer ")
						if validateJWTToken(token, config.JWTSecret) {
							authenticated = true
							authType = "jwt"
							klog.CtxInfof(ctx, "JWT authentication successful for method: %s", methodName)
						} else {
							klog.CtxWarnf(ctx, "Invalid JWT token provided for method: %s", methodName)
						}
					}
				}
			}

			if !authenticated {
				klog.CtxWarnf(ctx, "Authentication failed for method: %s", methodName)
				return fmt.Errorf("authentication required: invalid or missing credentials")
			}

			// Store auth info in context for downstream use
			ctx = context.WithValue(ctx, "auth_type", authType)

			return next(ctx, req, resp)
		}
	}
}

// getMetadataValue gets a metadata value from context
// This is a simplified implementation - in practice, you'd use Kitex's metadata utilities
func getMetadataValue(ctx context.Context, key string) string {
	// In a real implementation, you would extract metadata from the Kitex context
	// For now, this is a placeholder that would work with proper metadata handling
	if value, ok := ctx.Value(key).(string); ok {
		return value
	}
	return ""
}

// GetAuthTypeFromKitexContext returns the authentication type from Kitex context
func GetAuthTypeFromKitexContext(ctx context.Context) string {
	if authType, ok := ctx.Value("auth_type").(string); ok {
		return authType
	}
	return ""
}
