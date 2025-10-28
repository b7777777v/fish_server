//go:build wireinject
// +build wireinject

package main

import (
	"github.com/b7777777v/fish_server/internal/app/admin"
	"github.com/b7777777v/fish_server/internal/biz"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/data"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/pkg/token"

	"github.com/google/wire"
)

func initApp(*conf.Config) (*admin.AdminApp, func(), error) {
	wire.Build(
		logger.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		token.ProviderSet,
		admin.NewAdminApp,
	)
	return nil, nil, nil
}
