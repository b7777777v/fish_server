// Package examples provides example tests demonstrating the testing framework usage.
// This is not an actual test package but serves as documentation and reference.
package examples

import (
	"testing"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/testing/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ========================================
// Example Tests Using New Mock Architecture
// ========================================

// ExampleBasicTest demonstrates basic usage of the testing framework
func ExampleBasicTest(t *testing.T) {
	// Setup test environment with default mocks
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Your test logic here
	room, err := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
	assert.NoError(t, err)
	assert.NotNil(t, room)
}

// ExampleCustomMockBehavior demonstrates how to customize mock behavior
func ExampleCustomMockBehavior(t *testing.T) {
	// Create environment without default mocks
	env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
		SkipDefaultMocks: true, // We'll set up custom mocks
	})
	defer env.AssertExpectations(t)

	playerID := int64(123)

	// Custom mock: return specific player
	customPlayer := &game.Player{
		ID:       playerID,
		UserID:   playerID,
		Nickname: "CustomTestPlayer",
		Balance:  50000, // Custom balance
		WalletID: 5,     // Custom wallet ID
		Status:   game.PlayerStatusIdle,
	}
	env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(customPlayer, nil).Once()

	// Act: Get player
	player, err := env.PlayerRepo.GetPlayer(env.Ctx, playerID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "CustomTestPlayer", player.Nickname)
	assert.Equal(t, int64(50000), player.Balance)
	assert.Equal(t, uint(5), player.WalletID)
}

// ExampleMockExpectationVerification demonstrates mock expectation verification
func ExampleMockExpectationVerification(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)

	playerID := int64(1)
	newBalance := int64(90000)

	// Setup expectation: UpdatePlayerBalance should be called exactly once
	env.PlayerRepo.On("UpdatePlayerBalance", env.Ctx, playerID, newBalance).
		Return(nil).Once()

	// Act
	err := env.PlayerRepo.UpdatePlayerBalance(env.Ctx, playerID, newBalance)
	assert.NoError(t, err)

	// Verify: This will fail if UpdatePlayerBalance wasn't called exactly once
	env.AssertExpectations(t)
}

// ExampleRTPControllerTest demonstrates RTP controller testing
func ExampleRTPControllerTest(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Create test fixtures
	fixtures := testhelper.NewFishTypeFixtures()

	t.Run("RTP below target - should force win", func(t *testing.T) {
		// Setup: Create inventory with low RTP
		lowRTPInventory := testhelper.NewTestInventory("novice", 10000, 8000) // RTP = 80%

		// Mock expectation
		env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
			Return(lowRTPInventory, nil).Once()

		// Act: Check if kill is approved
		targetRTP := 0.95 // 95%
		fishValue := fixtures.SmallFish.BaseValue
		win := env.RTPController.ApproveKill(game.RoomTypeNovice, targetRTP, fishValue)

		// Assert: Should approve kill when RTP is low
		assert.True(t, win, "Should force a win when RTP is below target (80% < 95%)")
	})
}

// ExampleGameFlowTest demonstrates a complete game flow test
func ExampleGameFlowTest(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	playerID := int64(1)

	// Setup player mock
	testPlayer := testhelper.NewTestPlayerWithBalance(playerID, 100000)
	env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil)

	// Setup inventory mock
	inventory := testhelper.NewTestInventory("novice", 0, 0)
	env.InventoryRepo.On("GetInventory", env.Ctx, string(game.RoomTypeNovice)).
		Return(inventory, nil).Maybe()
	env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
		Return(nil).Maybe()

	// 1. Create Room
	room, err := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
	assert.NoError(t, err)

	// 2. Join Room
	err = env.GameUsecase.JoinRoom(env.Ctx, room.ID, playerID)
	assert.NoError(t, err)

	// 3. Fire Bullet
	bullet, err := env.GameUsecase.FireBullet(env.Ctx, room.ID, playerID, 1.0, 10, game.Position{X: 600, Y: 750}, 0)
	assert.NoError(t, err)
	assert.NotNil(t, bullet)

	// Verify bet was recorded
	inv := env.InventoryManager.GetInventory(game.RoomTypeNovice)
	assert.Equal(t, bullet.Cost, inv.TotalIn)
}

// ExampleUsingFixtures demonstrates using test data fixtures
func ExampleUsingFixtures(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Get standard fish type fixtures
	fixtures := testhelper.NewFishTypeFixtures()

	// Create fish using fixtures
	smallFish := testhelper.NewTestFish(1, fixtures.SmallFish)
	bossFish := testhelper.NewTestFish(2, fixtures.BossFish)

	assert.Equal(t, int64(10), smallFish.Value)
	assert.Equal(t, int64(1000), bossFish.Value)

	// Get all fish types
	allFishTypes := fixtures.AllFishTypes()
	assert.Len(t, allFishTypes, 4)

	// Use in mock expectations
	env.GameRepo.On("GetAllFishTypes", env.Ctx).
		Return(allFishTypes, nil)
}

// ExampleCustomRoomConfig demonstrates using custom room configuration
func ExampleCustomRoomConfig(t *testing.T) {
	customConfig := game.RoomConfig{
		MaxPlayers:   8,  // Custom max players
		MinBet:       10,
		MaxBet:       500,
		MinFishCount: 20,
		MaxFishCount: 40,
		TargetRTP:    0.95,
	}

	env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
		RoomConfig: &customConfig,
	})
	defer env.AssertExpectations(t)

	// Test with custom configuration
	assert.Equal(t, int32(8), env.RoomConfig.MaxPlayers)
	assert.Equal(t, int32(40), env.RoomConfig.MaxFishCount)
}

// ExampleErrorInjection demonstrates error injection testing
func ExampleErrorInjection(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
		SkipDefaultMocks: true,
	})
	defer env.AssertExpectations(t)

	// Simulate database error
	env.PlayerRepo.On("GetPlayer", env.Ctx, int64(999)).
		Return(nil, assert.AnError)

	// Test error handling
	player, err := env.PlayerRepo.GetPlayer(env.Ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, player)
}

// ExampleSubtests demonstrates organizing tests with subtests
func ExampleSubtests(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("create room", func(t *testing.T) {
		room, err := env.RoomManager.CreateRoom(game.RoomTypeNovice, 1)
		assert.NoError(t, err)
		assert.NotNil(t, room)
	})

	t.Run("create advanced room", func(t *testing.T) {
		room, err := env.RoomManager.CreateRoom(game.RoomTypeAdvanced, 1)
		assert.NoError(t, err)
		assert.Equal(t, game.RoomTypeAdvanced, room.Type)
	})
}

// ExampleDynamicReturnValues demonstrates using functions for dynamic return values
func ExampleDynamicReturnValues(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
		SkipDefaultMocks: true,
	})
	defer env.AssertExpectations(t)

	// Dynamic return values based on input
	env.PlayerRepo.On("GetPlayer", mock.Anything, mock.Anything).
		Return(func(ctx any, playerID int64) *game.Player {
			return &game.Player{
				ID:       playerID,
				Nickname: "Player_" + string(rune(playerID)),
				Balance:  playerID * 1000,
			}
		}, nil)

	player1, _ := env.PlayerRepo.GetPlayer(env.Ctx, 1)
	player2, _ := env.PlayerRepo.GetPlayer(env.Ctx, 2)

	assert.Equal(t, int64(1000), player1.Balance)
	assert.Equal(t, int64(2000), player2.Balance)
}
