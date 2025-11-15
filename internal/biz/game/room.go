package game

import (
	"fmt"
	"math"
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
	config := rm.getRoomConfig(roomType)

	// 使用配置中的 MaxPlayers，如果配置中有的话，否则使用传入的参数
	seatCount := maxPlayers
	if config.MaxPlayers > 0 {
		seatCount = config.MaxPlayers
	}

	room := &Room{
		ID:         roomID,
		Name:       fmt.Sprintf("%s房間", roomType),
		Type:       roomType,
		MaxPlayers: seatCount,
		Players:    make(map[int64]*Player),
		Seats:      make([]int64, seatCount), // 初始化座位切片，默认值为0表示空座位
		Fishes:     make(map[int64]*Fish),
		Bullets:    make(map[int64]*Bullet),
		Status:     RoomStatusWaiting,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Config:     config,
	}

	rm.rooms[roomID] = room
	rm.logger.Infof("Created room: %s, type: %s, seats: %d", roomID, roomType, seatCount)

	// 立即啟動遊戲循環，不等待玩家加入
	// 魚應該一直游動，不管有沒有玩家
	room.Status = RoomStatusPlaying
	go rm.startRoomGameLoop(room)
	rm.logger.Infof("Game loop started for room: %s", roomID)

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

	// 使用新的座位管理检查房间是否已满
	if room.IsFull() {
		return fmt.Errorf("room is full, no available seats")
	}

	// 檢查玩家是否已在其他房間
	for _, existingRoom := range rm.rooms {
		if _, playerExists := existingRoom.Players[player.ID]; playerExists {
			return fmt.Errorf("player already in room: %s", existingRoom.ID)
		}
	}

	// 分配座位
	seatID, err := room.AllocateSeat(player.ID)
	if err != nil {
		return fmt.Errorf("failed to allocate seat: %w", err)
	}

	player.RoomID = roomID
	player.SeatID = seatID
	player.Status = PlayerStatusPlaying
	player.JoinTime = time.Now()
	room.Players[player.ID] = player
	room.UpdatedAt = time.Now()

	// 遊戲循環已經在房間創建時啟動，不需要在這裡再次啟動

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

	// 释放座位
	if player.SeatID >= 0 && player.SeatID < len(room.Seats) {
		if err := room.ReleaseSeat(player.SeatID); err != nil {
			rm.logger.Warnf("Failed to release seat %d for player %d: %v", player.SeatID, playerID, err)
		} else {
			rm.logger.Debugf("Released seat %d for player %d", player.SeatID, playerID)
		}
	}

	delete(room.Players, playerID)
	player.RoomID = ""
	player.SeatID = -1 // 重置座位ID
	player.Status = PlayerStatusIdle
	room.UpdatedAt = time.Now()

	// 遊戲循環會繼續運行，即使沒有玩家
	// 魚會繼續游動，等待新玩家加入
	rm.logger.Infof("Player %d left room %s, remaining players: %d", playerID, roomID, len(room.Players))

	return nil
}

// FireBullet 玩家開火
func (rm *RoomManager) FireBullet(roomID string, playerID int64, direction float64, power int32, position Position) (*Bullet, error) {
	// First check room and player existence with read lock
	rm.mu.RLock()
	room, exists := rm.rooms[roomID]
	if !exists {
		rm.mu.RUnlock()
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	player, playerExists := room.Players[playerID]
	if !playerExists {
		rm.mu.RUnlock()
		return nil, fmt.Errorf("player not in room")
	}

	// Calculate bullet cost
	bulletCost := int64(float64(power) * room.Config.BulletCostMultiplier)
	if player.Balance < bulletCost {
		rm.mu.RUnlock()
		return nil, fmt.Errorf("insufficient balance")
	}
	rm.mu.RUnlock()

	// Create bullet outside of lock
	bulletID := time.Now().UnixNano()
	bullet := &Bullet{
		ID:        bulletID,
		PlayerID:  playerID,
		Position:  position, // 使用客戶端發送的位置
		Direction: direction,
		Speed:     500.0, // 固定速度
		Power:     power,
		Cost:      bulletCost,
		CreatedAt: time.Now(),
		Status:    BulletStatusFlying,
	}

	// Now acquire write lock for the actual modifications
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Double-check room and player still exist
	room, exists = rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	player, playerExists = room.Players[playerID]
	if !playerExists {
		return nil, fmt.Errorf("player not in room")
	}

	// Final balance check
	if player.Balance < bulletCost {
		return nil, fmt.Errorf("insufficient balance")
	}

	// Apply changes
	player.Balance -= bulletCost
	room.Bullets[bulletID] = bullet
	room.UpdatedAt = time.Now()

	// Release lock before external calls
	rm.mu.Unlock()

	// 將成本計入庫存系統 (external call without lock)
	rm.inventoryManager.AddBet(room.Type, bullet.Cost)

	// Re-acquire lock briefly for logging
	rm.mu.Lock()
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
			// 注意：即使沒有玩家，遊戲循環也應該繼續，只有房間狀態為 Closed 時才停止
			if room.Status == RoomStatusClosed {
				rm.logger.Infof("Game loop ended for room %s", room.ID)
				return
			}
		}
	}
}

