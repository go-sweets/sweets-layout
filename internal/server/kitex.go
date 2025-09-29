package server

import (
	"fmt"
	"net"
	"time"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	app_config "github.com/go-sweets/sweets-layout/internal/config"
	"github.com/go-sweets/sweets-layout/internal/di/providers"
	"github.com/go-sweets/sweets-layout/internal/middleware"
	hello_server "github.com/go-sweets/sweets-layout/api/gen/kitex/api/hello/hello"
)

func NewKitexServer(c *app_config.Config, registrar *providers.ServiceRegistrar) server.Server {
	addr := fmt.Sprintf(":%d", c.Server.RPCPort)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		panic(fmt.Errorf("failed to resolve TCP address: %w", err))
	}

	// Initialize interceptor configurations
	authConfig := middleware.NewKitexAuthConfig()
	loggingConfig := middleware.NewKitexLoggingConfig()
	tracingConfig := middleware.NewKitexTracingConfig()

	// Performance optimized options
	opts := []server.Option{
		server.WithServiceAddr(tcpAddr),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "hello",
		}),
		// Connection and timeout optimizations
		server.WithReadWriteTimeout(c.Server.ReadTimeout),
		server.WithExitWaitTime(time.Second * 5),
		// Note: Kitex handles connection limits through other mechanisms
		// Enable connection pooling
		server.WithMuxTransport(),
		// Add interceptors (middleware chain for Kitex)
		server.WithMiddleware(middleware.KitexTracingInterceptor(tracingConfig)),
		server.WithMiddleware(middleware.KitexLoggingInterceptor(loggingConfig)),
		server.WithMiddleware(middleware.KitexAuthInterceptor(authConfig)),
	}

	// Production optimizations
	if c.Server.Mode == "prod" {
		opts = append(opts,
			// Connection reuse optimization
			server.WithReusePort(true),
		)
	}

	// Development mode optimizations
	if c.Server.Mode == "dev" && c.Server.Debug {
		// Kitex already has appropriate default logging
		// Additional debug options can be added here if needed
	}

	// Feature flag based optimizations
	// TODO: Implement circuit breaker using Kitex-compatible middleware
	// if c.FeatureFlags.EnableCircuitBreaker {
	// 	opts = append(opts, server.WithMiddleware(middleware.KitexCircuitBreakerInterceptor()))
	// }

	// Create Kitex server using generated NewServer function
	// Use the HelloService from the registrar which delegates to handlers
	svr := hello_server.NewServer(registrar.HelloService, opts...)

	return svr
}
