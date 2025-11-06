package game

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
	"google.golang.org/protobuf/proto"
)

// ========================================
// RoomManager - 房間的 Goroutine 管理器
// ========================================

// RoomManager 管理單個房間的遊戲循環
type RoomManager struct {
	// 房間ID
	roomID string

	// 房間中的客戶端
	clients map[*Client]bool

	// 遊戲用例
	gameUsecase *game.GameUsecase

	// Hub 引用
	hub *Hub

	// 遊戲循環控制
	gameLoopTicker *time.Ticker
	gameLoopStop   chan bool
	tickDone       chan bool

	// 客戶端操作通道
	addClient    chan *Client
	removeClient chan *Client
	gameAction   chan *GameActionMessage

	// 遊戲狀態
	gameState *GameState
	
	// 廣播狀態追蹤
	lastBroadcast time.Time
	lastFishCount int
	lastBulletCount int
	lastFishSpawn time.Time

	// 日誌記錄器
	logger logger.Logger

	// 上下文和取消函數
	ctx    context.Context
	cancel context.CancelFunc
}

// GameState 房間遊戲狀態
type GameState struct {
	RoomID        string                 `json:"room_id"`
	Status        string                 `json:"status"` // waiting, playing, paused
	Players       map[string]*PlayerInfo `json:"players"`
	Fishes        map[int64]*FishInfo    `json:"fishes"`
	Bullets       map[int64]*BulletInfo  `json:"bullets"`
	LastUpdate    time.Time              `json:"last_update"`
	GameStartTime time.Time              `json:"game_start_time"`
}

// PlayerInfo 玩家信息
type PlayerInfo struct {
	ID       string       `json:"id"`
	PlayerID int64        `json:"player_id"`
	Nickname string       `json:"nickname"`
	Balance  int64        `json:"balance"`
	Position GamePosition `json:"position"`
	Cannon   CannonInfo   `json:"cannon"`
	Status   string       `json:"status"`
	JoinTime time.Time    `json:"join_time"`
}

// FishInfo 魚類信息
type FishInfo struct {
	ID        int64        `json:"id"`
	Type      int32        `json:"type"`
	Position  GamePosition `json:"position"`
	Direction float64      `json:"direction"`
	Speed     float64      `json:"speed"`
	Health    int32        `json:"health"`
	MaxHealth int32        `json:"max_health"`
	Value     int64        `json:"value"`
	Status    string       `json:"status"`
	SpawnTime time.Time    `json:"spawn_time"`
}

// BulletInfo 子彈信息
type BulletInfo struct {
	ID        int64        `json:"id"`
	PlayerID  string       `json:"player_id"`
	Position  GamePosition `json:"position"`
	Direction float64      `json:"direction"`
	Speed     float64      `json:"speed"`
	Power     int32        `json:"power"`
	CreatedAt time.Time    `json:"created_at"`
}

// GamePosition 遊戲位置信息 (避免與 mock_protobuf.go 中的 Position 衝突)
type GamePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// CannonInfo 砲台信息
type CannonInfo struct {
	Type      int32   `json:"type"`      // 砲台類型
	Level     int32   `json:"level"`     // 砲台等級
	Power     int32   `json:"power"`     // 威力
	FireRate  float64 `json:"fire_rate"` // 射擊速度
	Direction float64 `json:"direction"` // 瞄準方向
}

// NewRoomManager 創建房間管理器
func NewRoomManager(roomID string, gameUsecase *game.GameUsecase, hub *Hub, logger logger.Logger) *RoomManager {
	ctx, cancel := context.WithCancel(context.Background())

	rm := &RoomManager{
		roomID:         roomID,
		clients:        make(map[*Client]bool),
		gameUsecase:    gameUsecase,
		hub:            hub,
		gameLoopTicker: time.NewTicker(1 * time.Second), // 1 FPS for testing
		gameLoopStop:   make(chan bool),
		tickDone:       make(chan bool, 1),
		addClient:      make(chan *Client),
		removeClient:   make(chan *Client),
		gameAction:     make(chan *GameActionMessage),
		gameState:      NewGameState(roomID),
		logger:         logger.With("component", "room_manager", "room_id", roomID),
		ctx:            ctx,
		cancel:         cancel,
	}
	rm.tickDone <- true // 預先填充信號，允許第一次 gameLoop 執行
	return rm
}

