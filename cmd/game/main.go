// cmd/game/main.go
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

func main() {
	// 1. 解析命令列參數
	confPath := flag.String("conf", "configs/config.yaml", "config file path")
	flag.Parse()

	// 2. 初始化日誌
	// 在 app 啟動前，我們需要一個臨時的 logger
	log := logger.New(os.Stdout, "info", "console")

	// 3. 加載設定檔
	cfg, err := conf.NewConfig(*confPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 4. 透過 wire 初始化 App
	app, cleanup, err := initApp(cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	defer cleanup() // 確保在 main 函式結束時，執行清理工作 (如關閉資料庫連線)

	// 5. 驗證配置
	if cfg.Server == nil || cfg.Server.Game == nil {
		log.Fatalf("Game server configuration is missing")
	}

	// 6. 啟動遊戲應用程序
	go func() {
		log.Info("Starting Game App")
		if err := app.Run(); err != nil {
			log.Fatalf("Failed to start game app: %v", err)
		}
	}()

	// 7. 等待中斷訊號以進行優雅關閉
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")
	if err := app.Stop(); err != nil {
		log.Errorf("Error stopping game app: %v", err)
	}
	log.Info("Server exited")
}
