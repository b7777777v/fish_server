// internal/app/wire.go
package app

import (
	"github.com/b7777777v/fish_server/internal/app/admin"
	"github.com/b7777777v/fish_server/internal/app/game"
	"github.com/b7777777v/fish_server/internal/biz"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/data"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/pkg/token"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	game.ProviderSet,
)

var AdminProviderSet = wire.NewSet(
	// conf, logger, data, biz
	conf.ProviderSet,
	logger.ProviderSet,
	data.ProviderSet,
	biz.ProviderSet,

	// app
	admin.ProviderSet,
	game.ProviderSet,
	token.ProviderSet,
)
