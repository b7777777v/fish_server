package lobby

import (
	"context"
	"fmt"
)

// LobbyUsecase implements lobby business logic
// 此檔案包含大廳相關的核心業務邏輯，包括：
// - 遊戲房間列表查詢
// - 玩家狀態顯示
// - 公告與訊息管理

// LobbyUsecase 定義大廳業務邏輯的介面
type LobbyUsecase interface {
	// GetRoomList 獲取遊戲房間列表
	GetRoomList(ctx context.Context) ([]*RoomInfo, error)

	// GetPlayerStatus 獲取玩家狀態
	GetPlayerStatus(ctx context.Context, userID int64) (*PlayerStatus, error)

	// GetAnnouncements 獲取公告列表
	GetAnnouncements(ctx context.Context, limit int) ([]*Announcement, error)

	// CreateAnnouncement 建立新公告（管理員功能）
	CreateAnnouncement(ctx context.Context, title, content string, priority int) error

	// UpdateAnnouncement 更新公告（管理員功能）
	UpdateAnnouncement(ctx context.Context, id int64, title, content string, priority int) error

	// DeleteAnnouncement 刪除公告（管理員功能）
	DeleteAnnouncement(ctx context.Context, id int64) error
}

// RoomInfo 房間資訊
type RoomInfo struct {
	RoomID          string `json:"room_id"`
	RoomName        string `json:"room_name"`
	BetMultiplier   int    `json:"bet_multiplier"`   // 下注倍率
	MinCoins        int64  `json:"min_coins"`        // 最低金幣要求
	CurrentPlayers  int    `json:"current_players"`  // 當前玩家數
	MaxPlayers      int    `json:"max_players"`      // 最大玩家數
	GameServerID    string `json:"game_server_id"`   // Game Server 實例 ID
}

// PlayerStatus 玩家狀態
type PlayerStatus struct {
	UserID    int64  `json:"user_id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	Coins     int64  `json:"coins"`      // 金幣數量
	Level     int    `json:"level"`      // 等級
	EXP       int64  `json:"exp"`        // 經驗值
}

// Announcement 公告
type Announcement struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Priority  int    `json:"priority"`   // 優先級（數字越大越重要）
	CreatedAt string `json:"created_at"`
}

// WalletRepo 定義錢包資料訪問介面
type WalletRepo interface {
	GetBalance(ctx context.Context, userID int64) (int64, error)
}

// PlayerRepo 定義玩家資料訪問介面
type PlayerRepo interface {
	GetPlayerInfo(ctx context.Context, userID int64) (*PlayerStatus, error)
}

// lobbyUsecase 實現 LobbyUsecase 介面
type lobbyUsecase struct {
	lobbyRepo  LobbyRepo
	roomCache  RoomCache
	walletRepo WalletRepo
	playerRepo PlayerRepo
}

// NewLobbyUsecase 建立新的 LobbyUsecase 實例
func NewLobbyUsecase(lobbyRepo LobbyRepo, roomCache RoomCache, walletRepo WalletRepo, playerRepo PlayerRepo) LobbyUsecase {
	return &lobbyUsecase{
		lobbyRepo:  lobbyRepo,
		roomCache:  roomCache,
		walletRepo: walletRepo,
		playerRepo: playerRepo,
	}
}

// GetRoomList 獲取遊戲房間列表
func (uc *lobbyUsecase) GetRoomList(ctx context.Context) ([]*RoomInfo, error) {
	rooms, err := uc.roomCache.GetAllRooms(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms from cache: %w", err)
	}

	return rooms, nil
}

// GetPlayerStatus 獲取玩家狀態
func (uc *lobbyUsecase) GetPlayerStatus(ctx context.Context, userID int64) (*PlayerStatus, error) {
	// 從資料庫獲取玩家資訊
	playerStatus, err := uc.playerRepo.GetPlayerInfo(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player info: %w", err)
	}

	// 獲取玩家金幣餘額
	coins, err := uc.walletRepo.GetBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet balance: %w", err)
	}

	playerStatus.Coins = coins

	return playerStatus, nil
}

// GetAnnouncements 獲取公告列表
func (uc *lobbyUsecase) GetAnnouncements(ctx context.Context, limit int) ([]*Announcement, error) {
	if limit <= 0 {
		limit = 10 // 預設返回 10 條公告
	}

	announcements, err := uc.lobbyRepo.GetAnnouncements(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get announcements: %w", err)
	}

	return announcements, nil
}

// CreateAnnouncement 建立新公告（管理員功能）
func (uc *lobbyUsecase) CreateAnnouncement(ctx context.Context, title, content string, priority int) error {
	err := uc.lobbyRepo.CreateAnnouncement(ctx, title, content, priority)
	if err != nil {
		return fmt.Errorf("failed to create announcement: %w", err)
	}

	return nil
}

// UpdateAnnouncement 更新公告（管理員功能）
func (uc *lobbyUsecase) UpdateAnnouncement(ctx context.Context, id int64, title, content string, priority int) error {
	err := uc.lobbyRepo.UpdateAnnouncement(ctx, id, title, content, priority)
	if err != nil {
		return fmt.Errorf("failed to update announcement: %w", err)
	}

	return nil
}

// DeleteAnnouncement 刪除公告（管理員功能）
func (uc *lobbyUsecase) DeleteAnnouncement(ctx context.Context, id int64) error {
	err := uc.lobbyRepo.DeleteAnnouncement(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete announcement: %w", err)
	}

	return nil
}
