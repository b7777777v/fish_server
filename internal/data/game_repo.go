package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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

// SaveRoom 保存房間信息 (Upsert)
func (r *gameRepo) SaveRoom(ctx context.Context, room *game.Room) error {
	r.logger.Debugf("Saving room: %s", room.ID)

	configBytes, err := json.Marshal(room.Config)
	if err != nil {
		r.logger.Errorf("failed to marshal room config: %v", err)
		return err
	}

	query := `
		INSERT INTO rooms (id, name, type, status, max_players, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			status = EXCLUDED.status,
			config = EXCLUDED.config,
			updated_at = NOW()
	`

	_, err = r.data.db.Exec(ctx, query, room.ID, room.Name, room.Type, room.Status, room.MaxPlayers, configBytes)
	if err != nil {
		r.logger.Errorf("failed to save room: %v", err)
		return err
	}

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
		if json.Unmarshal([]byte(roomJSON), &room) == nil {
			r.logger.Debugf("Cache hit for room: %s", roomID)
			return &room, nil
		}
	}

	// 2. 快取未命中，從資料庫讀取
	r.logger.Debugf("Cache miss for room: %s. Fetching from DB.", roomID)
	query := `SELECT id, name, type, status, max_players, config, created_at, updated_at FROM rooms WHERE id = $1`
	var configBytes []byte
	room := &game.Room{
		Players: make(map[int64]*game.Player),
		Fishes:  make(map[int64]*game.Fish),
		Bullets: make(map[int64]*game.Bullet),
	}

	err = r.data.db.QueryRow(ctx, query, roomID).Scan(
		&room.ID, &room.Name, &room.Type, &room.Status, &room.MaxPlayers, &configBytes, &room.CreatedAt, &room.UpdatedAt,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, fmt.Errorf("room with id %s not found", roomID)
		}
		r.logger.Errorf("failed to get room from db: %v", err)
		return nil, err
	}

	if err := json.Unmarshal(configBytes, &room.Config); err != nil {
		r.logger.Errorf("failed to unmarshal room config: %v", err)
		return nil, err
	}

	// 3. 將數據寫入快取
	roomBytes, err := json.Marshal(room)
	if err == nil {
		r.data.redis.Set(ctx, cacheKey, roomBytes, 5*time.Minute)
	}

	return room, nil
}

