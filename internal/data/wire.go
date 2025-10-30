// internal/data/wire.go
package data

import (
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	ProvideLogger,
	NewData,
	NewPlayerRepo,
	NewWalletRepo,
)

// ProvideLogger provides a logger for data layer
func ProvideLogger(logger logger.Logger) logger.Logger {
	return logger.With("layer", "data")
}
