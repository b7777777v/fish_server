package postgres

import (
	"context"
	"database/sql"

	"github.com/b7777777v/fish_server/internal/biz/lobby"
)

// TODO: 實現大廳資料庫訪問層（PostgreSQL）
// 此檔案實現 LobbyRepo 介面，提供與 PostgreSQL 資料庫的互動功能

// lobbyRepo 實現 lobby.LobbyRepo 介面
type lobbyRepo struct {
	db *sql.DB
}

// NewLobbyRepo 建立新的 LobbyRepo 實例
func NewLobbyRepo(db *sql.DB) lobby.LobbyRepo {
	return &lobbyRepo{
		db: db,
	}
}

// GetAnnouncements 獲取公告列表
func (r *lobbyRepo) GetAnnouncements(ctx context.Context, limit int) ([]*lobby.Announcement, error) {
	// TODO: 實現獲取公告列表
	// 1. 從 announcements 表查詢最新的公告
	// 2. 按優先級和建立時間排序
	// 3. 限制返回數量
	panic("not implemented")
}

// CreateAnnouncement 建立新公告
func (r *lobbyRepo) CreateAnnouncement(ctx context.Context, title, content string, priority int) error {
	// TODO: 實現建立公告
	// 插入新公告到 announcements 表
	panic("not implemented")
}

// UpdateAnnouncement 更新公告
func (r *lobbyRepo) UpdateAnnouncement(ctx context.Context, id int64, title, content string, priority int) error {
	// TODO: 實現更新公告
	// 更新 announcements 表中指定 ID 的公告
	panic("not implemented")
}

// DeleteAnnouncement 刪除公告
func (r *lobbyRepo) DeleteAnnouncement(ctx context.Context, id int64) error {
	// TODO: 實現刪除公告
	// 從 announcements 表中刪除指定 ID 的公告
	panic("not implemented")
}
