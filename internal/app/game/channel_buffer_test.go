package game

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/pkg/token"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChannelBuffers 測試所有關鍵通道都有緩衝區,避免阻塞
func TestChannelBuffers(t *testing.T) {
	log := logger.New(os.Stdout, "debug", "console")
	gameRepo := &MockGameRepo{}
	playerRepo := &MockPlayerRepo{}
	inventoryRepo := NewMockInventoryRepo()

	bizPlayerRepo := NewMockBizPlayerRepo()
	jwtConfig := &conf.JWT{Secret: "test-secret", Expire: 3600}
	tokenHelper := token.NewTokenHelper(jwtConfig)
	playerUsecase := player.NewPlayerUsecase(bizPlayerRepo, tokenHelper, log)

	testRoomConfig := game.RoomConfig{
		MinBet:               1,
		MaxBet:               100,
		BulletCostMultiplier: 1.0,
		FishSpawnRate:        0.3,
		MaxFishCount:         20,
		RoomWidth:            1200,
		RoomHeight:           800,
		TargetRTP:            0.96,
	}

	spawner := game.NewFishSpawner(log, testRoomConfig)
	mathModel := game.NewMathModel(log)
	inventoryManager, err := game.NewInventoryManager(inventoryRepo, log)
	require.NoError(t, err)

	rtpController := game.NewRTPController(inventoryManager, log)
	roomManager := game.NewRoomManager(log, spawner, mathModel, inventoryManager, rtpController)
	gameUsecase := game.NewGameUsecase(gameRepo, playerRepo, roomManager, spawner, mathModel, inventoryManager, rtpController, log)

	t.Run("Hub channels have buffers", func(t *testing.T) {
		hub := NewHub(gameUsecase, playerUsecase, log)

		// 測試通道容量
		assert.Greater(t, cap(hub.register), 0, "register channel should be buffered")
		assert.Greater(t, cap(hub.unregister), 0, "unregister channel should be buffered")
		assert.Greater(t, cap(hub.joinRoom), 0, "joinRoom channel should be buffered")
		assert.Greater(t, cap(hub.leaveRoom), 0, "leaveRoom channel should be buffered")
		assert.Greater(t, cap(hub.gameAction), 0, "gameAction channel should be buffered")
		assert.Greater(t, cap(hub.broadcast), 0, "broadcast channel should be buffered")

		// 驗證預期的緩衝區大小
		assert.Equal(t, 10, cap(hub.register), "register buffer should be 10")
		assert.Equal(t, 10, cap(hub.unregister), "unregister buffer should be 10")
		assert.Equal(t, 10, cap(hub.joinRoom), "joinRoom buffer should be 10")
		assert.Equal(t, 10, cap(hub.leaveRoom), "leaveRoom buffer should be 10")
		assert.Equal(t, 100, cap(hub.gameAction), "gameAction buffer should be 100")
		assert.Equal(t, 100, cap(hub.broadcast), "broadcast buffer should be 100")
	})

	t.Run("RoomManager channels have buffers", func(t *testing.T) {
		hub := NewHub(gameUsecase, playerUsecase, log)
		roomManager := NewRoomManager("test_room", gameUsecase, hub, log)

		// 測試通道容量
		assert.Greater(t, cap(roomManager.addClient), 0, "addClient channel should be buffered")
		assert.Greater(t, cap(roomManager.removeClient), 0, "removeClient channel should be buffered")
		assert.Greater(t, cap(roomManager.gameAction), 0, "gameAction channel should be buffered")

		// 驗證預期的緩衝區大小
		assert.Equal(t, 10, cap(roomManager.addClient), "addClient buffer should be 10")
		assert.Equal(t, 10, cap(roomManager.removeClient), "removeClient buffer should be 10")
		assert.Equal(t, 100, cap(roomManager.gameAction), "gameAction buffer should be 100")
	})
}

