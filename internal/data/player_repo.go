// internal/data/player_repo.go
package data

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
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

// FindByUsername 根據用戶名查找玩家
func (r *playerRepo) FindByUsername(ctx context.Context, username string) (*player.Player, error) {
	// 在這裡，我們將編寫從資料庫查詢用戶的邏輯
	// 為了演示，我們先返回一個固定的用戶數據
	// TODO: 實現真實的資料庫查詢
	// 使用新的postgres客戶端
	// var user UserPO
	// result := r.data.db.DB.WithContext(ctx).Where("username = ?", username).First(&user)
	// if result.Error != nil {
	// 	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
	// 		return nil, nil
	// 	}
	// 	r.logger.Errorf("failed to find user by username: %v", result.Error)
	// 	return nil, result.Error
	// }
	// return &player.Player{
	// 	ID:           user.ID,
	// 	Username:     user.Username,
	// 	PasswordHash: user.PasswordHash,
	// }, nil

	// 為了演示，我們先返回一個固定的用戶數據
	if username == "test" {
		return &player.Player{
			ID:           1,
			Username:     "test",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "password" 的 bcrypt hash
		}, nil
	}
	return nil, nil // 在真實情境中，這裡應該回傳 not found 錯誤
}
