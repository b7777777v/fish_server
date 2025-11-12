package game

import (
	"context"
	"sync"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
	"google.golang.org/protobuf/proto"
)

// ========================================
// Hub - 管理所有 WebSocket 連接和房間
// ========================================

// Hub 管理所有客戶端連接和房間
type Hub struct {
	// 註冊的客戶端
	clients map[*Client]bool

	// 按房間分組的客戶端
	rooms map[string]map[*Client]bool

	// 房間管理器
	roomManagers map[string]*RoomManager

	// 遊戲用例
	gameUsecase   *game.GameUsecase
	playerUsecase *player.PlayerUsecase

	// 通道
	register   chan *Client
	unregister chan *Client
	joinRoom   chan *JoinRoomMessage
	leaveRoom  chan *LeaveRoomMessage
	gameAction chan *GameActionMessage
	broadcast  chan *BroadcastMessage

	// 互斥鎖
	mu sync.RWMutex

	// 日誌記錄器
	logger logger.Logger

	// 統計信息
	stats *HubStats

	// 上下文和取消函數
	ctx    context.Context
	cancel context.CancelFunc
}

// HubStats Hub 統計信息
type HubStats struct {
	TotalConnections  int64     `json:"total_connections"`
	ActiveConnections int       `json:"active_connections"`
	ActiveRooms       int       `json:"active_rooms"`
	TotalMessages     int64     `json:"total_messages"`
	LastActivity      time.Time `json:"last_activity"`
	StartTime         time.Time `json:"start_time"`
}

// JoinRoomMessage 加入房間消息
type JoinRoomMessage struct {
	Client *Client
	RoomID string
}

// LeaveRoomMessage 離開房間消息
type LeaveRoomMessage struct {
	Client *Client
	RoomID string
}

// GameActionMessage 遊戲操作消息
type GameActionMessage struct {
	Client    *Client
	RoomID    string
	Action    string
	Data      interface{}
	Timestamp time.Time
}

// BroadcastMessage 廣播消息
type BroadcastMessage struct {
	RoomID  string // 空字符串表示全局廣播
	Message []byte
	Exclude *Client // 排除的客戶端
}

// NewHub 創建新的 Hub
func NewHub(gameUsecase *game.GameUsecase, playerUsecase *player.PlayerUsecase, logger logger.Logger) *Hub {
	ctx, cancel := context.WithCancel(context.Background())

	return &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		roomManagers:  make(map[string]*RoomManager),
		gameUsecase:   gameUsecase,
		playerUsecase: playerUsecase,
		register:      make(chan *Client, 10),             // 添加緩衝區避免阻塞
		unregister:    make(chan *Client, 10),             // 添加緩衝區避免阻塞
		joinRoom:      make(chan *JoinRoomMessage, 10),    // 添加緩衝區避免阻塞
		leaveRoom:     make(chan *LeaveRoomMessage, 10),   // 添加緩衝區避免阻塞
		gameAction:    make(chan *GameActionMessage, 100), // 添加緩衝區避免阻塞
		broadcast:     make(chan *BroadcastMessage, 100),  // 添加緩衝區避免阻塞
		logger:        logger.With("component", "hub"),
		stats: &HubStats{
			StartTime: time.Now(),
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run 啟動 Hub 主循環
func (h *Hub) Run() {
	h.logger.Info("Hub started")

	// 啟動統計更新定時器
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// 添加 recover 機制防止 Hub 崩潰
	defer func() {
		if r := recover(); r != nil {
			h.logger.Errorf("Hub crashed with panic: %v", r)
			// 重新啟動 Hub
			go h.Run()
		}
	}()

	iterationCount := 0
	for {
		iterationCount++
		if iterationCount%100 == 0 {
			h.logger.Infof("[HUB] Main loop iteration %d, checking channels...", iterationCount)
			h.logger.Infof("[HUB] Channel status - broadcast: %d/%d, gameAction: %d/%d",
				len(h.broadcast), cap(h.broadcast), len(h.gameAction), cap(h.gameAction))
		}

		func() {
			// 為每個操作添加 recover
			defer func() {
				if r := recover(); r != nil {
					h.logger.Errorf("Recovered from panic in Hub operation: %v", r)
				}
			}()

			select {
			case client := <-h.register:
				h.logger.Infof("[HUB] Processing register for client %s", client.ID)
				h.handleRegister(client)

			case client := <-h.unregister:
				h.logger.Infof("[HUB] Processing unregister for client %s", client.ID)
				h.handleUnregister(client)

			case msg := <-h.joinRoom:
				h.logger.Infof("[HUB] Processing joinRoom for client %s, room %s", msg.Client.ID, msg.RoomID)
				h.handleJoinRoom(msg)
				h.logger.Infof("[HUB] Completed joinRoom for client %s", msg.Client.ID)

			case msg := <-h.leaveRoom:
				h.logger.Infof("[HUB] Processing leaveRoom for client %s", msg.Client.ID)
				h.handleLeaveRoom(msg)

			case msg := <-h.gameAction:
				h.logger.Infof("[HUB] Processing gameAction: %s for client %s", msg.Action, msg.Client.ID)
				h.handleGameAction(msg)

			case msg := <-h.broadcast:
				h.logger.Infof("[HUB] Processing broadcast for room %s, size=%d", msg.RoomID, len(msg.Message))
				h.handleBroadcast(msg)
				h.logger.Infof("[HUB] Completed broadcast for room %s", msg.RoomID)

			case <-ticker.C:
				h.updateStats()

			case <-h.ctx.Done():
				h.logger.Info("Hub shutting down")
				return
			}
		}()
	}
}

// handleRegister 處理客戶端註冊
func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	h.stats.TotalConnections++
	h.stats.ActiveConnections = len(h.clients)
	h.stats.LastActivity = time.Now()

	h.logger.Infof("Client registered: %s (total: %d)", client.ID, len(h.clients))

	// 發送歡迎消息
	welcomeMsg := &pb.GameMessage{
		Type: pb.MessageType_WELCOME,
		Data: &pb.GameMessage_Welcome{
			Welcome: &pb.WelcomeMessage{
				ClientId:   client.ID,
				ServerTime: time.Now().Unix(),
			},
		},
	}
	client.sendProtobuf(welcomeMsg)
}

// handleUnregister 處理客戶端註銷
func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		// 從全局客戶端列表移除
		delete(h.clients, client)
		close(client.send)

		// 從房間移除
		if client.RoomID != "" {
			h.removeClientFromRoom(client, client.RoomID)
		}

		h.stats.ActiveConnections = len(h.clients)
		h.stats.LastActivity = time.Now()

		h.logger.Infof("Client unregistered: %s (total: %d)", client.ID, len(h.clients))
	}
}

