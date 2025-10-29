// cmd/game/main.go
package main

import (
	"flag"
	"fmt"
	"net"
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

	// 5. 建立 TCP 監聽
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Game.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 6. 在一個新的 goroutine 中啟動 gRPC 伺服器
	go func() {
		log.Infof("gRPC server listening on: %s", lis.Addr().String())
		if err := app.GrpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve gRPC: %v", err)
		}
	}()

	// 7. 等待中斷訊號以進行優雅關閉
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gRPC server...")
	app.GrpcServer.GracefulStop()
	log.Info("gRPC server stopped")
}
