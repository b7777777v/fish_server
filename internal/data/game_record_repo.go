package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/jackc/pgx/v5"
)

// ========================================
// GameRecordPO - 遊戲記錄持久化對象
// ========================================

// GameRecordPO 遊戲記錄的持久化對象
type GameRecordPO struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	RoomID    string     `json:"room_id"`
	SessionID string     `json:"session_id"`

	// 時間相關
	StartTime       time.Time  `json:"start_time"`
	EndTime         *time.Time `json:"end_time"`
	DurationSeconds int        `json:"duration_seconds"`

	// 財務統計
	TotalBets float64 `json:"total_bets"`
	TotalWins float64 `json:"total_wins"`
	NetProfit float64 `json:"net_profit"`

	// 遊戲統計
	BulletsFired int64   `json:"bullets_fired"`
	BulletsHit   int64   `json:"bullets_hit"`
	FishCaught   int64   `json:"fish_caught"`
	HitRate      float64 `json:"hit_rate"`

	// 獎勵統計
	MaxSingleWin float64 `json:"max_single_win"`
	BonusCount   int     `json:"bonus_count"`

	// 狀態
	Status string `json:"status"`

	// 額外數據
	Metadata string `json:"metadata"` // JSONB 字段

	// 時間戳
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ========================================
// gameRecordRepo - 遊戲記錄倉庫實現
// ========================================

type gameRecordRepo struct {
	data   *Data
	logger logger.Logger
}

// NewGameRecordRepo 創建遊戲記錄倉庫
func NewGameRecordRepo(data *Data, logger logger.Logger) game.GameRecordRepo {
	return &gameRecordRepo{
		data:   data,
		logger: logger.With("module", "data/game_record_repo"),
	}
}

// po2do 將持久化對象轉換為領域對象
func (r *gameRecordRepo) po2do(po *GameRecordPO) (*game.GameRecord, error) {
	var metadata map[string]interface{}
	if po.Metadata != "" {
		if err := json.Unmarshal([]byte(po.Metadata), &metadata); err != nil {
			r.logger.Warnf("Failed to unmarshal metadata: %v", err)
			metadata = make(map[string]interface{})
		}
	} else {
		metadata = make(map[string]interface{})
	}

	return &game.GameRecord{
		ID:              po.ID,
		UserID:          po.UserID,
		RoomID:          po.RoomID,
		SessionID:       po.SessionID,
		StartTime:       po.StartTime,
		EndTime:         po.EndTime,
		DurationSeconds: po.DurationSeconds,
		TotalBets:       po.TotalBets,
		TotalWins:       po.TotalWins,
		NetProfit:       po.NetProfit,
		BulletsFired:    po.BulletsFired,
		BulletsHit:      po.BulletsHit,
		FishCaught:      po.FishCaught,
		HitRate:         po.HitRate,
		MaxSingleWin:    po.MaxSingleWin,
		BonusCount:      po.BonusCount,
		Status:          game.GameRecordStatus(po.Status),
		Metadata:        metadata,
		CreatedAt:       po.CreatedAt,
		UpdatedAt:       po.UpdatedAt,
	}, nil
}

