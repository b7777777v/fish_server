package game

import (
	"context"
	"net/http"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// ========================================
// GameApp - 遊戲應用程序
// ========================================

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
	config *conf.Server
	
	// 日誌記錄器
	logger logger.Logger
	
	// 上下文和取消函數
	ctx    context.Context
	cancel context.CancelFunc
}

// NewGameApp 創建遊戲應用程序
func NewGameApp(
	gameUsecase *game.GameUsecase,
	config *conf.Server,
	logger logger.Logger,
) *GameApp {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 創建 Hub
	hub := NewHub(gameUsecase, logger)
	
	// 創建 WebSocket 處理器
	wsHandler := NewWebSocketHandler(hub, logger)
	
	// 創建消息處理器
	messageHandler := NewMessageHandler(gameUsecase, hub, logger)
	
	app := &GameApp{
		hub:            hub,
		wsHandler:      wsHandler,
		messageHandler: messageHandler,
		gameUsecase:    gameUsecase,
		config:         config,
		logger:         logger.With("component", "game_app"),
		ctx:            ctx,
		cancel:         cancel,
	}
	
	// 設置 HTTP 服務器
	app.setupHTTPServer()
	
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
	
	// CORS 中間件
	handler := app.corsMiddleware(mux)
	
	app.httpServer = &http.Server{
		Addr:         ":9090", // 默認端口，應該從配置讀取
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	if app.config != nil && app.config.Game != nil {
		app.httpServer.Addr = ":" + string(rune(app.config.Game.Port))
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

// handleHealth 健康檢查處理器
func (app *GameApp) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"game","timestamp":` + 
		string(rune(time.Now().Unix())) + `}`))
}

// handleStatus 狀態處理器
func (app *GameApp) handleStatus(w http.ResponseWriter, r *http.Request) {
	stats := app.hub.GetStats()
	
	response := map[string]interface{}{
		"status":             "running",
		"service":            "game",
		"timestamp":          time.Now().Unix(),
		"active_connections": stats.ActiveConnections,
		"active_rooms":       stats.ActiveRooms,
		"total_connections":  stats.TotalConnections,
		"total_messages":     stats.TotalMessages,
		"start_time":         stats.StartTime.Unix(),
		"last_activity":      stats.LastActivity.Unix(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// 簡單的 JSON 編碼
	jsonResponse := `{`
	first := true
	for key, value := range response {
		if !first {
			jsonResponse += ","
		}
		jsonResponse += `"` + key + `":` + formatValue(value)
		first = false
	}
	jsonResponse += `}`
	
	w.Write([]byte(jsonResponse))
}

// formatValue 格式化值為 JSON
func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return `"` + v + `"`
	case int64:
		return string(rune(v + '0'))
	case int:
		return string(rune(v + '0'))
	case time.Time:
		return string(rune(v.Unix() + '0'))
	default:
		return `"` + string(rune(0)) + `"`
	}
}

// handleRooms 房間信息處理器
func (app *GameApp) handleRooms(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	rooms, err := app.gameUsecase.GetRoomList(ctx, "")
	if err != nil {
		app.logger.Errorf("Failed to get room list: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to get room list"}`))
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// 簡單的房間列表響應
	response := `{"rooms":[`
	for i, room := range rooms {
		if i > 0 {
			response += ","
		}
		response += `{"id":"` + room.ID + `","name":"` + room.Name + 
			`","type":"` + string(room.Type) + `","players":` + 
			string(rune(len(room.Players)+'0')) + `}`
	}
	response += `]}`
	
	w.Write([]byte(response))
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