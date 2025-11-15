package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/gin-gonic/gin"
)

// FishTideHandler 處理魚潮相關的 HTTP 請求
type FishTideHandler struct {
	repo    game.FishTideRepo
	manager game.FishTideManager
}

// NewFishTideHandler 建立新的 FishTideHandler
func NewFishTideHandler(repo game.FishTideRepo, manager game.FishTideManager) *FishTideHandler {
	return &FishTideHandler{
		repo:    repo,
		manager: manager,
	}
}

// RegisterFishTideRoutes 註冊魚潮相關的路由
func RegisterFishTideRoutes(r *gin.Engine, handler *FishTideHandler, lobbyHandler *LobbyHandler) {
	admin := r.Group("/api/v1/admin")
	admin.Use(lobbyHandler.adminAuthMiddleware()) // 應用管理員認證中間件
	{
		admin.GET("/fish-tides", handler.handleGetFishTides)
		admin.POST("/fish-tides", handler.handleCreateFishTide)
		admin.PUT("/fish-tides/:id", handler.handleUpdateFishTide)
		admin.DELETE("/fish-tides/:id", handler.handleDeleteFishTide)
		admin.POST("/fish-tides/:id/start", handler.handleStartFishTide)
		admin.POST("/fish-tides/:id/stop", handler.handleStopFishTide)
	}
}

// CreateFishTideRequest 建立魚潮請求
type CreateFishTideRequest struct {
	Name            string  `json:"name" binding:"required"`
	FishTypeID      int32   `json:"fish_type_id" binding:"required"`
	FishCount       int     `json:"fish_count" binding:"required,min=1"`
	DurationSeconds int     `json:"duration_seconds" binding:"required,min=1"`
	IntervalMs      int     `json:"interval_ms" binding:"required,min=1"`
	SpeedMultiplier float64 `json:"speed_multiplier" binding:"required,min=0.1"`
	TriggerRule     string  `json:"trigger_rule" binding:"required"`
	IsActive        bool    `json:"is_active"`
}

// UpdateFishTideRequest 更新魚潮請求
type UpdateFishTideRequest struct {
	Name            string  `json:"name"`
	FishTypeID      int32   `json:"fish_type_id"`
	FishCount       int     `json:"fish_count" binding:"omitempty,min=1"`
	DurationSeconds int     `json:"duration_seconds" binding:"omitempty,min=1"`
	IntervalMs      int     `json:"interval_ms" binding:"omitempty,min=1"`
	SpeedMultiplier float64 `json:"speed_multiplier" binding:"omitempty,min=0.1"`
	TriggerRule     string  `json:"trigger_rule"`
	IsActive        *bool   `json:"is_active"` // 使用指針以區分未設置和false
}

// TriggerFishTideRequest 觸發魚潮請求
type TriggerFishTideRequest struct {
	RoomID string `json:"room_id" binding:"required"`
}

// handleGetFishTides 獲取所有魚潮配置
func (h *FishTideHandler) handleGetFishTides(c *gin.Context) {
	tides, err := h.repo.GetActiveTides(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tides": tides,
		"count": len(tides),
	})
}

// handleCreateFishTide 建立新的魚潮配置
func (h *FishTideHandler) handleCreateFishTide(c *gin.Context) {
	var req CreateFishTideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 建立魚潮實體
	tide := &game.FishTide{
		Name:            req.Name,
		FishTypeID:      req.FishTypeID,
		FishCount:       req.FishCount,
		Duration:        time.Duration(req.DurationSeconds) * time.Second,
		SpawnInterval:   time.Duration(req.IntervalMs) * time.Millisecond,
		SpeedMultiplier: req.SpeedMultiplier,
		TriggerRule:     req.TriggerRule,
		IsActive:        req.IsActive,
	}

	if err := h.repo.CreateTide(c.Request.Context(), tide); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "fish tide created successfully",
		"tide":    tide,
	})
}

// handleUpdateFishTide 更新魚潮配置
func (h *FishTideHandler) handleUpdateFishTide(c *gin.Context) {
	// 解析 tide_id
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tide id"})
		return
	}

	var req UpdateFishTideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 獲取現有魚潮配置
	tide, err := h.repo.GetTideByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 更新字段（僅更新提供的字段）
	if req.Name != "" {
		tide.Name = req.Name
	}
	if req.FishTypeID != 0 {
		tide.FishTypeID = req.FishTypeID
	}
	if req.FishCount != 0 {
		tide.FishCount = req.FishCount
	}
	if req.DurationSeconds != 0 {
		tide.Duration = time.Duration(req.DurationSeconds) * time.Second
	}
	if req.IntervalMs != 0 {
		tide.SpawnInterval = time.Duration(req.IntervalMs) * time.Millisecond
	}
	if req.SpeedMultiplier != 0 {
		tide.SpeedMultiplier = req.SpeedMultiplier
	}
	if req.TriggerRule != "" {
		tide.TriggerRule = req.TriggerRule
	}
	if req.IsActive != nil {
		tide.IsActive = *req.IsActive
	}

	if err := h.repo.UpdateTide(c.Request.Context(), tide); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "fish tide updated successfully",
		"tide":    tide,
	})
}

// handleDeleteFishTide 刪除魚潮配置
func (h *FishTideHandler) handleDeleteFishTide(c *gin.Context) {
	// 解析 tide_id
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tide id"})
		return
	}

	if err := h.repo.DeleteTide(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "fish tide deleted successfully",
	})
}

// handleStartFishTide 手動觸發魚潮
func (h *FishTideHandler) handleStartFishTide(c *gin.Context) {
	// 解析 tide_id
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tide id"})
		return
	}

	var req TriggerFishTideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.manager.StartTide(c.Request.Context(), req.RoomID, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "fish tide started successfully",
		"room_id": req.RoomID,
		"tide_id": id,
	})
}

// handleStopFishTide 手動停止魚潮
func (h *FishTideHandler) handleStopFishTide(c *gin.Context) {
	var req TriggerFishTideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.manager.StopTide(c.Request.Context(), req.RoomID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "fish tide stopped successfully",
		"room_id": req.RoomID,
	})
}
