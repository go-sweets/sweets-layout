//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/go-sweets/sweets-layout/internal/config"
	"github.com/go-sweets/sweets-layout/internal/db"
	"github.com/go-sweets/sweets-layout/internal/server"
	"github.com/go-sweets/sweets-layout/internal/service"
	"github.com/go-sweets/sweets-layout/internal/svc"

	"github.com/google/wire"
)

// initApp init app application.
func initApp(c *config.Config) (*server.AppServer, error) {
	panic(wire.Build(svc.ProviderSet, service.ProviderSet, server.ProviderSet, server.NewApp))
}

func wireMigrate(c *config.Config) *db.Migrator {
	panic(wire.Build(db.ProviderSet))
}
