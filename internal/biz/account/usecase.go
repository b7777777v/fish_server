package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
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

// TokenService 定義 Token 生成服務介面
type TokenService interface {
	GenerateTokenWithClaims(userID int64, isGuest bool) (string, error)
}

// WalletCreator 定義錢包創建服務介面
type WalletCreator interface {
	CreateWallet(ctx context.Context, userID uint, currency string) error
}

// accountUsecase 實現 AccountUsecase 介面
type accountUsecase struct {
	repo          AccountRepo
	tokenService  TokenService
	oauthService  OAuthService
	walletCreator WalletCreator
}

// NewAccountUsecase 建立新的 AccountUsecase 實例
func NewAccountUsecase(repo AccountRepo, tokenService TokenService, oauthService OAuthService, walletCreator WalletCreator) AccountUsecase {
	return &accountUsecase{
		repo:          repo,
		tokenService:  tokenService,
		oauthService:  oauthService,
		walletCreator: walletCreator,
	}
}

// Register 註冊新使用者
func (uc *accountUsecase) Register(ctx context.Context, username, password string) (*User, error) {
	// 檢查使用者名稱是否已存在
	existingUser, _, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 生成密碼雜湊
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 建立使用者
	user := &User{
		Username: username,
		Nickname: username, // 預設暱稱為使用者名稱
		IsGuest:  false,
	}

	createdUser, err := uc.repo.CreateUser(ctx, user, string(passwordHash))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 自動創建初始錢包（CNY幣種）
	if uc.walletCreator != nil {
		err = uc.walletCreator.CreateWallet(ctx, uint(createdUser.ID), "CNY")
		if err != nil {
			// 錢包創建失敗記錄錯誤，但不影響註冊流程
			// TODO: 可以考慮使用消息隊列異步創建
			fmt.Printf("Warning: failed to create initial wallet for user %d: %v\n", createdUser.ID, err)
		}
	}

	return createdUser, nil
}

// Login 使用者登入（使用者名稱+密碼）
func (uc *accountUsecase) Login(ctx context.Context, username, password string) (string, error) {
	// 根據使用者名稱獲取使用者
	user, passwordHash, err := uc.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return "", errors.New("invalid username or password")
	}

	// 驗證密碼
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	// 生成 JWT token
	token, err := uc.tokenService.GenerateTokenWithClaims(user.ID, false)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// GuestLogin 遊客登入
func (uc *accountUsecase) GuestLogin(ctx context.Context) (string, error) {
	// 生成唯一的遊客 ID（使用時間戳和隨機數）
	guestNickname := fmt.Sprintf("Guest_%d", generateGuestID())

	// 建立遊客使用者
	user := &User{
		Nickname: guestNickname,
		IsGuest:  true,
	}

	createdUser, err := uc.repo.CreateUser(ctx, user, "")
	if err != nil {
		return "", fmt.Errorf("failed to create guest user: %w", err)
	}

	// 生成 JWT token（包含 is_guest: true）
	token, err := uc.tokenService.GenerateTokenWithClaims(createdUser.ID, true)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// OAuthLogin 第三方 OAuth 登入
func (uc *accountUsecase) OAuthLogin(ctx context.Context, provider string, code string) (string, error) {
	// 使用 OAuth 服務獲取使用者資訊
	oauthUserInfo, err := uc.oauthService.GetUserInfo(ctx, provider, code)
	if err != nil {
		return "", fmt.Errorf("failed to get oauth user info: %w", err)
	}

	// 根據第三方平台 ID 查找使用者
	user, err := uc.repo.GetUserByThirdParty(ctx, provider, oauthUserInfo.ThirdPartyID)
	if err != nil {
		return "", fmt.Errorf("failed to get user by third party: %w", err)
	}

	// 如果使用者不存在，建立新使用者
	if user == nil {
		user = &User{
			Nickname:          oauthUserInfo.Nickname,
			AvatarURL:         oauthUserInfo.AvatarURL,
			IsGuest:           false,
			ThirdPartyProvider: provider,
			ThirdPartyID:      oauthUserInfo.ThirdPartyID,
		}

		user, err = uc.repo.CreateUser(ctx, user, "")
		if err != nil {
			return "", fmt.Errorf("failed to create oauth user: %w", err)
		}
	}

	// 生成 JWT token
	token, err := uc.tokenService.GenerateTokenWithClaims(user.ID, false)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// GetUserByID 根據 ID 獲取使用者資料
func (uc *accountUsecase) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// UpdateUser 更新使用者資料
func (uc *accountUsecase) UpdateUser(ctx context.Context, userID int64, nickname, avatarURL string) error {
	// 先獲取使用者，確保使用者存在
	user, err := uc.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 更新使用者資料
	if nickname != "" {
		user.Nickname = nickname
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}

	err = uc.repo.UpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// generateGuestID 生成遊客 ID（使用納秒級時間戳）
func generateGuestID() int64 {
	return time.Now().UnixNano() / 1000000 // 毫秒級時間戳
}

// walletCreatorAdapter 是 WalletCreator 介面的適配器
type walletCreatorAdapter struct {
	createWalletFunc func(ctx context.Context, userID uint, currency string) error
}

func (a *walletCreatorAdapter) CreateWallet(ctx context.Context, userID uint, currency string) error {
	return a.createWalletFunc(ctx, userID, currency)
}

// NewWalletCreatorFromUsecase 創建 WalletCreator 適配器（用於 Wire 依賴注入）
func NewWalletCreatorFromUsecase(uc interface {
	CreateWallet(ctx context.Context, userID uint, currency string) error
}) WalletCreator {
	return &walletCreatorAdapter{
		createWalletFunc: uc.CreateWallet,
	}
}
