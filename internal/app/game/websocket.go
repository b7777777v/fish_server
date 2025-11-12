package game

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/account"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/pkg/token"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

// ========================================
// WebSocket 連接管理
// ========================================

const (
	// WebSocket 配置
	writeWait      = 10 * time.Second    // 寫入超時
	pongWait       = 60 * time.Second    // Pong 超時
	pingPeriod     = (pongWait * 9) / 10 // Ping 間隔
	maxMessageSize = 512                 // 最大消息大小
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 在生產環境中應該檢查來源
		return true
	},
}

// Client 表示一個 WebSocket 客戶端連接
type Client struct {
	// WebSocket 連接
	conn *websocket.Conn

	// 客戶端信息
	ID       string `json:"id"`
	PlayerID int64  `json:"player_id"`
	RoomID   string `json:"room_id"`

	// 消息通道
	send chan []byte

	// Hub 引用
	hub *Hub

	// 日誌記錄器
	logger logger.Logger

	// 連接時間
	connectedAt time.Time

	// 最後活動時間
	lastActivity time.Time
}

// NewClient 創建新的客戶端
func NewClient(conn *websocket.Conn, hub *Hub, logger logger.Logger) *Client {
	return &Client{
		conn:         conn,
		send:         make(chan []byte, 256),
		hub:          hub,
		logger:       logger.With("component", "websocket_client"),
		connectedAt:  time.Now(),
		lastActivity: time.Now(),
	}
}

// sendProtobuf 將 Protobuf 消息序列化後發送到客戶端
func (c *Client) sendProtobuf(msg *pb.GameMessage) {
	bytes, err := proto.Marshal(msg)
	if err != nil {
		c.logger.Errorf("Failed to marshal protobuf message: %v", err)
		return
	}
	
	// 使用非阻塞發送避免阻塞房間管理器
	select {
	case c.send <- bytes:
		// 成功發送
	default:
		c.logger.Warnf("Client %s send channel full, dropping protobuf message", c.ID)
		// 嘗試清空一些舊消息
		select {
		case <-c.send:
			// 丟棄一個舊消息，然後重試
			select {
			case c.send <- bytes:
			default:
				c.logger.Errorf("Client %s send channel still full after cleanup", c.ID)
			}
		default:
		}
	}
}

// sendError 將錯誤消息發送到客戶端
func (c *Client) sendError(message string) {
	c.sendErrorPB(message)
}

// sendJSON 將 interface{} 序列化為 JSON 後發送到客戶端
func (c *Client) sendJSON(v interface{}) {
	bytes, err := json.Marshal(v)
	if err != nil {
		c.logger.Errorf("Failed to marshal JSON message: %v", err)
		c.sendError("Internal server error: could not serialize JSON response")
		return
	}
	
	// 使用非阻塞發送避免阻塞房間管理器
	select {
	case c.send <- bytes:
		// 成功發送
	default:
		c.logger.Warnf("Client %s send channel full, dropping message", c.ID)
		// 嘗試清空一些舊消息
		select {
		case <-c.send:
			// 丟棄一個舊消息，然後重試
			select {
			case c.send <- bytes:
			default:
				c.logger.Errorf("Client %s send channel still full after cleanup", c.ID)
			}
		default:
		}
	}
}

// WebSocketHandler WebSocket 升級處理器
type WebSocketHandler struct {
	hub            *Hub
	tokenHelper    *token.TokenHelper
	accountUsecase account.AccountUsecase
	logger         logger.Logger
}

// NewWebSocketHandler 創建 WebSocket 處理器
func NewWebSocketHandler(hub *Hub, tokenHelper *token.TokenHelper, accountUsecase account.AccountUsecase, logger logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:            hub,
		tokenHelper:    tokenHelper,
		accountUsecase: accountUsecase,
		logger:         logger.With("component", "websocket_handler"),
	}
}

