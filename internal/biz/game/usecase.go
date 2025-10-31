package game

import (
	"context"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// ========================================
// GameUsecase 遊戲業務邏輯用例
// ========================================

// GameRepo 遊戲數據倉庫接口
type GameRepo interface {
	// 房間相關
	SaveRoom(ctx context.Context, room *Room) error
	GetRoom(ctx context.Context, roomID string) (*Room, error)
	ListRooms(ctx context.Context, roomType RoomType) ([]*Room, error)
	DeleteRoom(ctx context.Context, roomID string) error
	
	// 遊戲統計
	SaveGameStatistics(ctx context.Context, playerID int64, stats *GameStatistics) error
	GetGameStatistics(ctx context.Context, playerID int64) (*GameStatistics, error)
	
	// 遊戲事件
	SaveGameEvent(ctx context.Context, event *GameEvent) error
	GetGameEvents(ctx context.Context, roomID string, limit int) ([]*GameEvent, error)
}

// PlayerRepo 玩家數據倉庫接口
type PlayerRepo interface {
	GetPlayer(ctx context.Context, playerID int64) (*Player, error)
	UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error
	UpdatePlayerStatus(ctx context.Context, playerID int64, status PlayerStatus) error
}

// GameUsecase 遊戲用例
type GameUsecase struct {
	gameRepo    GameRepo
	playerRepo  PlayerRepo
	roomManager *RoomManager
	spawner     *FishSpawner
	mathModel   *MathModel
	logger      logger.Logger
}

// NewGameUsecase 創建遊戲用例
func NewGameUsecase(
	gameRepo GameRepo,
	playerRepo PlayerRepo,
	roomManager *RoomManager,
	spawner *FishSpawner,
	mathModel *MathModel,
	logger logger.Logger,
) *GameUsecase {
	return &GameUsecase{
		gameRepo:    gameRepo,
		playerRepo:  playerRepo,
		roomManager: roomManager,
		spawner:     spawner,
		mathModel:   mathModel,
		logger:      logger.With("component", "game_usecase"),
	}
}

// ========================================
// 房間管理相關用例
// ========================================

// CreateRoom 創建遊戲房間
func (gu *GameUsecase) CreateRoom(ctx context.Context, roomType RoomType, maxPlayers int32) (*Room, error) {
	room, err := gu.roomManager.CreateRoom(roomType, maxPlayers)
	if err != nil {
		gu.logger.Errorf("Failed to create room: %v", err)
		return nil, err
	}
	
	// 初始化房間魚類
	initialFishes := gu.spawner.BatchSpawnFish(5, room.Config)
	for _, fish := range initialFishes {
		room.Fishes[fish.ID] = fish
	}
	
	// 保存房間到數據庫
	if err := gu.gameRepo.SaveRoom(ctx, room); err != nil {
		gu.logger.Errorf("Failed to save room to database: %v", err)
		return nil, err
	}
	
	// 記錄事件
	event := &GameEvent{
		ID:        time.Now().UnixNano(),
		Type:      EventFishSpawn,
		RoomID:    room.ID,
		Data:      map[string]interface{}{"initial_fish_count": len(initialFishes)},
		Timestamp: time.Now(),
	}
	gu.gameRepo.SaveGameEvent(ctx, event)
	
	gu.logger.Infof("Created room %s with %d initial fishes", room.ID, len(initialFishes))
	return room, nil
}

// JoinRoom 玩家加入房間
func (gu *GameUsecase) JoinRoom(ctx context.Context, roomID string, playerID int64) error {
	// 獲取玩家信息
	player, err := gu.playerRepo.GetPlayer(ctx, playerID)
	if err != nil {
		gu.logger.Errorf("Failed to get player %d: %v", playerID, err)
		return err
	}
	
	// 檢查玩家餘額
	if player.Balance < 100 { // 最小餘額要求
		return fmt.Errorf("insufficient balance to join room")
	}
	
	// 加入房間
	if err := gu.roomManager.JoinRoom(roomID, player); err != nil {
		gu.logger.Errorf("Failed to join room %s: %v", roomID, err)
		return err
	}
	
	// 更新玩家狀態
	if err := gu.playerRepo.UpdatePlayerStatus(ctx, playerID, PlayerStatusPlaying); err != nil {
		gu.logger.Errorf("Failed to update player status: %v", err)
		return err
	}
	
	// 記錄事件
	event := &GameEvent{
		ID:        time.Now().UnixNano(),
		Type:      EventPlayerJoin,
		RoomID:    roomID,
		PlayerID:  playerID,
		Data:      map[string]interface{}{"player_nickname": player.Nickname},
		Timestamp: time.Now(),
	}
	gu.gameRepo.SaveGameEvent(ctx, event)
	
	gu.logger.Infof("Player %d joined room %s", playerID, roomID)
	return nil
}

// LeaveRoom 玩家離開房間
func (gu *GameUsecase) LeaveRoom(ctx context.Context, roomID string, playerID int64) error {
	if err := gu.roomManager.LeaveRoom(roomID, playerID); err != nil {
		gu.logger.Errorf("Failed to leave room %s: %v", roomID, err)
		return err
	}
	
	// 更新玩家狀態
	if err := gu.playerRepo.UpdatePlayerStatus(ctx, playerID, PlayerStatusIdle); err != nil {
		gu.logger.Errorf("Failed to update player status: %v", err)
	}
	
	// 記錄事件
	event := &GameEvent{
		ID:        time.Now().UnixNano(),
		Type:      EventPlayerLeave,
		RoomID:    roomID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
	}
	gu.gameRepo.SaveGameEvent(ctx, event)
	
	gu.logger.Infof("Player %d left room %s", playerID, roomID)
	return nil
}

// GetRoomList 獲取房間列表
func (gu *GameUsecase) GetRoomList(ctx context.Context, roomType RoomType) ([]*Room, error) {
	// 先從內存獲取
	rooms := gu.roomManager.GetRoomList()
	
	// 過濾房間類型
	if roomType != "" {
		filteredRooms := make([]*Room, 0)
		for _, room := range rooms {
			if room.Type == roomType {
				filteredRooms = append(filteredRooms, room)
			}
		}
		return filteredRooms, nil
	}
	
	return rooms, nil
}

// ========================================
// 遊戲玩法相關用例
// ========================================

// FireBullet 玩家開火
func (gu *GameUsecase) FireBullet(ctx context.Context, roomID string, playerID int64, direction float64, power int32) (*Bullet, error) {
	// 檢查參數
	if power < 1 || power > 100 {
		return nil, fmt.Errorf("invalid bullet power: %d", power)
	}
	
	// 發射子彈
	bullet, err := gu.roomManager.FireBullet(roomID, playerID, direction, power)
	if err != nil {
		gu.logger.Errorf("Failed to fire bullet: %v", err)
		return nil, err
	}
	
	// 更新玩家餘額到數據庫
	room, _ := gu.roomManager.GetRoom(roomID)
	if room != nil {
		if player, exists := room.Players[playerID]; exists {
			gu.playerRepo.UpdatePlayerBalance(ctx, playerID, player.Balance)
		}
	}
	
	// 記錄事件
	event := &GameEvent{
		ID:        time.Now().UnixNano(),
		Type:      EventBulletFire,
		RoomID:    roomID,
		PlayerID:  playerID,
		Data: map[string]interface{}{
			"bullet_id": bullet.ID,
			"direction": direction,
			"power":     power,
			"cost":      bullet.Cost,
		},
		Timestamp: time.Now(),
	}
	gu.gameRepo.SaveGameEvent(ctx, event)
	
	gu.logger.Debugf("Player %d fired bullet in room %s, power: %d, cost: %d", 
		playerID, roomID, power, bullet.Cost)
	
	return bullet, nil
}

// HitFish 處理子彈命中魚
func (gu *GameUsecase) HitFish(ctx context.Context, roomID string, bulletID int64, fishID int64) (*HitResult, error) {
	// 處理命中
	hitResult, err := gu.roomManager.ProcessBulletHit(roomID, bulletID, fishID)
	if err != nil {
		gu.logger.Errorf("Failed to process bullet hit: %v", err)
		return nil, err
	}
	
	room, _ := gu.roomManager.GetRoom(roomID)
	if room == nil {
		return hitResult, nil
	}
	
	// 如果命中成功，更新相關數據
	if hitResult.Success {
		// 查找子彈所屬玩家
		var playerID int64
		for _, bullet := range room.Bullets {
			if bullet.ID == bulletID {
				playerID = bullet.PlayerID
				break
			}
		}
		
		if playerID > 0 {
			// 更新玩家餘額到數據庫
			if player, exists := room.Players[playerID]; exists && hitResult.Reward > 0 {
				gu.playerRepo.UpdatePlayerBalance(ctx, playerID, player.Balance)
			}
			
			// 記錄命中事件
			event := &GameEvent{
				ID:       time.Now().UnixNano(),
				Type:     EventBulletHit,
				RoomID:   roomID,
				PlayerID: playerID,
				Data: map[string]interface{}{
					"bullet_id":    bulletID,
					"fish_id":      fishID,
					"damage":       hitResult.Damage,
					"reward":       hitResult.Reward,
					"is_critical":  hitResult.IsCritical,
					"multiplier":   hitResult.Multiplier,
				},
				Timestamp: time.Now(),
			}
			gu.gameRepo.SaveGameEvent(ctx, event)
			
			// 如果魚死亡，記錄魚死亡事件
			if hitResult.Reward > 0 {
				fishEvent := &GameEvent{
					ID:       time.Now().UnixNano() + 1,
					Type:     EventFishDie,
					RoomID:   roomID,
					PlayerID: playerID,
					Data: map[string]interface{}{
						"fish_id": fishID,
						"reward":  hitResult.Reward,
					},
					Timestamp: time.Now(),
				}
				gu.gameRepo.SaveGameEvent(ctx, fishEvent)
			}
		}
	}
	
	return hitResult, nil
}

// ========================================
// 遊戲信息查詢用例
// ========================================

// GetRoomState 獲取房間狀態
func (gu *GameUsecase) GetRoomState(ctx context.Context, roomID string) (*Room, error) {
	return gu.roomManager.GetRoom(roomID)
}

// GetFishTypes 獲取魚類型列表
func (gu *GameUsecase) GetFishTypes(ctx context.Context) []FishType {
	return gu.spawner.GetFishTypes()
}

// GetPlayerStatistics 獲取玩家遊戲統計
func (gu *GameUsecase) GetPlayerStatistics(ctx context.Context, playerID int64) (*GameStatistics, error) {
	stats, err := gu.gameRepo.GetGameStatistics(ctx, playerID)
	if err != nil {
		// 如果沒有統計數據，返回空統計
		return &GameStatistics{}, nil
	}
	return stats, nil
}

// GetGameEvents 獲取遊戲事件
func (gu *GameUsecase) GetGameEvents(ctx context.Context, roomID string, limit int) ([]*GameEvent, error) {
	if limit <= 0 {
		limit = 50 // 默認限制
	}
	return gu.gameRepo.GetGameEvents(ctx, roomID, limit)
}

// ========================================
// 管理員功能用例
// ========================================

// SpawnSpecialFish 生成特殊魚類（管理員功能）
func (gu *GameUsecase) SpawnSpecialFish(ctx context.Context, roomID string, fishTypeID int32) (*Fish, error) {
	room, err := gu.roomManager.GetRoom(roomID)
	if err != nil {
		return nil, err
	}
	
	fish := gu.spawner.SpawnSpecificFish(fishTypeID, room.Config)
	if fish == nil {
		return nil, fmt.Errorf("failed to spawn fish type %d", fishTypeID)
	}
	
	// 添加到房間
	room.Fishes[fish.ID] = fish
	
	// 記錄事件
	event := &GameEvent{
		ID:       time.Now().UnixNano(),
		Type:     EventFishSpawn,
		RoomID:   roomID,
		Data: map[string]interface{}{
			"fish_id":      fish.ID,
			"fish_type_id": fishTypeID,
			"special":      true,
		},
		Timestamp: time.Now(),
	}
	gu.gameRepo.SaveGameEvent(ctx, event)
	
	gu.logger.Infof("Spawned special fish %d (type %d) in room %s", fish.ID, fishTypeID, roomID)
	return fish, nil
}

// GetMathModelStats 獲取數學模型統計
func (gu *GameUsecase) GetMathModelStats(ctx context.Context) map[string]interface{} {
	return gu.mathModel.GetModelStats()
}

// UpdateRoomConfig 更新房間配置（管理員功能）
func (gu *GameUsecase) UpdateRoomConfig(ctx context.Context, roomID string, config RoomConfig) error {
	room, err := gu.roomManager.GetRoom(roomID)
	if err != nil {
		return err
	}
	
	room.Config = config
	room.UpdatedAt = time.Now()
	
	// 保存到數據庫
	return gu.gameRepo.SaveRoom(ctx, room)
}