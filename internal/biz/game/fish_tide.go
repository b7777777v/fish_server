package game

import (
	"context"
	"errors"
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
	// TODO: 注入必要的依賴，如：
	// - FishTideRepo: 魚潮配置資料庫操作介面
	// - FishSpawner: 魚群生成器（用於生成魚潮中的魚）
	// - EventBroadcaster: 事件廣播器（通知客戶端魚潮開始/結束）
}

// NewFishTideManager 建立新的 FishTideManager 實例
func NewFishTideManager( /* TODO: 添加參數 */ ) FishTideManager {
	return &fishTideManager{
		// TODO: 初始化依賴
	}
}

// StartTide 開始一次魚潮事件
func (m *fishTideManager) StartTide(ctx context.Context, roomID string, tideID int64) error {
	// TODO: 實現魚潮開始邏輯
	// 1. 從資料庫獲取魚潮配置
	// 2. 驗證魚潮是否可以啟動（是否已有活躍魚潮）
	// 3. 廣播魚潮開始事件給房間內所有玩家
	// 4. 啟動魚潮生成邏輯（在指定的持續時間內，以指定的間隔生成魚）
	// 5. 設定定時器，在持續時間結束後自動停止魚潮
	return ErrFishTideNotImplemented
}

// StopTide 停止當前的魚潮事件
func (m *fishTideManager) StopTide(ctx context.Context, roomID string) error {
	// TODO: 實現魚潮停止邏輯
	// 1. 停止魚潮生成
	// 2. 廣播魚潮結束事件給房間內所有玩家
	// 3. 清理魚潮狀態
	return ErrFishTideNotImplemented
}

// GetActiveTide 獲取當前房間的活躍魚潮
func (m *fishTideManager) GetActiveTide(ctx context.Context, roomID string) (*FishTide, error) {
	// TODO: 實現獲取活躍魚潮邏輯
	// 返回當前房間正在進行的魚潮，如果沒有則返回 nil
	return nil, ErrFishTideNotImplemented
}

// ScheduleTides 排程魚潮（根據配置自動觸發）
func (m *fishTideManager) ScheduleTides(ctx context.Context, roomID string) error {
	// TODO: 實現魚潮排程邏輯
	// 1. 從資料庫獲取所有啟用的魚潮配置
	// 2. 根據觸發規則（如固定時間、隨機間隔）設定定時器
	// 3. 當觸發條件滿足時，自動呼叫 StartTide
	return ErrFishTideNotImplemented
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
