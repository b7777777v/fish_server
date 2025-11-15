package game

import (
	"github.com/google/wire"
)

// NewDefaultRoomConfig creates a default room configuration for Wire
func NewDefaultRoomConfig() RoomConfig {
	return RoomConfig{
		MinBet:               1,
		MaxBet:               1000,
		BulletCostMultiplier: 1.0,
		FishSpawnRate:        0.3,
		MinFishCount:         10,
		MaxFishCount:         20,
		RoomWidth:            1200,
		RoomHeight:           800,
		TargetRTP:            0.96,
	}
}

// ProviderSet is game providers.
var ProviderSet = wire.NewSet(
	NewGameUsecase,
	NewRoomManager,
	NewRTPController,
	NewInventoryManager,
	NewMathModel,
	NewFishSpawner,
	NewDefaultRoomConfig,
	NewFormationConfigService,
)