// do2po 將領域對象轉換為持久化對象
func (r *gameRecordRepo) do2po(do *game.GameRecord) (*GameRecordPO, error) {
	metadataBytes, err := json.Marshal(do.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return &GameRecordPO{
		ID:              do.ID,
		UserID:          do.UserID,
		RoomID:          do.RoomID,
		SessionID:       do.SessionID,
		StartTime:       do.StartTime,
		EndTime:         do.EndTime,
		DurationSeconds: do.DurationSeconds,
		TotalBets:       do.TotalBets,
		TotalWins:       do.TotalWins,
		NetProfit:       do.NetProfit,
		BulletsFired:    do.BulletsFired,
		BulletsHit:      do.BulletsHit,
		FishCaught:      do.FishCaught,
		HitRate:         do.HitRate,
		MaxSingleWin:    do.MaxSingleWin,
		BonusCount:      do.BonusCount,
		Status:          string(do.Status),
		Metadata:        string(metadataBytes),
		CreatedAt:       do.CreatedAt,
		UpdatedAt:       do.UpdatedAt,
	}, nil
}

// Create 創建遊戲記錄
func (r *gameRecordRepo) Create(ctx context.Context, record *game.GameRecord) error {
	po, err := r.do2po(record)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO game_records (
			user_id, room_id, session_id,
			start_time, end_time, duration_seconds,
			total_bets, total_wins, net_profit,
			bullets_fired, bullets_hit, fish_caught, hit_rate,
			max_single_win, bonus_count,
			status, metadata,
			created_at, updated_at
		) VALUES (
			$1, $2, $3,
			$4, $5, $6,
			$7, $8, $9,
			$10, $11, $12, $13,
			$14, $15,
			$16, $17,
			$18, $19
		) RETURNING id
	`

	err = r.data.DBManager().Write().QueryRow(
		ctx, query,
		po.UserID, po.RoomID, po.SessionID,
		po.StartTime, po.EndTime, po.DurationSeconds,
		po.TotalBets, po.TotalWins, po.NetProfit,
		po.BulletsFired, po.BulletsHit, po.FishCaught, po.HitRate,
		po.MaxSingleWin, po.BonusCount,
		po.Status, po.Metadata,
		po.CreatedAt, po.UpdatedAt,
	).Scan(&record.ID)

	if err != nil {
		r.logger.Errorf("Failed to create game record: %v", err)
		return err
	}

	r.logger.Infof("Created game record: id=%d, user_id=%d, room_id=%s", record.ID, record.UserID, record.RoomID)
	return nil
}

// Update 更新遊戲記錄
func (r *gameRecordRepo) Update(ctx context.Context, record *game.GameRecord) error {
	po, err := r.do2po(record)
	if err != nil {
		return err
	}

	query := `
		UPDATE game_records SET
			end_time = $1,
			duration_seconds = $2,
			total_bets = $3,
			total_wins = $4,
			net_profit = $5,
			bullets_fired = $6,
			bullets_hit = $7,
			fish_caught = $8,
			hit_rate = $9,
			max_single_win = $10,
			bonus_count = $11,
			status = $12,
			metadata = $13,
			updated_at = $14
		WHERE id = $15
	`

	result, err := r.data.DBManager().Write().Exec(
		ctx, query,
		po.EndTime, po.DurationSeconds,
		po.TotalBets, po.TotalWins, po.NetProfit,
		po.BulletsFired, po.BulletsHit, po.FishCaught, po.HitRate,
		po.MaxSingleWin, po.BonusCount,
		po.Status, po.Metadata,
		time.Now(), po.ID,
	)

	if err != nil {
		r.logger.Errorf("Failed to update game record: %v", err)
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("game record not found: id=%d", record.ID)
	}

	r.logger.Debugf("Updated game record: id=%d", record.ID)
	return nil
}

// FindByID 根據ID查詢遊戲記錄
func (r *gameRecordRepo) FindByID(ctx context.Context, id int64) (*game.GameRecord, error) {
	query := `
		SELECT
			id, user_id, room_id, session_id,
			start_time, end_time, duration_seconds,
			total_bets, total_wins, net_profit,
			bullets_fired, bullets_hit, fish_caught, hit_rate,
			max_single_win, bonus_count,
			status, metadata,
			created_at, updated_at
		FROM game_records
		WHERE id = $1
	`

	var po GameRecordPO
	err := r.data.DBManager().Read().QueryRow(ctx, query, id).Scan(
		&po.ID, &po.UserID, &po.RoomID, &po.SessionID,
		&po.StartTime, &po.EndTime, &po.DurationSeconds,
		&po.TotalBets, &po.TotalWins, &po.NetProfit,
		&po.BulletsFired, &po.BulletsHit, &po.FishCaught, &po.HitRate,
		&po.MaxSingleWin, &po.BonusCount,
		&po.Status, &po.Metadata,
		&po.CreatedAt, &po.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("game record not found")
		}
		r.logger.Errorf("Failed to find game record by id: %v", err)
		return nil, err
	}

	return r.po2do(&po)
}

// FindByUserID 根據用戶ID查詢遊戲記錄（分頁）
func (r *gameRecordRepo) FindByUserID(ctx context.Context, userID int64, limit, offset int) ([]*game.GameRecord, error) {
	query := `
		SELECT
			id, user_id, room_id, session_id,
			start_time, end_time, duration_seconds,
			total_bets, total_wins, net_profit,
			bullets_fired, bullets_hit, fish_caught, hit_rate,
			max_single_win, bonus_count,
			status, metadata,
			created_at, updated_at
		FROM game_records
		WHERE user_id = $1
		ORDER BY start_time DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.data.DBManager().Read().Query(ctx, query, userID, limit, offset)
	if err != nil {
		r.logger.Errorf("Failed to query game records by user_id: %v", err)
		return nil, err
	}
	defer rows.Close()

	var records []*game.GameRecord
	for rows.Next() {
		var po GameRecordPO
		err := rows.Scan(
			&po.ID, &po.UserID, &po.RoomID, &po.SessionID,
			&po.StartTime, &po.EndTime, &po.DurationSeconds,
			&po.TotalBets, &po.TotalWins, &po.NetProfit,
			&po.BulletsFired, &po.BulletsHit, &po.FishCaught, &po.HitRate,
			&po.MaxSingleWin, &po.BonusCount,
			&po.Status, &po.Metadata,
			&po.CreatedAt, &po.UpdatedAt,
		)
		if err != nil {
			r.logger.Errorf("Failed to scan game record row: %v", err)
			return nil, err
		}

		record, err := r.po2do(&po)
		if err != nil {
			r.logger.Errorf("Failed to convert PO to DO: %v", err)
			continue
		}
		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("Error iterating game record rows: %v", err)
		return nil, err
	}

	return records, nil
}

// FindBySessionID 根據會話ID查詢遊戲記錄
func (r *gameRecordRepo) FindBySessionID(ctx context.Context, sessionID string) ([]*game.GameRecord, error) {
	query := `
		SELECT
			id, user_id, room_id, session_id,
			start_time, end_time, duration_seconds,
			total_bets, total_wins, net_profit,
			bullets_fired, bullets_hit, fish_caught, hit_rate,
			max_single_win, bonus_count,
			status, metadata,
			created_at, updated_at
		FROM game_records
		WHERE session_id = $1
		ORDER BY start_time DESC
	`

	rows, err := r.data.DBManager().Read().Query(ctx, query, sessionID)
	if err != nil {
		r.logger.Errorf("Failed to query game records by session_id: %v", err)
		return nil, err
	}
	defer rows.Close()

	var records []*game.GameRecord
	for rows.Next() {
		var po GameRecordPO
		err := rows.Scan(
			&po.ID, &po.UserID, &po.RoomID, &po.SessionID,
			&po.StartTime, &po.EndTime, &po.DurationSeconds,
			&po.TotalBets, &po.TotalWins, &po.NetProfit,
			&po.BulletsFired, &po.BulletsHit, &po.FishCaught, &po.HitRate,
			&po.MaxSingleWin, &po.BonusCount,
			&po.Status, &po.Metadata,
			&po.CreatedAt, &po.UpdatedAt,
		)
		if err != nil {
			r.logger.Errorf("Failed to scan game record row: %v", err)
			return nil, err
		}

		record, err := r.po2do(&po)
		if err != nil {
			r.logger.Errorf("Failed to convert PO to DO: %v", err)
			continue
		}
		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("Error iterating game record rows: %v", err)
		return nil, err
	}

	return records, nil
}

// FindActiveByUserID 查找用戶進行中的遊戲記錄
func (r *gameRecordRepo) FindActiveByUserID(ctx context.Context, userID int64) (*game.GameRecord, error) {
	query := `
		SELECT
			id, user_id, room_id, session_id,
			start_time, end_time, duration_seconds,
			total_bets, total_wins, net_profit,
			bullets_fired, bullets_hit, fish_caught, hit_rate,
			max_single_win, bonus_count,
			status, metadata,
			created_at, updated_at
		FROM game_records
		WHERE user_id = $1 AND status = 'playing'
		ORDER BY start_time DESC
		LIMIT 1
	`

	var po GameRecordPO
	err := r.data.DBManager().Read().QueryRow(ctx, query, userID).Scan(
		&po.ID, &po.UserID, &po.RoomID, &po.SessionID,
		&po.StartTime, &po.EndTime, &po.DurationSeconds,
		&po.TotalBets, &po.TotalWins, &po.NetProfit,
		&po.BulletsFired, &po.BulletsHit, &po.FishCaught, &po.HitRate,
		&po.MaxSingleWin, &po.BonusCount,
		&po.Status, &po.Metadata,
		&po.CreatedAt, &po.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // 沒有進行中的遊戲，返回 nil 而不是錯誤
		}
		r.logger.Errorf("Failed to find active game record: %v", err)
		return nil, err
	}

	return r.po2do(&po)
}