// NewGameState 創建新的遊戲狀態
func NewGameState(roomID string) *GameState {
	return &GameState{
		RoomID:        roomID,
		Status:        "waiting",
		Players:       make(map[string]*PlayerInfo),
		Fishes:        make(map[int64]*FishInfo),
		Bullets:       make(map[int64]*BulletInfo),
		LastUpdate:    time.Now(),
		GameStartTime: time.Now(),
	}
}

// Run 啟動房間管理器主循環
func (rm *RoomManager) Run() {
	rm.logger.Infof("Room manager started for room: %s, game state: %s", rm.roomID, rm.gameState.Status)

	// 添加 recover 機制防止房間管理器崩潰
	defer func() {
		if r := recover(); r != nil {
			rm.logger.Errorf("Room manager crashed with panic: %v", r)
			// 嘗試重新啟動房間管理器
			go rm.Run()
		}
	}()

	// 確保 ticker 已啟動
	rm.logger.Infof("Room manager main loop starting for room: %s", rm.roomID)
	rm.logger.Infof("Ticker interval: %v, ticker active: %v", rm.gameLoopTicker, rm.gameLoopTicker != nil)

	for {
		select {
		case <-rm.gameLoopTicker.C:
			select {
			case <-rm.tickDone: // 嘗試消耗上一次循環留下的「完成」信號
				// 如果成功消耗，代表上一個循環已結束，可以開始新的循環
				go func() {
					defer func() {
						if r := recover(); r != nil {
							rm.logger.Errorf("Recovered from panic in gameLoop: %v", r)
						}
						rm.tickDone <- true // 確保無論如何都會在結束時發回「完成」信號
					}()
					rm.gameLoop()
				}()
			default:
				// 如果無法消耗信號，代表上一個循環仍在運行中
				rm.logger.Warnf("Skipping game loop tick, previous one has not finished.")
			}
			
		case client := <-rm.addClient:
			rm.logger.Debugf("Handling add client for room: %s", rm.roomID)
			func() {
				defer func() {
					if r := recover(); r != nil {
						rm.logger.Errorf("Recovered from panic in handleAddClient: %v", r)
					}
				}()
				rm.handleAddClient(client)
			}()

		case client := <-rm.removeClient:
			rm.logger.Debugf("Handling remove client for room: %s", rm.roomID)
			func() {
				defer func() {
					if r := recover(); r != nil {
						rm.logger.Errorf("Recovered from panic in handleRemoveClient: %v", r)
					}
				}()
				rm.handleRemoveClient(client)
			}()

		case action := <-rm.gameAction:
			rm.logger.Debugf("Handling game action for room: %s", rm.roomID)
			func() {
				defer func() {
					if r := recover(); r != nil {
						rm.logger.Errorf("Recovered from panic in handleGameAction: %v", r)
					}
				}()
				rm.handleGameAction(action)
			}()

		case <-rm.gameLoopStop:
			rm.logger.Infof("Room manager stopping for room: %s", rm.roomID)
			return

		case <-rm.ctx.Done():
			rm.logger.Infof("Room manager context cancelled for room: %s", rm.roomID)
			return
		}
	}
}

// AddClient 添加客戶端到房間
func (rm *RoomManager) AddClient(client *Client) {
	rm.addClient <- client
}

// RemoveClient 從房間移除客戶端
func (rm *RoomManager) RemoveClient(client *Client) {
	rm.removeClient <- client
}

// HandleGameAction 處理遊戲操作
func (rm *RoomManager) HandleGameAction(action *GameActionMessage) {
	rm.gameAction <- action
}

// Stop 停止房間管理器
func (rm *RoomManager) Stop() {
	rm.gameLoopTicker.Stop()
	rm.gameLoopStop <- true
	rm.cancel()
}

// handleAddClient 處理添加客戶端
func (rm *RoomManager) handleAddClient(client *Client) {
	rm.clients[client] = true

	// 添加玩家到遊戲狀態
	playerInfo := &PlayerInfo{
		ID:       client.ID,
		PlayerID: client.PlayerID,
		Nickname: client.ID,                    // 暫時使用 client.ID 作為昵稱
		Balance:  10000,                        // 初始餘額
		Position: GamePosition{X: 100, Y: 700}, // 固定位置
		Cannon: CannonInfo{
			Type:      1,
			Level:     1,
			Power:     10,
			FireRate:  1.0,
			Direction: 0.0,
		},
		Status:   "playing",
		JoinTime: time.Now(),
	}

	rm.gameState.Players[client.ID] = playerInfo

	// 如果是第一個玩家且遊戲未開始，開始遊戲
	if len(rm.gameState.Players) == 1 && rm.gameState.Status == "waiting" {
		rm.startGame()
	}

	// 發送當前遊戲狀態給新玩家
	rm.sendGameStateToClient(client)

	rm.logger.Infof("Client %s added to room %s, total players: %d",
		client.ID, rm.roomID, len(rm.gameState.Players))
}

