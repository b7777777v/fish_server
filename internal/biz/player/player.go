// internal/biz/player/player.go
package player

import "context"

// Player 是玩家的領域模型
type Player struct {
	ID           uint
	Username     string
	PasswordHash string // 資料庫中應儲存密碼的雜湊值，而非明文
}

// PlayerRepo 定義了玩家數據倉庫的接口
type PlayerRepo interface {
	FindByUsername(ctx context.Context, username string) (*Player, error)
	// 後續可以擴展其他方法，例如 Create, Update 等
}
