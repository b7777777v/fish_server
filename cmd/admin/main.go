// cmd/admin/main.go
package main

import (
	"log"
	"os"

	"github.com/b7777777v/fish_server/internal/conf"
)

func main() {
	// 設置配置文件路徑
	configPath := "./configs/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// 加載配置
	cfg, err := conf.NewConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configPath, err)
	}

	// 使用 wire 初始化應用程序
	app, cleanup, err := initApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize admin app: %v", err)
	}
	defer cleanup()

	// 設置清理函數
	app.SetCleanup(cleanup)

	// 運行應用程序
	if err := app.Run(); err != nil {
		log.Fatalf("Admin app failed to run: %v", err)
	}
}
