package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

// KitexTracingConfig holds Kitex tracing configuration
type KitexTracingConfig struct {
	// ServiceName for tracing
	ServiceName string
	// SampleRate for sampling (0.0 to 1.0)
	SampleRate float64
	// SkipMethods that should not be traced
	SkipMethods map[string]bool
}

// NewKitexTracingConfig creates a new Kitex tracing configuration
func NewKitexTracingConfig() *KitexTracingConfig {
	return &KitexTracingConfig{
		ServiceName: "sweets-rpc-service",
		SampleRate:  1.0, // 100% sampling for development
		SkipMethods: map[string]bool{
			"Ping":   true,
			"Health": true,
		},
	}
}

// TraceContext holds tracing information
type TraceContext struct {
	TraceID   string
	SpanID    string
	ParentID  string
	StartTime time.Time
	Tags      map[string]interface{}
}

// KitexTracingInterceptor returns a Kitex tracing interceptor
func KitexTracingInterceptor(config *KitexTracingConfig) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, req, resp interface{}) (err error) {
			// Get RPC info
			ri := rpcinfo.GetRPCInfo(ctx)
			if ri == nil {
				klog.CtxErrorf(ctx, "No RPC info found in context for tracing")
				return next(ctx, req, resp)
			}

			methodName := ri.Invocation().MethodName()
			serviceName := ri.Invocation().ServiceName()

			// Skip tracing for configured methods
			if config.SkipMethods[methodName] {
				return next(ctx, req, resp)
			}

			// Create trace context
			traceCtx := createTraceContext(ctx, serviceName, methodName)

			// Add trace context to request context
			ctx = context.WithValue(ctx, "trace_context", traceCtx)

			// Start span
			startSpan(ctx, traceCtx, config)

			// Process request
			defer func() {
				// Finish span
				finishSpan(ctx, traceCtx, err, config)
			}()

			err = next(ctx, req, resp)
			return err
		}
	}
}

// createTraceContext creates a new trace context
func createTraceContext(ctx context.Context, serviceName, methodName string) *TraceContext {
	// Generate trace ID and span ID (simplified implementation)
	traceID := generateTraceID()
	spanID := generateSpanID()

	// Try to get parent span ID from context
	parentID := ""
	if parentTrace := getTraceFromContext(ctx); parentTrace != nil {
		parentID = parentTrace.SpanID
		traceID = parentTrace.TraceID // Inherit trace ID
	}

	return &TraceContext{
		TraceID:   traceID,
		SpanID:    spanID,
		ParentID:  parentID,
		StartTime: time.Now(),
		Tags: map[string]interface{}{
			"service_name": serviceName,
			"method_name":  methodName,
			"component":    "kitex_rpc",
		},
	}
}

// startSpan starts a new span
func startSpan(ctx context.Context, traceCtx *TraceContext, config *KitexTracingConfig) {
	// Add authentication info to span if available
	if authType := GetAuthTypeFromKitexContext(ctx); authType != "" {
		traceCtx.Tags["auth_type"] = authType
	}

	// Log span start
	klog.CtxInfof(ctx, "Span started - TraceID: %s, SpanID: %s, ParentID: %s, Tags: %+v",
		traceCtx.TraceID, traceCtx.SpanID, traceCtx.ParentID, traceCtx.Tags)
}

// finishSpan finishes a span
func finishSpan(ctx context.Context, traceCtx *TraceContext, err error, config *KitexTracingConfig) {
	duration := time.Since(traceCtx.StartTime)

	// Add final tags
	traceCtx.Tags["duration_ms"] = float64(duration.Nanoseconds()) / 1e6
	traceCtx.Tags["success"] = err == nil

	if err != nil {
		traceCtx.Tags["error"] = err.Error()
		traceCtx.Tags["error_type"] = fmt.Sprintf("%T", err)
	}

	// Log span finish
	if err != nil {
		klog.CtxErrorf(ctx, "Span finished with error - TraceID: %s, SpanID: %s, Duration: %v, Error: %v",
			traceCtx.TraceID, traceCtx.SpanID, duration, err)
	} else {
		klog.CtxInfof(ctx, "Span finished successfully - TraceID: %s, SpanID: %s, Duration: %v",
			traceCtx.TraceID, traceCtx.SpanID, duration)
	}

	// In a real implementation, you would send the span to a tracing backend
	// like Jaeger, Zipkin, or OpenTelemetry collector
}

// generateTraceID generates a trace ID
func generateTraceID() string {
	return fmt.Sprintf("trace_%d_%s", time.Now().UnixNano(), randomString(8))
}

// generateSpanID generates a span ID
func generateSpanID() string {
	return fmt.Sprintf("span_%d_%s", time.Now().UnixNano(), randomString(6))
}

// getTraceFromContext gets trace context from request context
func getTraceFromContext(ctx context.Context) *TraceContext {
	if trace, ok := ctx.Value("trace_context").(*TraceContext); ok {
		return trace
	}
	return nil
}

// GetTraceIDFromKitexContext returns the trace ID from Kitex context
func GetTraceIDFromKitexContext(ctx context.Context) string {
	if trace := getTraceFromContext(ctx); trace != nil {
		return trace.TraceID
	}
	return ""
}

// GetSpanIDFromKitexContext returns the span ID from Kitex context
func GetSpanIDFromKitexContext(ctx context.Context) string {
	if trace := getTraceFromContext(ctx); trace != nil {
		return trace.SpanID
	}
	return ""
}

// PropagateTrace propagates trace context to a new context
func PropagateTrace(ctx context.Context, newCtx context.Context) context.Context {
	if trace := getTraceFromContext(ctx); trace != nil {
		return context.WithValue(newCtx, "trace_context", trace)
	}
	return newCtx
}
