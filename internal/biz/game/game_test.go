package game

import (
	"context"
	"os"
	"sync"
	"testing"

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
	return &Player{ID: playerID, UserID: playerID, Nickname: "TestPlayer", Balance: 100000, Status: PlayerStatusIdle}, nil
}
func (m *MockPlayerRepo) UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error {
	return nil
}
func (m *MockPlayerRepo) UpdatePlayerStatus(ctx context.Context, playerID int64, status PlayerStatus) error {
	return nil
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
	inventoryRepo := NewMockInventoryRepo()

	spawner := NewFishSpawner(log)
	mathModel := NewMathModel(log)
	inventoryManager, err := NewInventoryManager(inventoryRepo, log)
	assert.NoError(t, err)

	rtpController := NewRTPController(inventoryManager, log)
	roomManager := NewRoomManager(log, spawner, mathModel, inventoryManager, rtpController)
	gameUsecase := NewGameUsecase(gameRepo, playerRepo, roomManager, spawner, mathModel, inventoryManager, rtpController, log)

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
		inv := te.inventoryManager.GetInventory(RoomTypeNovice)
		inv.TotalIn = 100000
		inv.TotalOut = 110000 // RTP is 110%
		te.inventoryRepo.SaveInventory(te.ctx, inv)

		// With high RTP, the chance is reduced, not zero. We test this by running it many times.
		wins := 0
		for i := 0; i < 1000; i++ {
			if te.rtpController.ApproveKill(room.Type, room.Config.TargetRTP, fish.Value) {
				wins++
			}
		}
		// Base hit rate is high (e.g., 80%), adjusted should be lower.
		assert.Less(t, wins, 950, "Wins should be significantly less than base hit rate when RTP is high")
		te.log.Infof("High RTP test: %d wins in 1000 trials", wins)
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
	bullet, err := te.gameUsecase.FireBullet(te.ctx, room.ID, playerID, 1.0, 10)
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
	inv.TotalIn = 10000
	inv.TotalOut = 1000 // RTP = 10%
	te.inventoryRepo.SaveInventory(te.ctx, inv)

	hitResult, err := te.gameUsecase.HitFish(te.ctx, room.ID, bullet.ID, firstFish.ID)
	assert.NoError(t, err)
	assert.True(t, hitResult.Success, "Hit should be successful when RTP is very low")

	// Check that the win was recorded
	inv = te.inventoryManager.GetInventory(RoomTypeNovice)
	assert.Equal(t, int64(10000), inv.TotalIn)
	assert.Equal(t, int64(1000)+hitResult.Reward, inv.TotalOut)
}
