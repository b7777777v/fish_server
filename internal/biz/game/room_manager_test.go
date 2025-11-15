package game_test

import (
	"testing"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/testing/testhelper"
	"github.com/stretchr/testify/assert"
)

// TestRoomManager_CreateRoom tests room creation
func TestRoomManager_CreateRoom(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	tests := []struct {
		name      string
		roomType  game.RoomType
		roomNum   int32
		wantError bool
	}{
		{"create novice room", game.RoomTypeNovice, 1, false},
		{"create intermediate room", game.RoomTypeIntermediate, 1, false},
		{"create advanced room", game.RoomTypeAdvanced, 1, false},
		{"create VIP room", game.RoomTypeVIP, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, err := env.RoomManager.CreateRoom(tt.roomType, tt.roomNum)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, room)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, room)
				assert.Equal(t, tt.roomType, room.Type)
				assert.NotEmpty(t, room.ID)
			}
		})
	}
}

// TestRoomManager_GetRoom tests room retrieval
func TestRoomManager_GetRoom(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Create a room first
	createdRoom, err := env.RoomManager.CreateRoom(game.RoomTypeNovice, 1)
	assert.NoError(t, err)

	t.Run("get existing room", func(t *testing.T) {
		room, err := env.RoomManager.GetRoom(createdRoom.ID)
		assert.NoError(t, err)
		assert.NotNil(t, room)
		assert.Equal(t, createdRoom.ID, room.ID)
	})

	t.Run("get non-existing room", func(t *testing.T) {
		room, err := env.RoomManager.GetRoom("non-existing-room-id")
		assert.Error(t, err)
		assert.Nil(t, room)
	})
}

// TestRoomManager_GetRoomList tests listing all rooms
func TestRoomManager_GetRoomList(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Create multiple rooms
	_, _ = env.RoomManager.CreateRoom(game.RoomTypeNovice, 1)
	_, _ = env.RoomManager.CreateRoom(game.RoomTypeNovice, 2)
	_, _ = env.RoomManager.CreateRoom(game.RoomTypeAdvanced, 1)

	t.Run("list all rooms", func(t *testing.T) {
		rooms := env.RoomManager.GetRoomList()
		assert.GreaterOrEqual(t, len(rooms), 3)
	})
}

// TestRoomManager_JoinRoom tests player joining room
func TestRoomManager_JoinRoom(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	room, err := env.RoomManager.CreateRoom(game.RoomTypeNovice, 4)
	assert.NoError(t, err)

	t.Run("join room successfully", func(t *testing.T) {
		player := testhelper.NewTestPlayer(1)
		err := env.RoomManager.JoinRoom(room.ID, player)
		assert.NoError(t, err)
	})

	t.Run("join non-existing room", func(t *testing.T) {
		player := testhelper.NewTestPlayer(99)
		err := env.RoomManager.JoinRoom("non-existing", player)
		assert.Error(t, err)
	})
}

// TestRoomManager_LeaveRoom tests player leaving room
func TestRoomManager_LeaveRoom(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	room, _ := env.RoomManager.CreateRoom(game.RoomTypeNovice, 4)
	player := testhelper.NewTestPlayer(1)
	_ = env.RoomManager.JoinRoom(room.ID, player)

	t.Run("leave room successfully", func(t *testing.T) {
		err := env.RoomManager.LeaveRoom(room.ID, player.ID)
		assert.NoError(t, err)
	})

	t.Run("leave non-existing room", func(t *testing.T) {
		err := env.RoomManager.LeaveRoom("non-existing", player.ID)
		assert.Error(t, err)
	})
}

// TestRoomManager_FireBullet tests firing bullets
func TestRoomManager_FireBullet(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	room, _ := env.RoomManager.CreateRoom(game.RoomTypeNovice, 4)
	player := testhelper.NewTestPlayer(1)
	_ = env.RoomManager.JoinRoom(room.ID, player)

	t.Run("fire bullet successfully", func(t *testing.T) {
		bullet, err := env.RoomManager.FireBullet(room.ID, player.ID, 1.0, 10, game.Position{X: 600, Y: 750})
		assert.NoError(t, err)
		assert.NotNil(t, bullet)
		assert.Equal(t, player.ID, bullet.PlayerID)
	})

	t.Run("fire bullet in non-existing room", func(t *testing.T) {
		bullet, err := env.RoomManager.FireBullet("non-existing", player.ID, 1.0, 10, game.Position{X: 600, Y: 750})
		assert.Error(t, err)
		assert.Nil(t, bullet)
	})
}
