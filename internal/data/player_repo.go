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
	"github.com/jackc/pgx/v5"
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
		var player game.Player
		if json.Unmarshal([]byte(playerJSON), &player) == nil {
			r.logger.Debugf("Cache hit for player: %d", playerID)
			return &player, nil
		}
	}

	// 2. 快取未命中，從資料庫讀取
	r.logger.Debugf("Cache miss for player: %d. Fetching from DB.", playerID)
	query := `
		SELECT u.id, u.nickname, u.status, w.balance
		FROM users u
		LEFT JOIN wallets w ON u.id = w.user_id AND w.currency = 'CNY'
		WHERE u.id = $1
	`
	var po struct {
		ID       int64
		Nickname string
		Status   int
		Balance  *float64 // Use pointer to handle NULL from LEFT JOIN
	}
	// 讀操作使用 Read DB
	err = r.data.DBManager().Read().QueryRow(ctx, query, playerID).Scan(&po.ID, &po.Nickname, &po.Status, &po.Balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("player with id %d not found", playerID)
		}
		r.logger.Errorf("failed to get player from db: %v", err)
		return nil, err
	}

	balance := int64(0)
	if po.Balance != nil {
		balance = int64(*po.Balance * 100) // Assuming balance is stored as decimal, convert to cents
	}

	player := &game.Player{
		ID:       po.ID,
		Nickname: po.Nickname,
		Balance:  balance,
		Status:   game.PlayerStatusIdle, // Default status
	}
	if po.Status == 0 {
		player.Status = game.PlayerStatusOffline
	}

	// 3. 將數據寫入快取
	playerBytes, err := json.Marshal(player)
	if err == nil {
		r.data.redis.Set(ctx, cacheKey, playerBytes, 10*time.Minute)
	}

	return player, nil
}

// UpdatePlayerBalance 更新玩家餘額
func (r *gamePlayerRepo) UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error {
	r.logger.Debugf("Updating player %d balance to %d", playerID, balance)
	query := `UPDATE wallets SET balance = $1 WHERE user_id = $2 AND currency = 'CNY'`
	// Convert balance from cents to decimal for DB
	balanceDecimal := float64(balance) / 100.0
	// 寫操作使用 Write DB
	_, err := r.data.DBManager().Write().Exec(ctx, query, balanceDecimal, playerID)
	if err != nil {
		r.logger.Errorf("failed to update player balance: %v", err)
		return err
	}

	// 更新成功後，使快取失效
	cacheKey := fmt.Sprintf("player:%d", playerID)
	if err := r.data.redis.Del(ctx, cacheKey); err != nil {
		r.logger.Warnf("Failed to delete player cache on balance update: %v", err)
	}

	return nil
}

// UpdatePlayerStatus 更新玩家狀態
func (r *gamePlayerRepo) UpdatePlayerStatus(ctx context.Context, playerID int64, status game.PlayerStatus) error {
	r.logger.Debugf("Updating player %d status to %s", playerID, status)
	var statusInt int
	switch status {
	case game.PlayerStatusPlaying, game.PlayerStatusIdle:
		statusInt = 1
	case game.PlayerStatusOffline:
		statusInt = 0
	default:
		statusInt = 1 // Default to active
	}

	query := `UPDATE users SET status = $1 WHERE id = $2`
	// 寫操作使用 Write DB
	_, err := r.data.DBManager().Write().Exec(ctx, query, statusInt, playerID)
	if err != nil {
		r.logger.Errorf("failed to update player status: %v", err)
		return err
	}

	// 更新成功後，使快取失效
	cacheKey := fmt.Sprintf("player:%d", playerID)
	if err := r.data.redis.Del(ctx, cacheKey); err != nil {
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

	// 2. 快取未命中，從資料庫讀取
	r.logger.Debugf("Cache miss for user: %s. Fetching from DB.", username)
	query := `SELECT id, username, password_hash, nickname, status FROM users WHERE username = $1`
	var po struct {
		ID           int64
		Username     string
		PasswordHash string
		Nickname     string
		Status       int
	}
	// 讀操作使用 Read DB
	err = r.data.DBManager().Read().QueryRow(ctx, query, username).Scan(&po.ID, &po.Username, &po.PasswordHash, &po.Nickname, &po.Status)

	var p *player.Player
	if err == nil {
		p = &player.Player{
			ID:           uint(po.ID),
			Username:     po.Username,
			PasswordHash: po.PasswordHash,
		}
	} else if err != pgx.ErrNoRows {
		r.logger.Errorf("failed to find user by username: %v", err)
		return nil, err
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

	return p, nil // 找不到用戶時 p 為 nil
}

// Create 創建一個新玩家
func (r *playerRepo) Create(ctx context.Context, p *player.Player) (*player.Player, error) {
	// 在真實應用中，可能需要更複雜的預設值邏輯
	// 這裡我們只插入必要的欄位，並讓資料庫使用預設值
	query := `INSERT INTO users (username, nickname, password_hash, status) VALUES ($1, $2, $3, $4) RETURNING id`

	// 為新用戶設置一個空的密碼雜湊和預設狀態
	// 密碼為空，因此無法透過一般登入流程登入
	emptyPasswordHash := ""
	defaultStatus := 1 // 假設 1 為活躍狀態

	var newID uint
	// 寫操作使用 Write DB
	err := r.data.DBManager().Write().QueryRow(ctx, query, p.Username, p.Username, emptyPasswordHash, defaultStatus).Scan(&newID)
	if err != nil {
		r.logger.Errorf("failed to create player: %v", err)
		return nil, err
	}

	p.ID = newID
	r.logger.Infof("Created new player %s with ID %d", p.Username, p.ID)

	return p, nil
}
