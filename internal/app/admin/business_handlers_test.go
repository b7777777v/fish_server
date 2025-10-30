package admin

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/b7777777v/fish_server/internal/biz/wallet"
)

func TestGetWallet(t *testing.T) {
	service, _, mockWalletUC := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service.RegisterRoutes(r)
	
	t.Run("Success", func(t *testing.T) {
		// 準備測試數據
		testWallet := &wallet.Wallet{
			ID:        123,
			UserID:    456,
			Balance:   1000.0,
			Currency:  "USD",
			Status:    1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		// 設置模擬期望
		mockWalletUC.On("GetWallet", mock.Anything, uint(123)).Return(testWallet, nil).Once()
		
		// 執行請求
		req, _ := http.NewRequest("GET", "/admin/wallets/123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response WalletResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, uint(123), response.ID)
		assert.Equal(t, uint(456), response.UserID)
		assert.Equal(t, 1000.0, response.Balance)
		assert.Equal(t, "USD", response.Currency)
		assert.Equal(t, 1, response.Status)
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
	
	t.Run("Wallet not found", func(t *testing.T) {
		// 設置模擬期望
		mockWalletUC.On("GetWallet", mock.Anything, uint(999)).Return(nil, errors.New("wallet not found")).Once()
		
		// 執行請求
		req, _ := http.NewRequest("GET", "/admin/wallets/999", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusNotFound, w.Code)
		
		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Wallet not found", response.Error)
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
	
	t.Run("Invalid wallet ID", func(t *testing.T) {
		// 執行請求
		req, _ := http.NewRequest("GET", "/admin/wallets/invalid", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid wallet ID", response.Error)
	})
}

func TestGetWalletTransactions(t *testing.T) {
	service, _, mockWalletUC := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service.RegisterRoutes(r)
	
	t.Run("Success with default pagination", func(t *testing.T) {
		// 準備測試數據
		testTransactions := []*wallet.Transaction{
			{
				ID:            1,
				WalletID:      123,
				Amount:        100.0,
				BalanceBefore: 900.0,
				BalanceAfter:  1000.0,
				Type:          "deposit",
				Status:        1,
				ReferenceID:   "ref_001",
				Description:   "Test deposit",
				Metadata:      map[string]interface{}{"test": true},
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			},
		}
		
		// 設置模擬期望
		mockWalletUC.On("GetTransactions", mock.Anything, uint(123), 10, 0).Return(testTransactions, nil).Once()
		
		// 執行請求
		req, _ := http.NewRequest("GET", "/admin/wallets/123/transactions", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		transactions := response["transactions"].([]interface{})
		assert.Len(t, transactions, 1)
		assert.Equal(t, float64(1), response["total"])
		assert.Equal(t, float64(10), response["limit"])
		assert.Equal(t, float64(0), response["offset"])
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
	
	t.Run("Success with custom pagination", func(t *testing.T) {
		// 設置模擬期望
		mockWalletUC.On("GetTransactions", mock.Anything, uint(123), 5, 10).Return([]*wallet.Transaction{}, nil).Once()
		
		// 執行請求
		req, _ := http.NewRequest("GET", "/admin/wallets/123/transactions?limit=5&offset=10", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(5), response["limit"])
		assert.Equal(t, float64(10), response["offset"])
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
}

func TestFreezeWallet(t *testing.T) {
	service, _, mockWalletUC := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service.RegisterRoutes(r)
	
	t.Run("Success", func(t *testing.T) {
		// 設置模擬期望
		mockWalletUC.On("FreezeWallet", mock.Anything, uint(123)).Return(nil).Once()
		
		// 執行請求
		req, _ := http.NewRequest("POST", "/admin/wallets/123/freeze", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Wallet frozen successfully", response["message"])
		assert.Equal(t, float64(123), response["wallet_id"])
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
	
	t.Run("Freeze failed", func(t *testing.T) {
		// 設置模擬期望
		mockWalletUC.On("FreezeWallet", mock.Anything, uint(123)).Return(errors.New("freeze failed")).Once()
		
		// 執行請求
		req, _ := http.NewRequest("POST", "/admin/wallets/123/freeze", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to freeze wallet", response.Error)
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
}

func TestUnfreezeWallet(t *testing.T) {
	service, _, mockWalletUC := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service.RegisterRoutes(r)
	
	t.Run("Success", func(t *testing.T) {
		// 設置模擬期望
		mockWalletUC.On("UnfreezeWallet", mock.Anything, uint(123)).Return(nil).Once()
		
		// 執行請求
		req, _ := http.NewRequest("POST", "/admin/wallets/123/unfreeze", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Wallet unfrozen successfully", response["message"])
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
}

func TestDepositToWallet(t *testing.T) {
	service, _, mockWalletUC := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service.RegisterRoutes(r)
	
	t.Run("Success", func(t *testing.T) {
		// 準備請求數據
		requestData := WalletOperationRequest{
			Amount:      100.0,
			Type:        "admin_deposit",
			ReferenceID: "ref_001",
			Description: "Test deposit",
			Metadata:    map[string]interface{}{"test": true},
		}
		
		jsonData, _ := json.Marshal(requestData)
		
		// 設置模擬期望
		mockWalletUC.On("Deposit", mock.Anything, uint(123), 100.0, "admin_deposit", "ref_001", "Test deposit", mock.MatchedBy(func(metadata map[string]interface{}) bool {
			return metadata["admin_operation"] == true && metadata["test"] == true
		})).Return(nil).Once()
		
		// 執行請求
		req, _ := http.NewRequest("POST", "/admin/wallets/123/deposit", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Deposit successful", response["message"])
		assert.Equal(t, float64(100.0), response["amount"])
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
	
	t.Run("Invalid request body", func(t *testing.T) {
		// 執行請求（空的 JSON）
		req, _ := http.NewRequest("POST", "/admin/wallets/123/deposit", bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response.Error)
	})
	
	t.Run("Negative amount", func(t *testing.T) {
		// 準備請求數據
		requestData := WalletOperationRequest{
			Amount: -100.0,
		}
		
		jsonData, _ := json.Marshal(requestData)
		
		// 執行請求
		req, _ := http.NewRequest("POST", "/admin/wallets/123/deposit", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWithdrawFromWallet(t *testing.T) {
	service, _, mockWalletUC := setupTestAdminService()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	service.RegisterRoutes(r)
	
	t.Run("Success", func(t *testing.T) {
		// 準備請求數據
		requestData := WalletOperationRequest{
			Amount:      50.0,
			Description: "Test withdrawal",
		}
		
		jsonData, _ := json.Marshal(requestData)
		
		// 設置模擬期望
		mockWalletUC.On("Withdraw", mock.Anything, uint(123), 50.0, "admin_withdraw", "", "Test withdrawal", mock.MatchedBy(func(metadata map[string]interface{}) bool {
			return metadata["admin_operation"] == true
		})).Return(nil).Once()
		
		// 執行請求
		req, _ := http.NewRequest("POST", "/admin/wallets/123/withdraw", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Withdrawal successful", response["message"])
		assert.Equal(t, float64(50.0), response["amount"])
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
	
	t.Run("Insufficient funds", func(t *testing.T) {
		// 準備請求數據
		requestData := WalletOperationRequest{
			Amount: 1000.0,
		}
		
		jsonData, _ := json.Marshal(requestData)
		
		// 設置模擬期望
		mockWalletUC.On("Withdraw", mock.Anything, uint(123), 1000.0, "admin_withdraw", "", "Admin withdraw operation", mock.Anything).Return(errors.New("insufficient funds")).Once()
		
		// 執行請求
		req, _ := http.NewRequest("POST", "/admin/wallets/123/withdraw", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		
		// 驗證結果
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to withdraw", response.Error)
		
		// 驗證模擬調用
		mockWalletUC.AssertExpectations(t)
	})
}