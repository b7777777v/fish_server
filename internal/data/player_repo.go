// internal/data/player_repo.go
package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// playerRepo 實現了 biz.PlayerRepo 接口
type playerRepo struct {
	data   *Data
	logger logger.Logger
}

// NewPlayerRepo 創建一個 playerRepo
func NewPlayerRepo(data *Data, logger logger.Logger) player.PlayerRepo {
	return &playerRepo{
		data:   data,
		logger: logger.With("module", "data/player"),
	}
}

// gamePlayerRepo 實現了 biz/game.PlayerRepo 接口
type gamePlayerRepo struct {
	data   *Data
	logger logger.Logger
}

// NewGamePlayerRepo 創建一個用於遊戲業務的 PlayerRepo
func NewGamePlayerRepo(data *Data, logger logger.Logger) game.PlayerRepo {
	return &gamePlayerRepo{
		data:   data,
		logger: logger.With("component", "game_player_repo"),
	}
}

// GetPlayer 獲取玩家信息
func (r *gamePlayerRepo) GetPlayer(ctx context.Context, playerID int64) (*game.Player, error) {
	// 1. 從 Redis 讀取快取
	cacheKey := fmt.Sprintf("player:%d", playerID)
	playerJSON, err := r.data.redis.Get(ctx, cacheKey)
	if err == nil {
		// 快取命中，反序列化並返回
		var player game.Player
		err = json.Unmarshal([]byte(playerJSON), &player)
		if err == nil {
			r.logger.Debugf("Cache hit for player: %d", playerID)
			return &player, nil
		}
		r.logger.Warnf("Failed to unmarshal player from cache: %v", err)
	}
	if err != redis.Nil {
		r.logger.Errorf("Redis error on GetPlayer: %v", err)
	}

	// 2. 快取未命中，從資料庫讀取 (TODO)
	r.logger.Debugf("Cache miss for player: %d. Fetching from DB.", playerID)
	// TODO: 實現實際的數據庫查詢
	// 這裡暫時返回一個默認玩家
	player := &game.Player{
		ID:       playerID,
		Nickname: "Player" + fmt.Sprintf("%d", playerID),
		Balance:  10000, // 默認餘額
		Status:   game.PlayerStatusIdle,
	}

	// 3. 將從資料庫讀取的數據寫入快取
	playerBytes, err := json.Marshal(player)
	if err != nil {
		r.logger.Warnf("Failed to marshal player for cache: %v", err)
	} else {
		err = r.data.redis.Set(ctx, cacheKey, playerBytes, 10*time.Minute) // 10分鐘過期
		if err != nil {
			r.logger.Warnf("Failed to set player cache: %v", err)
		}
	}

	return player, nil
}

// UpdatePlayerBalance 更新玩家餘額
func (r *gamePlayerRepo) UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error {
	r.logger.Debugf("Updating player %d balance to %d", playerID, balance)
	// TODO: 實現實際的數據庫更新

	// 更新成功後，使快取失效
	cacheKey := fmt.Sprintf("player:%d", playerID)
	err := r.data.redis.Del(ctx, cacheKey)
	if err != nil {
		r.logger.Warnf("Failed to delete player cache on balance update: %v", err)
	}

	return nil
}

// UpdatePlayerStatus 更新玩家狀態
func (r *gamePlayerRepo) UpdatePlayerStatus(ctx context.Context, playerID int64, status game.PlayerStatus) error {
	r.logger.Debugf("Updating player %d status to %s", playerID, status)
	// TODO: 實現實際的數據庫更新

	// 更新成功後，使快取失效
	cacheKey := fmt.Sprintf("player:%d", playerID)
	err := r.data.redis.Del(ctx, cacheKey)
	if err != nil {
		r.logger.Warnf("Failed to delete player cache on status update: %v", err)
	}

	return nil
}

// FindByUsername 根據用戶名查找玩家
func (r *playerRepo) FindByUsername(ctx context.Context, username string) (*player.Player, error) {
	// 1. 從 Redis 讀取快取
	cacheKey := fmt.Sprintf("user:%s", username)
	playerJSON, err := r.data.redis.Get(ctx, cacheKey)
	if err == nil {
		// 快取命中
		var p player.Player
		if err = json.Unmarshal([]byte(playerJSON), &p); err == nil {
			r.logger.Debugf("Cache hit for user: %s", username)
			return &p, nil
		}
		r.logger.Warnf("Failed to unmarshal user from cache: %v", err)
	}
	if err != redis.Nil {
		r.logger.Errorf("Redis error on FindByUsername: %v", err)
	}

	// 2. 快取未命中，從資料庫讀取 (TODO)
	r.logger.Debugf("Cache miss for user: %s. Fetching from DB.", username)
	// TODO: 實現真實的資料庫查詢

	// 為了演示，我們先返回一個固定的用戶數據
	var p *player.Player
	if username == "test" {
		p = &player.Player{
			ID:           1,
			Username:     "test",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "password" 的 bcrypt hash
		}
	}

	// 3. 如果找到用戶，寫入快取
	if p != nil {
		playerBytes, err := json.Marshal(p)
		if err != nil {
			r.logger.Warnf("Failed to marshal user for cache: %v", err)
		} else {
			err = r.data.redis.Set(ctx, cacheKey, playerBytes, 1*time.Hour) // 1小時過期
			if err != nil {
				r.logger.Warnf("Failed to set user cache: %v", err)
			}
		}
	}

	return p, nil // 在真實情境中，找不到用戶時 p 為 nil
}
