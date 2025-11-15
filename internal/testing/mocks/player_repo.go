package mocks

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/stretchr/testify/mock"
)

// PlayerRepo is a mock implementation of game.PlayerRepo interface
type PlayerRepo struct {
	mock.Mock
}

// GetPlayer mocks the GetPlayer method
func (m *PlayerRepo) GetPlayer(ctx context.Context, playerID int64) (*game.Player, error) {
	args := m.Called(ctx, playerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Player), args.Error(1)
}

// UpdatePlayerBalance mocks the UpdatePlayerBalance method
func (m *PlayerRepo) UpdatePlayerBalance(ctx context.Context, playerID int64, balance int64) error {
	args := m.Called(ctx, playerID, balance)
	return args.Error(0)
}

// UpdatePlayerStatus mocks the UpdatePlayerStatus method
func (m *PlayerRepo) UpdatePlayerStatus(ctx context.Context, playerID int64, status game.PlayerStatus) error {
	args := m.Called(ctx, playerID, status)
	return args.Error(0)
}