// updateRoom 更新房間狀態
func (rm *RoomManager) updateRoom(room *Room) {
	now := time.Now()
	deltaTime := 0.1 // 100ms

	// Update formations outside of lock (they have their own synchronization)
	rm.spawner.UpdateFormations(deltaTime)

	// Try spawn formation outside of lock
	newFormation := rm.spawner.TrySpawnFormation(room.Config, len(room.Players))

	// Try spawn fish outside of lock
	var newFish *Fish
	var batchFishes []*Fish
	rm.mu.RLock()
	fishCount := len(room.Fishes)
	minFish := int(room.Config.MinFishCount)
	maxFish := int(room.Config.MaxFishCount)
	rm.mu.RUnlock()

	// 魚數量監控：低於最小值時強制補充
	if fishCount < minFish {
		// 計算需要補充的魚數量，補充到最大值的 75%
		targetFishCount := int(float64(maxFish) * 0.75)
		spawnCount := targetFishCount - fishCount
		if spawnCount > 0 {
			rm.logger.Warnf("Room %s fish count too low (%d < %d), spawning %d fish to reach %d",
				room.ID, fishCount, minFish, spawnCount, targetFishCount)
			batchFishes = rm.spawner.BatchSpawnFish(spawnCount, room.Config)
		}
	} else if fishCount < maxFish {
		// 正常情況下使用概率生成
		newFish = rm.spawner.TrySpawnFish(room.Config)
	}

	// Now acquire write lock for minimal time
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.logger.Debugf("[GAME_LOOP] Room %s: Total fishes=%d, Total bullets=%d",
		room.ID, len(room.Fishes), len(room.Bullets))

	// Get all fish IDs that are in formations
	fishInFormations := make(map[int64]bool)
	formations := rm.spawner.GetFormationManager().GetAllFormations()
	formationFishCount := 0
	for _, formation := range formations {
		for _, fish := range formation.Fishes {
			fishInFormations[fish.ID] = true
			formationFishCount++
		}
	}

	// Log formation status
	if len(formations) > 0 {
		rm.logger.Debugf("Active formations: %d, fish in formations: %d, total fish: %d",
			len(formations), formationFishCount, len(room.Fishes))
	}

	// Update fish positions only for fish NOT in formations
	// Fish in formations are updated by the formation system
	independentFishCount := 0
	for _, fish := range room.Fishes {
		if !fishInFormations[fish.ID] {
			rm.updateFishPosition(fish, room.Config)
			independentFishCount++
		}
	}

	// Remove dead fish (out of bounds or killed)
	for fishID, fish := range room.Fishes {
		if fish.Status == FishStatusDead {
			delete(room.Fishes, fishID)
		}
	}

	if independentFishCount > 0 {
		rm.logger.Debugf("Updated %d independent fish (not in formations)", independentFishCount)
	}

	// Update bullet positions and remove expired/out-of-bounds bullets
	for bulletID, bullet := range room.Bullets {
		// Update bullet position based on speed and direction
		rm.updateBulletPosition(bullet, deltaTime, room.Config)

		// Remove bullet if expired or out of bounds
		if now.Sub(bullet.CreatedAt) > 5*time.Second ||
			bullet.Position.X < -100 || bullet.Position.X > room.Config.RoomWidth+100 ||
			bullet.Position.Y < -100 || bullet.Position.Y > room.Config.RoomHeight+100 {
			delete(room.Bullets, bulletID)
		}
	}
	
	// Add new fish if spawned
	if newFish != nil {
		room.Fishes[newFish.ID] = newFish
	}

	// Add batch spawned fish (from low fish count replenishment)
	if len(batchFishes) > 0 {
		for _, fish := range batchFishes {
			room.Fishes[fish.ID] = fish
		}
		rm.logger.Infof("Replenished room %s with %d fish (new total: %d)",
			room.ID, len(batchFishes), len(room.Fishes))
	}

	// Add formation fishes if spawned
	if newFormation != nil {
		for _, fish := range newFormation.Fishes {
			room.Fishes[fish.ID] = fish
		}
		rm.logger.Infof("Spawned formation in room %s: %s with %d fishes",
			room.ID, newFormation.Type, len(newFormation.Fishes))
	}

	// Clean up completed formations
	rm.cleanupCompletedFormations(room)

	room.UpdatedAt = now
}

