package data

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/go-redis/redis/v8"
)

// gameRepo 實現了 biz/game.GameRepo 接口
type gameRepo struct {
	data   *Data
	logger logger.Logger
}

// NewGameRepo 創建一個 gameRepo
func NewGameRepo(data *Data, logger logger.Logger) game.GameRepo {
	return &gameRepo{
		data:   data,
		logger: logger.With("component", "game_repo"),
	}
}

// SaveRoom 保存房間信息
func (r *gameRepo) SaveRoom(ctx context.Context, room *game.Room) error {
	r.logger.Debugf("Saving room: %s", room.ID)
	// TODO: 實現實際的數據庫操作

	// 操作成功後，使快取失效
	cacheKey := fmt.Sprintf("room:%s", room.ID)
	if err := r.data.redis.Del(ctx, cacheKey); err != nil {
		r.logger.Warnf("Failed to delete room cache on save: %v", err)
	}
	return nil
}

// GetRoom 獲取房間信息
func (r *gameRepo) GetRoom(ctx context.Context, roomID string) (*game.Room, error) {
	// 1. 從 Redis 讀取快取
	cacheKey := fmt.Sprintf("room:%s", roomID)
	roomJSON, err := r.data.redis.Get(ctx, cacheKey)
	if err == nil {
		var room game.Room
		if err = json.Unmarshal([]byte(roomJSON), &room); err == nil {
			r.logger.Debugf("Cache hit for room: %s", roomID)
			return &room, nil
		}
		r.logger.Warnf("Failed to unmarshal room from cache: %v", err)
	}
	if err != redis.Nil {
		r.logger.Errorf("Redis error on GetRoom: %v", err)
	}

	// 2. 快取未命中，從資料庫讀取 (TODO)
	r.logger.Debugf("Cache miss for room: %s. Fetching from DB.", roomID)
	// TODO: 實現實際的數據庫查詢
	room := &game.Room{
		ID:      roomID,
		Name:    "Default Room",
		Type:    game.RoomType("normal"),
		Players: make(map[int64]*game.Player),
		Status:  game.RoomStatus("waiting"),
	}

	// 3. 將數據寫入快取
	roomBytes, err := json.Marshal(room)
	if err != nil {
		r.logger.Warnf("Failed to marshal room for cache: %v", err)
	} else {
		if err = r.data.redis.Set(ctx, cacheKey, roomBytes, 5*time.Minute); err != nil {
			r.logger.Warnf("Failed to set room cache: %v", err)
		}
	}

	return room, nil
}

// ListRooms 列出房間
func (r *gameRepo) ListRooms(ctx context.Context, roomType game.RoomType) ([]*game.Room, error) {
	r.logger.Debugf("Listing rooms of type: %s", roomType)
	// TODO: 實現實際的數據庫查詢
	// 列表查詢暫不實現快取
	return []*game.Room{}, nil
}

// DeleteRoom 刪除房間
func (r *gameRepo) DeleteRoom(ctx context.Context, roomID string) error {
	r.logger.Debugf("Deleting room: %s", roomID)
	// TODO: 實現實際的數據庫操作

	// 操作成功後，使快取失效
	cacheKey := fmt.Sprintf("room:%s", roomID)
	if err := r.data.redis.Del(ctx, cacheKey); err != nil {
		r.logger.Warnf("Failed to delete room cache on delete: %v", err)
	}
	return nil
}

// SaveGameStatistics 保存遊戲統計
func (r *gameRepo) SaveGameStatistics(ctx context.Context, playerID int64, stats *game.GameStatistics) error {
	r.logger.Debugf("Saving game statistics for player: %d", playerID)
	// TODO: 實現實際的數據庫操作

	// 操作成功後，使快取失效
	cacheKey := fmt.Sprintf("stats:%d", playerID)
	if err := r.data.redis.Del(ctx, cacheKey); err != nil {
		r.logger.Warnf("Failed to delete stats cache on save: %v", err)
	}
	return nil
}

// GetGameStatistics 獲取遊戲統計
func (r *gameRepo) GetGameStatistics(ctx context.Context, playerID int64) (*game.GameStatistics, error) {
	// 1. 從 Redis 讀取快取
	cacheKey := fmt.Sprintf("stats:%d", playerID)
	statsJSON, err := r.data.redis.Get(ctx, cacheKey)
	if err == nil {
		var stats game.GameStatistics
		if err = json.Unmarshal([]byte(statsJSON), &stats); err == nil {
			r.logger.Debugf("Cache hit for stats: %d", playerID)
			return &stats, nil
		}
		r.logger.Warnf("Failed to unmarshal stats from cache: %v", err)
	}
	if err != redis.Nil {
		r.logger.Errorf("Redis error on GetGameStatistics: %v", err)
	}

	// 2. 快取未命中，從資料庫讀取 (TODO)
	r.logger.Debugf("Cache miss for stats: %d. Fetching from DB.", playerID)
	// TODO: 實現實際的數據庫查詢
	stats := &game.GameStatistics{}

	// 3. 將數據寫入快取
	statsBytes, err := json.Marshal(stats)
	if err != nil {
		r.logger.Warnf("Failed to marshal stats for cache: %v", err)
	} else {
		if err = r.data.redis.Set(ctx, cacheKey, statsBytes, 15*time.Minute); err != nil {
			r.logger.Warnf("Failed to set stats cache: %v", err)
		}
	}

	return stats, nil
}

// SaveGameEvent 保存遊戲事件
func (r *gameRepo) SaveGameEvent(ctx context.Context, event *game.GameEvent) error {
	r.logger.Debugf("Saving game event: %s", event.Type)
	// TODO: 實現實際的數據庫操作
	return nil
}

// GetGameEvents 獲取遊戲事件
func (r *gameRepo) GetGameEvents(ctx context.Context, roomID string, limit int) ([]*game.GameEvent, error) {
	r.logger.Debugf("Getting game events for room: %s, limit: %d", roomID, limit)
	// TODO: 實現實際的數據庫查詢
	// 這裡暫時返回空列表
	return []*game.GameEvent{}, nil
}