// ServeWS 處理 WebSocket 升級和連接
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// 升級 HTTP 連接為 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	// 創建客戶端
	client := NewClient(conn, h.hub, h.logger)

	// 嘗試從 token 獲取用戶信息（支持遊客模式）
	var playerUsername string
	var userID int64

	// 1. 首先檢查是否有 token（從查詢參數或 Authorization header）
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		// 檢查 Authorization header
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 2. 如果有 token，解析並使用 token 中的用戶信息
	if tokenString != "" {
		claims, err := h.tokenHelper.ParseToken(tokenString)
		if err != nil {
			h.logger.Errorf("Invalid token: %v", err)
			conn.Close()
			return
		}

		userID = claims.UserID

		// 從 AccountUsecase 獲取用戶信息
		user, err := h.accountUsecase.GetUserByID(r.Context(), userID)
		if err != nil {
			h.logger.Errorf("Failed to get user %d: %v", userID, err)
			conn.Close()
			return
		}

		// 使用用戶的 nickname 作為玩家名稱
		playerUsername = user.Nickname

		// 根據 nickname 獲取或創建玩家
		_, err = h.hub.playerUsecase.GetOrCreateByUsername(r.Context(), playerUsername)
		if err != nil {
			h.logger.Errorf("Failed to get or create player for user %d: %v", userID, err)
			conn.Close()
			return
		}

		h.logger.Infof("WebSocket connection with token: userID=%d, nickname=%s, isGuest=%v",
			userID, playerUsername, claims.IsGuest)
	} else {
		// 3. 如果沒有 token，回退到舊的 player_id 模式（向後兼容）
		playerUsername = r.URL.Query().Get("player_id")
		if playerUsername == "" {
			h.logger.Error("WebSocket connection rejected: token or player_id is required")
			conn.Close()
			return
		}

		// 根據 player_id (username) 獲取或創建玩家
		player, err := h.hub.playerUsecase.GetOrCreateByUsername(r.Context(), playerUsername)
		if err != nil {
			h.logger.Errorf("Failed to get or create player %s: %v", playerUsername, err)
			conn.Close()
			return
		}

		userID = int64(player.ID)
		h.logger.Infof("WebSocket connection with player_id: player=%s", playerUsername)
	}

	// 設置客戶端信息
	client.ID = playerUsername
	client.PlayerID = userID
	client.RoomID = r.URL.Query().Get("room_id") // 可選的 room_id

	// 註冊客戶端到 Hub
	h.hub.register <- client

	h.logger.Infof("New WebSocket connection: player=%s, userID=%d, room=%s", client.ID, client.PlayerID, client.RoomID)

	// 啟動客戶端的讀寫 goroutines
	go client.writePump()
	go client.readPump()
}

// readPump 從 WebSocket 連接讀取消息
func (c *Client) readPump() {
	defer func() {
		if r := recover(); r != nil {
			c.logger.Errorf("Recovered from panic in readPump: %v", r)
		}
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// 設置讀取限制
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.lastActivity = time.Now()
		return nil
	})

	// 消息處理統計
	messageCount := 0
	errorCount := 0
	lastResetTime := time.Now()

	for {
		// 檢查錯誤率，如果太高則暫停處理
		if time.Since(lastResetTime) > time.Minute {
			if errorCount > 50 { // 每分鐘超過50個錯誤
				c.logger.Warnf("High error rate detected: %d errors in the last minute", errorCount)
				c.sendErrorPB("Too many errors, please check your message format")
				time.Sleep(5 * time.Second) // 暫停5秒
			}
			messageCount = 0
			errorCount = 0
			lastResetTime = time.Now()
		}

		// 讀取消息
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		messageCount++
		c.lastActivity = time.Now()

		// 處理不同類型的消息
		func() {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Errorf("Recovered from panic while processing message: %v", r)
					c.sendErrorPB("Error processing your message")
					errorCount++
				}
			}()

			switch messageType {
			case websocket.BinaryMessage:
				c.handleBinaryMessage(message)
			case websocket.TextMessage:
				c.logger.Warnf("Received text message, expected binary: %s", string(message))
				c.sendErrorPB("Text messages not supported, please use binary format")
				errorCount++
			default:
				c.logger.Warnf("Unknown message type: %d", messageType)
				c.sendErrorPB(fmt.Sprintf("Unsupported message type: %d", messageType))
				errorCount++
			}
		}()
	}
}

