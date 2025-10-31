package game

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// MockGameRepo 模擬遊戲倉庫
type MockGameRepo struct{}

func (m *MockGameRepo) SaveRoom(ctx context.Context, room *Room) error {
	return nil
}

func (m *MockGameRepo) GetRoom(ctx context.Context, roomID string) (*Room, error) {
	return nil, nil
}

func (m *MockGameRepo) ListRooms(ctx context.Context, roomType RoomType) ([]*Room, error) {
	return []*Room{}, nil
}

func (m *MockGameRepo) DeleteRoom(ctx context.Context, roomID string) error {
	return nil
}

func (m *MockGameRepo) SaveGameStatistics(ctx context.Context, playerID int64, stats *GameStatistics) error {
	return nil
}

func (m *MockGameRepo) GetGameStatistics(ctx context.Context, playerID int64) (*GameStatistics, error) {
	return &GameStatistics{}, nil
}

func (m *MockGameRepo) SaveGameEvent(ctx context.Context, event *GameEvent) error {
	return nil
}

func (m *MockGameRepo) GetGameEvents(ctx context.Context, roomID string, limit int) ([]*GameEvent, error) {
	return []*GameEvent{}, nil
}

// MockPlayerRepo 模擬玩家倉庫
type MockPlayerRepo struct{}

func (m *MockPlayerRepo) GetPlayer(ctx context.Context, playerID int64) (*Player, error) {
	return &Player{
		ID:       playerID,
		UserID:   playerID,
		Nickname: "TestPlayer",
		Balance:  10000, // 100元
		Status:   PlayerStatusIdle,
	}, nil
}

func (m *MockPlayerRepo) UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error {
	return nil
}

func (m *MockPlayerRepo) UpdatePlayerStatus(ctx context.Context, playerID int64, status PlayerStatus) error {
	return nil
}

// 測試遊戲核心組件
func TestGameComponents(t *testing.T) {
	// 創建日誌記錄器
	log := logger.New(os.Stdout, "debug", "console")

	t.Run("Test MathModel", func(t *testing.T) {
		mathModel := NewMathModel(log)
		
		// 測試命中計算
		bullet := &Bullet{
			ID:       1,
			PlayerID: 1,
			Power:    10,
			Cost:     10,
		}
		
		fish := &Fish{
			ID:     1,
			Type:   getDefaultFishTypes()[0], // 小丑魚
			Health: 1,
			Value:  5,
		}
		
		result := mathModel.CalculateHit(bullet, fish)
		
		if result == nil {
			t.Error("Expected hit result, got nil")
		}
		
		t.Logf("Hit result: success=%t, damage=%d, reward=%d", 
			result.Success, result.Damage, result.Reward)
	})

	t.Run("Test FishSpawner", func(t *testing.T) {
		spawner := NewFishSpawner(log)
		
		// 測試魚類型
		fishTypes := spawner.GetFishTypes()
		if len(fishTypes) == 0 {
			t.Error("Expected fish types, got empty slice")
		}
		
		// 測試生成魚
		config := RoomConfig{
			FishSpawnRate: 1.0, // 100%生成率
			RoomWidth:     1200,
			RoomHeight:    800,
		}
		
		// 重置生成時間以確保可以生成
		spawner.lastSpawnTime = time.Time{}
		
		fish := spawner.TrySpawnFish(config)
		if fish == nil {
			// 嘗試強制生成
			fish = spawner.SpawnSpecificFish(1, config)
		}
		
		if fish == nil {
			t.Error("Expected fish to be spawned")
		} else {
			t.Logf("Spawned fish: type=%s, health=%d, value=%d", 
				fish.Type.Name, fish.Health, fish.Value)
		}
	})

	t.Run("Test RoomManager", func(t *testing.T) {
		spawner := NewFishSpawner(log)
		mathModel := NewMathModel(log)
		roomManager := NewRoomManager(log, spawner, mathModel)
		
		// 測試創建房間
		room, err := roomManager.CreateRoom(RoomTypeNovice, 4)
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}
		
		if room.Type != RoomTypeNovice {
			t.Errorf("Expected room type %s, got %s", RoomTypeNovice, room.Type)
		}
		
		// 測試玩家加入房間
		player := &Player{
			ID:       1,
			UserID:   1,
			Nickname: "TestPlayer",
			Balance:  10000,
			Status:   PlayerStatusIdle,
		}
		
		err = roomManager.JoinRoom(room.ID, player)
		if err != nil {
			t.Fatalf("Failed to join room: %v", err)
		}
		
		if len(room.Players) != 1 {
			t.Errorf("Expected 1 player in room, got %d", len(room.Players))
		}
		
		t.Logf("Room created: ID=%s, Players=%d", room.ID, len(room.Players))
	})

	t.Run("Test GameUsecase", func(t *testing.T) {
		// 創建依賴
		gameRepo := &MockGameRepo{}
		playerRepo := &MockPlayerRepo{}
		spawner := NewFishSpawner(log)
		mathModel := NewMathModel(log)
		roomManager := NewRoomManager(log, spawner, mathModel)
		
		// 創建遊戲用例
		gameUsecase := NewGameUsecase(gameRepo, playerRepo, roomManager, spawner, mathModel, log)
		
		ctx := context.Background()
		
		// 測試創建房間
		room, err := gameUsecase.CreateRoom(ctx, RoomTypeNovice, 4)
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}
		
		// 測試玩家加入房間
		err = gameUsecase.JoinRoom(ctx, room.ID, 1)
		if err != nil {
			t.Fatalf("Failed to join room: %v", err)
		}
		
		// 測試開火
		bullet, err := gameUsecase.FireBullet(ctx, room.ID, 1, 1.0, 10)
		if err != nil {
			t.Fatalf("Failed to fire bullet: %v", err)
		}
		
		// 檢查房間中是否有魚可以命中
		roomState, _ := gameUsecase.GetRoomState(ctx, room.ID)
		if len(roomState.Fishes) > 0 {
			// 取第一條魚進行命中測試
			var firstFish *Fish
			for _, fish := range roomState.Fishes {
				firstFish = fish
				break
			}
			
			// 測試命中
			hitResult, err := gameUsecase.HitFish(ctx, room.ID, bullet.ID, firstFish.ID)
			if err != nil {
				t.Fatalf("Failed to hit fish: %v", err)
			}
			
			t.Logf("Hit result: success=%t, damage=%d, reward=%d", 
				hitResult.Success, hitResult.Damage, hitResult.Reward)
		}
		
		t.Logf("Game usecase test completed successfully")
	})
}

