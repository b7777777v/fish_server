package postgres

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/game"
)

// TODO: 實現魚潮資料庫訪問層
// 此檔案實現 FishTideRepo 介面，提供與 PostgreSQL 資料庫的互動功能

// fishTideRepo 實現 game.FishTideRepo 介面
type fishTideRepo struct {
	db *Client
}

// NewFishTideRepo 建立新的 FishTideRepo 實例
func NewFishTideRepo(db *Client) game.FishTideRepo {
	return &fishTideRepo{
		db: db,
	}
}

// GetTideByID 根據 ID 獲取魚潮配置
func (r *fishTideRepo) GetTideByID(ctx context.Context, id int64) (*game.FishTide, error) {
	// TODO: 實現根據 ID 查詢魚潮配置
	// 從 fish_tide_config 表查詢
	return nil, game.ErrFishTideNotImplemented
}

// GetActiveTides 獲取所有啟用的魚潮配置
func (r *fishTideRepo) GetActiveTides(ctx context.Context) ([]*game.FishTide, error) {
	// TODO: 實現獲取所有啟用的魚潮配置
	// 查詢條件：is_active = true
	return nil, game.ErrFishTideNotImplemented
}

// CreateTide 建立新的魚潮配置
func (r *fishTideRepo) CreateTide(ctx context.Context, tide *game.FishTide) error {
	// TODO: 實現建立魚潮配置
	// 插入新記錄到 fish_tide_config 表
	return game.ErrFishTideNotImplemented
}

// UpdateTide 更新魚潮配置
func (r *fishTideRepo) UpdateTide(ctx context.Context, tide *game.FishTide) error {
	// TODO: 實現更新魚潮配置
	// 更新 fish_tide_config 表中的記錄
	return game.ErrFishTideNotImplemented
}

// DeleteTide 刪除魚潮配置
func (r *fishTideRepo) DeleteTide(ctx context.Context, id int64) error {
	// TODO: 實現刪除魚潮配置
	// 從 fish_tide_config 表中刪除記錄
	return game.ErrFishTideNotImplemented
}
