package service

import (
	"context"

	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/application/handlers"
	hello_kitex "github.com/go-sweets/sweets-layout/api/gen/kitex/api/hello"
)

// HelloService acts as a service layer that delegates to handlers
type HelloService struct {
	helloHandler *handlers.HelloGrpcHandler
}

// NewHelloService creates a new HelloService that delegates to handlers
func NewHelloService(helloHandler *handlers.HelloGrpcHandler) *HelloService {
	return &HelloService{
		helloHandler: helloHandler,
	}
}

// SayHello delegates to the gRPC handler for Kitex service
func (service *HelloService) SayHello(ctx context.Context, req *hello_kitex.HelloReq) (*hello_kitex.HelloResp, error) {
	// Directly delegate to the handler's gRPC method
	return service.helloHandler.SayHello(ctx, req)
}

// SayHelloHTTP delegates to the handler's HTTP method
func (service *HelloService) SayHelloHTTP(ctx context.Context, req *handlers.HelloRequest) (*handlers.HelloResponse, error) {
	// Directly delegate to the handler's HTTP method
	return service.helloHandler.SayHelloHTTP(ctx, req)
}
