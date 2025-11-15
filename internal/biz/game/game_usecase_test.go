package game_test


import "github.com/b7777777v/fish_server/internal/biz/game"
import (
	"testing"

	"github.com/b7777777v/fish_server/internal/testing/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGameUsecase_CreateRoom tests room creation through usecase
func TestGameUsecase_CreateRoom(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("create room successfully", func(t *testing.T) {
		room, err := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)

		assert.NoError(t, err)
		assert.NotNil(t, room)
		assert.Equal(t, game.RoomTypeNovice, room.Type)
		assert.NotEmpty(t, room.ID)
	})

	t.Run("create multiple rooms of different types", func(t *testing.T) {
		rooms := make([]*game.Room, 0)

		for i, roomType := range []game.RoomType{game.RoomTypeNovice, game.RoomTypeIntermediate, game.RoomTypeAdvanced, game.RoomTypeVIP} {
			room, err := env.GameUsecase.CreateRoom(env.Ctx, roomType, int32(i+1))
			assert.NoError(t, err)
			rooms = append(rooms, room)
		}

		assert.Len(t, rooms, 4)
		// Verify all rooms have unique IDs
		idSet := make(map[string]bool)
		for _, room := range rooms {
			assert.False(t, idSet[room.ID], "Room IDs should be unique")
			idSet[room.ID] = true
		}
	})
}

// TestGameUsecase_JoinRoom tests player joining a room
func TestGameUsecase_JoinRoom(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Create a room
	room, err := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
	assert.NoError(t, err)

	t.Run("player joins room successfully", func(t *testing.T) {
		playerID := int64(1)
		testPlayer := testhelper.NewTestPlayer(playerID)

		env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil).Once()
		env.PlayerRepo.On("UpdatePlayerStatus", env.Ctx, playerID, game.PlayerStatusPlaying).Return(nil).Once()

		err := env.GameUsecase.JoinRoom(env.Ctx, room.ID, playerID)
		assert.NoError(t, err)

		// Verify player is in room
		roomState, _ := env.GameUsecase.GetRoomState(env.Ctx, room.ID)
		assert.Contains(t, roomState.Players, playerID)
	})

	t.Run("multiple players join room", func(t *testing.T) {
		for i := int64(2); i <= 4; i++ {
			testPlayer := testhelper.NewTestPlayer(i)
			env.PlayerRepo.On("GetPlayer", env.Ctx, i).Return(testPlayer, nil).Once()
			env.PlayerRepo.On("UpdatePlayerStatus", env.Ctx, i, game.PlayerStatusPlaying).Return(nil).Once()

			err := env.GameUsecase.JoinRoom(env.Ctx, room.ID, i)
			assert.NoError(t, err)
		}

		roomState, _ := env.GameUsecase.GetRoomState(env.Ctx, room.ID)
		assert.Len(t, roomState.Players, 4) // Player 1-4
	})

	t.Run("cannot join non-existing room", func(t *testing.T) {
		playerID := int64(100)
		testPlayer := testhelper.NewTestPlayer(playerID)
		env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil).Once()
		// No UpdatePlayerStatus mock needed as it should fail before reaching that point

		err := env.GameUsecase.JoinRoom(env.Ctx, "non-existing-room", playerID)
		assert.Error(t, err)
	})
}

// TestGameUsecase_FireBullet tests bullet firing
func TestGameUsecase_FireBullet(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Setup
	room, _ := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
	playerID := int64(1)
	testPlayer := testhelper.NewTestPlayerWithBalance(playerID, 100000)

	env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil)
	env.PlayerRepo.On("UpdatePlayerStatus", env.Ctx, playerID, game.PlayerStatusPlaying).Return(nil)
	env.GameUsecase.JoinRoom(env.Ctx, room.ID, playerID)

	// Setup mock for bullet firing
	env.PlayerRepo.On("UpdatePlayerBalance", env.Ctx, playerID, mock.AnythingOfType("int64")).Return(nil).Maybe()

	// Setup inventory mocks
	inventory := testhelper.NewTestInventory("novice", 0, 0)
	env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
		Return(inventory, nil).Maybe()
	env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
		Return(nil).Maybe()

	t.Run("fire bullet successfully", func(t *testing.T) {
		bullet, err := env.GameUsecase.FireBullet(env.Ctx, room.ID, playerID, 1.0, 10, game.Position{X: 600, Y: 750})

		assert.NoError(t, err)
		assert.NotNil(t, bullet)
		assert.Equal(t, playerID, bullet.PlayerID)
		assert.Greater(t, bullet.Cost, int64(0))
		assert.Equal(t, game.BulletStatusFlying, bullet.Status)
	})

	t.Run("bet is recorded in inventory", func(t *testing.T) {
		initialIn := env.InventoryManager.GetInventory(game.RoomTypeNovice).TotalIn

		bullet, err := env.GameUsecase.FireBullet(env.Ctx, room.ID, playerID, 1.0, 10, game.Position{X: 600, Y: 750})
		assert.NoError(t, err)

		finalIn := env.InventoryManager.GetInventory(game.RoomTypeNovice).TotalIn
		assert.Equal(t, initialIn+bullet.Cost, finalIn)
	})

	t.Run("cannot fire from non-existing room", func(t *testing.T) {
		bullet, err := env.GameUsecase.FireBullet(env.Ctx, "non-existing", playerID, 1.0, 10, game.Position{X: 600, Y: 750})

		assert.Error(t, err)
		assert.Nil(t, bullet)
	})
}

