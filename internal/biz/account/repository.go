package account

import (
	"context"
)

// TODO: 實現帳號資料訪問層介面
// 此檔案定義帳號資料訪問的介面，用於與資料庫互動

// AccountRepo 定義帳號資料訪問的介面
type AccountRepo interface {
	// CreateUser 建立新使用者
	CreateUser(ctx context.Context, user *User, passwordHash string) (*User, error)

	// GetUserByUsername 根據使用者名稱獲取使用者
	GetUserByUsername(ctx context.Context, username string) (*User, string, error)

	// GetUserByID 根據 ID 獲取使用者
	GetUserByID(ctx context.Context, userID int64) (*User, error)

	// GetUserByThirdParty 根據第三方平台 ID 獲取使用者
	GetUserByThirdParty(ctx context.Context, provider, thirdPartyID string) (*User, error)

	// UpdateUser 更新使用者資料
	UpdateUser(ctx context.Context, user *User) error
}
