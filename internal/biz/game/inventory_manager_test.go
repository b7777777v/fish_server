package game_test


import "github.com/b7777777v/fish_server/internal/biz/game"
import (
	"testing"

	"github.com/b7777777v/fish_server/internal/testing/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestInventoryManager_AddBet tests adding bets to inventory
func TestInventoryManager_AddBet(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Setup mocks
	initialInv := testhelper.NewTestInventory("novice", 0, 0)
	env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
		Return(initialInv, nil).Maybe()
	env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
		Return(nil).Maybe()

	t.Run("add single bet", func(t *testing.T) {
		env.InventoryManager.AddBet(game.RoomTypeNovice, 100)

		inv := env.InventoryManager.GetInventory(game.RoomTypeNovice)
		assert.Equal(t, int64(100), inv.TotalIn)
		assert.Equal(t, int64(0), inv.TotalOut)
		assert.Equal(t, 0.0, inv.CurrentRTP)
	})

	t.Run("add multiple bets", func(t *testing.T) {
		env.InventoryManager.AddBet(game.RoomTypeNovice, 50)
		env.InventoryManager.AddBet(game.RoomTypeNovice, 75)
		env.InventoryManager.AddBet(game.RoomTypeNovice, 125)

		inv := env.InventoryManager.GetInventory(game.RoomTypeNovice)
		assert.Equal(t, int64(350), inv.TotalIn) // 100 + 50 + 75 + 125
	})
}

// TestInventoryManager_AddWin tests adding wins to inventory
func TestInventoryManager_AddWin(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	initialInv := testhelper.NewTestInventory("intermediate", 1000, 0)
	env.InventoryRepo.On("GetInventory", env.Ctx, "intermediate").
		Return(initialInv, nil).Maybe()
	env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
		Return(nil).Maybe()

	// Add some initial bets
	env.InventoryManager.AddBet(game.RoomTypeIntermediate, 1000)

	t.Run("add single win", func(t *testing.T) {
		env.InventoryManager.AddWin(game.RoomTypeIntermediate, 500)

		inv := env.InventoryManager.GetInventory(game.RoomTypeIntermediate)
		assert.Equal(t, int64(2000), inv.TotalIn) // 1000 initial + 1000 bet
		assert.Equal(t, int64(500), inv.TotalOut)
		assert.Equal(t, 0.25, inv.CurrentRTP) // 500/2000 = 0.25
	})

	t.Run("add multiple wins", func(t *testing.T) {
		env.InventoryManager.AddWin(game.RoomTypeIntermediate, 300)
		env.InventoryManager.AddWin(game.RoomTypeIntermediate, 200)

		inv := env.InventoryManager.GetInventory(game.RoomTypeIntermediate)
		assert.Equal(t, int64(1000), inv.TotalOut) // 500 + 300 + 200
		assert.Equal(t, 0.5, inv.CurrentRTP) // 1000/2000 = 0.5
	})
}

// TestInventoryManager_RTPCalculation tests RTP calculation
func TestInventoryManager_RTPCalculation(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	tests := []struct {
		name        string
		totalIn     int64
		totalOut    int64
		expectedRTP float64
	}{
		{
			name:        "50% RTP",
			totalIn:     10000,
			totalOut:    5000,
			expectedRTP: 0.5,
		},
		{
			name:        "96% RTP",
			totalIn:     100000,
			totalOut:    96000,
			expectedRTP: 0.96,
		},
		{
			name:        "110% RTP (over payout)",
			totalIn:     50000,
			totalOut:    55000,
			expectedRTP: 1.1,
		},
		{
			name:        "zero RTP",
			totalIn:     10000,
			totalOut:    0,
			expectedRTP: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roomType := game.RoomType("test_" + tt.name)
			initialInv := testhelper.NewTestInventory(string(roomType), tt.totalIn, tt.totalOut)
			env.InventoryRepo.On("GetInventory", env.Ctx, string(roomType)).
				Return(initialInv, nil).Once()

			inv := env.InventoryManager.GetInventory(roomType)
			assert.Equal(t, tt.totalIn, inv.TotalIn)
			assert.Equal(t, tt.totalOut, inv.TotalOut)
			assert.InDelta(t, tt.expectedRTP, inv.CurrentRTP, 0.001)
		})
	}
}

