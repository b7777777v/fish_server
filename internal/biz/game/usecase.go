package game

import (
	"context"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/wallet"
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

	// 魚類類型
	GetAllFishTypes(ctx context.Context) ([]*FishType, error)
	SaveFishTypeCache(ctx context.Context, ft *FishType) error
}

// InventoryRepo defines the persistence interface for game inventories.
type InventoryRepo interface {
	GetInventory(ctx context.Context, inventoryID string) (*Inventory, error)
	SaveInventory(ctx context.Context, inventory *Inventory) error
	GetAllInventories(ctx context.Context) (map[string]*Inventory, error)
}

// PlayerRepo 玩家數據倉庫接口
type PlayerRepo interface {
	GetPlayer(ctx context.Context, playerID int64) (*Player, error)
	UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error
	UpdatePlayerStatus(ctx context.Context, playerID int64, status PlayerStatus) error
}

// GameUsecase 遊戲用例
type GameUsecase struct {
	gameRepo         GameRepo
	playerRepo       PlayerRepo
	gameRecordRepo   GameRecordRepo
	walletUC         *wallet.WalletUsecase
	roomManager      *RoomManager
	spawner          *FishSpawner
	mathModel        *MathModel
	inventoryManager *InventoryManager
	rtpController    *RTPController
	logger           logger.Logger
}

