//go:build wireinject
// +build wireinject

package main

import (
	"github.com/b7777777v/fish_server/internal/app"
	"github.com/b7777777v/fish_server/internal/app/admin"
	"github.com/b7777777v/fish_server/internal/conf"

	"github.com/google/wire"
)

// initApp 組裝完整的依賴圖，返回 AdminApp 實例
func initApp(*conf.Config) (*admin.AdminApp, func(), error) {
	wire.Build(
		app.AdminProviderSet,
	)
	return nil, nil, nil
}