// handleRemoveClient 處理移除客戶端
func (rm *RoomManager) handleRemoveClient(client *Client) {
	if _, ok := rm.clients[client]; ok {
		delete(rm.clients, client)
		delete(rm.gameState.Players, client.ID)

		// 如果沒有玩家了，暫停遊戲
		if len(rm.gameState.Players) == 0 {
			rm.pauseGame()
		}

		rm.logger.Infof("Client %s removed from room %s, remaining players: %d",
			client.ID, rm.roomID, len(rm.gameState.Players))
	}
}

// handleGameAction 處理遊戲操作
func (rm *RoomManager) handleGameAction(action *GameActionMessage) {
	defer func() {
		if r := recover(); r != nil {
			rm.logger.Errorf("Recovered from panic in handleGameAction: %v", r)
			if action.Client != nil {
				action.Client.sendError("Error processing game action")
			}
		}
	}()

	if action == nil {
		rm.logger.Warnf("Received nil game action")
		return
	}

	if action.Client == nil {
		rm.logger.Warnf("Received game action with nil client")
		return
	}

	switch action.Action {
	case "fire_bullet":
		rm.handleFireBullet(action)
	case "switch_cannon":
		rm.handleSwitchCannon(action)
	default:
		rm.logger.Warnf("Unknown game action: %s", action.Action)
		action.Client.sendError(fmt.Sprintf("Unknown action: %s", action.Action))
	}
}

// handleFireBullet 處理開火操作
func (rm *RoomManager) handleFireBullet(action *GameActionMessage) {
	client := action.Client

	// 檢查玩家是否在房間中
	playerInfo, exists := rm.gameState.Players[client.ID]
	if !exists {
		client.sendError("Player not in game")
		return
	}

	// 解析消息
	gameMsg, ok := action.Data.(*pb.GameMessage)
	if !ok {
		client.sendError("Invalid message format")
		return
	}

	// 從消息中獲取開火參數
	fireData := gameMsg.GetFireBullet()
	direction := 0.0 // 默認方向
	power := playerInfo.Cannon.Power

	if fireData != nil {
		direction = fireData.Direction
		power = fireData.Power
	}

	// 調用業務邏輯層開火
	bullet, err := rm.gameUsecase.FireBullet(rm.ctx, rm.roomID, client.PlayerID, direction, power)
	if err != nil {
		rm.logger.Errorf("Failed to fire bullet: %v", err)
		client.sendError("Failed to fire bullet")
		return
	}

	// 添加子彈到遊戲狀態
	bulletInfo := &BulletInfo{
		ID:        bullet.ID,
		PlayerID:  client.ID,
		Position:  GamePosition{X: playerInfo.Position.X, Y: playerInfo.Position.Y},
		Direction: direction,
		Speed:     bullet.Speed,
		Power:     bullet.Power,
		CreatedAt: bullet.CreatedAt,
	}

	rm.gameState.Bullets[bullet.ID] = bulletInfo

	// 更新玩家餘額
	playerInfo.Balance -= bullet.Cost

	// 發送開火響應給客戶端
	fireResponse := &pb.GameMessage{
		Type: pb.MessageType_FIRE_BULLET_RESPONSE,
		Data: &pb.GameMessage_FireBulletResponse{
			FireBulletResponse: &pb.FireBulletResponse{
				Success:   true,
				BulletId:  bullet.ID,
				Cost:      bullet.Cost,
				Timestamp: time.Now().Unix(),
			},
		},
	}
	client.sendProtobuf(fireResponse)

	// 廣播開火事件給其他玩家
	fireEvent := &pb.GameMessage{
		Type: pb.MessageType_BULLET_FIRED,
		Data: &pb.GameMessage_BulletFired{
			BulletFired: &pb.BulletFiredEvent{
				PlayerId:  client.PlayerID,
				BulletId:  bullet.ID,
				Direction: direction,
				Power:     bullet.Power,
				Position:  &pb.Position{X: bulletInfo.Position.X, Y: bulletInfo.Position.Y},
				Timestamp: time.Now().Unix(),
			},
		},
	}

	// 序列化並廣播事件
	eventData, err := proto.Marshal(fireEvent)
	if err != nil {
		rm.logger.Errorf("Failed to marshal fire event: %v", err)
	} else {
		rm.hub.BroadcastToRoom(rm.roomID, eventData, client) // 排除發送者
	}

	rm.logger.Infof("Player %s fired bullet %d in room %s", client.ID, bullet.ID, rm.roomID)
}

