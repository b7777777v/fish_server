// internal/app/wire.go
package app

import (
	"github.com/b7777777v/fish_server/internal/app/admin"
	"github.com/b7777777v/fish_server/internal/app/game"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	game.NewGameServer,
	game.NewGameApp,
)

var AdminProviderSet = wire.NewSet(
	admin.ProviderSet,
)
