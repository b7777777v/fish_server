package game

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
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
	c.send <- bytes
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
	c.send <- bytes
}

// WebSocketHandler WebSocket 升級處理器
type WebSocketHandler struct {
	hub    *Hub
	logger logger.Logger
}

// NewWebSocketHandler 創建 WebSocket 處理器
func NewWebSocketHandler(hub *Hub, logger logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		hub:    hub,
		logger: logger.With("component", "websocket_handler"),
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

	// 從查詢參數獲取玩家信息
	playerUsername := r.URL.Query().Get("player_id")
	if playerUsername == "" {
		h.logger.Error("WebSocket connection rejected: player_id is required")
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

	// 設置客戶端信息
	client.ID = player.Username      // string ID
	client.PlayerID = int64(player.ID) // numeric ID
	client.RoomID = r.URL.Query().Get("room_id") // 可選的 room_id

	// 註冊客戶端到 Hub
	h.hub.register <- client

	h.logger.Infof("New WebSocket connection: player=%s, room=%s", client.ID, client.RoomID)

	// 啟動客戶端的讀寫 goroutines
	go client.writePump()
	go client.readPump()
}

// readPump 從 WebSocket 連接讀取消息
func (c *Client) readPump() {
	defer func() {
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

	for {
		// 讀取消息
		messageType, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		c.lastActivity = time.Now()

		// 處理不同類型的消息
		switch messageType {
		case websocket.BinaryMessage:
			c.handleBinaryMessage(message)
		default:
			c.logger.Warnf("Unknown message type: %d", messageType)
		}
	}
}

// writePump 向 WebSocket 連接寫入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 關閉了通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// 添加排隊的消息到當前寫入器
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleBinaryMessage 處理二進制消息（Protobuf 格式）
func (c *Client) handleBinaryMessage(message []byte) {
	// 解析 Protobuf 消息
	var gameMsg pb.GameMessage
	if err := proto.Unmarshal(message, &gameMsg); err != nil {
		c.logger.Errorf("Failed to parse protobuf message: %v", err)
		return
	}

	// 根據消息類型處理
	switch gameMsg.Type {
	case pb.MessageType_FIRE_BULLET:
		c.handleFireBullet(&gameMsg)
	case pb.MessageType_SWITCH_CANNON:
		c.handleSwitchCannon(&gameMsg)
	case pb.MessageType_JOIN_ROOM:
		c.handleJoinRoomPB(&gameMsg)
	case pb.MessageType_LEAVE_ROOM:
		c.handleLeaveRoomPB(&gameMsg)
	case pb.MessageType_GET_PLAYER_INFO:
		c.handleGetPlayerInfo(&gameMsg)
	case pb.MessageType_GET_ROOM_LIST:
		c.handleGetRoomList(&gameMsg)
	case pb.MessageType_HEARTBEAT:
		c.handleHeartbeat(&gameMsg)
	default:
		c.logger.Warnf("Unknown protobuf message type: %v", gameMsg.Type)
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

	// TODO: 實際應用中應從 usecase/service 獲取真實玩家數據
	// 這裡我們返回一個模擬的響應
	playerInfo := &pb.PlayerInfoResponse{
		PlayerId:  c.PlayerID, // 假設 c.PlayerID 已經在連接時被正確設置
		Nickname:  "MockPlayer",
		Balance:   10000,
		Level:     10,
		Exp:       500,
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
