package data

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
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
	// 這裡暫時使用內存存儲或 Redis
	return nil
}

// GetRoom 獲取房間信息
func (r *gameRepo) GetRoom(ctx context.Context, roomID string) (*game.Room, error) {
	r.logger.Debugf("Getting room: %s", roomID)
	// TODO: 實現實際的數據庫查詢
	// 這裡暫時返回一個默認房間
	return &game.Room{
		ID:      roomID,
		Name:    "Default Room",
		Type:    game.RoomType("normal"),
		Players: make(map[int64]*game.Player),
		Status:  game.RoomStatus("waiting"),
	}, nil
}

// ListRooms 列出房間
func (r *gameRepo) ListRooms(ctx context.Context, roomType game.RoomType) ([]*game.Room, error) {
	r.logger.Debugf("Listing rooms of type: %s", roomType)
	// TODO: 實現實際的數據庫查詢
	// 這裡暫時返回一個空列表
	return []*game.Room{}, nil
}

// DeleteRoom 刪除房間
func (r *gameRepo) DeleteRoom(ctx context.Context, roomID string) error {
	r.logger.Debugf("Deleting room: %s", roomID)
	// TODO: 實現實際的數據庫操作
	return nil
}

// SaveGameStatistics 保存遊戲統計
func (r *gameRepo) SaveGameStatistics(ctx context.Context, playerID int64, stats *game.GameStatistics) error {
	r.logger.Debugf("Saving game statistics for player: %d", playerID)
	// TODO: 實現實際的數據庫操作
	return nil
}

// GetGameStatistics 獲取遊戲統計
func (r *gameRepo) GetGameStatistics(ctx context.Context, playerID int64) (*game.GameStatistics, error) {
	r.logger.Debugf("Getting game statistics for player: %d", playerID)
	// TODO: 實現實際的數據庫查詢
	// 這裡暫時返回默認統計
	return &game.GameStatistics{}, nil
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