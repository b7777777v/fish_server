//go:build wireinject
// +build wireinject

package main

import (
	"github.com/b7777777v/fish_server/internal/app"
	"github.com/b7777777v/fish_server/internal/app/admin"
	"github.com/b7777777v/fish_server/internal/biz"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/data"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/pkg/token"

	"github.com/google/wire"
)

// initApp 組裝完整的依賴圖，返回 AdminApp 實例
func initApp(*conf.Config) (*admin.AdminApp, func(), error) {
	wire.Build(
		// 基礎設施層
		conf.ProviderSet,   // 配置提供者
		logger.ProviderSet, // 日誌提供者
		token.ProviderSet,  // Token 提供者
		
		// 數據層
		data.ProviderSet,   // 數據庫、Redis 等數據訪問層
		
		// 業務層
		biz.ProviderSet,    // 業務邏輯用例層
		
		// 應用層
		app.AdminProviderSet, // Admin 應用層提供者
	)
	return nil, nil, nil
}