// handleSwitchCannon 處理切換砲台操作
func (rm *RoomManager) handleSwitchCannon(action *GameActionMessage) {
	client := action.Client

	// 檢查玩家是否在房間中
	playerInfo, exists := rm.gameState.Players[client.ID]
	if !exists {
		client.sendError("Player not in game")
		return
	}

	// 解析 Protobuf 消息獲取砲台信息
	gameMsg, ok := action.Data.(*pb.GameMessage)
	if !ok {
		client.sendError("Invalid message format")
		return
	}

	switchData := gameMsg.GetSwitchCannon()
	newCannonType := int32(2) // 默認值
	newCannonLevel := int32(1)

	if switchData != nil {
		newCannonType = switchData.CannonType
		newCannonLevel = switchData.Level
	}

	// 更新砲台信息
	playerInfo.Cannon.Type = newCannonType
	playerInfo.Cannon.Level = newCannonLevel
	playerInfo.Cannon.Power = newCannonLevel * 10 // 根據等級計算威力

	// 發送切換砲台響應給客戶端
	switchResponse := &pb.GameMessage{
		Type: pb.MessageType_SWITCH_CANNON_RESPONSE,
		Data: &pb.GameMessage_SwitchCannonResponse{
			SwitchCannonResponse: &pb.SwitchCannonResponse{
				Success:     true,
				CannonType:  newCannonType,
				Level:       newCannonLevel,
				Power:       playerInfo.Cannon.Power,
				Timestamp:   time.Now().Unix(),
			},
		},
	}
	client.sendProtobuf(switchResponse)

	// 廣播砲台切換事件給其他玩家
	cannonEvent := &pb.GameMessage{
		Type: pb.MessageType_CANNON_SWITCHED,
		Data: &pb.GameMessage_CannonSwitched{
			CannonSwitched: &pb.CannonSwitchedEvent{
				PlayerId:    client.PlayerID,
				CannonType:  newCannonType,
				Level:       newCannonLevel,
				Power:       playerInfo.Cannon.Power,
				Timestamp:   time.Now().Unix(),
			},
		},
	}

	// 序列化並廣播事件
	eventData, err := proto.Marshal(cannonEvent)
	if err != nil {
		rm.logger.Errorf("Failed to marshal cannon switch event: %v", err)
	} else {
		rm.hub.BroadcastToRoom(rm.roomID, eventData, client) // 排除發送者
	}

	rm.logger.Infof("Player %s switched cannon to type %d level %d in room %s",
		client.ID, newCannonType, newCannonLevel, rm.roomID)
}