// 測試遊戲流程
func TestGameFlow(t *testing.T) {
	// 創建日誌記錄器
	log := logger.New(os.Stdout, "debug", "console")

	// 創建完整的遊戲環境
	gameRepo := &MockGameRepo{}
	playerRepo := &MockPlayerRepo{}
	spawner := NewFishSpawner(log)
	mathModel := NewMathModel(log)
	roomManager := NewRoomManager(log, spawner, mathModel)
	gameUsecase := NewGameUsecase(gameRepo, playerRepo, roomManager, spawner, mathModel, log)
	
	ctx := context.Background()
	
	// 模擬完整遊戲流程
	t.Run("Complete Game Flow", func(t *testing.T) {
		// 1. 創建房間
		room, err := gameUsecase.CreateRoom(ctx, RoomTypeNovice, 4)
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}
		t.Logf("✓ Room created: %s", room.ID)
		
		// 2. 玩家加入房間
		playerID := int64(1)
		err = gameUsecase.JoinRoom(ctx, room.ID, playerID)
		if err != nil {
			t.Fatalf("Failed to join room: %v", err)
		}
		t.Logf("✓ Player %d joined room", playerID)
		
		// 3. 模擬遊戲循環
		for i := 0; i < 5; i++ {
			// 開火
			bullet, err := gameUsecase.FireBullet(ctx, room.ID, playerID, float64(i)*0.5, 10)
			if err != nil {
				t.Fatalf("Failed to fire bullet: %v", err)
			}
			t.Logf("✓ Bullet fired: ID=%d, Cost=%d", bullet.ID, bullet.Cost)
			
			// 獲取當前房間狀態
			currentRoom, _ := gameUsecase.GetRoomState(ctx, room.ID)
			
			// 嘗試命中魚
			if len(currentRoom.Fishes) > 0 {
				var targetFish *Fish
				for _, fish := range currentRoom.Fishes {
					targetFish = fish
					break
				}
				
				hitResult, err := gameUsecase.HitFish(ctx, room.ID, bullet.ID, targetFish.ID)
				if err != nil {
					t.Fatalf("Failed to hit fish: %v", err)
				}
				
				if hitResult.Success {
					t.Logf("✓ Hit fish! Damage=%d, Reward=%d, Critical=%t", 
						hitResult.Damage, hitResult.Reward, hitResult.IsCritical)
				} else {
					t.Logf("✗ Missed fish")
				}
			}
			
			// 短暫暫停模擬真實遊戲
			time.Sleep(10 * time.Millisecond)
		}
		
		// 4. 獲取最終統計
		stats, err := gameUsecase.GetPlayerStatistics(ctx, playerID)
		if err != nil {
			t.Fatalf("Failed to get player statistics: %v", err)
		}
		t.Logf("✓ Final statistics: %+v", stats)
		
		// 5. 玩家離開房間
		err = gameUsecase.LeaveRoom(ctx, room.ID, playerID)
		if err != nil {
			t.Fatalf("Failed to leave room: %v", err)
		}
		t.Logf("✓ Player left room")
		
		t.Logf("🎉 Complete game flow test passed!")
	})
}