// cmd/game/main.go
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

func main() {

	// 傳入空字串，讓 NewConfig 根據環境變數自動選擇設定檔
	// 例如，未設置時預設載入 config.dev.yaml
	configPath := ""
	// 在 app 啟動前，我們需要一個臨時的 logger
	log := logger.New(os.Stdout, "info", "console")
	// 加載配置
	cfg, err := conf.NewConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", configPath, err)
	}

	// 為 game server 設置專屬的日誌文件路徑
	if cfg.Log != nil && cfg.Log.FilePath != "" {
		// 將日誌路徑修改為 game 專用（例如：logs/game-dev.log -> logs/game-server-dev.log）
		cfg.Log.FilePath = "logs/game-server.log"
		if cfg.Environment == "dev" || cfg.Environment == "development" {
			cfg.Log.FilePath = "logs/game-server-dev.log"
		} else if cfg.Environment == "staging" || cfg.Environment == "stag" {
			cfg.Log.FilePath = "logs/game-server-staging.log"
		} else if cfg.Environment == "prod" || cfg.Environment == "production" {
			cfg.Log.FilePath = "logs/game-server-prod.log"
		}
	}

	// 透過 wire 初始化 App
	app, cleanup, err := initApp(cfg)
	if err != nil {
		log.Fatalf("failed to init app: %v", err)
	}
	defer cleanup() // 確保在 main 函式結束時，執行清理工作 (如關閉資料庫連線)
	// 驗證配置
	if cfg.Server == nil || cfg.Server.Game == nil {
		log.Fatalf("Game server configuration is missing")
	}
	// 啟動遊戲應用程序
	go func() {
		log.Info("Starting Game App")
		if err := app.Run(); err != nil {
			log.Fatalf("Failed to start game app: %v", err)
		}
	}()
	// 等待中斷訊號以進行優雅關閉
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")
	if err := app.Stop(); err != nil {
		log.Errorf("Error stopping game app: %v", err)
	}
	log.Info("Server exited")
}