// updateFishPosition 更新魚的位置
func (rm *RoomManager) updateFishPosition(fish *Fish, config RoomConfig) {
	// 簡單的直線移動
	deltaTime := 0.1 // 100ms

	// 使用三角函數計算基於方向的移動
	// Direction 是弧度值
	fish.Position.X += fish.Speed * deltaTime * math.Cos(fish.Direction)
	fish.Position.Y += fish.Speed * deltaTime * math.Sin(fish.Direction)

	// 邊界檢查，魚游出邊界後移除（由spawner重新生成新魚）
	// 不重置位置，而是標記為已離開
	if fish.Position.X > config.RoomWidth+50 || fish.Position.X < -50 ||
		fish.Position.Y > config.RoomHeight+50 || fish.Position.Y < -50 {
		fish.Status = FishStatusDead
	}
}

// updateBulletPosition 更新子彈的位置
func (rm *RoomManager) updateBulletPosition(bullet *Bullet, deltaTime float64, config RoomConfig) {
	// 根據子彈的方向和速度更新位置
	// Speed 是像素/秒，deltaTime 是秒
	// Direction 是弧度值
	bullet.Position.X += bullet.Speed * deltaTime * math.Cos(bullet.Direction)
	bullet.Position.Y += bullet.Speed * deltaTime * math.Sin(bullet.Direction)
}

// cleanupCompletedFormations 清理已完成的阵型
func (rm *RoomManager) cleanupCompletedFormations(room *Room) {
	formations := rm.spawner.GetFormationManager().GetAllFormations()

	for _, formation := range formations {
		if formation.Status == FormationStatusComplete {
			// 移除阵型中的鱼
			for _, fish := range formation.Fishes {
				delete(room.Fishes, fish.ID)
			}

			// 从管理器中移除阵型
			rm.spawner.GetFormationManager().RemoveFormation(formation.ID)

			rm.logger.Infof("Cleaned up completed formation %s (type: %s) with %d fishes",
				formation.ID, formation.Type, len(formation.Fishes))
		}
	}
}

// getRoomConfig 獲取房間配置
func (rm *RoomManager) getRoomConfig(roomType RoomType) RoomConfig {
	configs := map[RoomType]RoomConfig{
		RoomTypeNovice: {
			MaxPlayers:           4,    // 4人座位
			MinBet:               10,   // 0.1元
			MaxBet:               100,  // 1元
			BulletCostMultiplier: 1.0,
			FishSpawnRate:        0.3,
			MinFishCount:         10,  // 最小魚數量（低於此值強制補充）
			MaxFishCount:         20,  // 最大魚數量
			RoomWidth:            1200,
			RoomHeight:           800,
			TargetRTP:            0.97, // 新手房RTP略高
		},
		RoomTypeIntermediate: {
			MaxPlayers:           4,    // 4人座位
			MinBet:               100,  // 1元
			MaxBet:               1000, // 10元
			BulletCostMultiplier: 2.0,
			FishSpawnRate:        0.4,
			MinFishCount:         12,  // 最小魚數量（低於此值強制補充）
			MaxFishCount:         25,  // 最大魚數量
			RoomWidth:            1200,
			RoomHeight:           800,
			TargetRTP:            0.96,
		},
		RoomTypeAdvanced: {
			MaxPlayers:           4,    // 4人座位
			MinBet:               1000,  // 10元
			MaxBet:               10000, // 100元
			BulletCostMultiplier: 5.0,
			FishSpawnRate:        0.5,
			MinFishCount:         15,  // 最小魚數量（低於此值強制補充）
			MaxFishCount:         30,  // 最大魚數量
			RoomWidth:            1200,
			RoomHeight:           800,
			TargetRTP:            0.95,
		},
		RoomTypeVIP: {
			MaxPlayers:           4,    // 4人座位
			MinBet:               10000, // 100元
			MaxBet:               100000, // 1000元
			BulletCostMultiplier: 10.0,
			FishSpawnRate:        0.6,
			MinFishCount:         18,  // 最小魚數量（低於此值強制補充）
			MaxFishCount:         35,  // 最大魚數量
			RoomWidth:            1200,
			RoomHeight:           800,
			TargetRTP:            0.94, // VIP房RTP略低
		},
	}

	return configs[roomType]
}

// ========================================
// 魚群陣型管理相關方法
// ========================================

