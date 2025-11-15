package mocks

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/stretchr/testify/mock"
)

// InventoryRepo is a mock implementation of game.InventoryRepo interface
type InventoryRepo struct {
	mock.Mock
}

// GetInventory mocks the GetInventory method
func (m *InventoryRepo) GetInventory(ctx context.Context, inventoryID string) (*game.Inventory, error) {
	args := m.Called(ctx, inventoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*game.Inventory), args.Error(1)
}

// SaveInventory mocks the SaveInventory method
func (m *InventoryRepo) SaveInventory(ctx context.Context, inventory *game.Inventory) error {
	args := m.Called(ctx, inventory)
	return args.Error(0)
}

// GetAllInventories mocks the GetAllInventories method
func (m *InventoryRepo) GetAllInventories(ctx context.Context) (map[string]*game.Inventory, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*game.Inventory), args.Error(1)
}