// gameLoop 遊戲主循環
func (rm *RoomManager) gameLoop() {
	if rm.gameState.Status != "playing" {
		// 記錄非運行狀態
		rm.logger.Warnf("Game loop called but status is '%s' in room %s", rm.gameState.Status, rm.roomID)
		return
	}
	
	// 記錄遊戲循環執行（移除時間條件以確保能看到）
	rm.logger.Infof("Game loop tick: %d fishes, %d bullets, %d players", 
		len(rm.gameState.Fishes), len(rm.gameState.Bullets), len(rm.gameState.Players))
	
	// 詳細記錄魚類和子彈狀態供前端調試
	if len(rm.gameState.Fishes) > 0 {
		for fishID, fish := range rm.gameState.Fishes {
			rm.logger.Debugf("Fish %d: type=%d, pos=(%.1f,%.1f), dir=%.2f, speed=%.1f, hp=%d/%d", 
				fishID, fish.Type, fish.Position.X, fish.Position.Y, fish.Direction, fish.Speed, fish.Health, fish.MaxHealth)
		}
	}
	
	if len(rm.gameState.Bullets) > 0 {
		for bulletID, bullet := range rm.gameState.Bullets {
			rm.logger.Debugf("Bullet %d: player=%s, pos=(%.1f,%.1f), dir=%.2f, speed=%.1f, power=%d", 
				bulletID, bullet.PlayerID, bullet.Position.X, bullet.Position.Y, bullet.Direction, bullet.Speed, bullet.Power)
		}
	}

	now := time.Now()
	deltaTime := now.Sub(rm.gameState.LastUpdate).Seconds()

	// 更新子彈位置
	rm.updateBullets(deltaTime)

	// 更新魚類位置
	rm.updateFishes(deltaTime)

	// 檢測碰撞
	rm.checkCollisions()

	// 生成新魚類
	rm.spawnFishes()

	// 清理過期對象
	rm.cleanupExpiredObjects()

	// 更新時間戳
	rm.gameState.LastUpdate = now

	// 定期廣播遊戲狀態或當狀態發生變化時廣播
	if now.Sub(rm.lastBroadcast) >= time.Second || len(rm.gameState.Fishes) != rm.lastFishCount || len(rm.gameState.Bullets) != rm.lastBulletCount {
		rm.broadcastGameStateProtobuf()
		rm.lastBroadcast = now
		rm.lastFishCount = len(rm.gameState.Fishes)
		rm.lastBulletCount = len(rm.gameState.Bullets)
	}
}

// updateBullets 更新子彈位置
func (rm *RoomManager) updateBullets(deltaTime float64) {
	for bulletID, bullet := range rm.gameState.Bullets {
		// 簡單的直線移動
		bullet.Position.X += bullet.Speed * deltaTime * 0.866 // cos(30°)
		bullet.Position.Y -= bullet.Speed * deltaTime * 0.5   // sin(30°)

		// 檢查是否出界
		if bullet.Position.Y < 0 || bullet.Position.X < 0 || bullet.Position.X > 1200 {
			delete(rm.gameState.Bullets, bulletID)
		}
	}
}

// updateFishes 更新魚類位置
func (rm *RoomManager) updateFishes(deltaTime float64) {
	// 更新現有魚類的位置
	for _, fish := range rm.gameState.Fishes {
		// 簡單的橫向移動
		fish.Position.X -= fish.Speed * deltaTime
		
		// 如果魚游出屏幕左側，則移除
		if fish.Position.X < -100 {
			delete(rm.gameState.Fishes, fish.ID)
			rm.logger.Debugf("Fish %d swam off screen and was removed", fish.ID)
		}
	}
}

// checkCollisions 檢測碰撞
func (rm *RoomManager) checkCollisions() {
	for bulletID, bullet := range rm.gameState.Bullets {
		for fishID, fish := range rm.gameState.Fishes {
			// 簡單的距離檢測
			dx := bullet.Position.X - fish.Position.X
			dy := bullet.Position.Y - fish.Position.Y
			distance := dx*dx + dy*dy

			// 碰撞半徑
			collisionRadius := 50.0 * 50.0

			if distance < collisionRadius {
				// 處理碰撞
				rm.handleCollision(bulletID, fishID)
				break
			}
		}
	}
}

// handleCollision 處理碰撞
func (rm *RoomManager) handleCollision(bulletID int64, fishID int64) {
	bullet, bulletExists := rm.gameState.Bullets[bulletID]
	fish, fishExists := rm.gameState.Fishes[fishID]

	if !bulletExists || !fishExists {
		return
	}

	// 調用業務邏輯處理命中
	hitResult, err := rm.gameUsecase.HitFish(rm.ctx, rm.roomID, bulletID, fishID)
	if err != nil {
		rm.logger.Errorf("Failed to process hit: %v", err)
		return
	}

	// 移除子彈
	delete(rm.gameState.Bullets, bulletID)

	if hitResult.Success {
		// 更新魚的血量或移除
		if hitResult.Damage >= fish.Health {
			delete(rm.gameState.Fishes, fishID)
		} else {
			fish.Health -= hitResult.Damage
		}

		// 更新玩家餘額
		if playerInfo, exists := rm.gameState.Players[bullet.PlayerID]; exists {
			playerInfo.Balance += hitResult.Reward
		}

		// 廣播命中事件
		hitEvent := map[string]interface{}{
			"type":        "fish_hit",
			"player_id":   bullet.PlayerID,
			"fish_id":     fishID,
			"bullet_id":   bulletID,
			"damage":      hitResult.Damage,
			"reward":      hitResult.Reward,
			"is_critical": hitResult.IsCritical,
			"timestamp":   time.Now().Unix(),
		}

		rm.broadcastToRoom(hitEvent, nil)

		rm.logger.Debugf("Fish %d hit by player %s, damage: %d, reward: %d",
			fishID, bullet.PlayerID, hitResult.Damage, hitResult.Reward)
	}
}

