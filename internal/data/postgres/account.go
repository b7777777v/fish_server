package postgres

import (
	"context"
	"database/sql"

	"github.com/b7777777v/fish_server/internal/biz/account"
)

// TODO: 實現帳號資料庫訪問層
// 此檔案實現 AccountRepo 介面，提供與 PostgreSQL 資料庫的互動功能

// accountRepo 實現 account.AccountRepo 介面
type accountRepo struct {
	db *sql.DB
}

// NewAccountRepo 建立新的 AccountRepo 實例
func NewAccountRepo(db *sql.DB) account.AccountRepo {
	return &accountRepo{
		db: db,
	}
}

// CreateUser 建立新使用者
func (r *accountRepo) CreateUser(ctx context.Context, user *account.User, passwordHash string) (*account.User, error) {
	// TODO: 實現新使用者建立邏輯
	// 1. 插入資料到 users 表
	// 2. 返回建立的使用者資料（包含自動生成的 ID）
	panic("not implemented")
}

// GetUserByUsername 根據使用者名稱獲取使用者
func (r *accountRepo) GetUserByUsername(ctx context.Context, username string) (*account.User, string, error) {
	// TODO: 實現根據使用者名稱查詢使用者
	// 返回：使用者資料、密碼雜湊、錯誤
	panic("not implemented")
}

// GetUserByID 根據 ID 獲取使用者
func (r *accountRepo) GetUserByID(ctx context.Context, userID int64) (*account.User, error) {
	// TODO: 實現根據 ID 查詢使用者
	panic("not implemented")
}

// GetUserByThirdParty 根據第三方平台 ID 獲取使用者
func (r *accountRepo) GetUserByThirdParty(ctx context.Context, provider, thirdPartyID string) (*account.User, error) {
	// TODO: 實現根據第三方平台資訊查詢使用者
	// 查詢條件：third_party_provider = ? AND third_party_id = ?
	panic("not implemented")
}

// UpdateUser 更新使用者資料
func (r *accountRepo) UpdateUser(ctx context.Context, user *account.User) error {
	// TODO: 實現使用者資料更新
	// 可更新欄位：nickname, avatar_url
	panic("not implemented")
}