// GetUserTotalStats 獲取用戶遊戲總統計
func (r *gameRecordRepo) GetUserTotalStats(ctx context.Context, userID int64) (*game.UserGameStats, error) {
	query := `
		SELECT
			COUNT(*) as total_games,
			COALESCE(SUM(total_bets), 0) as total_bets,
			COALESCE(SUM(total_wins), 0) as total_wins,
			COALESCE(SUM(net_profit), 0) as net_profit,
			COALESCE(AVG(duration_seconds), 0) as avg_game_duration,
			COALESCE(SUM(bullets_fired), 0) as total_bullets_fired,
			COALESCE(SUM(fish_caught), 0) as total_fish_caught,
			COALESCE(AVG(hit_rate), 0) as avg_hit_rate,
			COALESCE(MAX(max_single_win), 0) as max_single_win
		FROM game_records
		WHERE user_id = $1 AND status IN ('finished', 'abandoned')
	`

	var stats game.UserGameStats
	err := r.data.DBManager().Read().QueryRow(ctx, query, userID).Scan(
		&stats.TotalGames,
		&stats.TotalBets,
		&stats.TotalWins,
		&stats.NetProfit,
		&stats.AvgGameDuration,
		&stats.TotalBulletsFired,
		&stats.TotalFishCaught,
		&stats.AvgHitRate,
		&stats.MaxSingleWin,
	)

	if err != nil {
		r.logger.Errorf("Failed to get user total stats: %v", err)
		return nil, err
	}

	return &stats, nil
}
