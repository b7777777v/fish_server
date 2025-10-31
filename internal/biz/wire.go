// internal/biz/wire.go
package biz

import (
	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/biz/wallet"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	game.ProviderSet,
	player.NewPlayerUsecase,
	wallet.NewWalletUsecase,
)
