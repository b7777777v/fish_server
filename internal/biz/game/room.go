package game

import (
	"fmt"
	"sync"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// ========================================
// Room 遊戲房間管理
// ========================================

// RoomManager 房間管理器
type RoomManager struct {
	rooms            map[string]*Room
	mu               sync.RWMutex
	logger           logger.Logger
	spawner          *FishSpawner
	mathModel        *MathModel
	inventoryManager *InventoryManager
	rtpController    *RTPController
}

// NewRoomManager 創建房間管理器
func NewRoomManager(logger logger.Logger, spawner *FishSpawner, mathModel *MathModel, im *InventoryManager, rc *RTPController) *RoomManager {
	return &RoomManager{
		rooms:            make(map[string]*Room),
		logger:           logger.With("component", "room_manager"),
		spawner:          spawner,
		mathModel:        mathModel,
		inventoryManager: im,
		rtpController:    rc,
	}
}

// CreateRoom 創建房間
func (rm *RoomManager) CreateRoom(roomType RoomType, maxPlayers int32) (*Room, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	roomID := fmt.Sprintf("room_%s_%d", roomType, time.Now().Unix())
	
	room := &Room{
		ID:         roomID,
		Name:       fmt.Sprintf("%s房間", roomType),
		Type:       roomType,
		MaxPlayers: maxPlayers,
		Players:    make(map[int64]*Player),
		Fishes:     make(map[int64]*Fish),
		Bullets:    make(map[int64]*Bullet),
		Status:     RoomStatusWaiting,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Config:     rm.getRoomConfig(roomType),
	}

	rm.rooms[roomID] = room
	rm.logger.Infof("Created room: %s, type: %s", roomID, roomType)
	
	return room, nil
}

// GetRoom 獲取房間
func (rm *RoomManager) GetRoom(roomID string) (*Room, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}
	
	return room, nil
}

// JoinRoom 玩家加入房間
func (rm *RoomManager) JoinRoom(roomID string, player *Player) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	if len(room.Players) >= int(room.MaxPlayers) {
		return fmt.Errorf("room is full")
	}

	// 檢查玩家是否已在其他房間
	for _, existingRoom := range rm.rooms {
		if _, playerExists := existingRoom.Players[player.ID]; playerExists {
			return fmt.Errorf("player already in room: %s", existingRoom.ID)
		}
	}

	player.RoomID = roomID
	player.Status = PlayerStatusPlaying
	player.JoinTime = time.Now()
	room.Players[player.ID] = player
	room.UpdatedAt = time.Now()

	// 如果房間人數達到要求，開始遊戲
	if len(room.Players) >= 1 && room.Status == RoomStatusWaiting {
		room.Status = RoomStatusPlaying
		go rm.startRoomGameLoop(room)
	}

	rm.logger.Infof("Player %d joined room %s", player.ID, roomID)
	return nil
}

// LeaveRoom 玩家離開房間
func (rm *RoomManager) LeaveRoom(roomID string, playerID int64) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	player, playerExists := room.Players[playerID]
	if !playerExists {
		return fmt.Errorf("player not in room")
	}

	delete(room.Players, playerID)
	player.RoomID = ""
	player.Status = PlayerStatusIdle
	room.UpdatedAt = time.Now()

	// 如果房間沒有玩家了，關閉房間
	if len(room.Players) == 0 {
		room.Status = RoomStatusClosed
		rm.logger.Infof("Room %s closed due to no players", roomID)
	}

	rm.logger.Infof("Player %d left room %s", playerID, roomID)
	return nil
}

// FireBullet 玩家開火
func (rm *RoomManager) FireBullet(roomID string, playerID int64, direction float64, power int32) (*Bullet, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	player, playerExists := room.Players[playerID]
	if !playerExists {
		return nil, fmt.Errorf("player not in room")
	}

	// 計算子彈成本
	bulletCost := int64(float64(power) * room.Config.BulletCostMultiplier)
	if player.Balance < bulletCost {
		return nil, fmt.Errorf("insufficient balance")
	}

	// 創建子彈
	bulletID := time.Now().UnixNano()
	bullet := &Bullet{
		ID:        bulletID,
		PlayerID:  playerID,
		Position:  Position{X: 100, Y: 300}, // 固定發射位置
		Direction: direction,
		Speed:     500.0, // 固定速度
		Power:     power,
		Cost:      bulletCost,
		CreatedAt: time.Now(),
		Status:    BulletStatusFlying,
	}

	// 扣除玩家餘額
	player.Balance -= bulletCost
	room.Bullets[bulletID] = bullet
	room.UpdatedAt = time.Now()

	// 將成本計入庫存系統
	rm.inventoryManager.AddBet(room.Type, bullet.Cost)

	rm.logger.Infof("Player %d fired bullet in room %s, cost: %d", playerID, roomID, bulletCost)
	return bullet, nil
}

