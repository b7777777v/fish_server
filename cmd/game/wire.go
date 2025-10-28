//go:build wireinject
// +build wireinject

package main

import (
	"fish_server/internal/app/game"
	"fish_server/internal/biz"
	"fish_server/internal/conf"
	"fish_server/internal/data"
	"fish_server/internal/pkg/logger"
	"fish_server/internal/pkg/token"

	"github.com/google/wire"
)

func initApp(*conf.Config) (*game.GameApp, func(), error) {
	wire.Build(
		logger.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		token.ProviderSet,
		game.NewGameApp,
	)
	return nil, nil, nil
}
