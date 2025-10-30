package admin

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// AdminApp 管理後台應用程序
type AdminApp struct {
	server  *Server
	logger  logger.Logger
	cleanup func()
}

// NewAdminApp 創建管理後台應用程序
func NewAdminApp(
	server *Server,
	logger logger.Logger,
) *AdminApp {
	return &AdminApp{
		server:  server,
		logger:  logger.With("module", "app/admin"),
		cleanup: nil, // cleanup 將由外部設置
	}
}

// SetCleanup 設置清理函數
func (app *AdminApp) SetCleanup(cleanup func()) {
	app.cleanup = cleanup
}

// Run 運行管理後台應用程序
func (app *AdminApp) Run() error {
	app.logger.Info("Starting Fish Server Admin...")

	// 創建一個用於接收系統信號的通道
	quit := make(chan os.Signal, 1)
	// 監聽中斷信號和終止信號
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 在 goroutine 中啟動服務器
	serverError := make(chan error, 1)
	go func() {
		if err := app.server.Start(); err != nil {
			serverError <- err
		}
	}()

	app.logger.Infof("Admin server started on %s", app.server.GetAddr())
	app.logger.Info("Press Ctrl+C to gracefully shutdown the server...")

	// 等待退出信號或服務器錯誤
	select {
	case err := <-serverError:
		app.logger.Errorf("Server failed to start: %v", err)
		return err
	case sig := <-quit:
		app.logger.Infof("Received signal: %v", sig)
		app.logger.Info("Shutting down server...")

		// 創建一個帶超時的 context 用於優雅關閉
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 停止服務器
		if err := app.server.Stop(ctx); err != nil {
			app.logger.Errorf("Server forced to shutdown: %v", err)
			return err
		}

		app.logger.Info("Server exited gracefully")
	}

	return nil
}

// Stop 停止應用程序
func (app *AdminApp) Stop() error {
	app.logger.Info("Stopping admin application...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if app.server != nil {
		if err := app.server.Stop(ctx); err != nil {
			app.logger.Errorf("Error stopping server: %v", err)
			return err
		}
	}
	
	if app.cleanup != nil {
		app.cleanup()
		app.logger.Info("Cleanup completed")
	}
	
	app.logger.Info("Admin application stopped")
	return nil
}

// GetServerAddr 獲取服務器地址
func (app *AdminApp) GetServerAddr() string {
	if app.server != nil {
		return app.server.GetAddr()
	}
	return ""
}