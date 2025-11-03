package game

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
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
	playerID := r.URL.Query().Get("player_id")
	roomID := r.URL.Query().Get("room_id")
	
	if playerID != "" {
		// TODO: [Security] Implement JWT token validation.
		// The current implementation is simplified for development and accepts player_id directly from the query parameters,
		// which is insecure and should be replaced with a proper authentication mechanism.
		// The token should be passed via query param or header, then validated here to extract the player's identity.

		// 這裡應該驗證 JWT token 並解析玩家ID
		// 簡化實現直接使用查詢參數
		client.ID = playerID
		client.RoomID = roomID
	}
	
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
		case websocket.TextMessage:
			c.handleTextMessage(message)
		case websocket.BinaryMessage:
			c.handleBinaryMessage(message)
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

// handleTextMessage 處理文本消息（JSON 格式）
// TODO: [Refactor] This server handles both JSON (text) and Protobuf (binary) messages.
// The primary protocol should be Protobuf. This JSON handling logic might be for debugging or legacy purposes.
// Consider unifying the protocol to only support Protobuf to reduce complexity and maintenance overhead.
func (c *Client) handleTextMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		c.logger.Errorf("Failed to parse JSON message: %v", err)
		return
	}
	
	// 處理 JSON 格式的消息
	msgType, ok := msg["type"].(string)
	if !ok {
		c.logger.Error("Message missing type field")
		return
	}
	
	switch msgType {
	case "join_room":
		c.handleJoinRoom(msg)
	case "leave_room":
		c.handleLeaveRoom(msg)
	case "heartbeat":
		c.handleHeartbeat(msg)
	default:
		c.logger.Warnf("Unknown message type: %s", msgType)
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
	default:
		c.logger.Warnf("Unknown protobuf message type: %v", gameMsg.Type)
	}
}

// handleJoinRoom 處理加入房間請求（JSON）
func (c *Client) handleJoinRoom(msg map[string]interface{}) {
	roomID, ok := msg["room_id"].(string)
	if !ok {
		c.sendError("Invalid room_id")
		return
	}
	
	c.RoomID = roomID
	
	// 通知 Hub 客戶端要加入房間
	c.hub.joinRoom <- &JoinRoomMessage{
		Client: c,
		RoomID: roomID,
	}
}

// handleLeaveRoom 處理離開房間請求（JSON）
func (c *Client) handleLeaveRoom(msg map[string]interface{}) {
	if c.RoomID == "" {
		c.sendError("Not in any room")
		return
	}
	
	// 通知 Hub 客戶端要離開房間
	c.hub.leaveRoom <- &LeaveRoomMessage{
		Client: c,
		RoomID: c.RoomID,
	}
	
	c.RoomID = ""
}

// handleHeartbeat 處理心跳消息
func (c *Client) handleHeartbeat(msg map[string]interface{}) {
	response := map[string]interface{}{
		"type":      "heartbeat_response",
		"timestamp": time.Now().Unix(),
	}
	
	c.sendJSON(response)
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
	// TODO: [Implementation] The RoomID is hardcoded to "default".
	// It should be parsed from the protobuf message `msg.GetJoinRoom().GetRoomId()`.
	// Need to add nil checks for safety.

	// 從 Protobuf 消息中解析房間ID
	// 這裡需要根據實際的 Protobuf 定義來實現
	c.hub.joinRoom <- &JoinRoomMessage{
		Client: c,
		RoomID: "default", // 暫時使用默認值
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

// sendJSON 發送 JSON 消息
func (c *Client) sendJSON(data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		c.logger.Errorf("Failed to marshal JSON: %v", err)
		return
	}
	
	select {
	case c.send <- message:
	default:
		close(c.send)
	}
}

// sendProtobuf 發送 Protobuf 消息
func (c *Client) sendProtobuf(msg proto.Message) {
	data, err := proto.Marshal(msg)
	if err != nil {
		c.logger.Errorf("Failed to marshal protobuf: %v", err)
		return
	}
	
	select {
	case c.send <- data:
	default:
		close(c.send)
	}
}

// sendError 發送錯誤消息（JSON）
func (c *Client) sendError(message string) {
	errorMsg := map[string]interface{}{
		"type":    "error",
		"message": message,
	}
	c.sendJSON(errorMsg)
}

// sendErrorPB 發送錯誤消息
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