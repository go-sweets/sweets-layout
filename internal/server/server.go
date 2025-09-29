package server

import (
	"context"
	"fmt"

	hertz_server "github.com/cloudwego/hertz/pkg/app/server"
	kitex_server "github.com/cloudwego/kitex/server"
	"github.com/go-sweets/sweets-layout/internal/config"
	"github.com/go-sweets/sweets-layout/internal/di/providers"
	"github.com/google/wire"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewHertzServer, NewKitexServer, NewApp)

// AppServer represents the application server with both HTTP and RPC servers
type AppServer struct {
	Config          *config.Config
	HertzServer     *hertz_server.Hertz
	KitexServer     kitex_server.Server
	ServiceRegistrar *providers.ServiceRegistrar
}

// NewApp creates a new AppServer instance
func NewApp(config *config.Config, registrar *providers.ServiceRegistrar, hs *hertz_server.Hertz, ks kitex_server.Server) (*AppServer, error) {
	return &AppServer{
		Config:          config,
		ServiceRegistrar: registrar,
		HertzServer:     hs,
		KitexServer:     ks,
	}, nil
}

// Run starts both the HTTP and RPC servers
func (a *AppServer) Run() error {
	// Start Hertz HTTP server in goroutine
	go func() {
		httpAddr := fmt.Sprintf(":%d", a.Config.Server.HTTPPort)
		fmt.Printf("Starting Hertz HTTP server on %s\n", httpAddr)
		a.HertzServer.Spin()
	}()

	// Start Kitex RPC server (blocking)
	rpcAddr := fmt.Sprintf(":%d", a.Config.Server.RPCPort)
	fmt.Printf("Starting Kitex RPC server on %s\n", rpcAddr)
	return a.KitexServer.Run()
}

// Stop gracefully stops both servers
func (a *AppServer) Stop() error {
	// Stop servers gracefully
	ctx := context.Background()
	if err := a.HertzServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown Hertz server: %w", err)
	}

	if err := a.KitexServer.Stop(); err != nil {
		return fmt.Errorf("failed to stop Kitex server: %w", err)
	}

	return nil
}
