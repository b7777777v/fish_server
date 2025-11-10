package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TODO: 實現帳號模組的 HTTP API handlers
// 此檔案提供帳號相關的 RESTful API 端點

// RegisterAccountRoutes 註冊帳號相關的路由
func RegisterAccountRoutes(r *gin.Engine /* TODO: 添加 AccountUsecase 參數 */) {
	// TODO: 實現路由註冊
	// 建議路由結構：
	// POST   /api/v1/auth/register          - 使用者註冊
	// POST   /api/v1/auth/login             - 使用者登入
	// POST   /api/v1/auth/guest-login       - 遊客登入
	// POST   /api/v1/auth/oauth/callback    - OAuth 回調
	// GET    /api/v1/user/profile           - 獲取使用者資料（需要認證）
	// PUT    /api/v1/user/profile           - 更新使用者資料（需要認證）
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
func handleRegister(c *gin.Context) {
	// TODO: 實現註冊邏輯
	// 1. 綁定並驗證請求
	// 2. 呼叫 AccountUsecase.Register
	// 3. 返回 JWT token
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleLogin 處理使用者登入
func handleLogin(c *gin.Context) {
	// TODO: 實現登入邏輯
	// 1. 綁定並驗證請求
	// 2. 呼叫 AccountUsecase.Login
	// 3. 返回 JWT token
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleGuestLogin 處理遊客登入
func handleGuestLogin(c *gin.Context) {
	// TODO: 實現遊客登入邏輯
	// 1. 呼叫 AccountUsecase.GuestLogin
	// 2. 返回 JWT token（包含 is_guest: true）
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleOAuthCallback 處理 OAuth 回調
func handleOAuthCallback(c *gin.Context) {
	// TODO: 實現 OAuth 回調邏輯
	// 1. 綁定並驗證請求
	// 2. 呼叫 AccountUsecase.OAuthLogin
	// 3. 返回 JWT token
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleGetProfile 獲取使用者資料
func handleGetProfile(c *gin.Context) {
	// TODO: 實現獲取使用者資料邏輯
	// 1. 從 JWT 中解析 user_id
	// 2. 呼叫 AccountUsecase.GetUserByID
	// 3. 返回使用者資料
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleUpdateProfile 更新使用者資料
func handleUpdateProfile(c *gin.Context) {
	// TODO: 實現更新使用者資料邏輯
	// 1. 從 JWT 中解析 user_id
	// 2. 綁定並驗證請求
	// 3. 呼叫 AccountUsecase.UpdateUser
	// 4. 返回成功訊息
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
