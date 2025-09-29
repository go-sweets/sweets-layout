package handlers

import (
	"context"
	"fmt"

	hello_kitex "github.com/go-sweets/sweets-layout/api/gen/kitex/api/hello"
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/domain/repositories"
	"github.com/go-sweets/sweets-layout/internal/middleware"
)

// HelloRequest represents a generic hello request
type HelloRequest struct {
	Id int64
}

// HelloResponse represents a generic hello response
type HelloResponse struct {
	Id      int64
	Message string
}

// HelloGrpcHandler handles both gRPC and HTTP requests for the Hello service
type HelloGrpcHandler struct {
	repo repositories.UserRepository
}

// NewHelloGrpcHandler creates a new HelloGrpcHandler instance
func NewHelloGrpcHandler(userRepo repositories.UserRepository) *HelloGrpcHandler {
	return &HelloGrpcHandler{
		repo: userRepo,
	}
}

// ProcessHello contains the core business logic, protocol-agnostic
func (handler *HelloGrpcHandler) ProcessHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	// Core business logic: Get user from repository
	user, err := handler.repo.GetUser(req.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Create response with business logic
	return &HelloResponse{
		Id:      req.Id,
		Message: fmt.Sprintf("Hello %s from go-sweets server!", user.NickName),
	}, nil
}

// SayHello implements the Kitex/gRPC Hello service interface
func (handler *HelloGrpcHandler) SayHello(ctx context.Context, req *hello_kitex.HelloReq) (*hello_kitex.HelloResp, error) {
	// Validate the gRPC request
	if err := middleware.ValidateKitexHelloReq(req); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Convert to generic request and process
	genericReq := &HelloRequest{Id: req.Id}
	genericResp, err := handler.ProcessHello(ctx, genericReq)
	if err != nil {
		return nil, err
	}

	// Convert generic response to gRPC response
	resp := &hello_kitex.HelloResp{
		Id:      genericResp.Id,
		Message: genericResp.Message,
	}

	// Validate the response before returning
	if err := middleware.ValidateKitexHelloResp(resp); err != nil {
		return nil, fmt.Errorf("response validation failed: %w", err)
	}

	return resp, nil
}

// SayHelloHTTP handles HTTP requests directly
func (handler *HelloGrpcHandler) SayHelloHTTP(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	// HTTP-specific validation can be done here if needed
	if req.Id <= 0 {
		return nil, fmt.Errorf("invalid ID: %d", req.Id)
	}

	// Use the same core business logic
	return handler.ProcessHello(ctx, req)
}
