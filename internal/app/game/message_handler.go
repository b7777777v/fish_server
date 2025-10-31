package game

import (
	"context"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
	"google.golang.org/protobuf/proto"
)

// ========================================
// MessageHandler - 完整的 Protobuf 消息處理器
// ========================================

// MessageHandler Protobuf 消息處理器
type MessageHandler struct {
	gameUsecase *game.GameUsecase
	hub         *Hub
	logger      logger.Logger
}

// NewMessageHandler 創建 Protobuf 消息處理器
func NewMessageHandler(gameUsecase *game.GameUsecase, hub *Hub, logger logger.Logger) *MessageHandler {
	return &MessageHandler{
		gameUsecase: gameUsecase,
		hub:         hub,
		logger:      logger.With("component", "message_handler"),
	}
}

// HandleMessage 處理 Protobuf 消息的主入口
func (mh *MessageHandler) HandleMessage(client *Client, message *pb.GameMessage) {
	mh.logger.Debugf("Handling message type: %v from client: %s", message.Type, client.ID)
	
	// 更新客戶端活動時間
	client.lastActivity = time.Now()
	
	// 根據消息類型路由到具體處理器
	switch message.Type {
	case pb.MessageType_FIRE_BULLET:
		mh.handleFireBullet(client, message)
	case pb.MessageType_SWITCH_CANNON:
		mh.handleSwitchCannon(client, message)
	case pb.MessageType_JOIN_ROOM:
		mh.handleJoinRoom(client, message)
	case pb.MessageType_LEAVE_ROOM:
		mh.handleLeaveRoom(client, message)
	case pb.MessageType_HEARTBEAT:
		mh.handleHeartbeat(client, message)
	case pb.MessageType_GET_ROOM_LIST:
		mh.handleGetRoomList(client, message)
	case pb.MessageType_GET_PLAYER_INFO:
		mh.handleGetPlayerInfo(client, message)
	default:
		mh.logger.Warnf("Unknown message type: %v from client: %s", message.Type, client.ID)
		mh.sendErrorResponse(client, "Unknown message type")
	}
}

