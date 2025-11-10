// internal/data/wire.go
package data

import (
	"github.com/b7777777v/fish_server/internal/biz/account"
	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/lobby"
	"github.com/b7777777v/fish_server/internal/data/postgres" // Import postgres client
	"github.com/b7777777v/fish_server/internal/data/redis"    // Import redis client
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

	// Account and Lobby repo providers
	NewAccountRepo,
	wire.Bind(new(account.AccountRepo), new(*AccountRepo)),

	NewLobbyRepo,
	wire.Bind(new(lobby.LobbyRepo), new(*LobbyRepo)),

	NewRoomCache,
	wire.Bind(new(lobby.RoomCache), new(*RoomCache)),

	// Extractor functions for Postgres and Redis clients from Data struct
	ProvidePostgresClient,
	ProvideRedisClient,
)

// ProvidePostgresClient extracts *postgres.Client from *Data
func ProvidePostgresClient(data *Data) *postgres.Client {
	return data.db
}

// ProvideRedisClient extracts *redis.Client from *Data
func ProvideRedisClient(data *Data) *redis.Client {
	return data.redis
}