// handleJoinRoom 處理加入房間
func (h *Hub) handleJoinRoom(msg *JoinRoomMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := msg.Client
	roomID := msg.RoomID

	// 如果客戶端已經在其他房間，先離開
	if client.RoomID != "" && client.RoomID != roomID {
		h.removeClientFromRoom(client, client.RoomID)
	}

	// 添加到新房間
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}
	h.rooms[roomID][client] = true
	client.RoomID = roomID

	// 確保房間管理器存在
	if h.roomManagers[roomID] == nil {
		roomManager := NewRoomManager(roomID, h.gameUsecase, h, h.logger)
		h.roomManagers[roomID] = roomManager
		go roomManager.Run()
	}

	h.stats.ActiveRooms = len(h.rooms)
	h.stats.LastActivity = time.Now()

	h.logger.Infof("Client %s joined room %s", client.ID, roomID)

	// 通知房間管理器
	h.roomManagers[roomID].AddClient(client)

	// 發送加入成功消息
	joinMsg := &pb.GameMessage{
		Type: pb.MessageType_JOIN_ROOM_RESPONSE,
		Data: &pb.GameMessage_JoinRoomResponse{
			JoinRoomResponse: &pb.JoinRoomResponse{
				Success:     true,  // 明確設置 success 為 true
				RoomId:      roomID,
				PlayerCount: int32(len(h.rooms[roomID])),
				Timestamp:   time.Now().Unix(),
			},
		},
	}
	client.sendProtobuf(joinMsg)

	// 通知房間其他玩家
	playerJoinMsg := &pb.GameMessage{
		Type: pb.MessageType_PLAYER_JOINED,
		Data: &pb.GameMessage_PlayerJoined{
			PlayerJoined: &pb.PlayerJoinedMessage{
				PlayerId: client.ID,
				RoomId:   roomID,
			},
		},
	}

	// 重要：在調用 broadcastToRoom 之前釋放鎖，避免死鎖
	// broadcastToRoom -> broadcastToRoomBytes 會嘗試獲取 h.mu.RLock()
	// 而當前已經持有 h.mu.Lock()，Go 的 RWMutex 不支持鎖降級，會導致死鎖
	h.mu.Unlock()
	h.broadcastToRoom(roomID, playerJoinMsg, client)
	h.mu.Lock() // 重新獲取鎖以確保 defer 能正常工作
}