// writePump 向 WebSocket 連接寫入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.logger.Infof("[WRITEPUMP] WritePump stopped for client %s", c.ID)
	}()

	c.logger.Infof("[WRITEPUMP] WritePump started for client %s", c.ID)

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 關閉了通道
				c.logger.Warnf("[WRITEPUMP] Send channel closed for client %s", c.ID)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.logger.Infof("[WRITEPUMP] Received message from channel for client %s, size=%d bytes", c.ID, len(message))

			// 發送第一個消息
			if err := c.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				c.logger.Errorf("[WRITEPUMP] Failed to write message to WebSocket for client %s: %v", c.ID, err)
				return
			}
			c.logger.Infof("[WRITEPUMP] ✓ Successfully wrote message to WebSocket for client %s", c.ID)

			// 發送所有排隊的消息,每個作為獨立的 WebSocket 幀
			n := len(c.send)
			if n > 0 {
				c.logger.Infof("[WRITEPUMP] Processing %d queued messages for client %s", n, c.ID)
			}
			for i := 0; i < n; i++ {
				queuedMsg := <-c.send
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.conn.WriteMessage(websocket.BinaryMessage, queuedMsg); err != nil {
					c.logger.Errorf("[WRITEPUMP] Failed to write queued message %d to WebSocket for client %s: %v", i, c.ID, err)
					return
				}
			}
			if n > 0 {
				c.logger.Infof("[WRITEPUMP] ✓ Wrote %d queued messages to WebSocket for client %s", n, c.ID)
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.logger.Errorf("[WRITEPUMP] Failed to send ping to client %s: %v", c.ID, err)
				return
			}
		}
	}
}

// handleBinaryMessage 處理二進制消息（Protobuf 格式）
func (c *Client) handleBinaryMessage(message []byte) {
	// 添加 recover 機制防止 panic 導致整個連接崩潰
	defer func() {
		if r := recover(); r != nil {
			c.logger.Errorf("Recovered from panic in handleBinaryMessage: %v", r)
			c.sendErrorPB("Internal server error occurred while processing message")
		}
	}()

	// 基本消息大小檢查
	if len(message) == 0 {
		c.logger.Warnf("Received empty message")
		c.sendErrorPB("Empty message received")
		return
	}

	if len(message) > 1024*1024 { // 1MB 限制
		c.logger.Warnf("Received oversized message: %d bytes", len(message))
		c.sendErrorPB("Message too large")
		return
	}

	// 解析 Protobuf 消息
	var gameMsg pb.GameMessage
	if err := proto.Unmarshal(message, &gameMsg); err != nil {
		c.logger.Errorf("Failed to parse protobuf message: %v", err)
		c.sendErrorPB("Invalid message format")
		return
	}

	// 消息類型驗證
	if gameMsg.Type == pb.MessageType_INVALID {
		c.logger.Warnf("Received invalid message type")
		c.sendErrorPB("Invalid message type")
		return
	}

	// 添加消息處理超時機制
	done := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.logger.Errorf("Recovered from panic in message handler: %v", r)
				c.sendErrorPB("Error processing message")
			}
			done <- true
		}()

		c.handleMessageByType(&gameMsg)
	}()

	// 5秒超時
	select {
	case <-done:
		// 處理完成
	case <-time.After(5 * time.Second):
		c.logger.Errorf("Message processing timeout for type: %v", gameMsg.Type)
		c.sendErrorPB("Message processing timeout")
	}
}

