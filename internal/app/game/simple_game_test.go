package game

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
)

// MockGameRepo 模擬遊戲倉庫
type MockGameRepo struct{}

func (m *MockGameRepo) SaveRoom(ctx context.Context, room *game.Room) error { return nil }
func (m *MockGameRepo) GetRoom(ctx context.Context, roomID string) (*game.Room, error) {
	return &game.Room{ID: roomID, Players: make(map[int64]*game.Player), Fishes: make(map[int64]*game.Fish), Bullets: make(map[int64]*game.Bullet)}, nil
}
func (m *MockGameRepo) ListRooms(ctx context.Context, roomType game.RoomType) ([]*game.Room, error) {
	return []*game.Room{}, nil
}
func (m *MockGameRepo) DeleteRoom(ctx context.Context, roomID string) error { return nil }
func (m *MockGameRepo) SaveGameStatistics(ctx context.Context, playerID int64, stats *game.GameStatistics) error {
	return nil
}
func (m *MockGameRepo) GetGameStatistics(ctx context.Context, playerID int64) (*game.GameStatistics, error) {
	return &game.GameStatistics{}, nil
}
func (m *MockGameRepo) SaveGameEvent(ctx context.Context, event *game.GameEvent) error { return nil }
func (m *MockGameRepo) GetGameEvents(ctx context.Context, roomID string, limit int) ([]*game.GameEvent, error) {
	return []*game.GameEvent{}, nil
}

// MockPlayerRepo 模擬玩家倉庫
type MockPlayerRepo struct{}

func (m *MockPlayerRepo) GetPlayer(ctx context.Context, playerID int64) (*game.Player, error) {
	return &game.Player{
		ID:       playerID,
		UserID:   playerID,
		Nickname: "TestPlayer",
		Balance:  10000,
		Status:   game.PlayerStatusIdle,
	}, nil
}
func (m *MockPlayerRepo) UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error {
	return nil
}
func (m *MockPlayerRepo) UpdatePlayerStatus(ctx context.Context, playerID int64, status game.PlayerStatus) error {
	return nil
}

// 簡化版測試 - 只測試核心功能
func TestSimpleGameComponents(t *testing.T) {
	// 創建日誌記錄器
	log := logger.New(os.Stdout, "debug", "console")

	// 創建依賴
	gameRepo := &MockGameRepo{}
	playerRepo := &MockPlayerRepo{}
	spawner := game.NewFishSpawner(log)
	mathModel := game.NewMathModel(log)
	roomManager := game.NewRoomManager(log, spawner, mathModel)
	gameUsecase := game.NewGameUsecase(gameRepo, playerRepo, roomManager, spawner, mathModel, log)

	t.Run("Test Hub", func(t *testing.T) {
		// 創建 Hub
		hub := NewHub(gameUsecase, log)
		go hub.Run()
		defer hub.Stop()

		// 創建測試客戶端
		client := &Client{
			ID:       "test_client_1",
			PlayerID: 1,
			send:     make(chan []byte, 256),
			hub:      hub,
			logger:   log,
		}

		// 測試客戶端註冊
		hub.register <- client
		time.Sleep(100 * time.Millisecond)

		stats := hub.GetStats()
		if stats.ActiveConnections != 1 {
			t.Errorf("Expected 1 active connection, got %d", stats.ActiveConnections)
		}

		// 測試客戶端註銷
		hub.unregister <- client
		time.Sleep(100 * time.Millisecond)

		stats = hub.GetStats()
		if stats.ActiveConnections != 0 {
			t.Errorf("Expected 0 active connections after unregister, got %d", stats.ActiveConnections)
		}

		t.Log("✓ Hub 基本功能測試通過")
	})

	t.Run("Test MessageHandler", func(t *testing.T) {
		// 創建 Hub 和消息處理器
		hub := NewHub(gameUsecase, log)
		go hub.Run()
		defer hub.Stop()

		messageHandler := NewMessageHandler(gameUsecase, hub, log)

		// 創建測試客戶端
		client := &Client{
			ID:       "test_player_1",
			PlayerID: 1,
			send:     make(chan []byte, 256),
			hub:      hub,
			logger:   log,
		}

		// 註冊客戶端
		hub.register <- client
		time.Sleep(50 * time.Millisecond)

		// 測試心跳消息
		heartbeatMsg := &pb.GameMessage{
			Type: pb.MessageType_HEARTBEAT,
			Data: &pb.GameMessage_Heartbeat{
				Heartbeat: &pb.HeartbeatMessage{
					Timestamp: time.Now().Unix(),
				},
			},
		}

		messageHandler.HandleMessage(client, heartbeatMsg)

		// 檢查是否有響應
		select {
		case <-client.send:
			t.Log("✓ 心跳消息處理成功")
		case <-time.After(1 * time.Second):
			t.Error("心跳消息處理超時")
		}

		t.Log("✓ 消息處理器基本功能測試通過")
	})

	t.Run("Test Room Operations", func(t *testing.T) {
		hub := NewHub(gameUsecase, log)
		go hub.Run()
		defer hub.Stop()

		messageHandler := NewMessageHandler(gameUsecase, hub, log)

		client := &Client{
			ID:       "test_player_room",
			PlayerID: 1,
			send:     make(chan []byte, 256),
			hub:      hub,
			logger:   log,
		}

		hub.register <- client
		time.Sleep(50 * time.Millisecond)

		// 測試加入房間
		joinMsg := &pb.GameMessage{
			Type: pb.MessageType_JOIN_ROOM,
			Data: &pb.GameMessage_JoinRoom{
				JoinRoom: &pb.JoinRoomRequest{
					RoomId: "test_room_001",
				},
			},
		}

		messageHandler.HandleMessage(client, joinMsg)
		time.Sleep(100 * time.Millisecond)

		if client.RoomID != "test_room_001" {
			t.Errorf("Expected client to be in room test_room_001, got: %s", client.RoomID)
		}

		// 測試離開房間
		leaveMsg := &pb.GameMessage{
			Type: pb.MessageType_LEAVE_ROOM,
			Data: &pb.GameMessage_LeaveRoom{
				LeaveRoom: &pb.LeaveRoomRequest{},
			},
		}

		messageHandler.HandleMessage(client, leaveMsg)
		time.Sleep(100 * time.Millisecond)

		if client.RoomID != "" {
			t.Errorf("Expected client to not be in any room, got: %s", client.RoomID)
		}

		t.Log("✓ 房間操作測試通過")
	})

	t.Log("🎉 所有簡化測試通過！")
}