// SpawnFormationInRoom 在指定房間生成魚群陣型
func (rm *RoomManager) SpawnFormationInRoom(roomID string, formationType FishFormationType, routeID string) (*FishFormation, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	formation := rm.spawner.TrySpawnFormation(room.Config, len(room.Players))
	if formation == nil {
		return nil, fmt.Errorf("failed to spawn formation")
	}

	// 將陣型中的魚添加到房間
	for _, fish := range formation.Fishes {
		room.Fishes[fish.ID] = fish
	}

	rm.logger.Infof("Manually spawned formation in room %s: %s", roomID, formation.Type)
	return formation, nil
}

// SpawnSpecialFormationInRoom 在房間生成特殊陣型
func (rm *RoomManager) SpawnSpecialFormationInRoom(roomID string, formationType FishFormationType, routeID string, fishTypeIDs []int32) (*FishFormation, error) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	formation := rm.spawner.SpawnSpecialFormation(formationType, routeID, fishTypeIDs, room.Config)
	if formation == nil {
		return nil, fmt.Errorf("failed to spawn special formation")
	}

	// 將陣型中的魚添加到房間
	for _, fish := range formation.Fishes {
		room.Fishes[fish.ID] = fish
	}

	rm.logger.Infof("Spawned special formation in room %s: %s with %d fishes", 
		roomID, formation.Type, len(formation.Fishes))
	return formation, nil
}

// GetFormationsInRoom 獲取房間中的所有陣型
func (rm *RoomManager) GetFormationsInRoom(roomID string) ([]*FishFormation, error) {
	room, exists := rm.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	formations := rm.spawner.GetFormationManager().GetAllFormations()
	
	// 篩選出該房間中的陣型（通過檢查魚是否在房間中）
	var roomFormations []*FishFormation
	for _, formation := range formations {
		for _, fish := range formation.Fishes {
			if _, exists := room.Fishes[fish.ID]; exists {
				roomFormations = append(roomFormations, formation)
				break
			}
		}
	}

	return roomFormations, nil
}

// StopFormationInRoom 停止房間中的指定陣型
func (rm *RoomManager) StopFormationInRoom(roomID string, formationID string) error {
	_, exists := rm.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	success := rm.spawner.GetFormationManager().StopFormation(formationID)
	if !success {
		return fmt.Errorf("formation not found or failed to stop: %s", formationID)
	}

	rm.logger.Infof("Stopped formation %s in room %s", formationID, roomID)
	return nil
}

// GetAvailableRoutes 獲取可用的路線列表
func (rm *RoomManager) GetAvailableRoutes() []*FishRoute {
	return rm.spawner.GetFormationManager().GetAllRoutes()
}

// GetRoutesByType 根據類型獲取路線
func (rm *RoomManager) GetRoutesByType(routeType FishRouteType) []*FishRoute {
	return rm.spawner.GetFormationManager().GetRoutesByType(routeType)
}

// CreateCustomRoute 創建自定義路線
func (rm *RoomManager) CreateCustomRoute(id, name string, points []Position, routeType FishRouteType, difficulty float64, looping bool) (*FishRoute, error) {
	route := rm.spawner.GetFormationManager().CreateCustomRoute(id, name, points, routeType, difficulty, looping)
	if route == nil {
		return nil, fmt.Errorf("failed to create route")
	}
	
	rm.logger.Infof("Created custom route: %s", route.Name)
	return route, nil
}

// RemoveCustomRoute 移除自定義路線
func (rm *RoomManager) RemoveCustomRoute(routeID string) error {
	success := rm.spawner.GetFormationManager().RemoveRoute(routeID)
	if !success {
		return fmt.Errorf("failed to remove route: %s", routeID)
	}
	
	rm.logger.Infof("Removed custom route: %s", routeID)
	return nil
}

// GetFormationStatistics 獲取陣型統計信息
func (rm *RoomManager) GetFormationStatistics(roomID string) (map[string]interface{}, error) {
	formations, err := rm.GetFormationsInRoom(roomID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_formations": len(formations),
		"formations_by_type": make(map[FishFormationType]int),
		"formations_by_status": make(map[FormationStatus]int),
		"total_formation_fishes": 0,
	}

	formationsByType := make(map[FishFormationType]int)
	formationsByStatus := make(map[FormationStatus]int)
	totalFishes := 0

	for _, formation := range formations {
		formationsByType[formation.Type]++
		formationsByStatus[formation.Status]++
		totalFishes += len(formation.Fishes)
	}

	stats["formations_by_type"] = formationsByType
	stats["formations_by_status"] = formationsByStatus
	stats["total_formation_fishes"] = totalFishes

	return stats, nil
}