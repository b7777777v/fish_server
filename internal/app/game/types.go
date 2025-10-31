package game

import (
	"time"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
)

// ========================================
// 遊戲相關類型定義
// ========================================

// GameConfig 遊戲配置
type GameConfig struct {
	// WebSocket 配置
	MaxConnections    int           `json:"max_connections"`
	PingInterval      time.Duration `json:"ping_interval"`
	PongTimeout       time.Duration `json:"pong_timeout"`
	WriteTimeout      time.Duration `json:"write_timeout"`
	ReadTimeout       time.Duration `json:"read_timeout"`
	MaxMessageSize    int64         `json:"max_message_size"`
	
	// 房間配置
	MaxRooms          int           `json:"max_rooms"`
	MaxPlayersPerRoom int           `json:"max_players_per_room"`
	RoomIdleTimeout   time.Duration `json:"room_idle_timeout"`
	
	// 遊戲循環配置
	GameLoopFPS       int           `json:"game_loop_fps"`
	StateUpdateFPS    int           `json:"state_update_fps"`
	
	// 性能配置
	MessageQueueSize  int           `json:"message_queue_size"`
	BroadcastBuffer   int           `json:"broadcast_buffer"`
}

// DefaultGameConfig 獲取默認遊戲配置
func DefaultGameConfig() *GameConfig {
	return &GameConfig{
		MaxConnections:    1000,
		PingInterval:      54 * time.Second,
		PongTimeout:       60 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadTimeout:       60 * time.Second,
		MaxMessageSize:    512,
		MaxRooms:          100,
		MaxPlayersPerRoom: 4,
		RoomIdleTimeout:   5 * time.Minute,
		GameLoopFPS:       10,
		StateUpdateFPS:    1,
		MessageQueueSize:  256,
		BroadcastBuffer:   512,
	}
}

// ClientState 客戶端狀態
type ClientState string

const (
	ClientStateConnected    ClientState = "connected"    // 已連接
	ClientStateAuthenticated ClientState = "authenticated" // 已認證
	ClientStateInRoom       ClientState = "in_room"      // 在房間中
	ClientStatePlaying      ClientState = "playing"      // 遊戲中
	ClientStateDisconnected ClientState = "disconnected" // 已斷線
)

// RoomState 房間狀態
type RoomState string

const (
	RoomStateWaiting RoomState = "waiting" // 等待玩家
	RoomStatePlaying RoomState = "playing" // 遊戲中
	RoomStatePaused  RoomState = "paused"  // 暫停
	RoomStateClosed  RoomState = "closed"  // 已關閉
)

// EventType 事件類型
type EventType string

const (
	EventTypePlayerJoin    EventType = "player_join"
	EventTypePlayerLeave   EventType = "player_leave"
	EventTypeGameStart     EventType = "game_start"
	EventTypeGameEnd       EventType = "game_end"
	EventTypeBulletFired   EventType = "bullet_fired"
	EventTypeFishHit       EventType = "fish_hit"
	EventTypeFishSpawned   EventType = "fish_spawned"
	EventTypeFishDied      EventType = "fish_died"
	EventTypePlayerReward  EventType = "player_reward"
	EventTypeCannonSwitch  EventType = "cannon_switch"
	EventTypeRoomUpdate    EventType = "room_update"
	EventTypeError         EventType = "error"
)

// WebSocketMessage WebSocket 消息包裝
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	MessageID string      `json:"message_id,omitempty"`
}

// ErrorResponse 錯誤響應
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// AuthInfo 認證信息
type AuthInfo struct {
	PlayerID  int64  `json:"player_id"`
	Token     string `json:"token"`
	Nickname  string `json:"nickname"`
	Level     int32  `json:"level"`
	Balance   int64  `json:"balance"`
}

