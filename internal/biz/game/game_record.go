package game

import (
	"context"
	"time"
)

// ========================================
// GameRecord - 遊戲記錄領域模型
// ========================================

// GameRecordStatus 遊戲記錄狀態
type GameRecordStatus string

const (
	GameRecordStatusPlaying   GameRecordStatus = "playing"   // 遊戲進行中
	GameRecordStatusFinished  GameRecordStatus = "finished"  // 正常結束
	GameRecordStatusAbandoned GameRecordStatus = "abandoned" // 中途放棄
)

// GameRecord 遊戲記錄領域模型
type GameRecord struct {
	ID        int64
	UserID    int64
	RoomID    string
	SessionID string // 遊戲會話ID

	// 時間相關
	StartTime       time.Time
	EndTime         *time.Time // 可為空，表示遊戲進行中
	DurationSeconds int        // 遊戲時長（秒）

	// 財務統計
	TotalBets  float64 // 總投注（所有子彈費用）
	TotalWins  float64 // 總獎勵（所有捕獲獎勵）
	NetProfit  float64 // 淨盈虧（total_wins - total_bets）

	// 遊戲統計
	BulletsFired int64   // 發射子彈數量
	BulletsHit   int64   // 命中子彈數量
	FishCaught   int64   // 捕獲魚數量
	HitRate      float64 // 命中率（百分比）

	// 獎勵統計
	MaxSingleWin float64 // 最大單次獎勵
	BonusCount   int     // 獎金次數（暴擊、特殊魚等）

	// 狀態
	Status GameRecordStatus

	// 額外數據
	Metadata map[string]interface{} // 例如：魚類型分佈、使用的砲台等級等

	// 時間戳
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewGameRecord 創建新的遊戲記錄
func NewGameRecord(userID int64, roomID string, sessionID string) *GameRecord {
	now := time.Now()
	return &GameRecord{
		UserID:       userID,
		RoomID:       roomID,
		SessionID:    sessionID,
		StartTime:    now,
		Status:       GameRecordStatusPlaying,
		TotalBets:    0,
		TotalWins:    0,
		NetProfit:    0,
		BulletsFired: 0,
		BulletsHit:   0,
		FishCaught:   0,
		HitRate:      0,
		MaxSingleWin: 0,
		BonusCount:   0,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// RecordBulletFired 記錄子彈發射
func (gr *GameRecord) RecordBulletFired(cost float64) {
	gr.BulletsFired++
	gr.TotalBets += cost
	gr.NetProfit = gr.TotalWins - gr.TotalBets
	gr.UpdatedAt = time.Now()
}

// RecordFishCaught 記錄捕獲魚
func (gr *GameRecord) RecordFishCaught(reward float64, isCritical bool) {
	gr.BulletsHit++
	gr.FishCaught++
	gr.TotalWins += reward
	gr.NetProfit = gr.TotalWins - gr.TotalBets

	// 更新最大單次獎勵
	if reward > gr.MaxSingleWin {
		gr.MaxSingleWin = reward
	}

	// 如果是暴擊或特殊獎勵，增加獎金次數
	if isCritical {
		gr.BonusCount++
	}

	// 更新命中率
	if gr.BulletsFired > 0 {
		gr.HitRate = float64(gr.BulletsHit) / float64(gr.BulletsFired) * 100
	}

	gr.UpdatedAt = time.Now()
}

// Finish 結束遊戲記錄
func (gr *GameRecord) Finish() {
	now := time.Now()
	gr.EndTime = &now
	gr.Status = GameRecordStatusFinished
	gr.DurationSeconds = int(now.Sub(gr.StartTime).Seconds())
	gr.UpdatedAt = now
}

// Abandon 放棄遊戲
func (gr *GameRecord) Abandon() {
	now := time.Now()
	gr.EndTime = &now
	gr.Status = GameRecordStatusAbandoned
	gr.DurationSeconds = int(now.Sub(gr.StartTime).Seconds())
	gr.UpdatedAt = now
}

// IsActive 是否還在進行中
func (gr *GameRecord) IsActive() bool {
	return gr.Status == GameRecordStatusPlaying
}

// ========================================
// GameRecordRepo - 遊戲記錄倉庫接口
// ========================================

// GameRecordRepo 定義了遊戲記錄數據倉庫的接口
type GameRecordRepo interface {
	// 創建遊戲記錄
	Create(ctx context.Context, record *GameRecord) error

	// 更新遊戲記錄
	Update(ctx context.Context, record *GameRecord) error

	// 查詢遊戲記錄
	FindByID(ctx context.Context, id int64) (*GameRecord, error)
	FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*GameRecord, error)
	FindBySessionID(ctx context.Context, sessionID string) ([]*GameRecord, error)
	FindActiveByUserID(ctx context.Context, userID int64) (*GameRecord, error) // 查找進行中的遊戲

	// 統計查詢
	GetUserTotalStats(ctx context.Context, userID int64) (*UserGameStats, error)
}

// UserGameStats 用戶遊戲統計
type UserGameStats struct {
	TotalGames       int64   // 總遊戲局數
	TotalBets        float64 // 總投注
	TotalWins        float64 // 總獎勵
	NetProfit        float64 // 總淨盈虧
	AvgGameDuration  int     // 平均遊戲時長（秒）
	TotalBulletsFired int64  // 總發射子彈數
	TotalFishCaught  int64   // 總捕獲魚數
	AvgHitRate       float64 // 平均命中率
	MaxSingleWin     float64 // 最大單次獎勵
}
