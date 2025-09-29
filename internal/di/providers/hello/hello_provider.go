package hello

import (
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/application/handlers"
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/domain/repositories"
	infra_repos "github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/infrastructure/repositories"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// HelloProviderSet provides Hello domain-specific dependencies
var HelloProviderSet = wire.NewSet(
	// Repository implementations
	NewUserRepository,

	// Application handlers
	handlers.NewHelloGrpcHandler,
)

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return infra_repos.NewUserRepository(db)
}