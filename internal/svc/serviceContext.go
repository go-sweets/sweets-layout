package svc

import (
	"github.com/go-sweets/sweets-layout/internal/config"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewServiceContext)

type ServiceContext struct {
	Config *config.Config
	DB     *gorm.DB
	Redis  *redis.Client
}

func NewServiceContext(c *config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DbConf.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.RedisConf.Addr,
		Password: c.RedisConf.Pass,
		DB:       int(c.RedisConf.DataBase),
	})
	return &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  rdb,
	}
}