// handleMessageByType 根據消息類型處理消息
func (c *Client) handleMessageByType(gameMsg *pb.GameMessage) {
	switch gameMsg.Type {
	case pb.MessageType_FIRE_BULLET:
		c.handleFireBullet(gameMsg)
	case pb.MessageType_SWITCH_CANNON:
		c.handleSwitchCannon(gameMsg)
	case pb.MessageType_JOIN_ROOM:
		c.handleJoinRoomPB(gameMsg)
	case pb.MessageType_LEAVE_ROOM:
		c.handleLeaveRoomPB(gameMsg)
	case pb.MessageType_GET_PLAYER_INFO:
		c.handleGetPlayerInfo(gameMsg)
	case pb.MessageType_GET_ROOM_LIST:
		c.handleGetRoomList(gameMsg)
	case pb.MessageType_HEARTBEAT:
		c.handleHeartbeat(gameMsg)
	// TODO: Uncomment after running `make proto` to regenerate protobuf code
	// case pb.MessageType_SELECT_SEAT:
	// 	c.handleSelectSeat(gameMsg)
	default:
		c.logger.Warnf("Unknown protobuf message type: %v", gameMsg.Type)
		c.sendErrorPB(fmt.Sprintf("Unsupported message type: %v", gameMsg.Type))
	}
}

// handleFireBullet 處理開火請求
func (c *Client) handleFireBullet(msg *pb.GameMessage) {
	if c.RoomID == "" {
		c.sendErrorPB("Not in any room")
		return
	}

	// 轉發到房間處理
	c.hub.gameAction <- &GameActionMessage{
		Client:    c,
		RoomID:    c.RoomID,
		Action:    "fire_bullet",
		Data:      msg,
		Timestamp: time.Now(),
	}
}

// handleSwitchCannon 處理切換砲台請求
func (c *Client) handleSwitchCannon(msg *pb.GameMessage) {
	if c.RoomID == "" {
		c.sendErrorPB("Not in any room")
		return
	}

	// 轉發到房間處理
	c.hub.gameAction <- &GameActionMessage{
		Client:    c,
		RoomID:    c.RoomID,
		Action:    "switch_cannon",
		Data:      msg,
		Timestamp: time.Now(),
	}
}

// handleJoinRoomPB 處理加入房間請求
func (c *Client) handleJoinRoomPB(msg *pb.GameMessage) {
	joinRoomMsg := msg.GetJoinRoom()
	if joinRoomMsg == nil {
		c.sendErrorPB("Invalid JoinRoom message")
		return
	}

	roomID := joinRoomMsg.GetRoomId()
	if roomID == "" {
		c.sendErrorPB("Room ID cannot be empty")
		return
	}

	c.hub.joinRoom <- &JoinRoomMessage{
		Client: c,
		RoomID: roomID,
	}
}

// handleLeaveRoomPB 處理離開房間請求
func (c *Client) handleLeaveRoomPB(msg *pb.GameMessage) {
	if c.RoomID == "" {
		c.sendErrorPB("Not in any room")
		return
	}

	c.hub.leaveRoom <- &LeaveRoomMessage{
		Client: c,
		RoomID: c.RoomID,
	}
}