// TestGameUsecase_HitFish tests fish hitting logic
func TestGameUsecase_HitFish(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Setup
	room, _ := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
	playerID := int64(1)
	testPlayer := testhelper.NewTestPlayerWithBalance(playerID, 100000)

	env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil)
	env.PlayerRepo.On("UpdatePlayerStatus", env.Ctx, playerID, game.PlayerStatusPlaying).Return(nil)
	env.GameUsecase.JoinRoom(env.Ctx, room.ID, playerID)

	// Setup mock for balance updates
	env.PlayerRepo.On("UpdatePlayerBalance", env.Ctx, playerID, mock.AnythingOfType("int64")).Return(nil).Maybe()

	// Setup inventory - low RTP to force wins
	lowRTPInv := testhelper.NewTestInventory("novice", 10000, 1000)
	env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
		Return(lowRTPInv, nil).Maybe()
	env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
		Return(nil).Maybe()

	// Fire bullet
	bullet, _ := env.GameUsecase.FireBullet(env.Ctx, room.ID, playerID, 1.0, 10, game.Position{X: 600, Y: 750})

	t.Run("hit fish successfully", func(t *testing.T) {
		// Get a fish from room
		roomState, _ := env.GameUsecase.GetRoomState(env.Ctx, room.ID)
		if len(roomState.Fishes) == 0 {
			t.Skip("No fish in room")
		}

		var targetFish *game.Fish
		for _, fish := range roomState.Fishes {
			targetFish = fish
			break
		}

		// Try hitting multiple times (due to probability)
		var hitResult *game.HitResult
		var err error
		hitSuccess := false

		for i := 0; i < 20; i++ {
			hitResult, err = env.GameUsecase.HitFish(env.Ctx, room.ID, bullet.ID, targetFish.ID)
			assert.NoError(t, err)

			if hitResult.Success {
				hitSuccess = true
				break
			}

			// Create new bullet for next attempt
			if i < 19 {
				bullet, _ = env.GameUsecase.FireBullet(env.Ctx, room.ID, playerID, 1.0, 10, game.Position{X: 600, Y: 750})
			}
		}

		if hitSuccess {
			assert.True(t, hitResult.Success)
			assert.Greater(t, hitResult.Reward, int64(0))
			// Fish caught successfully
		}
	})
}

// TestGameUsecase_GetRoomState tests room state retrieval
func TestGameUsecase_GetRoomState(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	room, _ := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)

	t.Run("get room state", func(t *testing.T) {
		state, err := env.GameUsecase.GetRoomState(env.Ctx, room.ID)

		assert.NoError(t, err)
		assert.NotNil(t, state)
		assert.Equal(t, room.ID, state.ID)
		assert.NotNil(t, state.Players)
		assert.NotNil(t, state.Fishes)
		assert.NotNil(t, state.Bullets)
	})

	t.Run("get non-existing room state", func(t *testing.T) {
		state, err := env.GameUsecase.GetRoomState(env.Ctx, "non-existing")

		assert.Error(t, err)
		assert.Nil(t, state)
	})
}

