package providers

import (
	"github.com/go-sweets/sweets-layout/internal/service"
	"github.com/google/wire"
)

// GrpcProviderSet provides gRPC/HTTP service handlers
var GrpcProviderSet = wire.NewSet(
	// Service layer that delegates to handlers
	service.NewHelloService,

	// Service Registrar for organizing handlers
	NewServiceRegistrar,
)

// ServiceRegistrar holds all service handlers for registration
type ServiceRegistrar struct {
	HelloService *service.HelloService
}

// NewServiceRegistrar creates a new service registrar
func NewServiceRegistrar(helloService *service.HelloService) *ServiceRegistrar {
	return &ServiceRegistrar{
		HelloService: helloService,
	}
}