// handleGetPlayerInfo 處理獲取玩家資訊請求
func (c *Client) handleGetPlayerInfo(msg *pb.GameMessage) {
	c.logger.Infof("Handling GetPlayerInfo request for player %s", c.ID)

	// 從房間管理器的遊戲狀態中獲取玩家數據
	var nickname string = "Player"
	var balance int64 = 0
	var level int32 = 1
	var exp int64 = 0

	// 獲取房間管理器
	c.hub.mu.RLock()
	roomManager, exists := c.hub.roomManagers[c.RoomID]
	c.hub.mu.RUnlock()

	if exists && roomManager != nil && roomManager.gameState != nil {
		// 從遊戲狀態中獲取玩家信息
		if playerInfo, ok := roomManager.gameState.Players[c.ID]; ok {
			nickname = playerInfo.Nickname
			balance = playerInfo.Balance
			// Level 和 Exp 暫時使用默認值，未來可以從玩家系統獲取
		}
	}

	playerInfo := &pb.PlayerInfoResponse{
		PlayerId:  c.PlayerID,
		Nickname:  nickname,
		Balance:   balance,
		Level:     level,
		Exp:       exp,
		RoomId:    c.RoomID,
		Timestamp: time.Now().Unix(),
	}

	responseMsg := &pb.GameMessage{
		Type: pb.MessageType_PLAYER_INFO_RESPONSE,
		Data: &pb.GameMessage_PlayerInfoResponse{
			PlayerInfoResponse: playerInfo,
		},
	}

	c.sendProtobuf(responseMsg)
}

func (c *Client) sendErrorPB(message string) {
	errorMsg := &pb.GameMessage{
		Type: pb.MessageType_ERROR,
		Data: &pb.GameMessage_Error{
			Error: &pb.ErrorMessage{
				Message:   message,
				Code:      "GENERAL_ERROR",
				Timestamp: time.Now().Unix(),
			},
		},
	}
	c.sendProtobuf(errorMsg)
}

// handleGetRoomList 處理獲取房間列表請求
func (c *Client) handleGetRoomList(msg *pb.GameMessage) {
	// 創建模擬房間列表
	roomList := []*pb.RoomInfo{
		{
			RoomId:      "101",
			Name:        "初級房間",
			Type:        "normal",
			PlayerCount: 2,
			MaxPlayers:  4,
			Status:      "active",
		},
		{
			RoomId:      "102",
			Name:        "中級房間",
			Type:        "medium",
			PlayerCount: 1,
			MaxPlayers:  4,
			Status:      "active",
		},
		{
			RoomId:      "103",
			Name:        "高級房間",
			Type:        "hard",
			PlayerCount: 0,
			MaxPlayers:  4,
			Status:      "active",
		},
	}

	responseMsg := &pb.GameMessage{
		Type: pb.MessageType_ROOM_LIST_RESPONSE,
		Data: &pb.GameMessage_RoomListResponse{
			RoomListResponse: &pb.RoomListResponse{
				Rooms:     roomList,
				Timestamp: time.Now().Unix(),
			},
		},
	}

	c.sendProtobuf(responseMsg)
}

// handleHeartbeat 處理心跳請求
func (c *Client) handleHeartbeat(msg *pb.GameMessage) {
	responseMsg := &pb.GameMessage{
		Type: pb.MessageType_HEARTBEAT_RESPONSE,
		Data: &pb.GameMessage_HeartbeatResponse{
			HeartbeatResponse: &pb.HeartbeatResponse{
				ServerTime: time.Now().Unix(),
				Timestamp:  time.Now().Unix(),
			},
		},
	}

	c.sendProtobuf(responseMsg)
}

// TODO: Uncomment after running `make proto` to regenerate protobuf code
// handleSelectSeat 處理選擇座位請求
// func (c *Client) handleSelectSeat(msg *pb.GameMessage) {
// 	if c.RoomID == "" {
// 		c.sendErrorPB("Not in any room")
// 		return
// 	}
//
// 	// 轉發到房間處理
// 	c.hub.gameAction <- &GameActionMessage{
// 		Client:    c,
// 		RoomID:    c.RoomID,
// 		Action:    "select_seat",
// 		Data:      msg,
// 		Timestamp: time.Now(),
// 	}
// }

// GetConnectionInfo 獲取連接信息
func (c *Client) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":            c.ID,
		"player_id":     c.PlayerID,
		"room_id":       c.RoomID,
		"connected_at":  c.connectedAt,
		"last_activity": c.lastActivity,
		"send_queue":    len(c.send),
	}
}