// spawnFishes 生成新魚類
func (rm *RoomManager) spawnFishes() {
	// 控制生成頻率
	if len(rm.gameState.Fishes) >= 20 {
		return
	}

	// 每5秒嘗試生成一條魚
	now := time.Now()
	if now.Sub(rm.lastFishSpawn) >= 5*time.Second {
		rm.lastFishSpawn = now
		
		// 創建模擬魚類
		fishID := now.UnixNano()
		fishInfo := &FishInfo{
			ID:        fishID,
			Type:      int32(1 + (fishID % 5)), // 魚類型 1-5
			Position:  GamePosition{X: 1200, Y: float64(100 + (fishID % 500))}, // 從右側進入
			Direction: 3.14, // 向左游
			Speed:     float64(50 + (fishID % 100)), // 速度 50-150
			Health:    int32(10 + (fishID % 90)), // 血量 10-100
			MaxHealth: int32(10 + (fishID % 90)),
			Value:     int64(100 + (fishID % 900)), // 價值 100-1000
			Status:    "alive",
			SpawnTime: now,
		}
		
		rm.gameState.Fishes[fishID] = fishInfo
		rm.logger.Infof("Spawned fish %d in room %s, total fishes: %d", fishID, rm.roomID, len(rm.gameState.Fishes))
		
		// 廣播魚類生成事件
		rm.broadcastFishSpawned(fishInfo)
	}
}

// cleanupExpiredObjects 清理過期對象
func (rm *RoomManager) cleanupExpiredObjects() {
	now := time.Now()

	// 清理超時的子彈（5秒）
	for bulletID, bullet := range rm.gameState.Bullets {
		if now.Sub(bullet.CreatedAt) > 5*time.Second {
			delete(rm.gameState.Bullets, bulletID)
		}
	}
}

// startGame 開始遊戲
func (rm *RoomManager) startGame() {
	rm.logger.Infof("Starting game in room: %s", rm.roomID)
	rm.gameState.Status = "playing"
	rm.gameState.GameStartTime = time.Now()

	// 初始化房間內的魚
	initialFishCount := 20
	for i := 0; i < initialFishCount; i++ {
		fishID := time.Now().UnixNano() + int64(i)
		fishInfo := &FishInfo{
			ID:        fishID,
			Type:      int32(1 + (fishID % 5)), // 魚類型 1-5
			Position:  GamePosition{X: float64(100 + (i * 60)), Y: float64(100 + (i%5)*80)}, // 分散初始位置
			Direction: 3.14, // 向左游
			Speed:     float64(50 + (fishID % 100)), // 速度 50-150
			Health:    int32(10 + (fishID % 90)), // 血量 10-100
			MaxHealth: int32(10 + (fishID % 90)),
			Value:     int64(100 + (fishID % 900)), // 價值 100-1000
			Status:    "alive",
			SpawnTime: time.Now(),
		}
		rm.gameState.Fishes[fishID] = fishInfo
	}
	rm.logger.Infof("Initialized room with %d fishes.", initialFishCount)

	rm.logger.Infof("Game state changed to 'playing' for room: %s", rm.roomID)
	rm.logger.Infof("Ticker should start working now. Status: %s", rm.gameState.Status)

	// 廣播遊戲開始事件
	startEvent := map[string]interface{}{
		"type":      "game_started",
		"room_id":   rm.roomID,
		"timestamp": time.Now().Unix(),
	}

	rm.broadcastToRoom(startEvent, nil)

	rm.logger.Infof("Game started in room: %s", rm.roomID)
}

// pauseGame 暫停遊戲
func (rm *RoomManager) pauseGame() {
	rm.gameState.Status = "waiting"

	// 廣播遊戲暫停事件
	pauseEvent := map[string]interface{}{
		"type":      "game_paused",
		"room_id":   rm.roomID,
		"timestamp": time.Now().Unix(),
	}

	rm.broadcastToRoom(pauseEvent, nil)

	rm.logger.Infof("Game paused in room: %s", rm.roomID)
}

