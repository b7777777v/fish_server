package data

import (
	"context"
	"database/sql"

	"github.com/b7777777v/fish_server/internal/biz/lobby"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/jackc/pgx/v5"
)

// lobbyPlayerRepo 實現 lobby.PlayerRepo 介面
type lobbyPlayerRepo struct {
	data   *Data
	logger logger.Logger
}

// NewLobbyPlayerRepo 建立新的 LobbyPlayerRepo 實例
func NewLobbyPlayerRepo(data *Data, logger logger.Logger) lobby.PlayerRepo {
	return &lobbyPlayerRepo{
		data:   data,
		logger: logger.With("module", "data/lobby_player_repo"),
	}
}

// GetPlayerInfo 獲取玩家資訊
func (r *lobbyPlayerRepo) GetPlayerInfo(ctx context.Context, userID int64) (*lobby.PlayerStatus, error) {
	query := `
		SELECT id, nickname, avatar_url, level, exp
		FROM users
		WHERE id = $1 AND is_active = true
	`

	var playerStatus lobby.PlayerStatus
	var avatarURL sql.NullString

	// 讀操作使用 Read DB
	err := r.data.DBManager().Read().QueryRow(ctx, query, userID).Scan(
		&playerStatus.UserID,
		&playerStatus.Nickname,
		&avatarURL,
		&playerStatus.Level,
		&playerStatus.EXP,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // 玩家不存在
		}
		r.logger.Errorf("failed to get player info: %v", err)
		return nil, err
	}

	// 處理可為 NULL 的欄位
	playerStatus.AvatarURL = avatarURL.String

	return &playerStatus, nil
}

// lobbyWalletRepo 實現 lobby.WalletRepo 介面
type lobbyWalletRepo struct {
	data   *Data
	logger logger.Logger
}

// NewLobbyWalletRepo 建立新的 LobbyWalletRepo 實例
func NewLobbyWalletRepo(data *Data, logger logger.Logger) lobby.WalletRepo {
	return &lobbyWalletRepo{
		data:   data,
		logger: logger.With("module", "data/lobby_wallet_repo"),
	}
}

// GetBalance 獲取玩家金幣餘額
func (r *lobbyWalletRepo) GetBalance(ctx context.Context, userID int64) (int64, error) {
	// 直接從 users 表獲取金幣數量
	query := `
		SELECT coins
		FROM users
		WHERE id = $1 AND is_active = true
	`

	var balance int64
	// 讀操作使用 Read DB
	err := r.data.DBManager().Read().QueryRow(ctx, query, userID).Scan(&balance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil // 玩家不存在，返回 0
		}
		r.logger.Errorf("failed to get balance: %v", err)
		return 0, err
	}

	return balance, nil
}
