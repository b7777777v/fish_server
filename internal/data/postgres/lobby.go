package postgres

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/lobby"
)

// TODO: 實現大廳資料庫訪問層（PostgreSQL）
// 此檔案實現 LobbyRepo 介面，提供與 PostgreSQL 資料庫的互動功能

// lobbyRepo 實現 lobby.LobbyRepo 介面
type lobbyRepo struct {
	dbManager *DBManager
}

// NewLobbyRepo 建立新的 LobbyRepo 實例
func NewLobbyRepo(dbManager *DBManager) lobby.LobbyRepo {
	return &lobbyRepo{
		dbManager: dbManager,
	}
}

// GetAnnouncements 獲取公告列表
func (r *lobbyRepo) GetAnnouncements(ctx context.Context, limit int) ([]*lobby.Announcement, error) {
	query := `
		SELECT id, title, content, priority, created_at
		FROM announcements
		WHERE is_active = true
		  AND (start_time IS NULL OR start_time <= NOW())
		  AND (end_time IS NULL OR end_time >= NOW())
		ORDER BY priority DESC, created_at DESC
		LIMIT $1
	`

	// 讀操作使用 Read DB
	rows, err := r.dbManager.Read().Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announcements []*lobby.Announcement
	for rows.Next() {
		var ann lobby.Announcement
		err := rows.Scan(&ann.ID, &ann.Title, &ann.Content, &ann.Priority, &ann.CreatedAt)
		if err != nil {
			return nil, err
		}
		announcements = append(announcements, &ann)
	}

	return announcements, rows.Err()
}

// CreateAnnouncement 建立新公告
func (r *lobbyRepo) CreateAnnouncement(ctx context.Context, title, content string, priority int) error {
	query := `
		INSERT INTO announcements (title, content, priority)
		VALUES ($1, $2, $3)
	`

	// 寫操作使用 Write DB
	_, err := r.dbManager.Write().Exec(ctx, query, title, content, priority)
	return err
}

// UpdateAnnouncement 更新公告
func (r *lobbyRepo) UpdateAnnouncement(ctx context.Context, id int64, title, content string, priority int) error {
	query := `
		UPDATE announcements
		SET title = $1, content = $2, priority = $3, updated_at = NOW()
		WHERE id = $4
	`

	// 寫操作使用 Write DB
	_, err := r.dbManager.Write().Exec(ctx, query, title, content, priority, id)
	return err
}

// DeleteAnnouncement 刪除公告
func (r *lobbyRepo) DeleteAnnouncement(ctx context.Context, id int64) error {
	query := `
		DELETE FROM announcements
		WHERE id = $1
	`

	// 寫操作使用 Write DB
	_, err := r.dbManager.Write().Exec(ctx, query, id)
	return err
}