// handleLeaveRoom 處理離開房間
func (h *Hub) handleLeaveRoom(msg *LeaveRoomMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client := msg.Client
	roomID := msg.RoomID

	h.removeClientFromRoom(client, roomID)

	h.logger.Infof("Client %s left room %s", client.ID, roomID)

	// 發送離開成功消息
	leaveMsg := &pb.GameMessage{
		Type: pb.MessageType_LEAVE_ROOM_RESPONSE,
		Data: &pb.GameMessage_LeaveRoomResponse{
			LeaveRoomResponse: &pb.LeaveRoomResponse{
				Success:   true,  // 明確設置 success 為 true
				RoomId:    roomID,
				Timestamp: time.Now().Unix(),
			},
		},
	}
	client.sendProtobuf(leaveMsg)
}

// removeClientFromRoom 從房間移除客戶端
// 注意：此函數假設調用者已經持有 h.mu.Lock()
func (h *Hub) removeClientFromRoom(client *Client, roomID string) {
	if room, ok := h.rooms[roomID]; ok {
		if _, ok := room[client]; ok {
			delete(room, client)

			// 通知房間管理器
			if roomManager, ok := h.roomManagers[roomID]; ok {
				roomManager.RemoveClient(client)
			}

			// 如果房間空了，清理房間
			if len(room) == 0 {
				delete(h.rooms, roomID)
				if roomManager, ok := h.roomManagers[roomID]; ok {
					roomManager.Stop()
					delete(h.roomManagers, roomID)
				}
			} else {
				// 通知房間其他玩家
				playerLeaveMsg := &pb.GameMessage{
					Type: pb.MessageType_PLAYER_LEFT,
					Data: &pb.GameMessage_PlayerLeft{
						PlayerLeft: &pb.PlayerLeftMessage{
							PlayerId: client.ID,
							RoomId:   roomID,
						},
					},
				}

				// 重要：臨時釋放鎖避免死鎖
				// broadcastToRoom -> broadcastToRoomBytes 需要 h.mu.RLock()
				h.mu.Unlock()
				h.broadcastToRoom(roomID, playerLeaveMsg, client)
				h.mu.Lock()
			}
		}
	}

	client.RoomID = ""
	h.stats.ActiveRooms = len(h.rooms)
}

// handleGameAction 處理遊戲操作
func (h *Hub) handleGameAction(msg *GameActionMessage) {
	h.stats.TotalMessages++
	h.stats.LastActivity = time.Now()

	// 轉發到對應的房間管理器
	if roomManager, ok := h.roomManagers[msg.RoomID]; ok {
		roomManager.HandleGameAction(msg)
	} else {
		h.logger.Warnf("Room manager not found for room: %s", msg.RoomID)
		msg.Client.sendError("Room not found")
	}
}

