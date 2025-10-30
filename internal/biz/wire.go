// internal/biz/wire.go
package biz

import (
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/pkg/logger"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	ProvideLogger,
	player.NewPlayerUsecase,
	wallet.NewWalletUsecase,
)

// ProvideLogger provides a logger for biz layer
func ProvideLogger(logger logger.Logger) logger.Logger {
	return logger.With("layer", "biz")
}
