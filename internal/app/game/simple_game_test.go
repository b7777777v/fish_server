package game

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/pkg/token"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========================================
// Mocks
// ========================================

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
func (m *MockGameRepo) GetAllFishTypes(ctx context.Context) ([]*game.FishType, error) {
	// Return a default fish type for tests that might need it
	return []*game.FishType{{ID: 1, Name: "Test Fish"}}, nil
}
func (m *MockGameRepo) SaveFishTypeCache(ctx context.Context, ft *game.FishType) error {
	return nil
}

type MockPlayerRepo struct{}

func (m *MockPlayerRepo) GetPlayer(ctx context.Context, playerID int64) (*game.Player, error) {
	return &game.Player{ID: playerID, UserID: playerID, Nickname: "TestPlayer", Balance: 10000, WalletID: 1, Status: game.PlayerStatusIdle}, nil
}
func (m *MockPlayerRepo) UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error {
	return nil
}
func (m *MockPlayerRepo) UpdatePlayerStatus(ctx context.Context, playerID int64, status game.PlayerStatus) error {
	return nil
}

type MockWalletRepo struct{}

func (m *MockWalletRepo) FindByID(ctx context.Context, id uint) (*wallet.Wallet, error) {
	return &wallet.Wallet{ID: id, UserID: uint(id), Balance: 1000.00, Currency: "CNY", Status: 1}, nil
}
func (m *MockWalletRepo) FindByUserID(ctx context.Context, userID uint, currency string) (*wallet.Wallet, error) {
	return &wallet.Wallet{ID: 1, UserID: userID, Balance: 1000.00, Currency: currency, Status: 1}, nil
}
func (m *MockWalletRepo) FindAllByUserID(ctx context.Context, userID uint) ([]*wallet.Wallet, error) {
	return []*wallet.Wallet{{ID: 1, UserID: userID, Balance: 1000.00, Currency: "CNY", Status: 1}}, nil
}
func (m *MockWalletRepo) Create(ctx context.Context, w *wallet.Wallet) error {
	return nil
}
func (m *MockWalletRepo) Update(ctx context.Context, w *wallet.Wallet) error {
	return nil
}
func (m *MockWalletRepo) Deposit(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockWalletRepo) Withdraw(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	return nil
}
func (m *MockWalletRepo) CreateTransaction(ctx context.Context, tx *wallet.Transaction) error {
	return nil
}
func (m *MockWalletRepo) FindTransactionsByWalletID(ctx context.Context, walletID uint, limit, offset int) ([]*wallet.Transaction, error) {
	return []*wallet.Transaction{}, nil
}

type MockGameRecordRepo struct {
	mock.Mock
}

func (m *MockGameRecordRepo) Create(ctx context.Context, record *game.GameRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockGameRecordRepo) Update(ctx context.Context, record *game.GameRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockGameRecordRepo) FindByID(ctx context.Context, id int64) (*game.GameRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.GameRecord), args.Error(1)
}

func (m *MockGameRecordRepo) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*game.GameRecord, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game.GameRecord), args.Error(1)
}

func (m *MockGameRecordRepo) FindBySessionID(ctx context.Context, sessionID string) ([]*game.GameRecord, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game.GameRecord), args.Error(1)
}

func (m *MockGameRecordRepo) FindActiveByUserID(ctx context.Context, userID int64) (*game.GameRecord, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.GameRecord), args.Error(1)
}

func (m *MockGameRecordRepo) GetUserTotalStats(ctx context.Context, userID int64) (*game.UserGameStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.UserGameStats), args.Error(1)
}

type MockInventoryRepo struct {
	mu          sync.RWMutex
	inventories map[string]*game.Inventory
}

func NewMockInventoryRepo() *MockInventoryRepo {
	return &MockInventoryRepo{inventories: make(map[string]*game.Inventory)}
}
func (r *MockInventoryRepo) GetInventory(ctx context.Context, inventoryID string) (*game.Inventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if inv, ok := r.inventories[inventoryID]; ok {
		invCopy := *inv
		return &invCopy, nil
	}
	return &game.Inventory{ID: inventoryID}, nil
}
func (r *MockInventoryRepo) SaveInventory(ctx context.Context, inventory *game.Inventory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	invCopy := *inventory
	r.inventories[inventory.ID] = &invCopy
	return nil
}
func (r *MockInventoryRepo) GetAllInventories(ctx context.Context) (map[string]*game.Inventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inventoriesCopy := make(map[string]*game.Inventory, len(r.inventories))
	for id, inv := range r.inventories {
		invCopy := *inv
		inventoriesCopy[id] = &invCopy
	}
	return inventoriesCopy, nil
}

type MockBizPlayerRepo struct {
	mu        sync.Mutex
	players   map[string]*player.Player
	idCounter uint
}

func NewMockBizPlayerRepo() *MockBizPlayerRepo {
	return &MockBizPlayerRepo{players: make(map[string]*player.Player)}
}

func (m *MockBizPlayerRepo) FindByUsername(ctx context.Context, username string) (*player.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.players[username]; ok {
		return p, nil
	}
	return nil, nil // Return nil, nil for not found
}

