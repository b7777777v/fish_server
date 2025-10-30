package admin

import (
	"net/http"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse 健康檢查響應
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// ServerStatusResponse 伺服器狀態響應
type ServerStatusResponse struct {
	Status     string                 `json:"status"`
	Timestamp  time.Time              `json:"timestamp"`
	Uptime     string                 `json:"uptime"`
	Memory     MemoryStats            `json:"memory"`
	Goroutines int                    `json:"goroutines"`
	System     SystemStats            `json:"system"`
	Service    ServiceStats           `json:"service"`
	Database   map[string]interface{} `json:"database,omitempty"`
	Redis      map[string]interface{} `json:"redis,omitempty"`
}

// MemoryStats 記憶體統計
type MemoryStats struct {
	Alloc      string `json:"alloc"`       // 當前分配的記憶體
	TotalAlloc string `json:"total_alloc"` // 總分配記憶體
	Sys        string `json:"sys"`         // 系統記憶體
	NumGC      uint32 `json:"num_gc"`      // GC 次數
}

// SystemStats 系統統計
type SystemStats struct {
	NumCPU       int    `json:"num_cpu"`
	GOMAXPROCS   int    `json:"gomaxprocs"`
	GoVersion    string `json:"go_version"`
	Architecture string `json:"architecture"`
	OS           string `json:"os"`
}

// ServiceStats 服務統計
type ServiceStats struct {
	StartTime time.Time `json:"start_time"`
	Players   int64     `json:"players_count,omitempty"`
	Wallets   int64     `json:"wallets_count,omitempty"`
}

var (
	startTime = time.Now() // 服務啟動時間
	version   = "1.0.0"    // 可以通過編譯時注入
)

// RegisterRoutes 註冊管理後台路由
func (s *AdminService) RegisterRoutes(r *gin.Engine) {
	// 管理後台 API 組
	admin := r.Group("/admin")
	{
		// 健康檢查
		admin.GET("/health", s.HealthCheck)
		admin.GET("/health/live", s.LivenessCheck)
		admin.GET("/health/ready", s.ReadinessCheck)

		// 伺服器狀態
		admin.GET("/status", s.ServerStatus)
		admin.GET("/metrics", s.Metrics)

		// 玩家管理
		players := admin.Group("/players")
		{
			players.GET("/:id", s.GetPlayer)
			players.GET("/:id/wallets", s.GetPlayerWallets)
		}

		// 錢包管理
		wallets := admin.Group("/wallets")
		{
			wallets.GET("/:id", s.GetWallet)
			wallets.GET("/:id/transactions", s.GetWalletTransactions)
			wallets.POST("/:id/freeze", s.FreezeWallet)
			wallets.POST("/:id/unfreeze", s.UnfreezeWallet)
			wallets.POST("/:id/deposit", s.DepositToWallet)
			wallets.POST("/:id/withdraw", s.WithdrawFromWallet)
		}
	}

	// pprof 路由（性能分析）
	s.registerPprofRoutes(r)
}

// HealthCheck 一般健康檢查
func (s *AdminService) HealthCheck(c *gin.Context) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version,
		Checks: map[string]string{
			"service": "ok",
		},
	}

	c.JSON(http.StatusOK, response)
}

// LivenessCheck 存活檢查（用於 Kubernetes liveness probe）
func (s *AdminService) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"timestamp": time.Now(),
	})
}

// ReadinessCheck 就緒檢查（用於 Kubernetes readiness probe）
func (s *AdminService) ReadinessCheck(c *gin.Context) {
	// 這裡可以檢查依賴服務是否可用
	// 例如資料庫連接、Redis 連接等
	
	checks := map[string]string{
		"database": "ok", // 實際項目中應該檢查資料庫連接
		"redis":    "ok", // 實際項目中應該檢查 Redis 連接
	}

	allHealthy := true
	for _, status := range checks {
		if status != "ok" {
			allHealthy = false
			break
		}
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status":    map[bool]string{true: "ready", false: "not_ready"}[allHealthy],
		"timestamp": time.Now(),
		"checks":    checks,
	})
}

// ServerStatus 獲取伺服器狀態
func (s *AdminService) ServerStatus(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	uptime := time.Since(startTime)

	response := ServerStatusResponse{
		Status:     "running",
		Timestamp:  time.Now(),
		Uptime:     uptime.String(),
		Goroutines: runtime.NumGoroutine(),
		Memory: MemoryStats{
			Alloc:      formatBytes(m.Alloc),
			TotalAlloc: formatBytes(m.TotalAlloc),
			Sys:        formatBytes(m.Sys),
			NumGC:      m.NumGC,
		},
		System: SystemStats{
			NumCPU:       runtime.NumCPU(),
			GOMAXPROCS:   runtime.GOMAXPROCS(0),
			GoVersion:    runtime.Version(),
			Architecture: runtime.GOARCH,
			OS:           runtime.GOOS,
		},
		Service: ServiceStats{
			StartTime: startTime,
			// Players 和 Wallets 統計可以通過 usecase 獲取
		},
	}

	c.JSON(http.StatusOK, response)
}

// Metrics 獲取詳細指標（Prometheus 格式或 JSON 格式）
func (s *AdminService) Metrics(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := gin.H{
		"timestamp": time.Now(),
		"uptime_seconds": time.Since(startTime).Seconds(),
		"memory": gin.H{
			"alloc_bytes":       m.Alloc,
			"total_alloc_bytes": m.TotalAlloc,
			"sys_bytes":         m.Sys,
			"mallocs":           m.Mallocs,
			"frees":             m.Frees,
			"heap_alloc_bytes":  m.HeapAlloc,
			"heap_sys_bytes":    m.HeapSys,
			"heap_idle_bytes":   m.HeapIdle,
			"heap_inuse_bytes":  m.HeapInuse,
			"heap_released_bytes": m.HeapReleased,
			"heap_objects":      m.HeapObjects,
			"stack_inuse_bytes": m.StackInuse,
			"stack_sys_bytes":   m.StackSys,
			"gc_num":            m.NumGC,
			"gc_total_pause_ns": m.PauseTotalNs,
		},
		"goroutines": runtime.NumGoroutine(),
		"system": gin.H{
			"num_cpu":     runtime.NumCPU(),
			"gomaxprocs":  runtime.GOMAXPROCS(0),
			"go_version":  runtime.Version(),
			"arch":        runtime.GOARCH,
			"os":          runtime.GOOS,
		},
	}

	c.JSON(http.StatusOK, metrics)
}

// formatBytes 格式化位元組數
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatUint(bytes, 10) + " B"
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(bytes)/float64(div), 'f', 1, 64) + " " + 
		[]string{"", "K", "M", "G", "T", "P", "E", "Z", "Y"}[exp] + "B"
}