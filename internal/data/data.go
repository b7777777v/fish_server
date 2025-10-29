// internal/data/data.go
package data

import (
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Data .data 包含了所有數據源的客戶端，例如 db 和 redis
type Data struct {
	DB  *gorm.DB
	RDB *redis.Client
}

// NewData .創建一個新的 Data 結構
func NewData(c *conf.Data, logger logger.Logger) (*Data, func(), error) {
	db, err := gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{})
	if err != nil {
		logger.Errorf("failed opening connection to postgres: %v", err)
		return nil, nil, err
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Addr,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})

	cleanup := func() {
		logger.Info("closing the data resources")
		sqlDB, _ := db.DB()
		sqlDB.Close()
		rdb.Close()
	}

	return &Data{DB: db, RDB: rdb}, cleanup, nil
}
