// internal/data/wire.go
package data

import (
	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/data/postgres" // Import postgres client
	"github.com/b7777777v/fish_server/internal/data/redis"    // Import redis client
	"github.com/b7777777v/fish_server/internal/pkg/token"
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

	// Add FormationConfigRepo provider
	NewFormationConfigRepo,
	wire.Bind(new(game.FormationConfigRepo), new(*formationConfigRepo)), // Bind interface to implementation

	// Add RoomConfigRepo provider
	NewRoomConfigRepo,

	// Account and Lobby repo providers
	NewAccountRepo,
	NewLobbyRepo,
	NewRoomCache,
	NewLobbyPlayerRepo,
	NewLobbyWalletRepo,

	// Extractor functions for Postgres DBManager and Redis clients from Data struct
	ProvideDBManager,
	ProvideRedisClient,

	// Token cache provider
	redis.NewTokenCache,
	wire.Bind(new(token.TokenCache), new(*redis.TokenCache)),
)

// ProvideDBManager extracts *postgres.DBManager from *Data
func ProvideDBManager(data *Data) *postgres.DBManager {
	return data.dbManager
}

// ProvideRedisClient extracts *redis.Client from *Data
func ProvideRedisClient(data *Data) *redis.Client {
	return data.redis
}
