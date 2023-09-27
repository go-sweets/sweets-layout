package repositories

import (
	"github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/domain/entities"
	"github.com/go-sweets/sweets-layout/internal/svc"
)

type HelloRepository struct {
	Svc *svc.ServiceContext
}

func (repo *HelloRepository) GetUser(userId int64) (*entities.User, error) {
	user := entities.NewUser(userId, "mix-plus")
	return user, nil
}
