// internal/biz/player/usecase.go
package player

import (
	"context"
	"errors"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/pkg/token"

	"golang.org/x/crypto/bcrypt"
)

// ErrUserNotFoundOrPasswordIncorrect 表示用戶不存在或密碼錯誤
var ErrUserNotFoundOrPasswordIncorrect = errors.New("user not found or password incorrect")

// PlayerUsecase 是玩家相關的業務邏輯
type PlayerUsecase struct {
	repo        PlayerRepo
	tokenHelper *token.TokenHelper
	logger      logger.Logger
}

// NewPlayerUsecase 創建一個 PlayerUsecase
func NewPlayerUsecase(repo PlayerRepo, tokenHelper *token.TokenHelper, logger logger.Logger) *PlayerUsecase {
	return &PlayerUsecase{
		repo:        repo,
		tokenHelper: tokenHelper,
		logger:      logger.With("module", "biz/player"),
	}
}

// Login 處理玩家登入邏輯
func (uc *PlayerUsecase) Login(ctx context.Context, username, password string) (string, error) {
	// 1. 透過 repo 查找使用者
	player, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		// 如果是資料庫查詢出錯
		uc.logger.Errorf("failed to find player by username %s: %v", username, err)
		return "", err // 在真實應用中，可能需要轉換成更友好的錯誤類型
	}
	if player == nil {
		// 如果找不到使用者
		uc.logger.Warnf("login attempt failed: user %s not found", username)
		return "", ErrUserNotFoundOrPasswordIncorrect
	}

	// 2. 比對密碼
	err = bcrypt.CompareHashAndPassword([]byte(player.PasswordHash), []byte(password))
	if err != nil {
		// 密碼不匹配
		uc.logger.Warnf("login attempt failed: incorrect password for user %s", username)
		return "", ErrUserNotFoundOrPasswordIncorrect
	}

	// 3. 密碼正確，生成 JWT
	token, err := uc.tokenHelper.GenerateToken(player.ID)
	if err != nil {
		uc.logger.Errorf("failed to generate token for user %s: %v", username, err)
		return "", err
	}

	uc.logger.Infof("player %s (ID: %d) logged in successfully", player.Username, player.ID)
	return token, nil
}

// GetOrCreateByUsername 根據用戶名查找玩家，如果不存在則創建一個
func (uc *PlayerUsecase) GetOrCreateByUsername(ctx context.Context, username string) (*Player, error) {
	// 1. 嘗試查找用戶
	player, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		// 如果是資料庫本身錯誤，直接返回
		uc.logger.Errorf("failed to find player by username %s: %v", username, err)
		return nil, err
	}

	// 2. 如果用戶已存在，直接返回
	if player != nil {
		uc.logger.Debugf("Found existing player %s with ID %d", username, player.ID)
		return player, nil
	}

	// 3. 如果用戶不存在，創建新用戶
	uc.logger.Infof("Player %s not found, creating a new one.", username)
	newPlayer := &Player{
		Username: username,
		// PasswordHash is left empty as this user is created on-the-fly
	}

	createdPlayer, err := uc.repo.Create(ctx, newPlayer)
	if err != nil {
		uc.logger.Errorf("failed to create new player %s: %v", username, err)
		return nil, err
	}

	return createdPlayer, nil
}
