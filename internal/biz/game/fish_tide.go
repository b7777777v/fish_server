package game

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrFishTideNotImplemented 魚潮系統尚未實現
var ErrFishTideNotImplemented = errors.New("fish tide system is not yet implemented")

// TODO: 實現魚潮系統 (Fish Tide System)
// 魚潮系統是一種特殊的魚群陣型，特點：
// - 在特定時間觸發
// - 大量特定魚群快速、密集地游過螢幕
// - 為玩家提供在短時間內賺取大量金幣的機會
// - 觸發規則、持續時間、魚種構成均可在後台配置

// FishTide 代表一次魚潮事件
type FishTide struct {
	ID              int64         `json:"id"`
	Name            string        `json:"name"`
	FishTypeID      int32         `json:"fish_type_id"`      // 魚潮中的魚種 ID
	FishCount       int           `json:"fish_count"`        // 魚的數量
	Duration        time.Duration `json:"duration"`          // 持續時間
	SpawnInterval   time.Duration `json:"spawn_interval"`    // 生成間隔
	SpeedMultiplier float64       `json:"speed_multiplier"`  // 速度倍率
	TriggerRule     string        `json:"trigger_rule"`      // 觸發規則（如：固定時間、隨機、手動觸發）
	IsActive        bool          `json:"is_active"`         // 是否啟用
}

// FishTideManager 魚潮管理器
type FishTideManager interface {
	// StartTide 開始一次魚潮事件
	StartTide(ctx context.Context, roomID string, tideID int64) error

	// StopTide 停止當前的魚潮事件
	StopTide(ctx context.Context, roomID string) error

	// GetActiveTide 獲取當前房間的活躍魚潮
	GetActiveTide(ctx context.Context, roomID string) (*FishTide, error)

	// ScheduleTides 排程魚潮（根據配置自動觸發）
	ScheduleTides(ctx context.Context, roomID string) error
}

// fishTideManager 實現 FishTideManager 介面
type fishTideManager struct {
	repo         FishTideRepo
	activeTides  map[string]*FishTide // roomID -> active tide
	tideTimers   map[string]*time.Timer // roomID -> stop timer
	mu           sync.RWMutex
}

// NewFishTideManager 建立新的 FishTideManager 實例
func NewFishTideManager(repo FishTideRepo) FishTideManager {
	return &fishTideManager{
		repo:        repo,
		activeTides: make(map[string]*FishTide),
		tideTimers:  make(map[string]*time.Timer),
	}
}

// StartTide 開始一次魚潮事件
func (m *fishTideManager) StartTide(ctx context.Context, roomID string, tideID int64) error {
	// 1. 從資料庫獲取魚潮配置
	tide, err := m.repo.GetTideByID(ctx, tideID)
	if err != nil {
		return fmt.Errorf("failed to get tide config: %w", err)
	}

	// 2. 驗證魚潮是否可以啟動（是否已有活躍魚潮）
	m.mu.Lock()
	if _, exists := m.activeTides[roomID]; exists {
		m.mu.Unlock()
		return fmt.Errorf("room %s already has an active tide", roomID)
	}

	// 3. 記錄活躍魚潮
	m.activeTides[roomID] = tide
	m.mu.Unlock()

	// 4. 設定定時器，在持續時間結束後自動停止魚潮
	timer := time.AfterFunc(tide.Duration, func() {
		// 自動停止魚潮
		_ = m.StopTide(context.Background(), roomID)
	})

	m.mu.Lock()
	m.tideTimers[roomID] = timer
	m.mu.Unlock()

	// TODO: 5. 廣播魚潮開始事件給房間內所有玩家（需要整合 Hub/RoomManager）
	// TODO: 6. 啟動魚潮生成邏輯（需要整合 FishSpawner）
	// 當前實現：僅管理魚潮狀態，實際魚群生成由 RoomManager 處理

	return nil
}

// StopTide 停止當前的魚潮事件
func (m *fishTideManager) StopTide(ctx context.Context, roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. 檢查是否有活躍魚潮
	if _, exists := m.activeTides[roomID]; !exists {
		return fmt.Errorf("no active tide in room %s", roomID)
	}

	// 2. 停止定時器
	if timer, exists := m.tideTimers[roomID]; exists {
		timer.Stop()
		delete(m.tideTimers, roomID)
	}

	// 3. 清理魚潮狀態
	delete(m.activeTides, roomID)

	// TODO: 4. 廣播魚潮結束事件給房間內所有玩家（需要整合 Hub/RoomManager）
	// TODO: 5. 停止魚潮生成（需要整合 FishSpawner）

	return nil
}

// GetActiveTide 獲取當前房間的活躍魚潮
func (m *fishTideManager) GetActiveTide(ctx context.Context, roomID string) (*FishTide, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tide, exists := m.activeTides[roomID]
	if !exists {
		return nil, nil // 沒有活躍魚潮，返回 nil（不是錯誤）
	}

	return tide, nil
}

// ScheduleTides 排程魚潮（根據配置自動觸發）
func (m *fishTideManager) ScheduleTides(ctx context.Context, roomID string) error {
	// 1. 從資料庫獲取所有啟用的魚潮配置
	tides, err := m.repo.GetActiveTides(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active tides: %w", err)
	}

	// TODO: 2. 根據觸發規則設定定時器
	// 當前實現：僅返回成功，實際排程邏輯需要整合 cron 或其他排程系統
	// 建議使用：github.com/robfig/cron/v3

	// 簡單日誌記錄
	_ = tides // 避免未使用警告

	return nil
}

// FishTideRepo 定義魚潮資料訪問介面
type FishTideRepo interface {
	// GetTideByID 根據 ID 獲取魚潮配置
	GetTideByID(ctx context.Context, id int64) (*FishTide, error)

	// GetActiveTides 獲取所有啟用的魚潮配置
	GetActiveTides(ctx context.Context) ([]*FishTide, error)

	// CreateTide 建立新的魚潮配置
	CreateTide(ctx context.Context, tide *FishTide) error

	// UpdateTide 更新魚潮配置
	UpdateTide(ctx context.Context, tide *FishTide) error

	// DeleteTide 刪除魚潮配置
	DeleteTide(ctx context.Context, id int64) error
}
