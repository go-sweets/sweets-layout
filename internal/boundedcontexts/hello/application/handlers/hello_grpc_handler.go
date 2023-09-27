package handlers

import (
	"context"

	"github.com/go-sweets/sweets-layout/api/hello"
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/domain/repositories"
)

type IHelloGrpcHandler interface {
	SayHello(context.Context, *hello.HelloReq) (*hello.HelloResp, error)
}

type HelloGrpcHandler struct {
	repo repositories.IHelloRepository
}

func NewHelloGrpcHandler(helloRepo repositories.IHelloRepository) *HelloGrpcHandler {
	return &HelloGrpcHandler{
		repo: helloRepo,
	}
}

func (handler *HelloGrpcHandler) SayHello(ctx context.Context, in *hello.HelloReq) (*hello.HelloResp, error) {
	resp, err := handler.repo.GetUser(in.Id)
	if err != nil {
		return nil, err
	}
	return &hello.HelloResp{
		Id:      resp.ID,
		Message: resp.NickName,
	}, nil
}