// RoomInfo 房間信息
type RoomInfoResponse struct {
	RoomID      string                 `json:"room_id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Status      RoomState              `json:"status"`
	PlayerCount int                    `json:"player_count"`
	MaxPlayers  int                    `json:"max_players"`
	Players     map[string]*PlayerInfo `json:"players"`
	CreatedAt   time.Time              `json:"created_at"`
}

// GameStats 遊戲統計
type GameStats struct {
	// 連接統計
	TotalConnections  int64 `json:"total_connections"`
	ActiveConnections int   `json:"active_connections"`
	PeakConnections   int   `json:"peak_connections"`
	
	// 房間統計
	TotalRooms       int64 `json:"total_rooms"`
	ActiveRooms      int   `json:"active_rooms"`
	PlayingRooms     int   `json:"playing_rooms"`
	
	// 消息統計
	TotalMessages    int64 `json:"total_messages"`
	MessagesPerSec   int   `json:"messages_per_sec"`
	
	// 遊戲統計
	TotalGames       int64 `json:"total_games"`
	ActiveGames      int   `json:"active_games"`
	TotalBulletsFired int64 `json:"total_bullets_fired"`
	TotalFishCaught  int64 `json:"total_fish_caught"`
	
	// 性能統計
	AvgLatency       time.Duration `json:"avg_latency"`
	ServerUptime     time.Duration `json:"server_uptime"`
	LastUpdate       time.Time     `json:"last_update"`
}

// ProtobufMessageMap 消息類型映射
var ProtobufMessageMap = map[pb.MessageType]string{
	pb.MessageType_FIRE_BULLET:           "fire_bullet",
	pb.MessageType_SWITCH_CANNON:         "switch_cannon",
	pb.MessageType_JOIN_ROOM:             "join_room",
	pb.MessageType_LEAVE_ROOM:            "leave_room",
	pb.MessageType_HEARTBEAT:             "heartbeat",
	pb.MessageType_GET_ROOM_LIST:         "get_room_list",
	pb.MessageType_GET_PLAYER_INFO:       "get_player_info",
	pb.MessageType_FIRE_BULLET_RESPONSE:  "fire_bullet_response",
	pb.MessageType_SWITCH_CANNON_RESPONSE: "switch_cannon_response",
	pb.MessageType_JOIN_ROOM_RESPONSE:    "join_room_response",
	pb.MessageType_LEAVE_ROOM_RESPONSE:   "leave_room_response",
	pb.MessageType_HEARTBEAT_RESPONSE:    "heartbeat_response",
	pb.MessageType_ROOM_LIST_RESPONSE:    "room_list_response",
	pb.MessageType_PLAYER_INFO_RESPONSE:  "player_info_response",
	pb.MessageType_BULLET_FIRED:          "bullet_fired",
	pb.MessageType_CANNON_SWITCHED:       "cannon_switched",
	pb.MessageType_FISH_SPAWNED:          "fish_spawned",
	pb.MessageType_FISH_DIED:             "fish_died",
	pb.MessageType_PLAYER_REWARD:         "player_reward",
	pb.MessageType_ERROR:                 "error",
}

// GetMessageTypeName 獲取消息類型名稱
func GetMessageTypeName(msgType pb.MessageType) string {
	if name, exists := ProtobufMessageMap[msgType]; exists {
		return name
	}
	return "unknown"
}

// ValidationError 驗證錯誤
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// GameError 遊戲錯誤
type GameError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// 常見遊戲錯誤
var (
	ErrPlayerNotFound     = &GameError{Code: "PLAYER_NOT_FOUND", Message: "Player not found", Type: "CLIENT_ERROR"}
	ErrRoomNotFound       = &GameError{Code: "ROOM_NOT_FOUND", Message: "Room not found", Type: "CLIENT_ERROR"}
	ErrRoomFull           = &GameError{Code: "ROOM_FULL", Message: "Room is full", Type: "CLIENT_ERROR"}
	ErrInsufficientFunds  = &GameError{Code: "INSUFFICIENT_FUNDS", Message: "Insufficient balance", Type: "CLIENT_ERROR"}
	ErrInvalidParameters  = &GameError{Code: "INVALID_PARAMETERS", Message: "Invalid parameters", Type: "CLIENT_ERROR"}
	ErrNotInRoom          = &GameError{Code: "NOT_IN_ROOM", Message: "Player not in any room", Type: "CLIENT_ERROR"}
	ErrGameNotStarted     = &GameError{Code: "GAME_NOT_STARTED", Message: "Game has not started", Type: "CLIENT_ERROR"}
	ErrServerError        = &GameError{Code: "SERVER_ERROR", Message: "Internal server error", Type: "SERVER_ERROR"}
	ErrConnectionClosed   = &GameError{Code: "CONNECTION_CLOSED", Message: "Connection closed", Type: "CONNECTION_ERROR"}
	ErrMessageTooLarge    = &GameError{Code: "MESSAGE_TOO_LARGE", Message: "Message too large", Type: "PROTOCOL_ERROR"}
	ErrInvalidMessage     = &GameError{Code: "INVALID_MESSAGE", Message: "Invalid message format", Type: "PROTOCOL_ERROR"}
)