package admin

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// Server 管理後台 HTTP 服務器
type Server struct {
	conf    *conf.Server
	service *AdminService
	logger  logger.Logger
	engine  *gin.Engine
	server  *http.Server
}

// NewServer 創建一個新的管理後台服務器
func NewServer(
	conf *conf.Server,
	service *AdminService,
	logger logger.Logger,
) *Server {
	return &Server{
		conf:    conf,
		service: service,
		logger:  logger.With("module", "app/admin/server"),
	}
}

// Start 啟動管理後台服務器
func (s *Server) Start() error {
	// 根據環境設置 Gin 模式
	s.setupGinMode()
	
	// 創建 Gin 引擎
	s.engine = gin.New()
	
	// 添加中間件
	s.setupMiddleware()
	
	// 註冊路由
	s.setupRoutes()
	
	// 創建 HTTP 服務器
	s.server = s.createHTTPServer()
	
	s.logger.Infof("Starting admin server on port %d in %s environment", 
		s.conf.Admin.Port, s.service.config.Environment)
	
	// 啟動服務器
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Errorf("Failed to start admin server: %v", err)
		return err
	}
	
	return nil
}

// Stop 停止管理後台服務器
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping admin server...")
	
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	
	return nil
}

// setupMiddleware 設置中間件
func (s *Server) setupMiddleware() {
	// 自定義 Logger 中間件
	s.engine.Use(s.ginLogger())
	
	// Recovery 中間件
	s.engine.Use(gin.Recovery())
	
	// CORS 中間件
	s.engine.Use(s.corsMiddleware())
	
	// 安全標頭中間件
	s.engine.Use(s.securityHeadersMiddleware())
	
	// 請求大小限制中間件
	s.engine.Use(s.requestSizeLimitMiddleware(1 << 20)) // 1MB
}

// setupRoutes 設置路由
func (s *Server) setupRoutes() {
	// 根路由
	s.engine.GET("/", s.rootHandler)

	// 健康檢查（簡單版本，不需要認證）
	s.engine.GET("/ping", s.pingHandler)

	// 提供前端測試客戶端靜態文件
	// 可通過 http://localhost:6060/test-client 訪問遊戲測試客戶端
	s.engine.Static("/test-client", "./js")
	s.logger.Info("Game test client available at /test-client")

	// 註冊業務路由（包含 pprof 路由的條件性註冊）
	s.service.RegisterRoutes(s.engine)
}

// ginLogger 自定義 Gin 日誌中間件
func (s *Server) ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 開始時間
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// 處理請求
		c.Next()
		
		// 結束時間
		end := time.Now()
		latency := end.Sub(start)
		
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		
		if raw != "" {
			path = path + "?" + raw
		}
		
		// 根據狀態碼選擇日誌級別
		switch {
		case statusCode >= 400 && statusCode < 500:
			s.logger.Warnf("HTTP %d | %13v | %15s | %-7s %s",
				statusCode, latency, clientIP, method, path)
		case statusCode >= 500:
			s.logger.Errorf("HTTP %d | %13v | %15s | %-7s %s",
				statusCode, latency, clientIP, method, path)
		default:
			s.logger.Infof("HTTP %d | %13v | %15s | %-7s %s",
				statusCode, latency, clientIP, method, path)
		}
	}
}

// corsMiddleware CORS 中間件
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers")
		c.Header("Access-Control-Allow-Credentials", "true")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// securityHeadersMiddleware 安全標頭中間件
func (s *Server) securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 對於測試客戶端，放寬 CSP 策略以允許 WebSocket 連接和腳本執行
		if len(c.Request.URL.Path) >= 12 && c.Request.URL.Path[:12] == "/test-client" {
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; connect-src 'self' ws: wss:; img-src 'self' data:")
		} else {
			c.Header("Content-Security-Policy", "default-src 'self'")
		}

		c.Next()
	}
}

// requestSizeLimitMiddleware 請求大小限制中間件
func (s *Server) requestSizeLimitMiddleware(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request entity too large",
				"max_size": fmt.Sprintf("%d bytes", maxSize),
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// rootHandler 根路由處理器
func (s *Server) rootHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "Fish Server Admin API",
		"version": "1.0.0",
		"status":  "running",
		"time":    time.Now().Format(time.RFC3339),
		"endpoints": gin.H{
			"health":      "/admin/health",
			"status":      "/admin/status",
			"metrics":     "/admin/metrics",
			"players":     "/admin/players",
			"wallets":     "/admin/wallets",
			"debug":       "/debug/pprof",
			"test_client": "/test-client",
		},
	})
}

// pingHandler ping 處理器
func (s *Server) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// GetEngine 獲取 Gin 引擎（用於測試）
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}

// setupGinMode 根據環境設置 Gin 模式
func (s *Server) setupGinMode() {
	if s.service.config.Debug != nil && s.service.config.Debug.EnableGinDebug {
		gin.SetMode(gin.DebugMode)
		s.logger.Info("Gin running in debug mode")
	} else {
		gin.SetMode(gin.ReleaseMode)
		s.logger.Info("Gin running in release mode")
	}
}

// createHTTPServer 創建 HTTP 服務器
func (s *Server) createHTTPServer() *http.Server {
	// 根據環境設置不同的超時配置
	var readTimeout, writeTimeout, idleTimeout time.Duration
	var maxHeaderBytes int
	
	switch s.service.config.Environment {
	case "dev", "development":
		readTimeout = 30 * time.Second
		writeTimeout = 30 * time.Second
		idleTimeout = 120 * time.Second
		maxHeaderBytes = 2 << 20 // 2MB for development
	case "staging", "stag":
		readTimeout = 15 * time.Second
		writeTimeout = 15 * time.Second
		idleTimeout = 90 * time.Second
		maxHeaderBytes = 1 << 20 // 1MB
	default: // production
		readTimeout = 10 * time.Second
		writeTimeout = 10 * time.Second
		idleTimeout = 60 * time.Second
		maxHeaderBytes = 1 << 20 // 1MB
	}
	
	return &http.Server{
		Addr:           fmt.Sprintf(":%d", s.conf.Admin.Port),
		Handler:        s.engine,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		IdleTimeout:    idleTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
}

// GetAddr 獲取服務器地址
func (s *Server) GetAddr() string {
	if s.server != nil {
		return s.server.Addr
	}
	return fmt.Sprintf(":%d", s.conf.Admin.Port)
}