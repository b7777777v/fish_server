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
// 配置常量
// ========================================

const (
	// 默認砲台位置配置（畫布底部中央）
	DefaultCannonPositionX = 600.0
	DefaultCannonPositionY = 750.0
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
	case pb.MessageType_HIT_FISH:
		mh.handleHitFish(client, message)
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

	// 獲取子彈發射位置（優先使用客戶端提供的位置，否則使用默認砲台位置）
	position := game.Position{X: DefaultCannonPositionX, Y: DefaultCannonPositionY}
	if fireData.Position != nil {
		position = game.Position{
			X: fireData.Position.X,
			Y: fireData.Position.Y,
		}
	}

	bullet, err := mh.gameUsecase.FireBullet(ctx, client.RoomID, client.PlayerID,
		fireData.Direction, fireData.Power, position)
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

	// 開火後推送更新的餘額給客戶端
	mh.sendPlayerInfoUpdate(client)

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
    
    ctx := context.Background()
    if client.PlayerID != 0 {
        if err := mh.gameUsecase.JoinRoom(ctx, roomID, client.PlayerID); err != nil {
            mh.logger.Errorf("Failed to join room: %v", err)
            mh.sendErrorResponse(client, "Failed to join room")
            return
        }
    }
    
    client.RoomID = roomID
    mh.hub.joinRoom <- &JoinRoomMessage{Client: client, RoomID: roomID}
    
    room, err := mh.gameUsecase.GetRoom(ctx, roomID)
    if err != nil {
        mh.logger.Errorf("Failed to get room info after join: %v", err)
    }
    playerCount := int32(1)
    if room != nil {
        playerCount = int32(len(room.Players))
    } else {
        mh.hub.mu.RLock()
        if clients, ok := mh.hub.rooms[roomID]; ok {
            playerCount = int32(len(clients))
        }
        mh.hub.mu.RUnlock()
    }
    
    response := &pb.GameMessage{
        Type: pb.MessageType_JOIN_ROOM_RESPONSE,
        Data: &pb.GameMessage_JoinRoomResponse{
            JoinRoomResponse: &pb.JoinRoomResponse{
                Success:     true,
                RoomId:      roomID,
                Timestamp:   time.Now().Unix(),
                PlayerCount: playerCount,
            },
        },
    }
    client.sendProtobuf(response)
    
    go mh.broadcastRoomState(roomID)
    if playerCount == 1 {
        go mh.StartRoomStateUpdates(roomID)
        mh.logger.Infof("Started room state updates for room %s", roomID)
    }
    
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

// handleHitFish 處理擊中魚類消息
func (mh *MessageHandler) handleHitFish(client *Client, message *pb.GameMessage) {
	if client.RoomID == "" {
		mh.sendErrorResponse(client, "Not in any room")
		return
	}

	// 解析擊中數據
	hitData := message.GetHitFish()
	if hitData == nil {
		mh.sendErrorResponse(client, "Invalid hit fish data")
		return
	}

	// 驗證參數
	if hitData.GetBulletId() <= 0 || hitData.GetFishId() <= 0 {
		mh.sendErrorResponse(client, "Invalid bullet or fish ID")
		return
	}

	// 調用業務邏輯
	ctx := context.Background()
	hitResult, err := mh.gameUsecase.HitFish(ctx, client.RoomID, hitData.GetBulletId(), hitData.GetFishId())
	if err != nil {
		mh.logger.Errorf("Failed to process hit fish: %v", err)
		mh.sendErrorResponse(client, "Failed to process hit")
		return
	}

	// 構建響應消息
	response := &pb.GameMessage{
		Type: pb.MessageType_HIT_FISH_RESPONSE,
		Data: &pb.GameMessage_HitFishResponse{
			HitFishResponse: &pb.HitFishResponse{
				Success:    hitResult.Success,
				BulletId:   hitData.GetBulletId(),
				FishId:     hitData.GetFishId(),
				Damage:     hitResult.Damage,
				Reward:     hitResult.Reward,
				IsKilled:   hitResult.Reward > 0, // 有獎勵表示擊殺
				IsCritical: hitResult.IsCritical,
				Multiplier: hitResult.Multiplier,
				Timestamp:  time.Now().Unix(),
			},
		},
	}

	// 發送響應給客戶端
	client.sendProtobuf(response)

	// 如果擊殺了魚，廣播給房間所有玩家
	if hitResult.Reward > 0 {
		// 廣播魚死亡事件
		fishDiedMsg := &pb.GameMessage{
			Type: pb.MessageType_FISH_DIED,
			Data: &pb.GameMessage_FishDied{
				FishDied: &pb.FishDiedEvent{
					FishId:    hitData.GetFishId(),
					PlayerId:  client.PlayerID,
					Reward:    hitResult.Reward,
					Timestamp: time.Now().Unix(),
				},
			},
		}
		mh.broadcastToRoom(client.RoomID, fishDiedMsg, nil)

		// 廣播玩家獎勵事件
		rewardMsg := &pb.GameMessage{
			Type: pb.MessageType_PLAYER_REWARD,
			Data: &pb.GameMessage_PlayerReward{
				PlayerReward: &pb.PlayerRewardEvent{
					PlayerId:  client.PlayerID,
					Reward:    hitResult.Reward,
					Timestamp: time.Now().Unix(),
				},
			},
		}
		mh.broadcastToRoom(client.RoomID, rewardMsg, nil)

		// 擊殺後推送更新的餘額給客戶端
		mh.sendPlayerInfoUpdate(client)

		mh.logger.Infof("Player %d killed fish %d in room %s, reward: %d",
			client.PlayerID, hitData.GetFishId(), client.RoomID, hitResult.Reward)
	} else {
		mh.logger.Debugf("Player %d hit fish %d in room %s, damage: %d",
			client.PlayerID, hitData.GetFishId(), client.RoomID, hitResult.Damage)
	}
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
    ctx := context.Background()
    var nickname string
    var balance int64
    seatID := int32(-1)
    if client.PlayerID == 0 {
        nickname = client.ID
        if client.RoomID != "" {
            mh.hub.mu.RLock()
            rm := mh.hub.roomManagers[client.RoomID]
            if rm != nil {
                if pi, ok := rm.gameState.Players[client.ID]; ok {
                    balance = pi.Balance
                    seatID = int32(pi.SeatID)
                }
            }
            mh.hub.mu.RUnlock()
        }
    } else {
        player, err := mh.gameUsecase.GetPlayerInfo(ctx, client.PlayerID)
        if err != nil {
            mh.logger.Errorf("Failed to get player info: %v", err)
            mh.sendErrorResponse(client, "Failed to get player info")
            return
        }
        nickname = player.Nickname
        balance = player.Balance
        if client.RoomID != "" {
            room, err := mh.gameUsecase.GetRoom(ctx, client.RoomID)
            if err == nil && room != nil {
                seatID = int32(room.GetPlayerSeat(client.PlayerID))
            }
        }
    }
    response := &pb.GameMessage{
        Type: pb.MessageType_PLAYER_INFO_RESPONSE,
        Data: &pb.GameMessage_PlayerInfoResponse{
            PlayerInfoResponse: &pb.PlayerInfoResponse{
                PlayerId:  client.PlayerID,
                Nickname:  nickname,
                Balance:   balance,
                Level:     1,
                Exp:       0,
                RoomId:    client.RoomID,
                SeatId:    seatID,
                Timestamp: time.Now().Unix(),
            },
        },
    }
    client.sendProtobuf(response)
    mh.logger.Debugf("Sent player info: player=%d, balance=%d", client.PlayerID, balance)
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

// ========================================
// 房間狀態推送功能
// ========================================

// broadcastRoomState 廣播房間狀態更新
func (mh *MessageHandler) broadcastRoomState(roomID string) {
	room, err := mh.gameUsecase.GetRoom(context.Background(), roomID)
	if err != nil {
		mh.logger.Errorf("Failed to get room %s: %v", roomID, err)
		return
	}

	// 轉換魚類信息
	var fishInfos []*pb.FishInfo
	for _, fish := range room.Fishes {
		fishInfo := &pb.FishInfo{
			FishId:      fish.ID,
			FishType:    fish.Type.ID,
			Position:    &pb.Position{X: fish.Position.X, Y: fish.Position.Y},
			Direction:   fish.Direction,
			Speed:       fish.Speed,
			Health:      fish.Health,
			MaxHealth:   fish.MaxHealth,
			Value:       fish.Value,
			Status:      string(fish.Status),
			SpawnTime:   fish.SpawnTime.Unix(),
			InFormation: false, // 默认值，稍后会更新
			FormationId: "",
		}
		fishInfos = append(fishInfos, fishInfo)
	}

	// 轉換子彈信息
	var bulletInfos []*pb.BulletInfo
	for _, bullet := range room.Bullets {
		bulletInfo := &pb.BulletInfo{
			BulletId:  bullet.ID,
			PlayerId:  bullet.PlayerID,
			Position:  &pb.Position{X: bullet.Position.X, Y: bullet.Position.Y},
			Direction: bullet.Direction,
			Speed:     bullet.Speed,
			Power:     bullet.Power,
			Cost:      bullet.Cost,
			Status:    string(bullet.Status),
			CreatedAt: bullet.CreatedAt.Unix(),
		}
		bulletInfos = append(bulletInfos, bulletInfo)
	}

	// 獲取並轉換魚群陣型信息
	var formationInfos []*pb.FormationInfo
	formations, err := mh.gameUsecase.GetFormationsInRoom(context.Background(), roomID)
	if err == nil {
		for _, formation := range formations {
			var fishIds []int64
			for _, fish := range formation.Fishes {
				fishIds = append(fishIds, fish.ID)
				
				// 更新魚類的陣型信息
				for _, fishInfo := range fishInfos {
					if fishInfo.FishId == fish.ID {
						fishInfo.InFormation = true
						fishInfo.FormationId = formation.ID
					}
				}
			}

			// 轉換路徑控制點信息供前端渲染
			var routePoints []*pb.Position
			if formation.Route != nil {
				for _, point := range formation.Route.Points {
					routePoints = append(routePoints, &pb.Position{
						X: point.X,
						Y: point.Y,
					})
				}
			}

			formationInfo := &pb.FormationInfo{
				FormationId:     formation.ID,
				FormationType:   string(formation.Type),
				FishIds:         fishIds,
				CenterPosition:  &pb.Position{X: formation.Position.X, Y: formation.Position.Y},
				Direction:       formation.Direction,
				Speed:           formation.Speed,
				Status:          string(formation.Status),
				Progress:        formation.Progress,
				RouteId:         formation.Route.ID,
				RouteName:       formation.Route.Name,
				CreatedAt:       formation.CreatedAt.Unix(),
				Size: &pb.FormationSize{
					Width:  formation.Size.Width,
					Height: formation.Size.Height,
					Depth:  formation.Size.Depth,
				},
				Route: &pb.RouteInfo{
					RouteId:    formation.Route.ID,
					RouteName:  formation.Route.Name,
					RouteType:  string(formation.Route.Type),
					Points:     routePoints,
					Duration:   float64(formation.Route.Duration.Milliseconds()),
					Difficulty: formation.Route.Difficulty,
					Looping:    formation.Route.Looping,
				},
			}
			formationInfos = append(formationInfos, formationInfo)
		}
	}

	// 創建房間狀態更新消息
	roomStateUpdate := &pb.GameMessage{
		Type: pb.MessageType_ROOM_STATE_UPDATE,
		Data: &pb.GameMessage_RoomStateUpdate{
			RoomStateUpdate: &pb.RoomStateUpdate{
				RoomId:      roomID,
				Fishes:      fishInfos,
				Bullets:     bulletInfos,
				Formations:  formationInfos,
				PlayerCount: int32(len(room.Players)),
				Timestamp:   time.Now().Unix(),
				RoomStatus:  string(room.Status),
			},
		},
	}

	// 廣播給房間內所有玩家
	mh.broadcastToRoom(roomID, roomStateUpdate, nil)
}

// broadcastFormationSpawned 廣播魚群陣型生成事件
func (mh *MessageHandler) BroadcastFormationSpawned(roomID string, formation *game.FishFormation) {
	var fishInfos []*pb.FishInfo
	for _, fish := range formation.Fishes {
		fishInfo := &pb.FishInfo{
			FishId:      fish.ID,
			FishType:    fish.Type.ID,
			Position:    &pb.Position{X: fish.Position.X, Y: fish.Position.Y},
			Direction:   fish.Direction,
			Speed:       fish.Speed,
			Health:      fish.Health,
			MaxHealth:   fish.MaxHealth,
			Value:       fish.Value,
			Status:      string(fish.Status),
			SpawnTime:   fish.SpawnTime.Unix(),
			InFormation: true,
			FormationId: formation.ID,
		}
		fishInfos = append(fishInfos, fishInfo)
	}

	// 轉換路徑控制點信息供前端渲染
	var routePoints []*pb.Position
	if formation.Route != nil {
		for _, point := range formation.Route.Points {
			routePoints = append(routePoints, &pb.Position{
				X: point.X,
				Y: point.Y,
			})
		}
	}

	formationInfo := &pb.FormationInfo{
		FormationId:     formation.ID,
		FormationType:   string(formation.Type),
		CenterPosition:  &pb.Position{X: formation.Position.X, Y: formation.Position.Y},
		Direction:       formation.Direction,
		Speed:           formation.Speed,
		Status:          string(formation.Status),
		Progress:        formation.Progress,
		RouteId:         formation.Route.ID,
		RouteName:       formation.Route.Name,
		CreatedAt:       formation.CreatedAt.Unix(),
		Size: &pb.FormationSize{
			Width:  formation.Size.Width,
			Height: formation.Size.Height,
			Depth:  formation.Size.Depth,
		},
		Route: &pb.RouteInfo{
			RouteId:    formation.Route.ID,
			RouteName:  formation.Route.Name,
			RouteType:  string(formation.Route.Type),
			Points:     routePoints,
			Duration:   float64(formation.Route.Duration.Milliseconds()),
			Difficulty: formation.Route.Difficulty,
			Looping:    formation.Route.Looping,
		},
	}

	message := &pb.GameMessage{
		Type: pb.MessageType_FORMATION_SPAWNED,
		Data: &pb.GameMessage_FormationSpawned{
			FormationSpawned: &pb.FormationSpawnedEvent{
				RoomId:    roomID,
				Formation: formationInfo,
				Fishes:    fishInfos,
				Timestamp: time.Now().Unix(),
			},
		},
	}

	mh.broadcastToRoom(roomID, message, nil)
	mh.logger.Infof("Broadcasted formation spawned: %s in room %s", formation.Type, roomID)
}

// StartRoomStateUpdates 開始定期房間狀態更新
func (mh *MessageHandler) StartRoomStateUpdates(roomID string) {
	go func() {
		ticker := time.NewTicker(2 * time.Second) // 每2秒更新一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 檢查房間是否還存在
				_, err := mh.gameUsecase.GetRoom(context.Background(), roomID)
				if err != nil {
					mh.logger.Infof("Room %s no longer exists, stopping state updates", roomID)
					return
				}

				mh.broadcastRoomState(roomID)
			}
		}
	}()
}

