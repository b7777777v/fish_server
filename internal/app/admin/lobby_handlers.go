package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TODO: 實現大廳模組的 HTTP API handlers
// 此檔案提供大廳相關的 RESTful API 端點

// RegisterLobbyRoutes 註冊大廳相關的路由
func RegisterLobbyRoutes(r *gin.Engine /* TODO: 添加 LobbyUsecase 參數 */) {
	// TODO: 實現路由註冊
	// 建議路由結構：
	// GET    /api/v1/lobby/rooms            - 獲取房間列表
	// GET    /api/v1/lobby/player-status    - 獲取玩家狀態（需要認證）
	// GET    /api/v1/lobby/announcements    - 獲取公告列表
	//
	// 管理員路由：
	// POST   /api/v1/admin/announcements    - 建立公告（需要管理員權限）
	// PUT    /api/v1/admin/announcements/:id - 更新公告（需要管理員權限）
	// DELETE /api/v1/admin/announcements/:id - 刪除公告（需要管理員權限）
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
func handleGetRoomList(c *gin.Context) {
	// TODO: 實現獲取房間列表邏輯
	// 1. 呼叫 LobbyUsecase.GetRoomList
	// 2. 返回房間列表
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleGetPlayerStatus 獲取玩家狀態
func handleGetPlayerStatus(c *gin.Context) {
	// TODO: 實現獲取玩家狀態邏輯
	// 1. 從 JWT 中解析 user_id
	// 2. 呼叫 LobbyUsecase.GetPlayerStatus
	// 3. 返回玩家狀態
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleGetAnnouncements 獲取公告列表
func handleGetAnnouncements(c *gin.Context) {
	// TODO: 實現獲取公告列表邏輯
	// 1. 呼叫 LobbyUsecase.GetAnnouncements
	// 2. 返回公告列表
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleCreateAnnouncement 建立公告（管理員功能）
func handleCreateAnnouncement(c *gin.Context) {
	// TODO: 實現建立公告邏輯
	// 1. 驗證管理員權限
	// 2. 綁定並驗證請求
	// 3. 呼叫 LobbyUsecase.CreateAnnouncement
	// 4. 返回成功訊息
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleUpdateAnnouncement 更新公告（管理員功能）
func handleUpdateAnnouncement(c *gin.Context) {
	// TODO: 實現更新公告邏輯
	// 1. 驗證管理員權限
	// 2. 解析 announcement_id
	// 3. 綁定並驗證請求
	// 4. 呼叫 LobbyUsecase.UpdateAnnouncement
	// 5. 返回成功訊息
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// handleDeleteAnnouncement 刪除公告（管理員功能）
func handleDeleteAnnouncement(c *gin.Context) {
	// TODO: 實現刪除公告邏輯
	// 1. 驗證管理員權限
	// 2. 解析 announcement_id
	// 3. 呼叫 LobbyUsecase.DeleteAnnouncement
	// 4. 返回成功訊息
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
