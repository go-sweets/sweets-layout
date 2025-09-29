package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

// KitexLoggingConfig holds Kitex logging configuration
type KitexLoggingConfig struct {
	// LogRequest whether to log request payload
	LogRequest bool
	// LogResponse whether to log response payload
	LogResponse bool
	// SkipMethods that should not be logged
	SkipMethods map[string]bool
	// MaxPayloadSize maximum payload size to log
	MaxPayloadSize int
}

// NewKitexLoggingConfig creates a new Kitex logging configuration
func NewKitexLoggingConfig() *KitexLoggingConfig {
	return &KitexLoggingConfig{
		LogRequest:  true,
		LogResponse: true,
		SkipMethods: map[string]bool{
			"Ping":   true,
			"Health": true,
		},
		MaxPayloadSize: 1024, // 1KB max
	}
}

// KitexLoggingInterceptor returns a Kitex logging interceptor
func KitexLoggingInterceptor(config *KitexLoggingConfig) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			start := time.Now()

			// Get RPC info
			ri := rpcinfo.GetRPCInfo(ctx)
			if ri == nil {
				klog.CtxErrorf(ctx, "No RPC info found in context")
				return next(ctx, req, resp)
			}

			methodName := ri.Invocation().MethodName()
			serviceName := ri.Invocation().ServiceName()

			// Skip logging for configured methods
			if config.SkipMethods[methodName] {
				return next(ctx, req, resp)
			}

			// Prepare request log data
			requestData := map[string]interface{}{
				"service_name": serviceName,
				"method_name":  methodName,
				"timestamp":    start.Format(time.RFC3339),
			}

			// Add remote address if available
			if remoteAddr := ri.From(); remoteAddr != nil {
				requestData["remote_addr"] = remoteAddr.Address().String()
			}

			// Add authentication info if available
			if authType := GetAuthTypeFromKitexContext(ctx); authType != "" {
				requestData["auth_type"] = authType
			}

			// Log request payload if enabled
			if config.LogRequest && req != nil {
				if reqStr := formatPayload(req, config.MaxPayloadSize); reqStr != "" {
					requestData["request"] = reqStr
				}
			}

			// Log request
			klog.CtxInfof(ctx, "RPC Request: %+v", requestData)

			// Process request
			err = next(ctx, req, resp)

			// Calculate duration
			duration := time.Since(start)

			// Prepare response log data
			responseData := map[string]interface{}{
				"service_name": serviceName,
				"method_name":  methodName,
				"duration_ms":  float64(duration.Nanoseconds()) / 1e6,
				"timestamp":    time.Now().Format(time.RFC3339),
				"success":      err == nil,
			}

			// Log error if present
			if err != nil {
				responseData["error"] = err.Error()
			}

			// Log response payload if enabled and no error
			if config.LogResponse && resp != nil && err == nil {
				if respStr := formatPayload(resp, config.MaxPayloadSize); respStr != "" {
					responseData["response"] = respStr
				}
			}

			// Log response with appropriate level
			if err != nil {
				klog.CtxErrorf(ctx, "RPC Response (Error): %+v", responseData)
			} else {
				klog.CtxInfof(ctx, "RPC Response (Success): %+v", responseData)
			}

			return err
		}
	}
}

// formatPayload formats a payload for logging
func formatPayload(payload interface{}, maxSize int) string {
	if payload == nil {
		return ""
	}

	// Try to marshal to JSON first
	if jsonBytes, err := json.Marshal(payload); err == nil {
		if len(jsonBytes) <= maxSize {
			return string(jsonBytes)
		}
		// Truncate if too large
		return string(jsonBytes[:maxSize]) + "...[truncated]"
	}

	// Fallback to string representation
	str := formatValue(payload)
	if len(str) <= maxSize {
		return str
	}
	return str[:maxSize] + "...[truncated]"
}

// formatValue formats a value as string
func formatValue(v interface{}) string {
	if v == nil {
		return "nil"
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr:
		if val.IsNil() {
			return "nil"
		}
		return formatValue(val.Elem().Interface())
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", val.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", val.Float())
	case reflect.Bool:
		return fmt.Sprintf("%t", val.Bool())
	case reflect.Struct:
		return fmt.Sprintf("%+v", v)
	case reflect.Slice, reflect.Array:
		return fmt.Sprintf("%+v", v)
	case reflect.Map:
		return fmt.Sprintf("%+v", v)
	default:
		return fmt.Sprintf("%+v", v)
	}
}