// NewGameUsecase 創建遊戲用例
func NewGameUsecase(
	gameRepo GameRepo,
	playerRepo PlayerRepo,
	gameRecordRepo GameRecordRepo,
	walletUC *wallet.WalletUsecase,
	roomManager *RoomManager,
	spawner *FishSpawner,
	mathModel *MathModel,
	inventoryManager *InventoryManager,
	rtpController *RTPController,
	logger logger.Logger,
) *GameUsecase {
	return &GameUsecase{
		gameRepo:         gameRepo,
		playerRepo:       playerRepo,
		gameRecordRepo:   gameRecordRepo,
		walletUC:         walletUC,
		roomManager:      roomManager,
		spawner:          spawner,
		mathModel:        mathModel,
		inventoryManager: inventoryManager,
		rtpController:    rtpController,
		logger:           logger.With("component", "game_usecase"),
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
	
	// 記錄事件（魚類生成事件不關聯特定玩家，所以不設置 PlayerID）
	event := &GameEvent{
		ID:        time.Now().UnixNano(),
		Type:      EventFishSpawn,
		RoomID:    room.ID,
		PlayerID:  0, // 系統事件，不關聯玩家
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

	// 創建新的遊戲記錄（如果玩家沒有進行中的遊戲）
	activeRecord, err := gu.gameRecordRepo.FindActiveByUserID(ctx, playerID)
	if err != nil {
		gu.logger.Warnf("Failed to check active game record: %v", err)
	}

	if activeRecord == nil {
		// 生成會話ID（可以使用玩家ID + 時間戳）
		sessionID := fmt.Sprintf("session_%d_%d", playerID, time.Now().Unix())
		newRecord := NewGameRecord(playerID, roomID, sessionID)
		if err := gu.gameRecordRepo.Create(ctx, newRecord); err != nil {
			gu.logger.Errorf("Failed to create game record: %v", err)
			// 不阻塞加入房間流程，只記錄錯誤
		} else {
			gu.logger.Infof("Created game record for player %d: record_id=%d", playerID, newRecord.ID)
		}
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

	// 完成遊戲記錄
	activeRecord, err := gu.gameRecordRepo.FindActiveByUserID(ctx, playerID)
	if err != nil {
		gu.logger.Warnf("Failed to find active game record: %v", err)
	}

	if activeRecord != nil {
		activeRecord.Finish()
		if err := gu.gameRecordRepo.Update(ctx, activeRecord); err != nil {
			gu.logger.Errorf("Failed to finish game record: %v", err)
		} else {
			gu.logger.Infof("Finished game record for player %d: record_id=%d, profit=%.2f",
				playerID, activeRecord.ID, activeRecord.NetProfit)
		}
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
func (gu *GameUsecase) FireBullet(ctx context.Context, roomID string, playerID int64, direction float64, power int32, position Position, targetFishID int64) (*Bullet, error) {
	// 檢查參數
	if power < 1 || power > 100 {
		return nil, fmt.Errorf("invalid bullet power: %d", power)
	}

	// 發射子彈（內部會檢查餘額）
	bullet, err := gu.roomManager.FireBullet(roomID, playerID, direction, power, position, targetFishID)
	if err != nil {
		gu.logger.Errorf("Failed to fire bullet: %v", err)
		return nil, err
	}

	// 更新玩家餘額到數據庫並創建錢包交易記錄
	room, _ := gu.roomManager.GetRoom(roomID)
	if room != nil {
		if player, exists := room.Players[playerID]; exists {
			// 創建錢包交易記錄（如果玩家有錢包）
			var walletErr error
			if player.WalletID > 0 {
				bulletCost := float64(bullet.Cost) / 100.0 // 轉換為元
				walletErr = gu.walletUC.Withdraw(
					ctx,
					player.WalletID,
					bulletCost,
					"game_bullet_cost",
					fmt.Sprintf("game:%s:bullet:%d", roomID, bullet.ID),
					"子彈發射費用",
					map[string]interface{}{
						"room_id":      roomID,
						"bullet_id":    bullet.ID,
						"bullet_power": bullet.Power,
						"player_id":    playerID,
					},
				)
				if walletErr != nil {
					// 錢包操作失敗，需要回滾內存中的餘額扣除
					gu.logger.Errorf("Failed to create wallet transaction for bullet cost: %v, rolling back", walletErr)
					// 回滾內存餘額
					player.Balance += bullet.Cost
					// 不更新數據庫餘額，保持一致性
					return nil, fmt.Errorf("wallet operation failed: %w", walletErr)
				}
			}

			// 只有錢包操作成功（或玩家沒有錢包）才更新數據庫餘額
			gu.playerRepo.UpdatePlayerBalance(ctx, playerID, player.Balance)

			// 只有錢包操作成功才更新遊戲記錄
			if walletErr == nil {
				activeRecord, err := gu.gameRecordRepo.FindActiveByUserID(ctx, playerID)
				if err != nil {
					gu.logger.Warnf("Failed to find active game record: %v", err)
				}

				if activeRecord != nil {
					bulletCost := float64(bullet.Cost) / 100.0 // 轉換為元
					activeRecord.RecordBulletFired(bulletCost)
					if err := gu.gameRecordRepo.Update(ctx, activeRecord); err != nil {
						gu.logger.Warnf("Failed to update game record: %v", err)
					}
				}
			}
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
			// 更新玩家餘額到數據庫並創建錢包交易記錄
			if player, exists := room.Players[playerID]; exists && hitResult.Reward > 0 {
				// 創建錢包交易記錄（如果玩家有錢包且獲得獎勵）
				var walletErr error
				if player.WalletID > 0 {
					reward := float64(hitResult.Reward) / 100.0 // 轉換為元
					walletErr = gu.walletUC.Deposit(
						ctx,
						player.WalletID,
						reward,
						"game_fish_reward",
						fmt.Sprintf("game:%s:fish:%d", roomID, fishID),
						"捕魚獎勵",
						map[string]interface{}{
							"room_id":      roomID,
							"fish_id":      fishID,
							"bullet_id":    bulletID,
							"damage":       hitResult.Damage,
							"is_critical":  hitResult.IsCritical,
							"multiplier":   hitResult.Multiplier,
							"player_id":    playerID,
						},
					)
					if walletErr != nil {
						// 錢包操作失敗，需要回滾內存中的獎勵增加
						gu.logger.Errorf("Failed to create wallet transaction for fish reward: %v, rolling back", walletErr)
						// 回滾內存餘額
						player.Balance -= hitResult.Reward
						// 記錄錯誤但不阻塞遊戲流程（因為魚已經死亡）
						gu.logger.Warnf("Rolled back reward for player %d, amount: %d", playerID, hitResult.Reward)
						// 將 hitResult 的 Reward 設為 0，表示未實際獲得獎勵
						hitResult.Reward = 0
					}
				}

				// 只有錢包操作成功（或玩家沒有錢包）才更新數據庫餘額
				gu.playerRepo.UpdatePlayerBalance(ctx, playerID, player.Balance)

				// 只有錢包操作成功且有實際獎勵才更新遊戲記錄
				if walletErr == nil && hitResult.Reward > 0 {
					activeRecord, err := gu.gameRecordRepo.FindActiveByUserID(ctx, playerID)
					if err != nil {
						gu.logger.Warnf("Failed to find active game record: %v", err)
					}

					if activeRecord != nil {
						reward := float64(hitResult.Reward) / 100.0 // 轉換為元
						activeRecord.RecordFishCaught(reward, hitResult.IsCritical)
						if err := gu.gameRecordRepo.Update(ctx, activeRecord); err != nil {
							gu.logger.Warnf("Failed to update game record: %v", err)
						}
					}
				}
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

// GetPlayerInfo 獲取玩家信息（包含最新餘額）
func (gu *GameUsecase) GetPlayerInfo(ctx context.Context, playerID int64) (*Player, error) {
	player, err := gu.playerRepo.GetPlayer(ctx, playerID)
	if err != nil {
		gu.logger.Errorf("Failed to get player %d: %v", playerID, err)
		return nil, err
	}
	return player, nil
}

// GetGameEvents 獲取遊戲事件
func (gu *GameUsecase) GetGameEvents(ctx context.Context, roomID string, limit int) ([]*GameEvent, error) {
	if limit <= 0 {
		limit = 50 // 默認限制
	}
	return gu.gameRepo.GetGameEvents(ctx, roomID, limit)
}

// LoadAndCacheFishTypes 從數據庫加載所有魚類類型並緩存到 Redis
func (gu *GameUsecase) LoadAndCacheFishTypes(ctx context.Context) error {
	fishTypes, err := gu.gameRepo.GetAllFishTypes(ctx)
	if err != nil {
		gu.logger.Errorf("Failed to get all fish types from DB: %v", err)
		return err
	}

	for _, ft := range fishTypes {
		if err := gu.gameRepo.SaveFishTypeCache(ctx, ft); err != nil {
			// 即使單個失敗也繼續嘗試緩存其他的
			gu.logger.Warnf("Failed to cache fish type %d: %v", ft.ID, err)
		}
	}

	gu.logger.Infof("Successfully loaded and cached %d fish types.", len(fishTypes))
	return nil
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

// GetMathModelConfig 獲取數學模型配置
func (gu *GameUsecase) GetMathModelConfig(ctx context.Context) ModelConfig {
	return gu.mathModel.GetModelConfig()
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

// GetRoom 獲取房間詳細信息
func (gu *GameUsecase) GetRoom(ctx context.Context, roomID string) (*Room, error) {
	room, err := gu.roomManager.GetRoom(roomID)
	if err != nil {
		gu.logger.Errorf("Failed to get room %s: %v", roomID, err)
		return nil, err
	}
	
	gu.logger.Debugf("Retrieved room: %s", roomID)
	return room, nil
}

// GetFormationsInRoom 獲取房間中的陣型
func (gu *GameUsecase) GetFormationsInRoom(ctx context.Context, roomID string) ([]*FishFormation, error) {
	formations, err := gu.roomManager.GetFormationsInRoom(roomID)
	if err != nil {
		gu.logger.Errorf("Failed to get formations in room %s: %v", roomID, err)
		return nil, err
	}

	gu.logger.Debugf("Retrieved %d formations from room: %s", len(formations), roomID)
	return formations, nil
}

// ========================================
// 陣型配置管理
// ========================================

// GetFormationConfig 獲取當前陣型配置
func (gu *GameUsecase) GetFormationConfig() FormationSpawnConfig {
	return gu.spawner.GetFormationConfig()
}

// UpdateFormationConfig 更新陣型配置
func (gu *GameUsecase) UpdateFormationConfig(config FormationSpawnConfig) {
	gu.spawner.UpdateFormationConfig(config)
	gu.logger.Infof("Updated formation spawn config")
}

// SetFormationDifficulty 設置陣型難度
func (gu *GameUsecase) SetFormationDifficulty(difficulty string) error {
	gu.spawner.SetFormationDifficulty(difficulty)
	gu.logger.Infof("Set formation difficulty to: %s", difficulty)
	return nil
}

// SetFormationSpawnRate 設置陣型生成率
func (gu *GameUsecase) SetFormationSpawnRate(minInterval, maxInterval int, baseChance float64) {
	gu.spawner.SetFormationSpawnRate(minInterval, maxInterval, baseChance)
	gu.logger.Infof("Updated formation spawn rate")
}

// GetFormationSpawnStats 獲取陣型生成統計
func (gu *GameUsecase) GetFormationSpawnStats() map[string]interface{} {
	return gu.spawner.GetFormationSpawnStats()
}

// EnableFormationSpawn 啟用/禁用陣型生成
func (gu *GameUsecase) EnableFormationSpawn(enabled bool) {
	gu.spawner.EnableFormationSpawn(enabled)
	gu.logger.Infof("Formation spawn enabled: %v", enabled)
}

// TriggerSpecialFormationEvent 觸發特殊陣型事件
func (gu *GameUsecase) TriggerSpecialFormationEvent(multiplier float64, duration time.Duration) {
	gu.spawner.TriggerSpecialEvent(multiplier, duration)
	gu.logger.Infof("Triggered special formation event: multiplier=%.2f, duration=%v", multiplier, duration)
}

// GetRoomsFromDB 直接從資料庫獲取房間列表（按類型）
func (gu *GameUsecase) GetRoomsFromDB(ctx context.Context, roomType RoomType) ([]*Room, error) {
    return gu.gameRepo.ListRooms(ctx, roomType)
}