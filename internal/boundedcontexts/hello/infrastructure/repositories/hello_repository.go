package repositories

import (
	"context"
	"fmt"

	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/domain/entities"
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/domain/repositories"
	"gorm.io/gorm"
)

// UserRepository implements the UserRepository interface
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetUser finds a user by numeric ID (for backward compatibility)
func (repo *UserRepository) GetUser(userId int64) (*entities.User, error) {
	// For backward compatibility, create a user with the old ID
	// In a real implementation, this would fetch from database
	user, err := entities.NewUser(
		fmt.Sprintf("user%d@example.com", userId),
		fmt.Sprintf("user_%d", userId),
		"+8613800138000",
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// FindByID finds a user by ID
func (repo *UserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	// In a real implementation, this would query the database
	// For now, return a mock user
	return entities.NewUser("mock@example.com", "MockUser", "+8613800138000")
}

// FindByEmail finds a user by email
func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	// In a real implementation, this would query the database
	// For now, return nil to indicate not found
	return nil, nil
}

// Save saves a new user
func (repo *UserRepository) Save(ctx context.Context, user *entities.User) error {
	// In a real implementation, this would save to database using repo.db
	return nil
}

// Update updates an existing user
func (repo *UserRepository) Update(ctx context.Context, user *entities.User) error {
	// In a real implementation, this would update the database using repo.db
	return nil
}

// Delete deletes a user
func (repo *UserRepository) Delete(ctx context.Context, id string) error {
	// In a real implementation, this would delete from database using repo.db
	return nil
}