package admin

import (
	"net/http"
	"strings"

	"github.com/b7777777v/fish_server/internal/biz/account"
	"github.com/b7777777v/fish_server/internal/pkg/token"
	"github.com/gin-gonic/gin"
)

// AccountHandler 處理帳號相關的 HTTP 請求
type AccountHandler struct {
	accountUsecase account.AccountUsecase
	tokenHelper    *token.TokenHelper
}

// NewAccountHandler 建立新的 AccountHandler
func NewAccountHandler(accountUsecase account.AccountUsecase, tokenHelper *token.TokenHelper) *AccountHandler {
	return &AccountHandler{
		accountUsecase: accountUsecase,
		tokenHelper:    tokenHelper,
	}
}

// RegisterAccountRoutes 註冊帳號相關的路由
func RegisterAccountRoutes(r *gin.Engine, handler *AccountHandler) {
	api := r.Group("/api/v1")

	// 認證路由（不需要登入）
	auth := api.Group("/auth")
	{
		auth.POST("/register", handler.handleRegister)
		auth.POST("/login", handler.handleLogin)
		auth.POST("/guest-login", handler.handleGuestLogin)
		auth.POST("/oauth/callback", handler.handleOAuthCallback)
	}

	// 使用者路由（需要認證）
	user := api.Group("/user")
	user.Use(handler.authMiddleware())
	{
		user.GET("/profile", handler.handleGetProfile)
		user.PUT("/profile", handler.handleUpdateProfile)
	}
}

// authMiddleware JWT 認證中間件
func (h *AccountHandler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 Authorization header 獲取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 驗證 token
		claims, err := h.tokenHelper.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// 將 user_id、is_guest 和 nickname 存入 context
		c.Set("user_id", claims.UserID)
		c.Set("is_guest", claims.IsGuest)
		c.Set("nickname", claims.Nickname)
		c.Next()
	}
}

// RegisterRequest 註冊請求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登入請求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// OAuthCallbackRequest OAuth 回調請求
type OAuthCallbackRequest struct {
	Provider string `json:"provider" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

// UpdateProfileRequest 更新資料請求
type UpdateProfileRequest struct {
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

// handleRegister 處理使用者註冊
func (h *AccountHandler) handleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 記錄詳細的綁定錯誤
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"message": "請求數據格式不正確或缺少必要字段",
		})
		return
	}

	// 呼叫 AccountUsecase.Register
	user, err := h.accountUsecase.Register(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"message": "註冊失敗",
		})
		return
	}

	// 生成 JWT token
	token, err := h.tokenHelper.GenerateTokenWithClaims(user.ID, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// handleLogin 處理使用者登入
func (h *AccountHandler) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 呼叫 AccountUsecase.Login
	token, err := h.accountUsecase.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// handleGuestLogin 處理遊客登入
func (h *AccountHandler) handleGuestLogin(c *gin.Context) {
	// 呼叫 AccountUsecase.GuestLogin
	token, err := h.accountUsecase.GuestLogin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// handleOAuthCallback 處理 OAuth 回調
func (h *AccountHandler) handleOAuthCallback(c *gin.Context) {
	var req OAuthCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 呼叫 AccountUsecase.OAuthLogin
	token, err := h.accountUsecase.OAuthLogin(c.Request.Context(), req.Provider, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

// handleGetProfile 獲取使用者資料
func (h *AccountHandler) handleGetProfile(c *gin.Context) {
	// 從 context 中獲取 user_id 和 is_guest
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	isGuest, _ := c.Get("is_guest")

	// 如果是遊客，返回虛擬用戶信息（不查詢數據庫）
	if isGuest != nil && isGuest.(bool) {
		nickname, _ := c.Get("nickname")
		guestUser := gin.H{
			"id":       userID,
			"nickname": nickname,
			"is_guest": true,
		}
		c.JSON(http.StatusOK, gin.H{
			"user": guestUser,
		})
		return
	}

	// 一般用戶：呼叫 AccountUsecase.GetUserByID
	user, err := h.accountUsecase.GetUserByID(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// handleUpdateProfile 更新使用者資料
func (h *AccountHandler) handleUpdateProfile(c *gin.Context) {
	// 從 context 中獲取 user_id
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 呼叫 AccountUsecase.UpdateUser
	err := h.accountUsecase.UpdateUser(c.Request.Context(), userID.(int64), req.Nickname, req.AvatarURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "profile updated successfully",
	})
}
