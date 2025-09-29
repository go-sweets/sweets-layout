package providers

import (
	"github.com/go-sweets/sweets-layout/internal/di/providers/hello"
	"github.com/google/wire"
)

// ProviderSet aggregates all provider sets for dependency injection
var ProviderSet = wire.NewSet(
	// Database connection providers
	DBProviderSet,

	// Domain-specific providers
	hello.HelloProviderSet,

	// gRPC/HTTP handler providers
	GrpcProviderSet,
)