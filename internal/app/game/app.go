package game

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// =======================================
// GameApp - 遊戲應用程序
// =======================================

// GameApp 遊戲應用程序
type GameApp struct {
	// HTTP 服務器
	httpServer *http.Server

	// WebSocket Hub
	hub *Hub

	// WebSocket 處理器
	wsHandler *WebSocketHandler

	// 消息處理器
	messageHandler *MessageHandler

	// 遊戲用例
	gameUsecase *game.GameUsecase

	// 配置
	config *conf.Config

	// 日誌記錄器
	logger logger.Logger

	// 上下文和取消函數
	ctx    context.Context
	cancel context.CancelFunc
}

// NewGameApp 創建遊戲應用程序
func NewGameApp(
	gameUsecase *game.GameUsecase,
	config *conf.Config, // Changed: Accept full config
	logger logger.Logger,
	hub *Hub,
	wsHandler *WebSocketHandler,
	messageHandler *MessageHandler,
) *GameApp {
	ctx, cancel := context.WithCancel(context.Background())

	app := &GameApp{
		hub:            hub,
		wsHandler:      wsHandler,
		messageHandler: messageHandler,
		gameUsecase:    gameUsecase,
		config:         config, // Changed: Store full config
		logger:         logger.With("component", "game_app"),
		ctx:            ctx,
		cancel:         cancel,
	}

	// 設置 HTTP 服務器
	app.setupHTTPServer()

	// 異步加載魚類數據到緩存
	go func() {
		if err := app.gameUsecase.LoadAndCacheFishTypes(context.Background()); err != nil {
			app.logger.Errorf("Failed to load and cache fish types on startup: %v", err)
		}
	}()

	return app
}

// setupHTTPServer 設置 HTTP 服務器
func (app *GameApp) setupHTTPServer() {
	mux := http.NewServeMux()

	// WebSocket 端點
	mux.HandleFunc("/ws", app.wsHandler.ServeWS)

	// 健康檢查端點
	mux.HandleFunc("/health", app.handleHealth)

	// 狀態端點
	mux.HandleFunc("/status", app.handleStatus)

	// 房間信息端點
	mux.HandleFunc("/rooms", app.handleRooms)

	// Register pprof handlers if enabled
	if app.config != nil && app.config.Debug != nil && app.config.Debug.EnablePprof {
		app.logger.Info("Registering pprof handlers on /debug/pprof")
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	// CORS 中間件
	handler := app.corsMiddleware(mux)

	app.httpServer = &http.Server{
		Addr:         ":9090", // 默認端口
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if app.config != nil && app.config.Server != nil && app.config.Server.Game != nil {
		app.httpServer.Addr = fmt.Sprintf(":%d", app.config.Server.Game.Port)
	}
}

// corsMiddleware CORS 中間件
func (app *GameApp) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// respondJSON is a helper to write JSON responses
func (app *GameApp) respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		app.logger.Errorf("Failed to write JSON response: %v", err)
	}
}

// respondError is a helper to write JSON error responses
func (app *GameApp) respondError(w http.ResponseWriter, statusCode int, message string) {
	app.respondJSON(w, statusCode, map[string]string{"error": message})
}


// handleHealth 健康檢查處理器
func (app *GameApp) handleHealth(w http.ResponseWriter, r *http.Request) {
	app.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "game",
		"timestamp": time.Now().Unix(),
	})
}

// handleStatus 狀態處理器
func (app *GameApp) handleStatus(w http.ResponseWriter, r *http.Request) {
	stats := app.hub.GetStats()

	app.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":             "running",
		"service":            "game",
		"timestamp":          time.Now().Unix(),
		"active_connections": stats.ActiveConnections,
		"active_rooms":       stats.ActiveRooms,
		"total_connections":  stats.TotalConnections,
		"total_messages":     stats.TotalMessages,
		"start_time":         stats.StartTime.Unix(),
		"last_activity":      stats.LastActivity.Unix(),
	})
}

// handleRooms 房間信息處理器
func (app *GameApp) handleRooms(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	rooms, err := app.gameUsecase.GetRoomList(ctx, "")
	if err != nil {
		app.logger.Errorf("Failed to get room list: %v", err)
		app.respondError(w, http.StatusInternalServerError, "Failed to get room list")
		return
	}

	// Create a serializable representation of the rooms
	type roomInfo struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Players int    `json:"players"`
	}

	roomInfos := make([]roomInfo, 0, len(rooms))
	for _, room := range rooms {
		roomInfos = append(roomInfos, roomInfo{
			ID:      room.ID,
			Name:    room.Name,
			Type:    string(room.Type),
			Players: len(room.Players),
		})
	}

	app.respondJSON(w, http.StatusOK, map[string]interface{}{"rooms": roomInfos})
}


// Run 運行遊戲應用程序
func (app *GameApp) Run() error {
	app.logger.Infof("Starting Game App on %s", app.httpServer.Addr)

	// 啟動 Hub
	go app.hub.Run()

	// 啟動 HTTP 服務器
	if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		app.logger.Errorf("Failed to start game server: %v", err)
		return err
	}

	return nil
}

// Stop 停止遊戲應用程序
func (app *GameApp) Stop() error {
	app.logger.Info("Stopping Game App")

	// 停止 Hub
	app.hub.Stop()

	// 停止 HTTP 服務器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.httpServer.Shutdown(ctx); err != nil {
		app.logger.Errorf("Failed to shutdown game server: %v", err)
		return err
	}

	app.cancel()
	return nil
}

// GetStats 獲取應用程序統計信息
func (app *GameApp) GetStats() map[string]interface{} {
	hubStats := app.hub.GetStats()

	return map[string]interface{}{
		"service":            "game",
		"status":             "running",
		"active_connections": hubStats.ActiveConnections,
		"active_rooms":       hubStats.ActiveRooms,
		"total_connections":  hubStats.TotalConnections,
		"total_messages":     hubStats.TotalMessages,
		"start_time":         hubStats.StartTime,
		"last_activity":      hubStats.LastActivity,
	}
}

// GetGameUsecase 獲取遊戲用例（用於 Admin Service）
func (app *GameApp) GetGameUsecase() *game.GameUsecase {
	return app.gameUsecase
}