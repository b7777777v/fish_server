package game

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// ========================================
// Mocks
// ========================================

type MockGameRepo struct{}

func (m *MockGameRepo) SaveRoom(ctx context.Context, room *Room) error            { return nil }
func (m *MockGameRepo) GetRoom(ctx context.Context, roomID string) (*Room, error) { return nil, nil }
func (m *MockGameRepo) ListRooms(ctx context.Context, roomType RoomType) ([]*Room, error) {
	return []*Room{}, nil
}
func (m *MockGameRepo) DeleteRoom(ctx context.Context, roomID string) error { return nil }
func (m *MockGameRepo) SaveGameStatistics(ctx context.Context, playerID int64, stats *GameStatistics) error {
	return nil
}
func (m *MockGameRepo) GetGameStatistics(ctx context.Context, playerID int64) (*GameStatistics, error) {
	return &GameStatistics{}, nil
}
func (m *MockGameRepo) SaveGameEvent(ctx context.Context, event *GameEvent) error { return nil }
func (m *MockGameRepo) GetGameEvents(ctx context.Context, roomID string, limit int) ([]*GameEvent, error) {
	return []*GameEvent{}, nil
}
func (m *MockGameRepo) GetAllFishTypes(ctx context.Context) ([]*FishType, error) {
	// Return a default fish type for tests that might need it
	return []*FishType{{ID: 1, Name: "Test Fish"}}, nil
}
func (m *MockGameRepo) SaveFishTypeCache(ctx context.Context, ft *FishType) error {
	return nil
}

type MockPlayerRepo struct{}

func (m *MockPlayerRepo) GetPlayer(ctx context.Context, playerID int64) (*Player, error) {
	return &Player{ID: playerID, UserID: playerID, Nickname: "TestPlayer", Balance: 100000, WalletID: 1, Status: PlayerStatusIdle}, nil
}
func (m *MockPlayerRepo) UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error {
	return nil
}
func (m *MockPlayerRepo) UpdatePlayerStatus(ctx context.Context, playerID int64, status PlayerStatus) error {
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

type MockInventoryRepo struct {
	mu          sync.RWMutex
	inventories map[string]*Inventory
}

func NewMockInventoryRepo() *MockInventoryRepo {
	return &MockInventoryRepo{inventories: make(map[string]*Inventory)}
}
func (r *MockInventoryRepo) GetInventory(ctx context.Context, inventoryID string) (*Inventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if inv, ok := r.inventories[inventoryID]; ok {
		invCopy := *inv
		return &invCopy, nil
	}
	return &Inventory{ID: inventoryID}, nil
}
func (r *MockInventoryRepo) SaveInventory(ctx context.Context, inventory *Inventory) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	invCopy := *inventory
	r.inventories[inventory.ID] = &invCopy
	return nil
}
func (r *MockInventoryRepo) GetAllInventories(ctx context.Context) (map[string]*Inventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	inventoriesCopy := make(map[string]*Inventory, len(r.inventories))
	for id, inv := range r.inventories {
		invCopy := *inv
		inventoriesCopy[id] = &invCopy
	}
	return inventoriesCopy, nil
}

// ========================================
// Test Setup Helper
// ========================================

type testEnvironment struct {
	ctx              context.Context
	log              logger.Logger
	gameRepo         *MockGameRepo
	playerRepo       *MockPlayerRepo
	inventoryRepo    *MockInventoryRepo
	spawner          *FishSpawner
	mathModel        *MathModel
	inventoryManager *InventoryManager
	rtpController    *RTPController
	roomManager      *RoomManager
	gameUsecase      *GameUsecase
}

func setupTestEnvironment(t *testing.T) *testEnvironment {
	log := logger.New(os.Stdout, "debug", "console")
	gameRepo := &MockGameRepo{}
	playerRepo := &MockPlayerRepo{}
	walletRepo := &MockWalletRepo{}
	inventoryRepo := NewMockInventoryRepo()

	// Create wallet usecase
	walletUC := wallet.NewWalletUsecase(walletRepo, log)

	// Create a test room config
	testRoomConfig := RoomConfig{
		MinBet:               1,
		MaxBet:               100,
		BulletCostMultiplier: 1.0,
		FishSpawnRate:        0.3,
		MaxFishCount:         20,
		RoomWidth:            1200,
		RoomHeight:           800,
		TargetRTP:            0.96,
	}

	spawner := NewFishSpawner(log, testRoomConfig)
	mathModel := NewMathModel(log)
	inventoryManager, err := NewInventoryManager(inventoryRepo, log)
	assert.NoError(t, err)

	rtpController := NewRTPController(inventoryManager, log)
	roomManager := NewRoomManager(log, spawner, mathModel, inventoryManager, rtpController)
	gameUsecase := NewGameUsecase(gameRepo, playerRepo, walletUC, roomManager, spawner, mathModel, inventoryManager, rtpController, log)

	return &testEnvironment{
		ctx:              context.Background(),
		log:              log,
		gameRepo:         gameRepo,
		playerRepo:       playerRepo,
		inventoryRepo:    inventoryRepo,
		spawner:          spawner,
		mathModel:        mathModel,
		inventoryManager: inventoryManager,
		rtpController:    rtpController,
		roomManager:      roomManager,
		gameUsecase:      gameUsecase,
	}
}

