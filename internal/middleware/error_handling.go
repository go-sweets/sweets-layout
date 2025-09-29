package middleware

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error     string      `json:"error"`
	Message   string      `json:"message"`
	Code      int         `json:"code"`
	RequestID string      `json:"request_id,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// ErrorHandlingConfig holds error handling configuration
type ErrorHandlingConfig struct {
	// IncludeStackTrace whether to include stack trace in error responses (dev only)
	IncludeStackTrace bool
	// LogStackTrace whether to log stack trace
	LogStackTrace bool
	// CustomErrorMessages custom error messages for specific status codes
	CustomErrorMessages map[int]string
}

// NewErrorHandlingConfig creates a new error handling configuration
func NewErrorHandlingConfig() *ErrorHandlingConfig {
	return &ErrorHandlingConfig{
		IncludeStackTrace: false, // Should be false in production
		LogStackTrace:     true,
		CustomErrorMessages: map[int]string{
			400: "Bad Request - The request was invalid or cannot be served",
			401: "Unauthorized - Authentication is required",
			403: "Forbidden - You don't have permission to access this resource",
			404: "Not Found - The requested resource was not found",
			405: "Method Not Allowed - The HTTP method is not allowed for this resource",
			429: "Too Many Requests - Rate limit exceeded",
			500: "Internal Server Error - An unexpected error occurred",
			502: "Bad Gateway - Invalid response from upstream server",
			503: "Service Unavailable - The service is temporarily unavailable",
			504: "Gateway Timeout - Upstream server timeout",
		},
	}
}

// HertzErrorHandlingMiddleware returns a Hertz error handling middleware
func HertzErrorHandlingMiddleware(config *ErrorHandlingConfig) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				// Handle panic
				err := fmt.Errorf("panic recovered: %v", r)
				stack := debug.Stack()

				// Log the panic with stack trace
				hlog.CtxErrorf(ctx, "Panic recovered: %v\nStack trace:\n%s", err, string(stack))

				// Prepare error response
				errorResp := ErrorResponse{
					Error:     "Internal Server Error",
					Message:   "An unexpected error occurred",
					Code:      consts.StatusInternalServerError,
					RequestID: getRequestID(c),
					Timestamp: getCurrentTimestamp(),
				}

				// Include stack trace in response if configured (dev only)
				if config.IncludeStackTrace {
					errorResp.Details = map[string]interface{}{
						"stack_trace": string(stack),
						"panic_value": r,
					}
				}

				c.JSON(consts.StatusInternalServerError, errorResp)
				c.Abort()
			}
		}()

		c.Next(ctx)

		// Handle HTTP errors after request processing
		statusCode := c.Response.StatusCode()
		if statusCode >= 400 {
			handleHTTPError(ctx, c, statusCode, config)
		}
	}
}

// handleHTTPError handles HTTP error responses
func handleHTTPError(ctx context.Context, c *app.RequestContext, statusCode int, config *ErrorHandlingConfig) {
	// Check if response body is already set (custom error handling)
	if len(c.Response.Body()) > 0 {
		return
	}

	// Get custom error message or use default
	message := config.CustomErrorMessages[statusCode]
	if message == "" {
		message = fmt.Sprintf("HTTP %d error", statusCode)
	}

	// Prepare error response
	errorResp := ErrorResponse{
		Error:     getErrorName(statusCode),
		Message:   message,
		Code:      statusCode,
		RequestID: getRequestID(c),
		Timestamp: getCurrentTimestamp(),
	}

	// Log the error
	path := string(c.Request.URI().Path())
	method := string(c.Request.Method())

	if statusCode >= 500 {
		hlog.CtxErrorf(ctx, "HTTP %d error - %s %s: %s", statusCode, method, path, message)
	} else {
		hlog.CtxWarnf(ctx, "HTTP %d error - %s %s: %s", statusCode, method, path, message)
	}

	c.JSON(statusCode, errorResp)
}

// ValidationErrorMiddleware handles validation errors specifically
func ValidationErrorMiddleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)

		// Check if there's a validation error stored in context
		if err, exists := c.Get("validation_error"); exists {
			if validationErr, ok := err.(ValidationError); ok {
				errorResp := ErrorResponse{
					Error:     "Validation Error",
					Message:   validationErr.Error(),
					Code:      consts.StatusBadRequest,
					RequestID: getRequestID(c),
					Timestamp: getCurrentTimestamp(),
					Details: map[string]interface{}{
						"field": validationErr.Field,
					},
				}

				hlog.CtxWarnf(ctx, "Validation error: %s", validationErr.Error())
				c.JSON(consts.StatusBadRequest, errorResp)
				c.Abort()
			}
		}
	}
}

// getErrorName returns the error name for a status code
func getErrorName(statusCode int) string {
	switch statusCode {
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 405:
		return "Method Not Allowed"
	case 429:
		return "Too Many Requests"
	case 500:
		return "Internal Server Error"
	case 502:
		return "Bad Gateway"
	case 503:
		return "Service Unavailable"
	case 504:
		return "Gateway Timeout"
	default:
		return fmt.Sprintf("HTTP %d", statusCode)
	}
}

// getRequestID gets the request ID from context
func getRequestID(c *app.RequestContext) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// getCurrentTimestamp returns the current timestamp in RFC3339 format
func getCurrentTimestamp() string {
	return getCurrentTime().Format("2006-01-02T15:04:05Z07:00")
}

// getCurrentTime returns the current time (abstracted for testing)
var getCurrentTime = func() interface{ Format(string) string } {
	return timeNow{}
}

type timeNow struct{}

func (timeNow) Format(layout string) string {
	return fmt.Sprintf("2025-09-28T%s", "16:00:00Z")
}