// ListRooms 列出房間
func (r *gameRepo) ListRooms(ctx context.Context, roomType game.RoomType) ([]*game.Room, error) {
	r.logger.Debugf("Listing rooms of type: %s", roomType)
	query := `SELECT id, name, type, status, max_players FROM rooms WHERE type = $1 AND status != 'closed'`
	rows, err := r.data.db.Query(ctx, query, roomType)
	if err != nil {
		r.logger.Errorf("failed to list rooms: %v", err)
		return nil, err
	}
	defer rows.Close()

	var rooms []*game.Room
	for rows.Next() {
		room := &game.Room{}
		if err := rows.Scan(&room.ID, &room.Name, &room.Type, &room.Status, &room.MaxPlayers); err != nil {
			r.logger.Errorf("failed to scan room row: %v", err)
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

// DeleteRoom 刪除房間
func (r *gameRepo) DeleteRoom(ctx context.Context, roomID string) error {
	r.logger.Debugf("Deleting room: %s", roomID)
	query := `DELETE FROM rooms WHERE id = $1`
	_, err := r.data.db.Exec(ctx, query, roomID)
	if err != nil {
		r.logger.Errorf("failed to delete room: %v", err)
		return err
	}

	// 操作成功後，使快取失效
	cacheKey := fmt.Sprintf("room:%s", roomID)
	if err := r.data.redis.Del(ctx, cacheKey); err != nil {
		r.logger.Warnf("Failed to delete room cache on delete: %v", err)
	}
	return nil
}

// SaveGameStatistics 保存遊戲統計 (Upsert)
func (r *gameRepo) SaveGameStatistics(ctx context.Context, playerID int64, stats *game.GameStatistics) error {
	r.logger.Debugf("Saving game statistics for player: %d", playerID)

	query := `
		INSERT INTO game_statistics (user_id, total_shots, total_hits, total_rewards, total_costs, fish_killed, play_time_seconds, last_played_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			total_shots = game_statistics.total_shots + EXCLUDED.total_shots,
			total_hits = game_statistics.total_hits + EXCLUDED.total_hits,
			total_rewards = game_statistics.total_rewards + EXCLUDED.total_rewards,
			total_costs = game_statistics.total_costs + EXCLUDED.total_costs,
			fish_killed = game_statistics.fish_killed + EXCLUDED.fish_killed,
			play_time_seconds = game_statistics.play_time_seconds + EXCLUDED.play_time_seconds,
			last_played_at = NOW(),
			updated_at = NOW()
	`

	_, err := r.data.db.Exec(ctx, query, playerID, stats.TotalShots, stats.TotalHits, stats.TotalRewards, stats.TotalCosts, stats.FishKilled, stats.PlayTime)
	if err != nil {
		r.logger.Errorf("failed to save game statistics: %v", err)
		return err
	}

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
		if json.Unmarshal([]byte(statsJSON), &stats) == nil {
			r.logger.Debugf("Cache hit for stats: %d", playerID)
			return &stats, nil
		}
	}

	// 2. 快取未命中，從資料庫讀取
	r.logger.Debugf("Cache miss for stats: %d. Fetching from DB.", playerID)
	query := `SELECT total_shots, total_hits, total_rewards, total_costs, fish_killed, play_time_seconds FROM game_statistics WHERE user_id = $1`
	stats := &game.GameStatistics{}

	var totalRewards, totalCosts float64
	err = r.data.db.QueryRow(ctx, query, playerID).Scan(
		&stats.TotalShots, &stats.TotalHits, &totalRewards, &totalCosts, &stats.FishKilled, &stats.PlayTime,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			// 如果沒有統計數據，返回一個空的統計對象，而不是錯誤
			return &game.GameStatistics{}, nil
		}
		r.logger.Errorf("failed to get game statistics: %v", err)
		return nil, err
	}

	// Convert decimal from db to int64
	stats.TotalRewards = int64(totalRewards * 100)
	stats.TotalCosts = int64(totalCosts * 100)

	// 3. 將數據寫入快取
	statsBytes, err := json.Marshal(stats)
	if err == nil {
		r.data.redis.Set(ctx, cacheKey, statsBytes, 15*time.Minute)
	}

	return stats, nil
}

// SaveGameEvent 保存遊戲事件
func (r *gameRepo) SaveGameEvent(ctx context.Context, event *game.GameEvent) error {
	r.logger.Debugf("Saving game event: %s", event.Type)

	dataBytes, err := json.Marshal(event.Data)
	if err != nil {
		r.logger.Errorf("failed to marshal event data: %v", err)
		return err
	}

	query := `
		INSERT INTO game_events (room_id, user_id, event_type, data, timestamp)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err = r.data.db.Exec(ctx, query, event.RoomID, event.PlayerID, event.Type, dataBytes, event.Timestamp)
	if err != nil {
		r.logger.Errorf("failed to save game event: %v", err)
		return err
	}

	return nil
}

// GetGameEvents 獲取遊戲事件
func (r *gameRepo) GetGameEvents(ctx context.Context, roomID string, limit int) ([]*game.GameEvent, error) {
	r.logger.Debugf("Getting game events for room: %s, limit: %d", roomID, limit)
	query := `
		SELECT id, room_id, user_id, event_type, data, timestamp
		FROM game_events
		WHERE room_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`
	rows, err := r.data.db.Query(ctx, query, roomID, limit)
	if err != nil {
		r.logger.Errorf("failed to get game events: %v", err)
		return nil, err
	}
	defer rows.Close()

	var events []*game.GameEvent
	for rows.Next() {
		event := &game.GameEvent{}
		var dataBytes []byte
		var userID sql.NullInt64 // Handle nullable user_id

		if err := rows.Scan(&event.ID, &event.RoomID, &userID, &event.Type, &dataBytes, &event.Timestamp); err != nil {
			r.logger.Errorf("failed to scan game event row: %v", err)
			return nil, err
		}

		if userID.Valid {
			event.PlayerID = userID.Int64
		}

		if err := json.Unmarshal(dataBytes, &event.Data); err != nil {
			r.logger.Warnf("failed to unmarshal event data: %v", err)
			// Continue with the field as nil or handle as needed
			event.Data = nil
		}

		events = append(events, event)
	}

	return events, nil
}