// sendPlayerInfoUpdate 發送玩家資訊更新（用於餘額變動後）
func (mh *MessageHandler) sendPlayerInfoUpdate(client *Client) {
    ctx := context.Background()
    var nickname string
    var balance int64
    seatID := int32(-1)
    if client.PlayerID == 0 {
        nickname = client.ID
        if client.RoomID != "" {
            mh.hub.mu.RLock()
            rm := mh.hub.roomManagers[client.RoomID]
            if rm != nil {
                if pi, ok := rm.gameState.Players[client.ID]; ok {
                    balance = pi.Balance
                    seatID = int32(pi.SeatID)
                }
            }
            mh.hub.mu.RUnlock()
        }
    } else {
        player, err := mh.gameUsecase.GetPlayerInfo(ctx, client.PlayerID)
        if err != nil {
            mh.logger.Errorf("Failed to get player info for update: %v", err)
            return
        }
        nickname = player.Nickname
        balance = player.Balance
        if client.RoomID != "" {
            room, err := mh.gameUsecase.GetRoom(ctx, client.RoomID)
            if err == nil && room != nil {
                seatID = int32(room.GetPlayerSeat(client.PlayerID))
            }
        }
    }
    response := &pb.GameMessage{
        Type: pb.MessageType_PLAYER_INFO_RESPONSE,
        Data: &pb.GameMessage_PlayerInfoResponse{
            PlayerInfoResponse: &pb.PlayerInfoResponse{
                PlayerId:  client.PlayerID,
                Nickname:  nickname,
                Balance:   balance,
                Level:     1,
                Exp:       0,
                RoomId:    client.RoomID,
                SeatId:    seatID,
                Timestamp: time.Now().Unix(),
            },
        },
    }
    client.sendProtobuf(response)
    mh.logger.Debugf("Sent player info update: player=%d, balance=%d", client.PlayerID, balance)
}