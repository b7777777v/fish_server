package game

import (
	"context"
	"encoding/json"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
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
	
	// 客戶端操作通道
	addClient    chan *Client
	removeClient chan *Client
	gameAction   chan *GameActionMessage
	
	// 遊戲狀態
	gameState *GameState
	
	// 日誌記錄器
	logger logger.Logger
	
	// 上下文和取消函數
	ctx    context.Context
	cancel context.CancelFunc
}

// GameState 房間遊戲狀態
type GameState struct {
	RoomID       string                 `json:"room_id"`
	Status       string                 `json:"status"` // waiting, playing, paused
	Players      map[string]*PlayerInfo `json:"players"`
	Fishes       map[int64]*FishInfo    `json:"fishes"`
	Bullets      map[int64]*BulletInfo  `json:"bullets"`
	LastUpdate   time.Time              `json:"last_update"`
	GameStartTime time.Time             `json:"game_start_time"`
}

// PlayerInfo 玩家信息
type PlayerInfo struct {
	ID       string    `json:"id"`
	PlayerID int64     `json:"player_id"`
	Nickname string    `json:"nickname"`
	Balance  int64     `json:"balance"`
	Position GamePosition  `json:"position"`
	Cannon   CannonInfo `json:"cannon"`
	Status   string    `json:"status"`
	JoinTime time.Time `json:"join_time"`
}

// FishInfo 魚類信息
type FishInfo struct {
	ID        int64     `json:"id"`
	Type      int32     `json:"type"`
	Position  GamePosition  `json:"position"`
	Direction float64   `json:"direction"`
	Speed     float64   `json:"speed"`
	Health    int32     `json:"health"`
	MaxHealth int32     `json:"max_health"`
	Value     int64     `json:"value"`
	Status    string    `json:"status"`
	SpawnTime time.Time `json:"spawn_time"`
}

// BulletInfo 子彈信息
type BulletInfo struct {
	ID        int64     `json:"id"`
	PlayerID  string    `json:"player_id"`
	Position  GamePosition  `json:"position"`
	Direction float64   `json:"direction"`
	Speed     float64   `json:"speed"`
	Power     int32     `json:"power"`
	CreatedAt time.Time `json:"created_at"`
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
	
	return &RoomManager{
		roomID:         roomID,
		clients:        make(map[*Client]bool),
		gameUsecase:    gameUsecase,
		hub:            hub,
		gameLoopTicker: time.NewTicker(100 * time.Millisecond), // 10 FPS
		gameLoopStop:   make(chan bool),
		addClient:      make(chan *Client),
		removeClient:   make(chan *Client),
		gameAction:     make(chan *GameActionMessage),
		gameState:      NewGameState(roomID),
		logger:         logger.With("component", "room_manager", "room_id", roomID),
		ctx:            ctx,
		cancel:         cancel,
	}
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
	rm.logger.Infof("Room manager started for room: %s", rm.roomID)
	
	for {
		select {
		case client := <-rm.addClient:
			rm.handleAddClient(client)
			
		case client := <-rm.removeClient:
			rm.handleRemoveClient(client)
			
		case action := <-rm.gameAction:
			rm.handleGameAction(action)
			
		case <-rm.gameLoopTicker.C:
			rm.gameLoop()
			
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
		Nickname: client.ID, // 暫時使用 client.ID 作為昵稱
		Balance:  10000,     // 初始餘額
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
	switch action.Action {
	case "fire_bullet":
		rm.handleFireBullet(action)
	case "switch_cannon":
		rm.handleSwitchCannon(action)
	default:
		rm.logger.Warnf("Unknown game action: %s", action.Action)
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
	
	// 廣播開火事件
	fireEvent := map[string]interface{}{
		"type":      "bullet_fired",
		"player_id": client.ID,
		"bullet":    bulletInfo,
		"timestamp": time.Now().Unix(),
	}
	
	rm.broadcastToRoom(fireEvent, nil)
	
	rm.logger.Debugf("Player %s fired bullet %d in room %s", client.ID, bullet.ID, rm.roomID)
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
	// 這裡應該從消息中解析新的砲台類型和等級
	newCannonType := int32(2) // 示例值
	newCannonLevel := int32(1)
	
	// 更新砲台信息
	playerInfo.Cannon.Type = newCannonType
	playerInfo.Cannon.Level = newCannonLevel
	playerInfo.Cannon.Power = newCannonLevel * 10 // 根據等級計算威力
	
	// 廣播砲台切換事件
	cannonEvent := map[string]interface{}{
		"type":      "cannon_switched",
		"player_id": client.ID,
		"cannon":    playerInfo.Cannon,
		"timestamp": time.Now().Unix(),
	}
	
	rm.broadcastToRoom(cannonEvent, nil)
	
	rm.logger.Debugf("Player %s switched cannon to type %d level %d in room %s", 
		client.ID, newCannonType, newCannonLevel, rm.roomID)
}

// gameLoop 遊戲主循環
func (rm *RoomManager) gameLoop() {
	if rm.gameState.Status != "playing" {
		return
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
	
	// 每秒廣播一次遊戲狀態
	if int(now.Unix())%1 == 0 {
		rm.broadcastGameState()
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
	// 從業務邏輯層獲取房間狀態
	roomState, err := rm.gameUsecase.GetRoomState(rm.ctx, rm.roomID)
	if err != nil {
		return
	}
	
	// 同步魚類狀態
	for _, fish := range roomState.Fishes {
		fishInfo := &FishInfo{
			ID:        fish.ID,
			Type:      fish.Type.ID,
			Position:  GamePosition{X: fish.Position.X, Y: fish.Position.Y},
			Direction: fish.Direction,
			Speed:     fish.Speed,
			Health:    fish.Health,
			MaxHealth: fish.MaxHealth,
			Value:     fish.Value,
			Status:    string(fish.Status),
			SpawnTime: fish.SpawnTime,
		}
		rm.gameState.Fishes[fish.ID] = fishInfo
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
	if time.Now().Unix()%5 == 0 {
		// 這裡應該調用業務邏輯層生成魚類
		// 暫時跳過實現
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
	rm.gameState.Status = "playing"
	rm.gameState.GameStartTime = time.Now()
	
	// 創建業務邏輯層的房間
	_, err := rm.gameUsecase.CreateRoom(rm.ctx, game.RoomTypeNovice, 4)
	if err != nil {
		rm.logger.Errorf("Failed to create game room: %v", err)
		return
	}
	
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
	stateMsg := map[string]interface{}{
		"type":       "game_state",
		"game_state": rm.gameState,
	}
	
	client.sendJSON(stateMsg)
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