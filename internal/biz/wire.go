// internal/biz/wire.go
package biz

import (
	"github.com/b7777777v/fish_server/internal/biz/player"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(player.NewPlayerUsecase)
