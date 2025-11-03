package game

import (
	"context"
	"sync"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// InventoryManager manages the game's financial inventories for different room types.
// It ensures that the game's RTP is tracked correctly.
type InventoryManager struct {
	inventories map[RoomType]*Inventory
	mu          sync.RWMutex
	repo        InventoryRepo
	logger      logger.Logger
}

// NewInventoryManager creates a new inventory manager.
func NewInventoryManager(repo InventoryRepo, logger logger.Logger) (*InventoryManager, error) {
	im := &InventoryManager{
		inventories: make(map[RoomType]*Inventory),
		repo:        repo,
		logger:      logger.With("component", "inventory_manager"),
	}

	// Load existing inventories from the repository on startup
	if err := im.loadAllInventories(); err != nil {
		return nil, err
	}

	return im, nil
}

// loadAllInventories loads all inventories from the repository into memory.
func (im *InventoryManager) loadAllInventories() error {
	ctx := context.Background()
	inventories, err := im.repo.GetAllInventories(ctx)
	if err != nil {
		im.logger.Errorf("Failed to load inventories from repo: %v", err)
		return err
	}

	im.mu.Lock()
	defer im.mu.Unlock()
	for roomTypeStr, inv := range inventories {
		roomType := RoomType(roomTypeStr)
		im.inventories[roomType] = inv
		im.logger.Infof("Loaded inventory for %s: TotalIn=%d, TotalOut=%d", roomType, inv.TotalIn, inv.TotalOut)
	}
	return nil
}

// GetInventory returns the inventory for a specific room type.
// If it doesn't exist, it creates a new one.
func (im *InventoryManager) GetInventory(roomType RoomType) *Inventory {
	im.mu.RLock()
	inv, exists := im.inventories[roomType]
	im.mu.RUnlock()

	if !exists {
		im.mu.Lock()
		// Double-check after acquiring write lock
		inv, exists = im.inventories[roomType]
		if !exists {
			inv = &Inventory{
				ID:        string(roomType),
				TotalIn:   0,
				TotalOut:  0,
				UpdatedAt: time.Now(),
			}
			im.inventories[roomType] = inv
			im.logger.Infof("Created new in-memory inventory for room type: %s", roomType)
		}
		im.mu.Unlock()
	}
	return inv
}

// AddBet records a player's bet, increasing TotalIn for the room type's inventory.
func (im *InventoryManager) AddBet(roomType RoomType, amount int64) {
	if amount <= 0 {
		return
	}

	inv := im.GetInventory(roomType)

	im.mu.Lock()
	defer im.mu.Unlock()

	inv.TotalIn += amount
	inv.UpdatedAt = time.Now()
	im.updateRTP(inv)

	// Persist changes periodically or based on a threshold (logic can be added here)
	im.repo.SaveInventory(context.Background(), inv) // For simplicity, save on every change
}

// AddWin records a player's win, increasing TotalOut for the room type's inventory.
func (im *InventoryManager) AddWin(roomType RoomType, amount int64) {
	if amount <= 0 {
		return
	}

	inv := im.GetInventory(roomType)

	im.mu.Lock()
	defer im.mu.Unlock()

	inv.TotalOut += amount
	inv.UpdatedAt = time.Now()
	im.updateRTP(inv)

	// Persist changes
	im.repo.SaveInventory(context.Background(), inv)
}

// updateRTP calculates and updates the current RTP for an inventory.
func (im *InventoryManager) updateRTP(inv *Inventory) {
	if inv.TotalIn == 0 {
		inv.CurrentRTP = 0
		return
	}
	inv.CurrentRTP = float64(inv.TotalOut) / float64(inv.TotalIn)
}