// TestNoChannelBlocking 測試通道不會阻塞
func TestNoChannelBlocking(t *testing.T) {
	log := logger.New(os.Stdout, "debug", "console")
	gameRepo := &MockGameRepo{}
	playerRepo := &MockPlayerRepo{}
	inventoryRepo := NewMockInventoryRepo()

	bizPlayerRepo := NewMockBizPlayerRepo()
	jwtConfig := &conf.JWT{Secret: "test-secret", Expire: 3600}
	tokenHelper := token.NewTokenHelper(jwtConfig)
	playerUsecase := player.NewPlayerUsecase(bizPlayerRepo, tokenHelper, log)

	testRoomConfig := game.RoomConfig{
		MinBet:               1,
		MaxBet:               100,
		BulletCostMultiplier: 1.0,
		FishSpawnRate:        0.3,
		MaxFishCount:         20,
		RoomWidth:            1200,
		RoomHeight:           800,
		TargetRTP:            0.96,
	}

	spawner := game.NewFishSpawner(log, testRoomConfig)
	mathModel := game.NewMathModel(log)
	inventoryManager, err := game.NewInventoryManager(inventoryRepo, log)
	require.NoError(t, err)

	rtpController := game.NewRTPController(inventoryManager, log)
	roomManager := game.NewRoomManager(log, spawner, mathModel, inventoryManager, rtpController)
	gameUsecase := game.NewGameUsecase(gameRepo, playerRepo, roomManager, spawner, mathModel, inventoryManager, rtpController, log)

	t.Run("Hub can handle burst of messages without blocking", func(t *testing.T) {
		hub := NewHub(gameUsecase, playerUsecase, log)
		go hub.Run()
		defer hub.Stop()

		// 創建多個客戶端
		clients := make([]*Client, 20)
		for i := 0; i < 20; i++ {
			clients[i] = &Client{
				ID:       "test_client_" + string(rune('A'+i)),
				PlayerID: int64(i + 1),
				send:     make(chan []byte, 256),
				hub:      hub,
				logger:   log,
			}
		}

		// 測試 burst 註冊不會阻塞
		start := time.Now()
		for _, client := range clients {
			hub.register <- client
		}
		duration := time.Since(start)

		// 即使 Hub.Run() 還沒處理,因為有緩衝區,發送應該是非阻塞的
		assert.Less(t, duration, 100*time.Millisecond, "Burst register should not block")

		// 等待處理完成
		time.Sleep(200 * time.Millisecond)
		stats := hub.GetStats()
		assert.Equal(t, 20, stats.ActiveConnections, "All clients should be registered")
	})

	t.Run("Multiple game actions don't block when sent concurrently", func(t *testing.T) {
		hub := NewHub(gameUsecase, playerUsecase, log)
		go hub.Run()
		defer hub.Stop()

		// 創建客戶端
		client := &Client{
			ID:       "test_player_1",
			PlayerID: 1,
			send:     make(chan []byte, 256),
			hub:      hub,
			logger:   log,
			RoomID:   "test_room",
		}

		// 並發發送多個遊戲操作到緩衝通道
		numActions := 50
		wg := sync.WaitGroup{}
		wg.Add(numActions)

		start := time.Now()
		for i := 0; i < numActions; i++ {
			go func(idx int) {
				defer wg.Done()

				fireMsg := &pb.GameMessage{
					Type: pb.MessageType_FIRE_BULLET,
					Data: &pb.GameMessage_FireBullet{
						FireBullet: &pb.FireBulletRequest{
							Direction: float64(idx),
							Power:     10,
						},
					},
				}

				action := &GameActionMessage{
					Client:    client,
					RoomID:    "test_room",
					Action:    "fire_bullet",
					Data:      fireMsg,
					Timestamp: time.Now(),
				}

				// 這應該不會阻塞,因為通道有緩衝 (100)
				select {
				case hub.gameAction <- action:
					// 成功發送
				case <-time.After(100 * time.Millisecond):
					// 如果阻塞,測試會失敗
					assert.Fail(t, "Sending to gameAction channel blocked")
				}
			}(i)
		}

		// 等待所有 goroutine 完成發送
		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()

		select {
		case <-done:
			duration := time.Since(start)
			// 發送到緩衝通道應該很快 (50個消息,緩衝區100)
			assert.Less(t, duration, 500*time.Millisecond, "Concurrent game actions should not block")
		case <-time.After(2 * time.Second):
			t.Fatal("Concurrent game actions caused timeout/deadlock")
		}
	})

	t.Run("RoomManager can receive actions without blocking", func(t *testing.T) {
		hub := NewHub(gameUsecase, playerUsecase, log)
		go hub.Run()
		defer hub.Stop()

		// 創建房間管理器
		rm := NewRoomManager("test_room_simple", gameUsecase, hub, log)
		go rm.Run()
		defer rm.Stop()

		// 創建客戶端
		client := &Client{
			ID:       "test_player_rm",
			PlayerID: 1,
			send:     make(chan []byte, 256),
			hub:      hub,
			logger:   log,
			RoomID:   "test_room_simple",
		}

		// 測試快速發送多個操作不會阻塞
		numActions := 20
		start := time.Now()
		for i := 0; i < numActions; i++ {
			fireMsg := &pb.GameMessage{
				Type: pb.MessageType_FIRE_BULLET,
				Data: &pb.GameMessage_FireBullet{
					FireBullet: &pb.FireBulletRequest{
						Direction: float64(i * 10),
						Power:     10,
					},
				},
			}

			action := &GameActionMessage{
				Client:    client,
				RoomID:    "test_room_simple",
				Action:    "fire_bullet",
				Data:      fireMsg,
				Timestamp: time.Now(),
			}

			// 這應該不會阻塞,因為通道有緩衝
			rm.HandleGameAction(action)
		}
		duration := time.Since(start)

		// 發送到緩衝通道應該很快
		assert.Less(t, duration, 100*time.Millisecond, "Sending actions should not block")
	})
}