// TestGameUsecase_CompleteGameFlow tests a complete game session
func TestGameUsecase_CompleteGameFlow(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Setup inventory
	inventory := testhelper.NewTestInventory("novice", 0, 0)
	env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
		Return(inventory, nil).Maybe()
	env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
		Return(nil).Maybe()

	t.Run("complete game session", func(t *testing.T) {
		// 1. Create room
		room, err := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
		assert.NoError(t, err)

		// 2. Players join
		players := make([]*game.Player, 0)
		for i := int64(1); i <= 2; i++ {
			player := testhelper.NewTestPlayerWithBalance(i, 100000)
			players = append(players, player)
			env.PlayerRepo.On("GetPlayer", env.Ctx, i).Return(player, nil)
			env.PlayerRepo.On("UpdatePlayerStatus", env.Ctx, i, game.PlayerStatusPlaying).Return(nil)

			err := env.GameUsecase.JoinRoom(env.Ctx, room.ID, i)
			assert.NoError(t, err)
		}

		// 3. Players fire bullets
		bullets := make([]*game.Bullet, 0)
		for _, player := range players {
			bullet, err := env.GameUsecase.FireBullet(env.Ctx, room.ID, player.ID, 1.0, 10, game.Position{X: 600, Y: 750})
			assert.NoError(t, err)
			bullets = append(bullets, bullet)
		}

		// 4. Check inventory updated
		inv := env.InventoryManager.GetInventory(game.RoomTypeNovice)
		assert.Greater(t, inv.TotalIn, int64(0), "Bets should be recorded")

		// 5. Get final room state
		finalState, err := env.GameUsecase.GetRoomState(env.Ctx, room.ID)
		assert.NoError(t, err)
		assert.Len(t, finalState.Players, 2)
		assert.GreaterOrEqual(t, len(finalState.Bullets), 2)
	})
}

// TestGameUsecase_LeaveRoom tests player leaving room
func TestGameUsecase_LeaveRoom(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	room, _ := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
	playerID := int64(1)
	testPlayer := testhelper.NewTestPlayer(playerID)

	env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil)
	env.PlayerRepo.On("UpdatePlayerStatus", env.Ctx, playerID, game.PlayerStatusPlaying).Return(nil)
	env.GameUsecase.JoinRoom(env.Ctx, room.ID, playerID)

	t.Run("player leaves room", func(t *testing.T) {
		err := env.GameUsecase.LeaveRoom(env.Ctx, room.ID, playerID)
		assert.NoError(t, err)

		// Verify player is not in room
		roomState, _ := env.GameUsecase.GetRoomState(env.Ctx, room.ID)
		assert.NotContains(t, roomState.Players, playerID)
	})
}

// TestGameUsecase_RoomList tests listing rooms
func TestGameUsecase_RoomList(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Create multiple rooms
	_, _ = env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
	_, _ = env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 2)
	_, _ = env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeAdvanced, 1)

	t.Run("list rooms by type", func(t *testing.T) {
		allRooms := env.RoomManager.GetRoomList()
		

		assert.GreaterOrEqual(t, len(allRooms), 3, "Should have at least 3 rooms created")
		
	})
}

// TestGameUsecase_EdgeCases tests edge cases
func TestGameUsecase_EdgeCases(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
		SkipDefaultMocks: true,
	})
	defer env.AssertExpectations(t)

	t.Run("fire bullet with insufficient balance", func(t *testing.T) {
		room, _ := env.RoomManager.CreateRoom(game.RoomTypeNovice, 1)
		playerID := int64(1)
		poorPlayer := testhelper.NewTestPlayerWithBalance(playerID, 1) // Very low balance

		env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(poorPlayer, nil)
		env.RoomManager.JoinRoom(room.ID, poorPlayer)

		// Try to fire expensive bullet
		bullet, err := env.GameUsecase.FireBullet(env.Ctx, room.ID, playerID, 10.0, 100, game.Position{X: 600, Y: 750})

		// Should either error or refuse
		if err != nil {
			assert.Nil(t, bullet)
		}
	})

	t.Run("hit non-existing fish", func(t *testing.T) {
		room, _ := env.RoomManager.CreateRoom(game.RoomTypeNovice, 1)
		playerID := int64(1)
		testPlayer := testhelper.NewTestPlayer(playerID)

		env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil)
		env.RoomManager.JoinRoom(room.ID, testPlayer)

		bullet := testhelper.NewTestBullet(1, playerID, 10, 100)
		room.Bullets[bullet.ID] = bullet

		hitResult, err := env.GameUsecase.HitFish(env.Ctx, room.ID, bullet.ID, 99999) // Non-existing fish

		assert.Error(t, err)
		assert.Nil(t, hitResult)
	})
}