// handleFireBullet 處理開火消息
func (mh *MessageHandler) handleFireBullet(client *Client, message *pb.GameMessage) {
	if client.RoomID == "" {
		mh.sendErrorResponse(client, "Not in any room")
		return
	}
	
	// 解析開火數據
	fireData := message.GetFireBullet()
	if fireData == nil {
		mh.sendErrorResponse(client, "Invalid fire bullet data")
		return
	}
	
	// 驗證參數
	if fireData.Power < 1 || fireData.Power > 100 {
		mh.sendErrorResponse(client, "Invalid bullet power")
		return
	}
	
	// 調用業務邏輯
	ctx := context.Background()
	bullet, err := mh.gameUsecase.FireBullet(ctx, client.RoomID, client.PlayerID, 
		fireData.Direction, fireData.Power)
	if err != nil {
		mh.logger.Errorf("Failed to fire bullet: %v", err)
		mh.sendErrorResponse(client, "Failed to fire bullet")
		return
	}
	
	// 構建響應消息
	response := &pb.GameMessage{
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
	
	// 發送響應給客戶端
	client.sendProtobuf(response)
	
	// 廣播給房間其他玩家
	broadcastMsg := &pb.GameMessage{
		Type: pb.MessageType_BULLET_FIRED,
		Data: &pb.GameMessage_BulletFired{
			BulletFired: &pb.BulletFiredEvent{
				PlayerId:  client.PlayerID,
				BulletId:  bullet.ID,
				Direction: fireData.Direction,
				Power:     fireData.Power,
				Position: &pb.Position{
					X: fireData.Position.X,
					Y: fireData.Position.Y,
				},
				Timestamp: time.Now().Unix(),
			},
		},
	}
	
	mh.broadcastToRoom(client.RoomID, broadcastMsg, client)
	
	mh.logger.Debugf("Player %d fired bullet %d in room %s", 
		client.PlayerID, bullet.ID, client.RoomID)
}

// handleSwitchCannon 處理切換砲台消息
func (mh *MessageHandler) handleSwitchCannon(client *Client, message *pb.GameMessage) {
	if client.RoomID == "" {
		mh.sendErrorResponse(client, "Not in any room")
		return
	}
	
	// 解析砲台數據
	cannonData := message.GetSwitchCannon()
	if cannonData == nil {
		mh.sendErrorResponse(client, "Invalid cannon data")
		return
	}
	
	// 驗證砲台類型和等級
	if cannonData.CannonType < 1 || cannonData.CannonType > 10 {
		mh.sendErrorResponse(client, "Invalid cannon type")
		return
	}
	
	if cannonData.Level < 1 || cannonData.Level > 10 {
		mh.sendErrorResponse(client, "Invalid cannon level")
		return
	}
	
	// 計算砲台威力
	power := cannonData.Level * 10
	
	// 構建響應消息
	response := &pb.GameMessage{
		Type: pb.MessageType_SWITCH_CANNON_RESPONSE,
		Data: &pb.GameMessage_SwitchCannonResponse{
			SwitchCannonResponse: &pb.SwitchCannonResponse{
				Success:     true,
				CannonType:  cannonData.CannonType,
				Level:       cannonData.Level,
				Power:       power,
				Timestamp:   time.Now().Unix(),
			},
		},
	}
	
	// 發送響應給客戶端
	client.sendProtobuf(response)
	
	// 廣播給房間其他玩家
	broadcastMsg := &pb.GameMessage{
		Type: pb.MessageType_CANNON_SWITCHED,
		Data: &pb.GameMessage_CannonSwitched{
			CannonSwitched: &pb.CannonSwitchedEvent{
				PlayerId:    client.PlayerID,
				CannonType:  cannonData.CannonType,
				Level:       cannonData.Level,
				Power:       power,
				Timestamp:   time.Now().Unix(),
			},
		},
	}
	
	mh.broadcastToRoom(client.RoomID, broadcastMsg, client)
	
	mh.logger.Debugf("Player %d switched cannon to type %d level %d in room %s", 
		client.PlayerID, cannonData.CannonType, cannonData.Level, client.RoomID)
}

// handleJoinRoom 處理加入房間消息
func (mh *MessageHandler) handleJoinRoom(client *Client, message *pb.GameMessage) {
	joinData := message.GetJoinRoom()
	if joinData == nil {
		mh.sendErrorResponse(client, "Invalid join room data")
		return
	}
	
	roomID := joinData.RoomId
	if roomID == "" {
		mh.sendErrorResponse(client, "Room ID is required")
		return
	}
	
	// 調用業務邏輯
	ctx := context.Background()
	err := mh.gameUsecase.JoinRoom(ctx, roomID, client.PlayerID)
	if err != nil {
		mh.logger.Errorf("Failed to join room: %v", err)
		mh.sendErrorResponse(client, "Failed to join room")
		return
	}
	
	// 更新客戶端房間ID
	client.RoomID = roomID
	
	// 通知 Hub
	mh.hub.joinRoom <- &JoinRoomMessage{
		Client: client,
		RoomID: roomID,
	}
	
	// 發送響應
	response := &pb.GameMessage{
		Type: pb.MessageType_JOIN_ROOM_RESPONSE,
		Data: &pb.GameMessage_JoinRoomResponse{
			JoinRoomResponse: &pb.JoinRoomResponse{
				Success:   true,
				RoomId:    roomID,
				Timestamp: time.Now().Unix(),
			},
		},
	}
	
	client.sendProtobuf(response)
	
	mh.logger.Infof("Player %d joined room %s", client.PlayerID, roomID)
}

// handleLeaveRoom 處理離開房間消息
func (mh *MessageHandler) handleLeaveRoom(client *Client, message *pb.GameMessage) {
	if client.RoomID == "" {
		mh.sendErrorResponse(client, "Not in any room")
		return
	}
	
	roomID := client.RoomID
	
	// 調用業務邏輯
	ctx := context.Background()
	err := mh.gameUsecase.LeaveRoom(ctx, roomID, client.PlayerID)
	if err != nil {
		mh.logger.Errorf("Failed to leave room: %v", err)
		mh.sendErrorResponse(client, "Failed to leave room")
		return
	}
	
	// 通知 Hub
	mh.hub.leaveRoom <- &LeaveRoomMessage{
		Client: client,
		RoomID: roomID,
	}
	
	// 清除客戶端房間ID
	client.RoomID = ""
	
	// 發送響應
	response := &pb.GameMessage{
		Type: pb.MessageType_LEAVE_ROOM_RESPONSE,
		Data: &pb.GameMessage_LeaveRoomResponse{
			LeaveRoomResponse: &pb.LeaveRoomResponse{
				Success:   true,
				RoomId:    roomID,
				Timestamp: time.Now().Unix(),
			},
		},
	}
	
	client.sendProtobuf(response)
	
	mh.logger.Infof("Player %d left room %s", client.PlayerID, roomID)
}

// handleHeartbeat 處理心跳消息
func (mh *MessageHandler) handleHeartbeat(client *Client, message *pb.GameMessage) {
	response := &pb.GameMessage{
		Type: pb.MessageType_HEARTBEAT_RESPONSE,
		Data: &pb.GameMessage_HeartbeatResponse{
			HeartbeatResponse: &pb.HeartbeatResponse{
				ServerTime: time.Now().Unix(),
				Timestamp:  time.Now().Unix(),
			},
		},
	}
	
	client.sendProtobuf(response)
}

// handleGetRoomList 處理獲取房間列表消息
func (mh *MessageHandler) handleGetRoomList(client *Client, message *pb.GameMessage) {
	ctx := context.Background()
	rooms, err := mh.gameUsecase.GetRoomList(ctx, "")
	if err != nil {
		mh.logger.Errorf("Failed to get room list: %v", err)
		mh.sendErrorResponse(client, "Failed to get room list")
		return
	}
	
	// 轉換房間數據到 Protobuf 格式
	var pbRooms []*pb.RoomInfo
	for _, room := range rooms {
		pbRoom := &pb.RoomInfo{
			RoomId:      room.ID,
			Name:        room.Name,
			Type:        string(room.Type),
			PlayerCount: int32(len(room.Players)),
			MaxPlayers:  room.MaxPlayers,
			Status:      string(room.Status),
		}
		pbRooms = append(pbRooms, pbRoom)
	}
	
	// 發送響應
	response := &pb.GameMessage{
		Type: pb.MessageType_ROOM_LIST_RESPONSE,
		Data: &pb.GameMessage_RoomListResponse{
			RoomListResponse: &pb.RoomListResponse{
				Rooms:     pbRooms,
				Timestamp: time.Now().Unix(),
			},
		},
	}
	
	client.sendProtobuf(response)
}

// handleGetPlayerInfo 處理獲取玩家信息消息
func (mh *MessageHandler) handleGetPlayerInfo(client *Client, message *pb.GameMessage) {
	response := &pb.GameMessage{
		Type: pb.MessageType_PLAYER_INFO_RESPONSE,
		Data: &pb.GameMessage_PlayerInfoResponse{
			PlayerInfoResponse: &pb.PlayerInfoResponse{
				PlayerId: client.PlayerID,
				Nickname: client.ID,
				Balance:  10000, // 模擬餘額
				Level:    1,
				Exp:      0,
				RoomId:   client.RoomID,
				Timestamp: time.Now().Unix(),
			},
		},
	}
	
	client.sendProtobuf(response)
}

// sendErrorResponse 發送錯誤響應
func (mh *MessageHandler) sendErrorResponse(client *Client, errorMsg string) {
	response := &pb.GameMessage{
		Type: pb.MessageType_ERROR,
		Data: &pb.GameMessage_Error{
			Error: &pb.ErrorMessage{
				Message:   errorMsg,
				Code:      "GENERAL_ERROR",
				Timestamp: time.Now().Unix(),
			},
		},
	}
	
	client.sendProtobuf(response)
}

// broadcastToRoom 向房間廣播 Protobuf 消息
func (mh *MessageHandler) broadcastToRoom(roomID string, message *pb.GameMessage, exclude *Client) {
	data, err := proto.Marshal(message)
	if err != nil {
		mh.logger.Errorf("Failed to marshal protobuf message: %v", err)
		return
	}
	
	mh.hub.BroadcastToRoom(roomID, data, exclude)
}

// broadcastGlobal 全局廣播 Protobuf 消息
func (mh *MessageHandler) broadcastGlobal(message *pb.GameMessage) {
	data, err := proto.Marshal(message)
	if err != nil {
		mh.logger.Errorf("Failed to marshal protobuf message: %v", err)
		return
	}
	
	mh.hub.BroadcastGlobal(data)
}