//go:build wireinject
// +build wireinject

package main

import (
	"github.com/b7777777v/fish_server/internal/app"
	game "github.com/b7777777v/fish_server/internal/app/game"
	"github.com/b7777777v/fish_server/internal/biz"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/data"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/pkg/token"

	"github.com/google/wire"
)

func initApp(*conf.Config) (*game.GameApp, func(), error) {
	wire.Build(
		conf.ProviderSet,
		logger.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		token.ProviderSet,
		app.ProviderSet,
		// game.NewGameApp,
	)
	return nil, nil, nil
}
