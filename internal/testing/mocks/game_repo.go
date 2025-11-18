// Package mocks provides mock implementations for testing using testify/mock
package mocks

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/stretchr/testify/mock"
)

// GameRepo is a mock implementation of game.GameRepo interface
type GameRepo struct {
	mock.Mock
}

// SaveRoom mocks the SaveRoom method
func (m *GameRepo) SaveRoom(ctx context.Context, room *game.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

// GetRoom mocks the GetRoom method
func (m *GameRepo) GetRoom(ctx context.Context, roomID string) (*game.Room, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Room), args.Error(1)
}

// ListRooms mocks the ListRooms method
func (m *GameRepo) ListRooms(ctx context.Context, roomType game.RoomType) ([]*game.Room, error) {
	args := m.Called(ctx, roomType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game.Room), args.Error(1)
}

// DeleteRoom mocks the DeleteRoom method
func (m *GameRepo) DeleteRoom(ctx context.Context, roomID string) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

// SaveGameStatistics mocks the SaveGameStatistics method
func (m *GameRepo) SaveGameStatistics(ctx context.Context, playerID int64, stats *game.GameStatistics) error {
	args := m.Called(ctx, playerID, stats)
	return args.Error(0)
}

// GetGameStatistics mocks the GetGameStatistics method
func (m *GameRepo) GetGameStatistics(ctx context.Context, playerID int64) (*game.GameStatistics, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.GameStatistics), args.Error(1)
}

// SaveGameEvent mocks the SaveGameEvent method
func (m *GameRepo) SaveGameEvent(ctx context.Context, event *game.GameEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// GetGameEvents mocks the GetGameEvents method
func (m *GameRepo) GetGameEvents(ctx context.Context, roomID string, limit int) ([]*game.GameEvent, error) {
	args := m.Called(ctx, roomID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game.GameEvent), args.Error(1)
}

// GetAllFishTypes mocks the GetAllFishTypes method
func (m *GameRepo) GetAllFishTypes(ctx context.Context) ([]*game.FishType, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*game.FishType), args.Error(1)
}

// SaveFishTypeCache mocks the SaveFishTypeCache method
func (m *GameRepo) SaveFishTypeCache(ctx context.Context, ft *game.FishType) error {
	args := m.Called(ctx, ft)
	return args.Error(0)
}

// SaveRoomToRedis mocks the SaveRoomToRedis method
func (m *GameRepo) SaveRoomToRedis(ctx context.Context, room *game.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

// DeleteRoomFromRedis mocks the DeleteRoomFromRedis method
func (m *GameRepo) DeleteRoomFromRedis(ctx context.Context, roomID string) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}

// IncrementRoomCount mocks the IncrementRoomCount method
func (m *GameRepo) IncrementRoomCount(ctx context.Context, roomType game.RoomType) error {
	args := m.Called(ctx, roomType)
	return args.Error(0)
}

// DecrementRoomCount mocks the DecrementRoomCount method
func (m *GameRepo) DecrementRoomCount(ctx context.Context, roomType game.RoomType) error {
	args := m.Called(ctx, roomType)
	return args.Error(0)
}

// GetRoomCount mocks the GetRoomCount method
func (m *GameRepo) GetRoomCount(ctx context.Context, roomType game.RoomType) (int64, error) {
	args := m.Called(ctx, roomType)
	return args.Get(0).(int64), args.Error(1)
}

// GetTotalRoomCount mocks the GetTotalRoomCount method
func (m *GameRepo) GetTotalRoomCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// GetAllRoomCounts mocks the GetAllRoomCounts method
func (m *GameRepo) GetAllRoomCounts(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}
