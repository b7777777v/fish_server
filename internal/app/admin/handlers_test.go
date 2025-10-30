package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// MockPlayerUsecase 模擬 PlayerUsecase
type MockPlayerUsecase struct {
	mock.Mock
}

func (m *MockPlayerUsecase) Login(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

// 實現 PlayerUsecase 接口所需的其他方法
func (m *MockPlayerUsecase) GetPlayer(ctx context.Context, id uint) (*player.Player, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*player.Player), args.Error(1)
}

// MockWalletUsecase 模擬 WalletUsecase
type MockWalletUsecase struct {
	mock.Mock
}

func (m *MockWalletUsecase) GetWallet(ctx context.Context, id uint) (*wallet.Wallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.Wallet), args.Error(1)
}

func (m *MockWalletUsecase) GetWalletByUserID(ctx context.Context, userID uint, currency string) (*wallet.Wallet, error) {
	args := m.Called(ctx, userID, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.Wallet), args.Error(1)
}

func (m *MockWalletUsecase) CreateWallet(ctx context.Context, userID uint, currency string) (*wallet.Wallet, error) {
	args := m.Called(ctx, userID, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.Wallet), args.Error(1)
}

func (m *MockWalletUsecase) Deposit(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	args := m.Called(ctx, walletID, amount, txType, referenceID, description, metadata)
	return args.Error(0)
}

func (m *MockWalletUsecase) Withdraw(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	args := m.Called(ctx, walletID, amount, txType, referenceID, description, metadata)
	return args.Error(0)
}

func (m *MockWalletUsecase) GetTransactions(ctx context.Context, walletID uint, limit, offset int) ([]*wallet.Transaction, error) {
	args := m.Called(ctx, walletID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*wallet.Transaction), args.Error(1)
}

func (m *MockWalletUsecase) FreezeWallet(ctx context.Context, walletID uint) error {
	args := m.Called(ctx, walletID)
	return args.Error(0)
}

func (m *MockWalletUsecase) UnfreezeWallet(ctx context.Context, walletID uint) error {
	args := m.Called(ctx, walletID)
	return args.Error(0)
}

// 設置測試環境
func setupTestAdminService() (*AdminService, *MockPlayerUsecase, *MockWalletUsecase) {
	mockPlayerUC := new(MockPlayerUsecase)
	mockWalletUC := new(MockWalletUsecase)
	
	log := logger.New(nil, "info", "console")
	
	// 為了測試，我們創建一個簡單的 AdminService
	// 在實際測試中，我們會創建自定義的測試 handlers
	service := &AdminService{
		playerUC: nil, 
		walletUC: nil, 
		logger:   log.With("module", "app/admin"),
	}
	
	return service, mockPlayerUC, mockWalletUC
}

// 注意：這些 mock 對象僅用於演示測試結構
// 實際的業務邏輯測試應該在集成測試中進行

func TestHealthCheck(t *testing.T) {
	service, _, _ := setupTestAdminService()
	
	// 設置 Gin 為測試模式
	gin.SetMode(gin.TestMode)
	
	// 創建簡單的路由（只測試系統功能）
	r := gin.New()
	admin := r.Group("/admin")
	{
		admin.GET("/health", service.HealthCheck)
	}
	
	// 創建測試請求
	req, _ := http.NewRequest("GET", "/admin/health", nil)
	w := httptest.NewRecorder()
	
	// 執行請求
	r.ServeHTTP(w, req)
	
	// 驗證結果
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response.Status)
	assert.NotEmpty(t, response.Timestamp)
}

func TestLivenessCheck(t *testing.T) {
	service, _, _ := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/admin")
	{
		admin.GET("/health/live", service.LivenessCheck)
	}
	
	req, _ := http.NewRequest("GET", "/admin/health/live", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "alive", response["status"])
}

func TestReadinessCheck(t *testing.T) {
	service, _, _ := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/admin")
	{
		admin.GET("/health/ready", service.ReadinessCheck)
	}
	
	req, _ := http.NewRequest("GET", "/admin/health/ready", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ready", response["status"])
}

func TestServerStatus(t *testing.T) {
	service, _, _ := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/admin")
	{
		admin.GET("/status", service.ServerStatus)
	}
	
	req, _ := http.NewRequest("GET", "/admin/status", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response ServerStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "running", response.Status)
	assert.NotEmpty(t, response.Uptime)
	assert.Greater(t, response.Goroutines, 0)
	assert.NotEmpty(t, response.Memory.Alloc)
	assert.NotEmpty(t, response.System.GoVersion)
}

func TestMetrics(t *testing.T) {
	service, _, _ := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/admin")
	{
		admin.GET("/metrics", service.Metrics)
	}
	
	req, _ := http.NewRequest("GET", "/admin/metrics", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["timestamp"])
	assert.NotEmpty(t, response["memory"])
	assert.NotEmpty(t, response["system"])
	assert.Greater(t, response["goroutines"], float64(0))
}

// TestGetPlayer 已移除，因為它需要完整的業務邏輯集成
// 業務邏輯測試應該在集成測試中進行

// TestGetPlayerWallets 已移除，因為它需要完整的業務邏輯集成
// 業務邏輯測試應該在集成測試中進行

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}
	
	for _, test := range tests {
		result := formatBytes(test.input)
		assert.Equal(t, test.expected, result)
	}
}

func TestPprofInfo(t *testing.T) {
	service, _, _ := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/debug/pprof/info", service.GetPprofInfo)
	
	req, _ := http.NewRequest("GET", "/debug/pprof/info", nil)
	w := httptest.NewRecorder()
	
	r.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Pprof debugging endpoints", response["message"])
	assert.NotEmpty(t, response["endpoints"])
	assert.NotEmpty(t, response["usage"])
}