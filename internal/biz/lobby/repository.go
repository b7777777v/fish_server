package lobby

import (
	"context"
)

// TODO: 實現大廳資料訪問層介面
// 此檔案定義大廳資料訪問的介面，用於與資料庫和快取互動

// LobbyRepo 定義大廳資料訪問的介面（PostgreSQL）
type LobbyRepo interface {
	// GetAnnouncements 獲取公告列表
	GetAnnouncements(ctx context.Context, limit int) ([]*Announcement, error)

	// CreateAnnouncement 建立新公告
	CreateAnnouncement(ctx context.Context, title, content string, priority int) error

	// UpdateAnnouncement 更新公告
	UpdateAnnouncement(ctx context.Context, id int64, title, content string, priority int) error

	// DeleteAnnouncement 刪除公告
	DeleteAnnouncement(ctx context.Context, id int64) error
}

// RoomCache 定義房間快取操作介面（Redis）
type RoomCache interface {
	// GetAllRooms 獲取所有房間資訊（從 Redis 讀取各 Game Server 上報的資料）
	GetAllRooms(ctx context.Context) ([]*RoomInfo, error)

	// UpdateRoomInfo 更新房間資訊（Game Server 使用）
	UpdateRoomInfo(ctx context.Context, gameServerID string, rooms []*RoomInfo) error
}
