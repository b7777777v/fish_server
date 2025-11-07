package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PlayerResponse 玩家信息響應
type PlayerResponse struct {
	ID           uint   `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Status       int    `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	WalletsCount int    `json:"wallets_count,omitempty"`
}

// WalletResponse 錢包信息響應
type WalletResponse struct {
	ID        uint    `json:"id"`
	UserID    uint    `json:"user_id"`
	Balance   float64 `json:"balance"`
	Currency  string  `json:"currency"`
	Status    int     `json:"status"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// TransactionResponse 交易記錄響應
type TransactionResponse struct {
	ID            uint                   `json:"id"`
	WalletID      uint                   `json:"wallet_id"`
	Amount        float64                `json:"amount"`
	BalanceBefore float64                `json:"balance_before"`
	BalanceAfter  float64                `json:"balance_after"`
	Type          string                 `json:"type"`
	Status        int                    `json:"status"`
	ReferenceID   string                 `json:"reference_id"`
	Description   string                 `json:"description"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	UpdatedAt     string                 `json:"updated_at"`
}

// WalletOperationRequest 錢包操作請求
type WalletOperationRequest struct {
	Amount      float64                `json:"amount" binding:"required,gt=0"`
	Type        string                 `json:"type,omitempty"`
	ReferenceID string                 `json:"reference_id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ErrorResponse 錯誤響應
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token         string `json:"token"`
	GameServerURL string `json:"game_server_url"`
}

// Login handles player login.
func (s *AdminService) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	token, err := s.playerUC.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		s.logger.Errorf("Failed to login player: %v", err)
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid credentials",
			Message: "Invalid username or password",
		})
		return
	}

	gameServerURL := "ws://localhost:9090/ws"
	if s.config.Server != nil && s.config.Server.Game != nil {
		gameServerURL = "ws://localhost:" + strconv.Itoa(s.config.Server.Game.Port) + "/ws"
	}

	response := LoginResponse{
		Token:         token,
		GameServerURL: gameServerURL,
	}

	c.JSON(http.StatusOK, response)
}

// GetPlayer 獲取玩家信息
func (s *AdminService) GetPlayer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid player ID",
			Message: "Player ID must be a valid number",
		})
		return
	}

	// 這裡需要在 PlayerUsecase 中添加 GetPlayer 方法
	// 暫時返回模擬數據
	s.logger.Infof("Getting player info for ID: %d", id)

	response := PlayerResponse{
		ID:        uint(id),
		Username:  "player_" + idStr, // 模擬數據
		Email:     "player" + idStr + "@example.com",
		Status:    1,
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-01T00:00:00Z",
	}

	c.JSON(http.StatusOK, response)
}

// GetPlayerWallets 獲取玩家的錢包列表
func (s *AdminService) GetPlayerWallets(c *gin.Context) {
	idStr := c.Param("id")
	userID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid player ID",
			Message: "Player ID must be a valid number",
		})
		return
	}

	s.logger.Infof("Getting wallets for player ID: %d", userID)

	// 這裡應該通過 WalletUsecase 獲取用戶的所有錢包
	// 暫時返回模擬數據
	wallets := []WalletResponse{
		{
			ID:        100 + uint(userID),
			UserID:    uint(userID),
			Balance:   1000.0,
			Currency:  "USD",
			Status:    1,
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
		},
		{
			ID:        200 + uint(userID),
			UserID:    uint(userID),
			Balance:   500.0,
			Currency:  "CNY",
			Status:    1,
			CreatedAt: "2024-01-01T00:00:00Z",
			UpdatedAt: "2024-01-01T00:00:00Z",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"wallets": wallets,
		"total":   len(wallets),
	})
}

// GetWallet 獲取錢包詳細信息
func (s *AdminService) GetWallet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid wallet ID",
			Message: "Wallet ID must be a valid number",
		})
		return
	}

	wallet, err := s.walletUC.GetWallet(c.Request.Context(), uint(id))
	if err != nil {
		s.logger.Errorf("Failed to get wallet %d: %v", id, err)
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Wallet not found",
			Message: "The specified wallet does not exist",
		})
		return
	}

	response := WalletResponse{
		ID:        wallet.ID,
		UserID:    wallet.UserID,
		Balance:   wallet.Balance,
		Currency:  wallet.Currency,
		Status:    int(wallet.Status),
		CreatedAt: wallet.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: wallet.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// GetWalletTransactions 獲取錢包交易記錄
func (s *AdminService) GetWalletTransactions(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid wallet ID",
			Message: "Wallet ID must be a valid number",
		})
		return
	}

	// 獲取分頁參數
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	transactions, err := s.walletUC.GetTransactions(c.Request.Context(), uint(id), limit, offset)
	if err != nil {
		s.logger.Errorf("Failed to get transactions for wallet %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get transactions",
			Message: "Unable to retrieve transaction history",
		})
		return
	}

	response := make([]TransactionResponse, len(transactions))
	for i, tx := range transactions {
		response[i] = TransactionResponse{
			ID:            tx.ID,
			WalletID:      tx.WalletID,
			Amount:        tx.Amount,
			BalanceBefore: tx.BalanceBefore,
			BalanceAfter:  tx.BalanceAfter,
			Type:          tx.Type,
			Status:        int(tx.Status),
			ReferenceID:   tx.ReferenceID,
			Description:   tx.Description,
			Metadata:      tx.Metadata,
			CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     tx.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": response,
		"total":        len(response),
		"limit":        limit,
		"offset":       offset,
	})
}

// CreatePlayer 創建新玩家
func (s *AdminService) CreatePlayer(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "Not Implemented"})
}

// UpdatePlayer 更新玩家信息
func (s *AdminService) UpdatePlayer(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "Not Implemented"})
}

// DeletePlayer 刪除玩家
func (s *AdminService) DeletePlayer(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "Not Implemented"})
}

// BanPlayer 封禁玩家
func (s *AdminService) BanPlayer(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "Not Implemented"})
}

// UnbanPlayer 解封玩家
func (s *AdminService) UnbanPlayer(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, ErrorResponse{Error: "Not Implemented"})
}

// FreezeWallet 凍結錢包
func (s *AdminService) FreezeWallet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid wallet ID",
			Message: "Wallet ID must be a valid number",
		})
		return
	}

	err = s.walletUC.FreezeWallet(c.Request.Context(), uint(id))
	if err != nil {
		s.logger.Errorf("Failed to freeze wallet %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to freeze wallet",
			Message: "Unable to freeze the specified wallet",
		})
		return
	}

	s.logger.Infof("Wallet %d has been frozen", id)
	c.JSON(http.StatusOK, gin.H{
		"message":   "Wallet frozen successfully",
		"wallet_id": id,
	})
}

// UnfreezeWallet 解凍錢包
func (s *AdminService) UnfreezeWallet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid wallet ID",
			Message: "Wallet ID must be a valid number",
		})
		return
	}

	err = s.walletUC.UnfreezeWallet(c.Request.Context(), uint(id))
	if err != nil {
		s.logger.Errorf("Failed to unfreeze wallet %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to unfreeze wallet",
			Message: "Unable to unfreeze the specified wallet",
		})
		return
	}

	s.logger.Infof("Wallet %d has been unfrozen", id)
	c.JSON(http.StatusOK, gin.H{
		"message":   "Wallet unfrozen successfully",
		"wallet_id": id,
	})
}

// DepositToWallet 向錢包存款
func (s *AdminService) DepositToWallet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid wallet ID",
			Message: "Wallet ID must be a valid number",
		})
		return
	}

	var req WalletOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 設置默認值
	if req.Type == "" {
		req.Type = "admin_deposit"
	}
	if req.Description == "" {
		req.Description = "Admin deposit operation"
	}
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["admin_operation"] = true

	err = s.walletUC.Deposit(c.Request.Context(), uint(id), req.Amount, req.Type, req.ReferenceID, req.Description, req.Metadata)
	if err != nil {
		s.logger.Errorf("Failed to deposit to wallet %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to deposit",
			Message: "Unable to process the deposit",
		})
		return
	}

	s.logger.Infof("Deposited %.2f to wallet %d", req.Amount, id)
	c.JSON(http.StatusOK, gin.H{
		"message":   "Deposit successful",
		"wallet_id": id,
		"amount":    req.Amount,
	})
}

// WithdrawFromWallet 從錢包提款
func (s *AdminService) WithdrawFromWallet(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid wallet ID",
			Message: "Wallet ID must be a valid number",
		})
		return
	}

	var req WalletOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// 設置默認值
	if req.Type == "" {
		req.Type = "admin_withdraw"
	}
	if req.Description == "" {
		req.Description = "Admin withdraw operation"
	}
	if req.Metadata == nil {
		req.Metadata = make(map[string]interface{})
	}
	req.Metadata["admin_operation"] = true

	err = s.walletUC.Withdraw(c.Request.Context(), uint(id), req.Amount, req.Type, req.ReferenceID, req.Description, req.Metadata)
	if err != nil {
		s.logger.Errorf("Failed to withdraw from wallet %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to withdraw",
			Message: "Unable to process the withdrawal",
		})
		return
	}

	s.logger.Infof("Withdrew %.2f from wallet %d", req.Amount, id)
	c.JSON(http.StatusOK, gin.H{
		"message":   "Withdrawal successful",
		"wallet_id": id,
		"amount":    req.Amount,
	})
}
