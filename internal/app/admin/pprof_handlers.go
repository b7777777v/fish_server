package admin

import (
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"
)

// registerPprofRoutes 註冊 pprof 路由用於性能分析
func (s *AdminService) registerPprofRoutes(r *gin.Engine) {
	// pprof 路由組
	pprofGroup := r.Group("/debug/pprof")
	{
		pprofGroup.GET("/", s.pprofIndex)
		pprofGroup.GET("/cmdline", s.pprofCmdline)
		pprofGroup.GET("/profile", s.pprofProfile)
		pprofGroup.POST("/symbol", s.pprofSymbol)
		pprofGroup.GET("/symbol", s.pprofSymbol)
		pprofGroup.GET("/trace", s.pprofTrace)
		pprofGroup.GET("/allocs", s.pprofAllocs)
		pprofGroup.GET("/block", s.pprofBlock)
		pprofGroup.GET("/goroutine", s.pprofGoroutine)
		pprofGroup.GET("/heap", s.pprofHeap)
		pprofGroup.GET("/mutex", s.pprofMutex)
		pprofGroup.GET("/threadcreate", s.pprofThreadCreate)
		pprofGroup.GET("/info", s.GetPprofInfo)
	}
}

// pprofIndex pprof 首頁
func (s *AdminService) pprofIndex(c *gin.Context) {
	s.logger.Info("Pprof index accessed")
	pprof.Index(c.Writer, c.Request)
}

// pprofCmdline 顯示運行程序的命令行
func (s *AdminService) pprofCmdline(c *gin.Context) {
	s.logger.Info("Pprof cmdline accessed")
	pprof.Cmdline(c.Writer, c.Request)
}

// pprofProfile CPU 性能分析
func (s *AdminService) pprofProfile(c *gin.Context) {
	s.logger.Info("Pprof profile accessed")
	pprof.Profile(c.Writer, c.Request)
}

// pprofSymbol 符號表
func (s *AdminService) pprofSymbol(c *gin.Context) {
	s.logger.Info("Pprof symbol accessed")
	pprof.Symbol(c.Writer, c.Request)
}

// pprofTrace 執行追蹤
func (s *AdminService) pprofTrace(c *gin.Context) {
	s.logger.Info("Pprof trace accessed")
	pprof.Trace(c.Writer, c.Request)
}

// pprofAllocs 記憶體分配分析
func (s *AdminService) pprofAllocs(c *gin.Context) {
	s.logger.Info("Pprof allocs accessed")
	pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
}

// pprofBlock 阻塞分析
func (s *AdminService) pprofBlock(c *gin.Context) {
	s.logger.Info("Pprof block accessed")
	pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
}

// pprofGoroutine goroutine 分析
func (s *AdminService) pprofGoroutine(c *gin.Context) {
	s.logger.Info("Pprof goroutine accessed")
	pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
}

// pprofHeap 堆記憶體分析
func (s *AdminService) pprofHeap(c *gin.Context) {
	s.logger.Info("Pprof heap accessed")
	pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
}

// pprofMutex 互斥鎖分析
func (s *AdminService) pprofMutex(c *gin.Context) {
	s.logger.Info("Pprof mutex accessed")
	pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
}

// pprofThreadCreate 線程創建分析
func (s *AdminService) pprofThreadCreate(c *gin.Context) {
	s.logger.Info("Pprof threadcreate accessed")
	pprof.Handler("threadcreate").ServeHTTP(c.Writer, c.Request)
}

// PprofMiddleware pprof 中間件，用於限制訪問
func (s *AdminService) PprofMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 在生產環境中，這裡應該添加身份驗證
		// 例如檢查 API key 或 JWT token
		
		// 記錄訪問
		s.logger.Warnf("Pprof endpoint accessed: %s from %s", c.Request.URL.Path, c.ClientIP())
		
		// 設置安全標頭
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		
		c.Next()
	})
}

// EnablePprofWithAuth 啟用帶身份驗證的 pprof
func (s *AdminService) EnablePprofWithAuth(r *gin.Engine, authKey string) {
	// 帶身份驗證的 pprof 路由
	authGroup := r.Group("/debug/pprof")
	authGroup.Use(s.pprofAuthMiddleware(authKey))
	authGroup.Use(s.PprofMiddleware())
	{
		authGroup.GET("/", s.pprofIndex)
		authGroup.GET("/cmdline", s.pprofCmdline)
		authGroup.GET("/profile", s.pprofProfile)
		authGroup.POST("/symbol", s.pprofSymbol)
		authGroup.GET("/symbol", s.pprofSymbol)
		authGroup.GET("/trace", s.pprofTrace)
		authGroup.GET("/allocs", s.pprofAllocs)
		authGroup.GET("/block", s.pprofBlock)
		authGroup.GET("/goroutine", s.pprofGoroutine)
		authGroup.GET("/heap", s.pprofHeap)
		authGroup.GET("/mutex", s.pprofMutex)
		authGroup.GET("/threadcreate", s.pprofThreadCreate)
		authGroup.GET("/info", s.GetPprofInfo)
	}
}

// pprofAuthMiddleware pprof 身份驗證中間件
func (s *AdminService) pprofAuthMiddleware(authKey string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 檢查 Authorization header
		auth := c.GetHeader("Authorization")
		if auth == "" {
			// 檢查查詢參數
			auth = c.Query("auth")
		}
		
		if auth != authKey {
			s.logger.Warnf("Unauthorized pprof access attempt from %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized access to debug endpoints",
			})
			c.Abort()
			return
		}
		
		c.Next()
	})
}

// GetPprofInfo 獲取 pprof 信息和使用說明
func (s *AdminService) GetPprofInfo(c *gin.Context) {
	info := gin.H{
		"message": "Pprof debugging endpoints",
		"endpoints": gin.H{
			"/debug/pprof/":            "Overview of available profiles",
			"/debug/pprof/cmdline":     "Command line that invoked the target",
			"/debug/pprof/profile":     "CPU profile (add ?seconds=N for N-second sample)",
			"/debug/pprof/symbol":      "Symbol table",
			"/debug/pprof/trace":       "Execution trace (add ?seconds=N for N-second trace)",
			"/debug/pprof/allocs":      "Memory allocation samples",
			"/debug/pprof/block":       "Stack traces that led to blocking on synchronization primitives",
			"/debug/pprof/goroutine":   "Stack traces of all current goroutines",
			"/debug/pprof/heap":        "Memory allocation samples of live objects",
			"/debug/pprof/mutex":       "Stack traces of holders of contended mutexes",
			"/debug/pprof/threadcreate": "Stack traces that led to thread creation",
		},
		"usage": gin.H{
			"go_tool":    "go tool pprof http://localhost:6060/debug/pprof/profile",
			"web_ui":     "http://localhost:6060/debug/pprof/",
			"curl_heap":  "curl http://localhost:6060/debug/pprof/heap > heap.prof",
			"curl_cpu":   "curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof",
		},
		"security_note": "These endpoints should be protected in production environments",
	}
	
	c.JSON(http.StatusOK, info)
}