// sendGameStateToClient 發送遊戲狀態給特定客戶端
func (rm *RoomManager) sendGameStateToClient(client *Client) {
	// 使用與廣播相同的 Protobuf 格式
	rm.sendGameStateProtobufToClient(client)
}

// sendGameStateProtobufToClient 使用 Protobuf 格式發送遊戲狀態給特定客戶端
func (rm *RoomManager) sendGameStateProtobufToClient(client *Client) {
	// 轉換魚類信息到 Protobuf 格式
	var fishInfos []*pb.FishInfo
	for _, fish := range rm.gameState.Fishes {
		fishInfos = append(fishInfos, &pb.FishInfo{
			FishId:    fish.ID,
			FishType:  fish.Type,
			Position:  &pb.Position{X: fish.Position.X, Y: fish.Position.Y},
			Direction: fish.Direction,
			Speed:     fish.Speed,
			Health:    fish.Health,
			MaxHealth: fish.MaxHealth,
			Value:     fish.Value,
			Status:    fish.Status,
			SpawnTime: fish.SpawnTime.Unix(),
		})
	}

	// 轉換子彈信息到 Protobuf 格式
	var bulletInfos []*pb.BulletInfo
	for _, bullet := range rm.gameState.Bullets {
		// 從玩家 ID 字符串獲取數字 ID
		var playerID int64
		if playerInfo, exists := rm.gameState.Players[bullet.PlayerID]; exists {
			playerID = playerInfo.PlayerID
		}
		
		bulletInfos = append(bulletInfos, &pb.BulletInfo{
			BulletId:  bullet.ID,
			PlayerId:  playerID,
			Position:  &pb.Position{X: bullet.Position.X, Y: bullet.Position.Y},
			Direction: bullet.Direction,
			Speed:     bullet.Speed,
			Power:     bullet.Power,
			CreatedAt: bullet.CreatedAt.Unix(),
		})
	}

	// 創建房間狀態更新消息
	roomStateUpdate := &pb.RoomStateUpdate{
		RoomId:       rm.roomID,
		Fishes:       fishInfos,
		Bullets:      bulletInfos,
		PlayerCount:  int32(len(rm.gameState.Players)),
		Timestamp:    time.Now().Unix(),
		RoomStatus:   rm.gameState.Status,
	}

	// 創建 GameMessage
	gameMessage := &pb.GameMessage{
		Type: pb.MessageType_ROOM_STATE_UPDATE,
		Data: &pb.GameMessage_RoomStateUpdate{
			RoomStateUpdate: roomStateUpdate,
		},
	}

	// 發送給特定客戶端
	client.sendProtobuf(gameMessage)
	rm.logger.Debugf("Sent room state update to client %s: %d fishes, %d bullets", 
		client.ID, len(fishInfos), len(bulletInfos))
}

// broadcastGameState 廣播遊戲狀態
func (rm *RoomManager) broadcastGameState() {
	stateMsg := map[string]interface{}{
		"type":       "game_state_update",
		"game_state": rm.gameState,
	}

	rm.broadcastToRoom(stateMsg, nil)
}

// broadcastToRoom 向房間廣播消息
func (rm *RoomManager) broadcastToRoom(message interface{}, exclude *Client) {
	data, err := json.Marshal(message)
	if err != nil {
		rm.logger.Errorf("Failed to marshal broadcast message: %v", err)
		return
	}

	rm.hub.BroadcastToRoom(rm.roomID, data, exclude)
}

