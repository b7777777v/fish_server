// internal/data/wire.go
package data

import (
	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewGameRepo,
	NewGamePlayerRepo,
	NewPlayerRepo,
	NewWalletRepo,

	// Add the new inventory repo provider
	NewInMemoryInventoryRepo,
	wire.Bind(new(game.InventoryRepo), new(*InMemoryInventoryRepo)),
)
