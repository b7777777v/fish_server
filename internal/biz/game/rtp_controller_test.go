package game_test


import "github.com/b7777777v/fish_server/internal/biz/game"
import (
	"testing"

	"github.com/b7777777v/fish_server/internal/testing/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestRTPController_ApproveKill tests RTP-based kill approval
func TestRTPController_ApproveKill(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("force win when RTP is below target", func(t *testing.T) {
		// Setup: Low RTP inventory (80% vs 95% target)
		lowRTPInv := testhelper.NewTestInventory("novice", 10000, 8000)
		env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
			Return(lowRTPInv, nil).Once()

		// Act
		approved := env.RTPController.ApproveKill(game.RoomTypeNovice, 0.95, 100)

		// Assert: Should force win when RTP is low
		assert.True(t, approved, "Should approve kill when RTP (80%) is below target (95%)")
	})

	t.Run("reduce win rate when RTP is above target", func(t *testing.T) {
		// Setup: High RTP inventory (110% vs 95% target)
		highRTPInv := testhelper.NewTestInventory("advanced", 200000, 220000)
		env.InventoryRepo.On("GetInventory", env.Ctx, "advanced").
			Return(highRTPInv, nil).Times(100)

		// Act: Run multiple trials
		wins := 0
		for i := 0; i < 100; i++ {
			if env.RTPController.ApproveKill(game.RoomTypeAdvanced, 0.95, 100) {
				wins++
			}
		}

		// Assert: Win rate should be reduced (not 100%, but not 0%)
		assert.Less(t, wins, 100, "Should not approve all kills when RTP is above target")
		assert.Greater(t, wins, 50, "Should still approve some kills")
	})

	t.Run("normal behavior with balanced RTP", func(t *testing.T) {
		// Setup: Balanced RTP (95% = target)
		balancedInv := testhelper.NewTestInventory("intermediate", 200000, 190000)
		env.InventoryRepo.On("GetInventory", env.Ctx, "intermediate").
			Return(balancedInv, nil).Times(100)

		// Act
		wins := 0
		for i := 0; i < 100; i++ {
			if env.RTPController.ApproveKill(game.RoomTypeIntermediate, 0.95, 100) {
				wins++
			}
		}

		// Assert: Win rate should be around base hit rate
		assert.Greater(t, wins, 0, "Should approve some kills")
		assert.Less(t, wins, 100, "Should not approve all kills")
	})

	t.Run("handle low total bet scenario", func(t *testing.T) {
		// Setup: Very low total bet (less than threshold)
		lowBetInv := testhelper.NewTestInventory("novice", 100, 50)
		env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
			Return(lowBetInv, nil).Once()

		// Act
		_ = env.RTPController.ApproveKill(game.RoomTypeNovice, 0.95, 100)

		// Assert: Should use default logic without RTP adjustment
		assert.NotPanics(t, func() {
			env.RTPController.ApproveKill(game.RoomTypeNovice, 0.95, 100)
		})
	})
}

// TestRTPController_AdjustReward tests reward adjustment based on RTP
// TestRTPController_AdjustReward is commented out - AdjustReward method does not exist
// func TestRTPController_AdjustReward(t *testing.T) { ... }

// TestRTPController_Integration tests RTP controller in a game scenario
func TestRTPController_Integration(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Setup: Start with low RTP
	inventory := testhelper.NewTestInventory("novice", 10000, 5000) // 50% RTP
	env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
		Return(inventory, nil).Maybe()
	env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
		Return(nil).Maybe()

	t.Run("RTP should influence game outcomes", func(t *testing.T) {
		// Simulate multiple bets
		env.InventoryManager.AddBet(game.RoomTypeNovice, 1000)
		env.InventoryManager.AddBet(game.RoomTypeNovice, 1000)
		env.InventoryManager.AddBet(game.RoomTypeNovice, 1000)

		// Check current RTP
		inv := env.InventoryManager.GetInventory(game.RoomTypeNovice)
		assert.Equal(t, int64(13000), inv.TotalIn) // 10000 + 3000

		// When RTP is low, should approve more kills
		approvals := 0
		for i := 0; i < 10; i++ {
			if env.RTPController.ApproveKill(game.RoomTypeNovice, 0.96, 100) {
				approvals++
			}
		}

		// Should approve most kills when RTP is low
		assert.Greater(t, approvals, 7, "Should approve most kills when RTP is low")
	})
}

// TestRTPController_EdgeCases tests edge cases and error scenarios
func TestRTPController_EdgeCases(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
		SkipDefaultMocks: true,
	})
	defer env.AssertExpectations(t)

	t.Run("handle zero total bet", func(t *testing.T) {
		zeroInv := testhelper.NewTestInventory("novice", 0, 0)
		env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
			Return(zeroInv, nil).Once()

		// Should not panic
		assert.NotPanics(t, func() {
			env.RTPController.ApproveKill(game.RoomTypeNovice, 0.96, 100)
		})
	})

	t.Run("handle extremely high RTP", func(t *testing.T) {
		// RTP = 500% (payout way too high)
		extremeInv := testhelper.NewTestInventory("vip", 10000, 50000)
		env.InventoryRepo.On("GetInventory", env.Ctx, "vip").
			Return(extremeInv, nil).Times(10)

		// Should heavily restrict wins
		wins := 0
		for i := 0; i < 10; i++ {
			if env.RTPController.ApproveKill(game.RoomTypeVIP, 0.94, 1000) {
				wins++
			}
		}

		assert.LessOrEqual(t, wins, 3, "Should heavily restrict wins when RTP is extremely high")
	})

// 	t.Run("handle zero reward", func(t *testing.T) {
// 		balancedInv := testhelper.NewTestInventory("novice", 10000, 9600)
// 		env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
// 			Return(balancedInv, nil).Once()
// 
// 		adjustedReward := env.RTPController.AdjustReward(game.RoomTypeNovice, 0.96, 0)
// 		assert.Equal(t, int64(0), adjustedReward, "Zero reward should remain zero")
// 	})
}
