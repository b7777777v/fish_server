package data

import (
	"context"
	"sync"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// InMemoryInventoryRepo is an in-memory implementation of the InventoryRepo interface.
// NOTE: This is for demonstration purposes. In a real production environment,
// this should be replaced with a persistent storage solution like Redis or a database.
type InMemoryInventoryRepo struct {
	mu          sync.RWMutex
	inventories map[string]*game.Inventory
	logger      logger.Logger
}

// NewInMemoryInventoryRepo creates a new in-memory inventory repository.
func NewInMemoryInventoryRepo(logger logger.Logger) *InMemoryInventoryRepo {
	return &InMemoryInventoryRepo{
		inventories: make(map[string]*game.Inventory),
		logger:      logger.With("component", "in_memory_inventory_repo"),
	}
}

// GetInventory retrieves an inventory by its ID.
func (r *InMemoryInventoryRepo) GetInventory(ctx context.Context, inventoryID string) (*game.Inventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if inv, ok := r.inventories[inventoryID]; ok {
		// Return a copy to prevent race conditions on the caller's side
		invCopy := *inv
		return &invCopy, nil
	}

	// In a real implementation, you might return a "not found" error.
	// Here, we return a new, empty inventory to simplify the business logic.
	return &game.Inventory{ID: inventoryID}, nil
}

// SaveInventory saves an inventory.
func (r *InMemoryInventoryRepo) SaveInventory(ctx context.Context, inventory *game.Inventory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Debugf("Saving inventory %s: TotalIn=%d, TotalOut=%d", inventory.ID, inventory.TotalIn, inventory.TotalOut)
	invCopy := *inventory
	r.inventories[inventory.ID] = &invCopy
	return nil
}

// GetAllInventories retrieves all inventories.
func (r *InMemoryInventoryRepo) GetAllInventories(ctx context.Context) (map[string]*game.Inventory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a deep copy of the map
	inventoriesCopy := make(map[string]*game.Inventory, len(r.inventories))
	for id, inv := range r.inventories {
		invCopy := *inv
		inventoriesCopy[id] = &invCopy
	}

	return inventoriesCopy, nil
}