// ProcessBulletHit 處理子彈命中
func (rm *RoomManager) ProcessBulletHit(roomID string, bulletID int64, fishID int64) (*HitResult, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	bullet, bulletExists := room.Bullets[bulletID]
	if !bulletExists {
		return nil, fmt.Errorf("bullet not found")
	}

	fish, fishExists := room.Fishes[fishID]
	if !fishExists {
		return nil, fmt.Errorf("fish not found")
	}

	player, playerExists := room.Players[bullet.PlayerID]
	if !playerExists {
		return nil, fmt.Errorf("player not found")
	}

	// 1. Calculate the potential outcome from the math model
	potentialHit := rm.mathModel.CalculatePotentialHit(bullet, fish)

	// Clean up bullet immediately
	bullet.Status = BulletStatusHit
	delete(room.Bullets, bulletID)
	room.UpdatedAt = time.Now()

	// 2. If the hit is a potential kill, ask the RTP controller for approval
	if potentialHit.Success { // Success from math model means a potential kill
		approved := rm.rtpController.ApproveKill(room.Type, room.Config.TargetRTP, potentialHit.Reward)

		if approved {
			// 3a. Kill is approved: Grant the reward
			fish.Status = FishStatusDead
			delete(room.Fishes, fishID)

			player.Balance += potentialHit.Reward
			rm.inventoryManager.AddWin(room.Type, potentialHit.Reward)

			rm.logger.Infof("RTP APPROVED kill. Player %d killed fish %d, reward: %d", player.ID, fishID, potentialHit.Reward)
			return potentialHit, nil
		} else {
			// 3b. Kill is denied by RTP controller: Downgrade to non-lethal damage
			fish.Health -= potentialHit.Damage
			// Ensure fish survives, maybe with 1 HP
			if fish.Health <= 0 {
				fish.Health = 1
			}

			rm.logger.Infof("RTP DENIED kill. Player %d hit fish %d, but reward was not approved.", player.ID, fishID)

			// Return a result indicating damage but no kill/reward
			return &HitResult{
				Success:    false,
				Damage:     potentialHit.Damage,
				Reward:     0,
				IsCritical: potentialHit.IsCritical,
				Multiplier: 0,
			}, nil
		}
	} else {
		// 4. Hit was not a potential kill from the start, just apply damage
		fish.Health -= potentialHit.Damage
		rm.logger.Debugf("Player %d hit fish %d, no kill. Damage: %d", player.ID, fishID, potentialHit.Damage)
		return potentialHit, nil
	}
}

// GetRoomList 獲取房間列表
func (rm *RoomManager) GetRoomList() []*Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rooms := make([]*Room, 0, len(rm.rooms))
	for _, room := range rm.rooms {
		if room.Status != RoomStatusClosed {
			rooms = append(rooms, room)
		}
	}
	
	return rooms
}

// startRoomGameLoop 開始房間遊戲循環
func (rm *RoomManager) startRoomGameLoop(room *Room) {
	ticker := time.NewTicker(100 * time.Millisecond) // 10 FPS
	defer ticker.Stop()

	rm.logger.Infof("Starting game loop for room %s", room.ID)

	for {
		select {
		case <-ticker.C:
			rm.updateRoom(room)
			
			// 檢查房間是否應該關閉
			if room.Status == RoomStatusClosed || len(room.Players) == 0 {
				rm.logger.Infof("Game loop ended for room %s", room.ID)
				return
			}
		}
	}
}

// updateRoom 更新房間狀態
func (rm *RoomManager) updateRoom(room *Room) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	now := time.Now()
	
	// 更新魚的位置
	for _, fish := range room.Fishes {
		rm.updateFishPosition(fish, room.Config)
	}
	
	// 移除超時的子彈
	for bulletID, bullet := range room.Bullets {
		if now.Sub(bullet.CreatedAt) > 5*time.Second {
			delete(room.Bullets, bulletID)
		}
	}
	
	// 生成新魚
	if len(room.Fishes) < int(room.Config.MaxFishCount) {
		if fish := rm.spawner.TrySpawnFish(room.Config); fish != nil {
			room.Fishes[fish.ID] = fish
		}
	}
	
	room.UpdatedAt = now
}

// updateFishPosition 更新魚的位置
func (rm *RoomManager) updateFishPosition(fish *Fish, config RoomConfig) {
	// 簡單的直線移動
	deltaTime := 0.1 // 100ms
	
	fish.Position.X += fish.Speed * deltaTime * fish.Direction
	fish.Position.Y += fish.Speed * deltaTime * 0.1 // 輕微的Y軸移動
	
	// 邊界檢查，魚游出邊界後重新生成位置
	if fish.Position.X > config.RoomWidth || fish.Position.X < 0 ||
		fish.Position.Y > config.RoomHeight || fish.Position.Y < 0 {
		// 重新設置魚的位置
		fish.Position.X = 0
		fish.Position.Y = config.RoomHeight / 2
		fish.Direction = 1.0 // 向右游
	}
}

// getRoomConfig 獲取房間配置
func (rm *RoomManager) getRoomConfig(roomType RoomType) RoomConfig {
	configs := map[RoomType]RoomConfig{
		RoomTypeNovice: {
			MinBet:               10,   // 0.1元
			MaxBet:               100,  // 1元
			BulletCostMultiplier: 1.0,
			FishSpawnRate:        0.3,
			MaxFishCount:         20,
			RoomWidth:            1200,
			RoomHeight:           800,
			TargetRTP:            0.97, // 新手房RTP略高
		},
		RoomTypeIntermediate: {
			MinBet:               100,  // 1元
			MaxBet:               1000, // 10元
			BulletCostMultiplier: 2.0,
			FishSpawnRate:        0.4,
			MaxFishCount:         25,
			RoomWidth:            1200,
			RoomHeight:           800,
			TargetRTP:            0.96,
		},
		RoomTypeAdvanced: {
			MinBet:               1000,  // 10元
			MaxBet:               10000, // 100元
			BulletCostMultiplier: 5.0,
			FishSpawnRate:        0.5,
			MaxFishCount:         30,
			RoomWidth:            1200,
			RoomHeight:           800,
			TargetRTP:            0.95,
		},
		RoomTypeVIP: {
			MinBet:               10000, // 100元
			MaxBet:               100000, // 1000元
			BulletCostMultiplier: 10.0,
			FishSpawnRate:        0.6,
			MaxFishCount:         35,
			RoomWidth:            1200,
			RoomHeight:           800,
			TargetRTP:            0.94, // VIP房RTP略低
		},
	}
	
	return configs[roomType]
}