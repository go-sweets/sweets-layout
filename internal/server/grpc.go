package server

import (
	"github.com/go-sweets/sweets-layout/api/hello"
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/application/handlers"
	"github.com/go-sweets/sweets-layout/internal/config"
	"github.com/go-sweets/sweets-layout/internal/service"
	"github.com/go-sweets/sweets-layout/internal/svc"
	service2 "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func NewGrpcServer(c *config.Config,
	svc *svc.ServiceContext,
	handler *handlers.HelloGrpcHandler,
) *zrpc.RpcServer {

	srv := service.NewHelloServer(svc, handler)
	s := zrpc.MustNewServer(zrpc.RpcServerConf{
		ListenOn:     c.GrpcConf.Addr,
		Auth:         c.GrpcConf.Auth,
		Timeout:      c.GrpcConf.Timeout,
		CpuThreshold: c.GrpcConf.CpuThreshold,
		Health:       c.GrpcConf.Health,
		ServiceConf: service2.ServiceConf{
			Mode:       c.GrpcConf.Mode,
			MetricsUrl: c.GrpcConf.MetricsUrl,
			Prometheus: c.GrpcConf.Prometheus,
			Telemetry:  c.GrpcConf.Telemetry,
		},
	}, func(g *grpc.Server) {
		// grpc register
		hello.RegisterHelloServer(g, srv)
	})
	s.AddOptions()

	return s
}
