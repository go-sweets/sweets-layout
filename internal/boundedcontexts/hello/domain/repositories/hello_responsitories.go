package repositories

import "github.com/go-sweets/sweets-layout/internal/boundedcontexts/hello/domain/entities"

type IHelloRepository interface {
	GetUser(userId int64) (*entities.User, error)
}