func (m *MockBizPlayerRepo) Create(ctx context.Context, p *player.Player) (*player.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idCounter++
	p.ID = m.idCounter
	m.players[p.Username] = p
	return p, nil
}

// ========================================
// Test Main Function
// ========================================

func TestSimpleGameComponents(t *testing.T) {
	// 1. Setup a complete but mocked dependency chain
	log := logger.New(os.Stdout, "debug", "console")
	gameRepo := &MockGameRepo{}
	playerRepo := &MockPlayerRepo{}
	inventoryRepo := NewMockInventoryRepo()

	// Setup for biz/player
	bizPlayerRepo := NewMockBizPlayerRepo()
	jwtConfig := &conf.JWT{Secret: "test-secret-for-test", Expire: 3600}
	tokenHelper := token.NewTokenHelper(jwtConfig)
	playerUsecase := player.NewPlayerUsecase(bizPlayerRepo, tokenHelper, log)

	// Create a test room config
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
	assert.NoError(t, err)

	rtpController := game.NewRTPController(inventoryManager, log)
	roomManager := game.NewRoomManager(log, spawner, mathModel, inventoryManager, rtpController)
	walletRepo := &MockWalletRepo{}
	walletUC := wallet.NewWalletUsecase(walletRepo, log)

	// Create MockGameRecordRepo
	gameRecordRepo := &MockGameRecordRepo{}
	gameRecordRepo.On("FindActiveByUserID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
	gameRecordRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	gameRecordRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

	gameUsecase := game.NewGameUsecase(gameRepo, playerRepo, gameRecordRepo, walletUC, roomManager, spawner, mathModel, inventoryManager, rtpController, log)

	// 2. Run tests for the app/game layer components
	t.Run("Test Hub", func(t *testing.T) {
		hub := NewHub(gameUsecase, playerUsecase, log)
		go hub.Run()
		defer hub.Stop()

		client := &Client{
			ID:       "test_client_1",
			PlayerID: 1,
			send:     make(chan []byte, 256),
			hub:      hub,
			logger:   log,
		}

		hub.register <- client
		time.Sleep(100 * time.Millisecond)
		stats := hub.GetStats()
		assert.Equal(t, 1, stats.ActiveConnections)

		hub.unregister <- client
		time.Sleep(100 * time.Millisecond)
		stats = hub.GetStats()
		assert.Equal(t, 0, stats.ActiveConnections)
	})

	t.Run("Test MessageHandler", func(t *testing.T) {
		hub := NewHub(gameUsecase, playerUsecase, log)
		go hub.Run()
		defer hub.Stop()

		messageHandler := NewMessageHandler(gameUsecase, hub, log)

		client := &Client{
			ID:       "test_player_1",
			PlayerID: 1,
			send:     make(chan []byte, 256),
			hub:      hub,
			logger:   log,
		}

		hub.register <- client
		time.Sleep(50 * time.Millisecond)

		heartbeatMsg := &pb.GameMessage{
			Type: pb.MessageType_HEARTBEAT,
			Data: &pb.GameMessage_Heartbeat{Heartbeat: &pb.HeartbeatMessage{Timestamp: time.Now().Unix()}},
		}

		messageHandler.HandleMessage(client, heartbeatMsg)

		select {
		case <-client.send:
			// Success
		case <-time.After(1 * time.Second):
			t.Error("Did not receive heartbeat response in time")
		}
	})

	t.Run("Test Room Operations via MessageHandler", func(t *testing.T) {
		// Create a fresh usecase for this test to avoid state leakage
		roomManager := game.NewRoomManager(log, spawner, mathModel, inventoryManager, rtpController)
		walletRepo := &MockWalletRepo{}
		walletUC := wallet.NewWalletUsecase(walletRepo, log)

		// Create MockGameRecordRepo
		gameRecordRepo2 := &MockGameRecordRepo{}
		gameRecordRepo2.On("FindActiveByUserID", mock.Anything, mock.Anything).Return(nil, errors.New("not found"))
		gameRecordRepo2.On("Create", mock.Anything, mock.Anything).Return(nil)
		gameRecordRepo2.On("Update", mock.Anything, mock.Anything).Return(nil)

		gameUsecase := game.NewGameUsecase(gameRepo, playerRepo, gameRecordRepo2, walletUC, roomManager, spawner, mathModel, inventoryManager, rtpController, log)
		room, err := gameUsecase.CreateRoom(context.Background(), "test_room_001", 4)
		assert.NoError(t, err)

		hub := NewHub(gameUsecase, playerUsecase, log)
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

		// Use the actual room ID from the created room
		joinMsg := &pb.GameMessage{
			Type: pb.MessageType_JOIN_ROOM,
			Data: &pb.GameMessage_JoinRoom{JoinRoom: &pb.JoinRoomRequest{RoomId: room.ID}},
		}

		messageHandler.HandleMessage(client, joinMsg)
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, room.ID, client.RoomID)

		leaveMsg := &pb.GameMessage{
			Type: pb.MessageType_LEAVE_ROOM,
			Data: &pb.GameMessage_LeaveRoom{LeaveRoom: &pb.LeaveRoomRequest{}},
		}

		messageHandler.HandleMessage(client, leaveMsg)
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, "", client.RoomID)
	})
}
