package hello

import (
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/application/handlers"
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/domain/repositories"
	repositories2 "github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/infrastructure/repositories"
	"github.com/go-sweets/sweets-layout/internal/svc"
)

func InjectHelloRepository(svc *svc.ServiceContext) repositories.IHelloRepository {
	return &repositories2.HelloRepository{Svc: svc}
}
func InjectHelloGrpcHandler(helloRepo repositories.IHelloRepository) *handlers.HelloGrpcHandler {
	return handlers.NewHelloGrpcHandler(helloRepo)
}