// TestInventoryManager_GetInventory tests inventory retrieval
func TestInventoryManager_GetInventory(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("get existing inventory", func(t *testing.T) {
		existingInv := testhelper.NewTestInventory("novice", 5000, 4800)
		env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
			Return(existingInv, nil).Once()

		inv := env.InventoryManager.GetInventory(game.RoomTypeNovice)
		assert.NotNil(t, inv)
		assert.Equal(t, int64(5000), inv.TotalIn)
		assert.Equal(t, int64(4800), inv.TotalOut)
		assert.Equal(t, 0.96, inv.CurrentRTP)
	})

	t.Run("get new inventory", func(t *testing.T) {
		newInv := testhelper.NewTestInventory("new_room", 0, 0)
		env.InventoryRepo.On("GetInventory", env.Ctx, "new_room").
			Return(newInv, nil).Once()

		inv := env.InventoryManager.GetInventory(game.RoomType("new_room"))
		assert.NotNil(t, inv)
		assert.Equal(t, int64(0), inv.TotalIn)
		assert.Equal(t, int64(0), inv.TotalOut)
	})
}

// TestInventoryManager_ConcurrentOperations tests thread-safe operations
func TestInventoryManager_ConcurrentOperations(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	initialInv := testhelper.NewTestInventory("concurrent", 0, 0)
	env.InventoryRepo.On("GetInventory", env.Ctx, "concurrent").
		Return(initialInv, nil).Maybe()
	env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
		Return(nil).Maybe()

	t.Run("concurrent bets", func(t *testing.T) {
		roomType := game.RoomType("concurrent")
		iterations := 100

		// Simulate concurrent bets
		done := make(chan bool)
		for i := 0; i < iterations; i++ {
			go func() {
				env.InventoryManager.AddBet(roomType, 10)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < iterations; i++ {
			<-done
		}

		// Verify total
		inv := env.InventoryManager.GetInventory(roomType)
		assert.Equal(t, int64(iterations*10), inv.TotalIn)
	})
}

// TestInventoryManager_EdgeCases tests edge cases
func TestInventoryManager_EdgeCases(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("add zero bet", func(t *testing.T) {
		initialInv := testhelper.NewTestInventory("zero_bet", 100, 50)
		env.InventoryRepo.On("GetInventory", env.Ctx, "zero_bet").
			Return(initialInv, nil).Maybe()
		env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
			Return(nil).Maybe()

		env.InventoryManager.AddBet(game.RoomType("zero_bet"), 0)

		inv := env.InventoryManager.GetInventory(game.RoomType("zero_bet"))
		assert.Equal(t, int64(100), inv.TotalIn) // Should remain unchanged
	})

	t.Run("add negative bet", func(t *testing.T) {
		initialInv := testhelper.NewTestInventory("negative_bet", 100, 50)
		env.InventoryRepo.On("GetInventory", env.Ctx, "negative_bet").
			Return(initialInv, nil).Maybe()
		env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
			Return(nil).Maybe()

		// Should handle gracefully (or reject, depending on implementation)
		assert.NotPanics(t, func() {
			env.InventoryManager.AddBet(game.RoomType("negative_bet"), -100)
		})
	})

	t.Run("RTP calculation with zero total in", func(t *testing.T) {
		zeroInv := testhelper.NewTestInventory("zero_in", 0, 100)
		env.InventoryRepo.On("GetInventory", env.Ctx, "zero_in").
			Return(zeroInv, nil).Once()

		inv := env.InventoryManager.GetInventory(game.RoomType("zero_in"))
		// RTP should be 0 or infinity, implementation should handle gracefully
		assert.NotPanics(t, func() {
			_ = inv.CurrentRTP
		})
	})
}

// TestInventoryManager_Integration tests inventory in game flow
func TestInventoryManager_Integration(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	initialInv := testhelper.NewTestInventory("integration", 0, 0)
	env.InventoryRepo.On("GetInventory", env.Ctx, "integration").
		Return(initialInv, nil).Maybe()
	env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
		Return(nil).Maybe()

	roomType := game.RoomType("integration")

	t.Run("simulate game session", func(t *testing.T) {
		// Player 1 bets and wins
		env.InventoryManager.AddBet(roomType, 100)
		env.InventoryManager.AddWin(roomType, 50)

		// Player 2 bets and loses
		env.InventoryManager.AddBet(roomType, 100)

		// Player 3 bets and wins big
		env.InventoryManager.AddBet(roomType, 100)
		env.InventoryManager.AddWin(roomType, 150)

		// Check final state
		inv := env.InventoryManager.GetInventory(roomType)
		assert.Equal(t, int64(300), inv.TotalIn)  // 100 + 100 + 100
		assert.Equal(t, int64(200), inv.TotalOut) // 50 + 150
		assert.InDelta(t, 0.6667, inv.CurrentRTP, 0.01) // 200/300 ≈ 0.67
	})
}
