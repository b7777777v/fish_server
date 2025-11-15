package admin

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/b7777777v/fish_server/internal/biz/lobby"
	"github.com/b7777777v/fish_server/internal/pkg/token"
	"github.com/gin-gonic/gin"
)

// LobbyHandler 處理大廳相關的 HTTP 請求
type LobbyHandler struct {
	lobbyUsecase lobby.LobbyUsecase
	tokenHelper  *token.TokenHelper
}

// NewLobbyHandler 建立新的 LobbyHandler
func NewLobbyHandler(lobbyUsecase lobby.LobbyUsecase, tokenHelper *token.TokenHelper) *LobbyHandler {
	return &LobbyHandler{
		lobbyUsecase: lobbyUsecase,
		tokenHelper:  tokenHelper,
	}
}

// adminAuthMiddleware 管理員認證中間件
// 驗證 JWT token 並檢查是否為管理員用戶
func (h *LobbyHandler) adminAuthMiddleware() gin.HandlerFunc {
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

		// 檢查是否為遊客（遊客不能是管理員）
		if claims.IsGuest {
			c.JSON(http.StatusForbidden, gin.H{"error": "guests are not allowed to access admin APIs"})
			c.Abort()
			return
		}

		// 檢查管理員權限
		// TODO: 可以根據業務需求實現更複雜的權限檢查：
		// 1. 從數據庫查詢用戶角色
		// 2. 使用 RBAC (Role-Based Access Control)
		// 3. 檢查環境變數中配置的管理員 ID 列表
		//
		// 當前簡單實現：UserID <= 10 的用戶被視為管理員
		// 生產環境應該使用更安全的權限系統
		if claims.UserID > 10 {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions - admin access required"})
			c.Abort()
			return
		}

		// 將 user_id 存入 context
		c.Set("user_id", claims.UserID)
		c.Set("is_admin", true)
		c.Next()
	}
}

// RegisterLobbyRoutes 註冊大廳相關的路由
func RegisterLobbyRoutes(r *gin.Engine, handler *LobbyHandler, accountHandler *AccountHandler) {
	api := r.Group("/api/v1")

	// 大廳路由（公開）
	lobby := api.Group("/lobby")
	{
		lobby.GET("/rooms", handler.handleGetRoomList)
		lobby.GET("/announcements", handler.handleGetAnnouncements)

		// 玩家狀態需要認證
		lobby.GET("/player-status", accountHandler.authMiddleware(), handler.handleGetPlayerStatus)
	}

	// 管理員路由（需要管理員權限）
	admin := api.Group("/admin")
	admin.Use(handler.adminAuthMiddleware()) // 應用管理員認證中間件
	{
		admin.POST("/announcements", handler.handleCreateAnnouncement)
		admin.PUT("/announcements/:id", handler.handleUpdateAnnouncement)
		admin.DELETE("/announcements/:id", handler.handleDeleteAnnouncement)
	}
}

// CreateAnnouncementRequest 建立公告請求
type CreateAnnouncementRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Priority int    `json:"priority"`
}

// UpdateAnnouncementRequest 更新公告請求
type UpdateAnnouncementRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Priority int    `json:"priority"`
}

// handleGetRoomList 獲取房間列表
func (h *LobbyHandler) handleGetRoomList(c *gin.Context) {
	rooms, err := h.lobbyUsecase.GetRoomList(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}

// handleGetPlayerStatus 獲取玩家狀態
func (h *LobbyHandler) handleGetPlayerStatus(c *gin.Context) {
	// 從 context 中獲取 user_id（由認證中間件設置）
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	playerStatus, err := h.lobbyUsecase.GetPlayerStatus(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player": playerStatus,
	})
}

// handleGetAnnouncements 獲取公告列表
func (h *LobbyHandler) handleGetAnnouncements(c *gin.Context) {
	// 從查詢參數獲取 limit，預設 10
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	announcements, err := h.lobbyUsecase.GetAnnouncements(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"announcements": announcements,
	})
}

// handleCreateAnnouncement 建立公告（管理員功能）
func (h *LobbyHandler) handleCreateAnnouncement(c *gin.Context) {
	var req CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.lobbyUsecase.CreateAnnouncement(c.Request.Context(), req.Title, req.Content, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "announcement created successfully",
	})
}

// handleUpdateAnnouncement 更新公告（管理員功能）
func (h *LobbyHandler) handleUpdateAnnouncement(c *gin.Context) {
	// 解析 announcement_id
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid announcement id"})
		return
	}

	var req UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.lobbyUsecase.UpdateAnnouncement(c.Request.Context(), id, req.Title, req.Content, req.Priority)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "announcement updated successfully",
	})
}

// handleDeleteAnnouncement 刪除公告（管理員功能）
func (h *LobbyHandler) handleDeleteAnnouncement(c *gin.Context) {
	// 解析 announcement_id
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid announcement id"})
		return
	}

	err = h.lobbyUsecase.DeleteAnnouncement(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "announcement deleted successfully",
	})
}