// broadcastGameStateProtobuf 使用 Protobuf 格式廣播遊戲狀態
func (rm *RoomManager) broadcastGameStateProtobuf() {
	// 轉換魚類信息到 Protobuf 格式
	var fishInfos []*pb.FishInfo
	for _, fish := range rm.gameState.Fishes {
		fishInfos = append(fishInfos, &pb.FishInfo{
			FishId:    fish.ID,
			FishType:  fish.Type,
			Position:  &pb.Position{X: fish.Position.X, Y: fish.Position.Y},
			Direction: fish.Direction,
			Speed:     fish.Speed,
			Health:    fish.Health,
			MaxHealth: fish.MaxHealth,
			Value:     fish.Value,
			Status:    fish.Status,
			SpawnTime: fish.SpawnTime.Unix(),
		})
	}

	// 轉換子彈信息到 Protobuf 格式
	var bulletInfos []*pb.BulletInfo
	for _, bullet := range rm.gameState.Bullets {
		// 從玩家 ID 字符串獲取數字 ID
		var playerID int64
		if playerInfo, exists := rm.gameState.Players[bullet.PlayerID]; exists {
			playerID = playerInfo.PlayerID
		}
		
		bulletInfos = append(bulletInfos, &pb.BulletInfo{
			BulletId:  bullet.ID,
			PlayerId:  playerID,
			Position:  &pb.Position{X: bullet.Position.X, Y: bullet.Position.Y},
			Direction: bullet.Direction,
			Speed:     bullet.Speed,
			Power:     bullet.Power,
			CreatedAt: bullet.CreatedAt.Unix(),
		})
	}

	// 創建房間狀態更新消息
	roomStateUpdate := &pb.RoomStateUpdate{
		RoomId:       rm.roomID,
		Fishes:       fishInfos,
		Bullets:      bulletInfos,
		PlayerCount:  int32(len(rm.gameState.Players)),
		Timestamp:    time.Now().Unix(),
		RoomStatus:   rm.gameState.Status,
	}

	// 創建 GameMessage
	gameMessage := &pb.GameMessage{
		Type: pb.MessageType_ROOM_STATE_UPDATE,
		Data: &pb.GameMessage_RoomStateUpdate{
			RoomStateUpdate: roomStateUpdate,
		},
	}

	// 序列化並廣播
	data, err := proto.Marshal(gameMessage)
	if err != nil {
		rm.logger.Errorf("Failed to marshal room state update: %v", err)
		return
	}

	rm.hub.BroadcastToRoom(rm.roomID, data, nil)
	rm.logger.Infof("Broadcasted room state update: %d fishes, %d bullets to room %s", len(fishInfos), len(bulletInfos), rm.roomID)
	
	// 詳細記錄廣播的遊戲狀態，方便前端調試
	rm.logger.Debugf("Broadcast details - Room: %s, Status: %s, Players: %d", 
		roomStateUpdate.RoomId, roomStateUpdate.RoomStatus, roomStateUpdate.PlayerCount)
	
	if len(fishInfos) > 0 {
		rm.logger.Debugf("Broadcasting %d fishes with detailed positions for frontend rendering", len(fishInfos))
		for i, fish := range fishInfos {
			if i < 3 { // 只記錄前3條魚避免日誌過多
				rm.logger.Debugf("  Fish[%d]: ID=%d, Type=%d, Pos=(%.1f,%.1f), Speed=%.1f, HP=%d", 
					i, fish.FishId, fish.FishType, fish.Position.X, fish.Position.Y, fish.Speed, fish.Health)
			}
		}
		if len(fishInfos) > 3 {
			rm.logger.Debugf("  ... and %d more fishes", len(fishInfos)-3)
		}
	}
	
	if len(bulletInfos) > 0 {
		rm.logger.Debugf("Broadcasting %d bullets with trajectories for frontend rendering", len(bulletInfos))
		for i, bullet := range bulletInfos {
			if i < 3 { // 只記錄前3發子彈避免日誌過多
				rm.logger.Debugf("  Bullet[%d]: ID=%d, Player=%d, Pos=(%.1f,%.1f), Dir=%.2f", 
					i, bullet.BulletId, bullet.PlayerId, bullet.Position.X, bullet.Position.Y, bullet.Direction)
			}
		}
		if len(bulletInfos) > 3 {
			rm.logger.Debugf("  ... and %d more bullets", len(bulletInfos)-3)
		}
	}
}

// broadcastFishSpawned 廣播魚類生成事件
func (rm *RoomManager) broadcastFishSpawned(fish *FishInfo) {
	fishSpawnedEvent := &pb.FishSpawnedEvent{
		FishId:    fish.ID,
		FishType:  fish.Type,
		Position:  &pb.Position{X: fish.Position.X, Y: fish.Position.Y},
		Timestamp: fish.SpawnTime.Unix(),
	}

	gameMessage := &pb.GameMessage{
		Type: pb.MessageType_FISH_SPAWNED,
		Data: &pb.GameMessage_FishSpawned{
			FishSpawned: fishSpawnedEvent,
		},
	}

	data, err := proto.Marshal(gameMessage)
	if err != nil {
		rm.logger.Errorf("Failed to marshal fish spawned event: %v", err)
		return
	}

	rm.hub.BroadcastToRoom(rm.roomID, data, nil)
}
