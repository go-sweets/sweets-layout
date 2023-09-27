package service

import (
	"context"

	"github.com/go-sweets/sweets-layout/api/hello"
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/application/handlers"
	"github.com/go-sweets/sweets-layout/internal/svc"
)

type HelloService struct {
	hello.UnimplementedHelloServer

	svcCtx       *svc.ServiceContext
	helloHandler *handlers.HelloGrpcHandler
}

func NewHelloServer(ctx *svc.ServiceContext,
	helloHandler *handlers.HelloGrpcHandler,
) *HelloService {
	return &HelloService{
		svcCtx:       ctx,
		helloHandler: helloHandler,
	}
}

func (service *HelloService) SayHello(ctx context.Context, in *hello.HelloReq) (*hello.HelloResp, error) {
	resp, err := service.helloHandler.SayHello(ctx, in)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
