package account

import (
	"context"
)

// TODO: 實現帳號模組的業務邏輯
// 此檔案包含帳號相關的核心業務邏輯，包括：
// - 使用者註冊
// - 使用者登入（密碼、遊客、第三方 OAuth）
// - 使用者資料管理
// - JWT Token 生成與驗證

// AccountUsecase 定義帳號業務邏輯的介面
type AccountUsecase interface {
	// Register 註冊新使用者
	Register(ctx context.Context, username, password string) (*User, error)

	// Login 使用者登入（使用者名稱+密碼）
	Login(ctx context.Context, username, password string) (string, error)

	// GuestLogin 遊客登入
	GuestLogin(ctx context.Context) (string, error)

	// OAuthLogin 第三方 OAuth 登入
	OAuthLogin(ctx context.Context, provider string, code string) (string, error)

	// GetUserByID 根據 ID 獲取使用者資料
	GetUserByID(ctx context.Context, userID int64) (*User, error)

	// UpdateUser 更新使用者資料
	UpdateUser(ctx context.Context, userID int64, nickname, avatarURL string) error
}

// User 代表使用者實體
type User struct {
	ID                int64  `json:"id"`
	Username          string `json:"username"`
	Nickname          string `json:"nickname"`
	AvatarURL         string `json:"avatar_url"`
	IsGuest           bool   `json:"is_guest"`
	ThirdPartyProvider string `json:"third_party_provider,omitempty"`
	ThirdPartyID      string `json:"third_party_id,omitempty"`
}

// accountUsecase 實現 AccountUsecase 介面
type accountUsecase struct {
	// TODO: 注入必要的依賴，如：
	// - AccountRepo: 帳號資料庫操作介面
	// - TokenService: JWT Token 生成服務
	// - OAuthService: OAuth 第三方登入服務
}

// NewAccountUsecase 建立新的 AccountUsecase 實例
func NewAccountUsecase( /* TODO: 添加參數 */ ) AccountUsecase {
	return &accountUsecase{
		// TODO: 初始化依賴
	}
}

// TODO: 實現 AccountUsecase 介面的所有方法