// ========================================
// Tests
// ========================================

func TestRTPController(t *testing.T) {
	te := setupTestEnvironment(t)

	room, err := te.roomManager.CreateRoom(RoomTypeNovice, 1)
	assert.NoError(t, err)
	room.Config.TargetRTP = 0.95 // 95%

	fish := &Fish{ID: 1, Type: te.spawner.GetFishTypes()[0], Health: 1, Value: 100}

	t.Run("RTP below target", func(t *testing.T) {
		inv := te.inventoryManager.GetInventory(RoomTypeNovice)
		inv.TotalIn = 10000
		inv.TotalOut = 8000 // RTP is 80%
		te.inventoryRepo.SaveInventory(te.ctx, inv)

		win := te.rtpController.ApproveKill(room.Type, room.Config.TargetRTP, fish.Value)
		assert.True(t, win, "Should force a win when RTP is low")
	})

	t.Run("RTP above target", func(t *testing.T) {
		// Create a fresh inventory for this test
		inv := te.inventoryManager.GetInventory(RoomTypeAdvanced) // Use different room type
		inv.TotalIn = 200000  // Must be > 100000 to trigger RTP logic
		inv.TotalOut = 220000 // RTP is 110%
		inv.CurrentRTP = 1.10 // Explicitly set the calculated RTP
		te.inventoryRepo.SaveInventory(te.ctx, inv)

		// With high RTP, the chance should be significantly reduced
		wins := 0
		for i := 0; i < 100; i++ { // Reduce test iterations to make it faster
			if te.rtpController.ApproveKill(RoomTypeAdvanced, 0.95, fish.Value) {
				wins++
			}
		}
		// When RTP is above target (110% vs 95%), wins should be much lower
		// The RTP controller should be conservative when payout is already high
		te.log.Infof("High RTP test: %d wins in 100 trials (RTP: 110%% vs target 95%%)", wins)
		
		// Since RTP is significantly above target (110% vs 95%), most kills should be denied
		// With 1.10 RTP vs 0.95 target, denial chance should be (1.10-0.95)/1.10 = ~13.6%
		// So we expect roughly 86-87 wins out of 100, definitely not 100
		assert.Less(t, wins, 100, "Should not approve all kills when RTP is above target")
		assert.Greater(t, wins, 50, "Should still approve some kills even when RTP is high")
	})
}

func TestInventoryManager(t *testing.T) {
	te := setupTestEnvironment(t)

	roomType := RoomTypeNovice
	te.inventoryManager.AddBet(roomType, 100)
	te.inventoryManager.AddWin(roomType, 50)

	inv := te.inventoryManager.GetInventory(roomType)
	assert.Equal(t, int64(100), inv.TotalIn)
	assert.Equal(t, int64(50), inv.TotalOut)
	assert.Equal(t, 0.5, inv.CurrentRTP)
}

func TestGameFlowWithRTP(t *testing.T) {
	te := setupTestEnvironment(t)

	// 1. Create Room & Player
	room, err := te.gameUsecase.CreateRoom(te.ctx, RoomTypeNovice, 1)
	assert.NoError(t, err)

	playerID := int64(1)
	err = te.gameUsecase.JoinRoom(te.ctx, room.ID, playerID)
	assert.NoError(t, err)

	// 2. Fire a bullet
	bullet, err := te.gameUsecase.FireBullet(te.ctx, room.ID, playerID, 1.0, 10, Position{X: 600, Y: 750})
	assert.NoError(t, err)

	// Check that the bet was recorded
	inv := te.inventoryManager.GetInventory(RoomTypeNovice)
	assert.Equal(t, bullet.Cost, inv.TotalIn)

	// 3. Hit a fish
	roomState, _ := te.gameUsecase.GetRoomState(te.ctx, room.ID)
	assert.NotEmpty(t, roomState.Fishes)

	var firstFish *Fish
	for _, f := range roomState.Fishes {
		firstFish = f
		break
	}

	// Force a win scenario by setting RTP low
	inv = te.inventoryManager.GetInventory(RoomTypeNovice)
	inv.TotalIn = 10000
	inv.TotalOut = 1000 // RTP = 10%
	te.inventoryRepo.SaveInventory(te.ctx, inv)

	// Try multiple times since there's still a random component
	var hitResult *HitResult
	var hitSuccess bool
	for i := 0; i < 10; i++ {
		hitResult, err = te.gameUsecase.HitFish(te.ctx, room.ID, bullet.ID, firstFish.ID)
		assert.NoError(t, err)
		if hitResult.Success {
			hitSuccess = true
			break
		}
		// If first attempt fails, create a new bullet for next attempt
		if i < 9 {
			bullet, err = te.gameUsecase.FireBullet(te.ctx, room.ID, playerID, 1.0, 10, Position{X: 600, Y: 750})
			assert.NoError(t, err)
		}
	}
	assert.True(t, hitSuccess, "Hit should be successful when RTP is very low (tried 10 times)")

	// Check that the win was recorded
	inv = te.inventoryManager.GetInventory(RoomTypeNovice)
	assert.Equal(t, int64(10000), inv.TotalIn)
	assert.Equal(t, int64(1000)+hitResult.Reward, inv.TotalOut)
}
