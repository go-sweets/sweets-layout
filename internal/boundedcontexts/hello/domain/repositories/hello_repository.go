package repositories

import (
	"context"
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/domain/entities"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// FindByID finds a user by ID
	FindByID(ctx context.Context, id string) (*entities.User, error)

	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*entities.User, error)

	// Save saves a new user
	Save(ctx context.Context, user *entities.User) error

	// Update updates an existing user
	Update(ctx context.Context, user *entities.User) error

	// Delete deletes a user
	Delete(ctx context.Context, id string) error

	// GetUser finds a user by numeric ID (for backward compatibility)
	GetUser(userId int64) (*entities.User, error)
}