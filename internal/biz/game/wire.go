package game

import (
	"github.com/google/wire"
)

// ProviderSet is game providers.
var ProviderSet = wire.NewSet(
	NewGameUsecase,
	NewRoomManager,
	NewRTPController,
	NewInventoryManager,
	// TODO: These are temporary mocks and should be replaced by real implementations
	NewMathModel,
	NewFishSpawner,
)
