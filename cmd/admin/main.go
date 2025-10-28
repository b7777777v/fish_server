// cmd/admin/main.go
package main

import (
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

func main() {
	// 加載配置
	cfg, err := conf.NewConfig("./configs/config.yaml")
	if err != nil {
		// 這裡還不能用我們的 logger，因為它還沒被創建
		// 使用標準庫 log 記錄致命錯誤
		panic(err)
	}

	// 創建 Logger
	appLogger, err := logger.NewLogger(cfg.Log)
	if err != nil {
		panic(err)
	}

	// 使用我們的 logger！
	appLogger.Infow("Config loaded successfully!",
		"admin_port", cfg.Server.Admin.Port,
		"db_driver", cfg.Data.Database.Driver,
	)

	appLogger.Debugw("This is a debug message with details.",
		"jwt_secret_length", len(cfg.JWT.Secret),
	)

	appLogger.Warnw("This is a warning message.")

	// 之後這裡會是啟動 Gin 伺服器的程式碼
}