// handleBroadcast 處理廣播消息
func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
	h.logger.Infof("[HUB] handleBroadcast called: roomID=%s, size=%d", msg.RoomID, len(msg.Message))

	if msg.RoomID == "" {
		// 全局廣播
		h.logger.Infof("[HUB] Global broadcast to %d clients", len(h.clients))
		h.mu.RLock()
		for client := range h.clients {
			if client != msg.Exclude {
				select {
				case client.send <- msg.Message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
		h.mu.RUnlock()
	} else {
		// 房間廣播
		h.logger.Infof("[HUB] Room broadcast, calling broadcastToRoomBytes for room %s", msg.RoomID)
		h.broadcastToRoomBytes(msg.RoomID, msg.Message, msg.Exclude)
		h.logger.Infof("[HUB] Room broadcast completed for room %s", msg.RoomID)
	}
}

// broadcastToRoom 向房間廣播 Protobuf 消息
func (h *Hub) broadcastToRoom(roomID string, message *pb.GameMessage, exclude *Client) {
	bytes, err := proto.Marshal(message)
	if err != nil {
		h.logger.Errorf("Failed to marshal message for broadcast: %v", err)
		return
	}
	h.broadcastToRoomBytes(roomID, bytes, exclude)
}

// broadcastToRoomBytes 向房間廣播字節消息
func (h *Hub) broadcastToRoomBytes(roomID string, message []byte, exclude *Client) {
	h.logger.Infof("[BROADCAST] Starting broadcast: %d bytes to room %s", len(message), roomID)
	h.mu.RLock()
	defer h.mu.RUnlock()

	if room, ok := h.rooms[roomID]; ok {
		sentCount := 0
		totalClients := len(room)
		h.logger.Infof("[BROADCAST] Room %s has %d clients", roomID, totalClients)

		for client := range room {
			if client != exclude {
				// 檢查 send channel 的容量和當前使用量
				channelLen := len(client.send)
				channelCap := cap(client.send)
				h.logger.Infof("[BROADCAST] Attempting to send to client %s (channel: %d/%d)",
					client.ID, channelLen, channelCap)

				select {
				case client.send <- message:
					sentCount++
					h.logger.Infof("[BROADCAST] ✓ Successfully sent to client %s", client.ID)
				default:
					h.logger.Errorf("[BROADCAST] ✗ Failed to send to client %s, channel full (%d/%d)",
						client.ID, channelLen, channelCap)
					// 不要立即關閉和刪除客戶端，給它一次機會
					// close(client.send)
					// delete(room, client)
				}
			} else {
				h.logger.Infof("[BROADCAST] Skipping excluded client %s", client.ID)
			}
		}
		h.logger.Infof("[BROADCAST] Sent message to %d/%d clients in room %s", sentCount, totalClients, roomID)
	} else {
		h.logger.Errorf("[BROADCAST] Room %s not found for broadcast", roomID)
	}
}

// updateStats 更新統計信息
func (h *Hub) updateStats() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.stats.ActiveConnections = len(h.clients)
	h.stats.ActiveRooms = len(h.rooms)

	h.logger.Debugf("Hub stats: connections=%d, rooms=%d, messages=%d",
		h.stats.ActiveConnections, h.stats.ActiveRooms, h.stats.TotalMessages)
}

// GetStats 獲取 Hub 統計信息
func (h *Hub) GetStats() *HubStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	statsCopy := *h.stats
	statsCopy.ActiveConnections = len(h.clients)
	statsCopy.ActiveRooms = len(h.rooms)

	return &statsCopy
}

// GetRoomClients 獲取房間客戶端列表
func (h *Hub) GetRoomClients(roomID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var clients []*Client
	if room, ok := h.rooms[roomID]; ok {
		for client := range room {
			clients = append(clients, client)
		}
	}

	return clients
}

// BroadcastToRoom 向房間廣播消息（外部接口）
func (h *Hub) BroadcastToRoom(roomID string, message []byte, exclude *Client) {
	h.logger.Infof("[BROADCAST] BroadcastToRoom called: room=%s, messageSize=%d", roomID, len(message))

	broadcastMsg := &BroadcastMessage{
		RoomID:  roomID,
		Message: message,
		Exclude: exclude,
	}

	// 使用非阻塞發送避免阻塞房間管理器
	select {
	case h.broadcast <- broadcastMsg:
		h.logger.Infof("[BROADCAST] Message queued to broadcast channel for room %s", roomID)
	default:
		h.logger.Warnf("[BROADCAST] Broadcast channel full, using direct broadcast for room %s", roomID)
		// 如果緩衝區滿了，嘗試直接廣播（繞過 channel）
		h.broadcastToRoomBytes(roomID, message, exclude)
	}
}

// BroadcastGlobal 全局廣播消息
func (h *Hub) BroadcastGlobal(message []byte) {
	h.broadcast <- &BroadcastMessage{
		RoomID:  "",
		Message: message,
	}
}

// CloseRoom 關閉指定房間並清理資源
func (h *Hub) CloseRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 停止房間管理器
	if roomManager, exists := h.roomManagers[roomID]; exists {
		roomManager.Stop()
		delete(h.roomManagers, roomID)
	}

	// 移除房間中的所有客戶端（但不關閉連接）
	if room, exists := h.rooms[roomID]; exists {
		for client := range room {
			client.RoomID = ""
		}
		delete(h.rooms, roomID)
	}

	h.stats.ActiveRooms = len(h.rooms)
	h.logger.Infof("Room %s closed and cleaned up", roomID)
}

// Stop 停止 Hub
func (h *Hub) Stop() {
	h.logger.Info("Stopping Hub")

	// 停止所有房間管理器
	h.mu.Lock()
	for _, roomManager := range h.roomManagers {
		roomManager.Stop()
	}
	h.mu.Unlock()

	// 關閉所有客戶端連接
	for client := range h.clients {
		close(client.send)
	}

	h.cancel()
}
