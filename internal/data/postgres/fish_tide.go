package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/jackc/pgx/v5"
)

// fishTideRepo 實現 game.FishTideRepo 介面
type fishTideRepo struct {
	db *Client
}

// NewFishTideRepo 建立新的 FishTideRepo 實例
func NewFishTideRepo(db *Client) game.FishTideRepo {
	return &fishTideRepo{
		db: db,
	}
}

// GetTideByID 根據 ID 獲取魚潮配置
func (r *fishTideRepo) GetTideByID(ctx context.Context, id int64) (*game.FishTide, error) {
	query := `
		SELECT id, name, fish_type_id, fish_count, duration_seconds,
		       spawn_interval_ms, speed_multiplier, trigger_rule, is_active
		FROM fish_tide_config
		WHERE id = $1
	`

	var tide game.FishTide
	var durationSeconds int
	var spawnIntervalMs int

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&tide.ID,
		&tide.Name,
		&tide.FishTypeID,
		&tide.FishCount,
		&durationSeconds,
		&spawnIntervalMs,
		&tide.SpeedMultiplier,
		&tide.TriggerRule,
		&tide.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("fish tide with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get fish tide by id: %w", err)
	}

	// 轉換時間單位
	tide.Duration = time.Duration(durationSeconds) * time.Second
	tide.SpawnInterval = time.Duration(spawnIntervalMs) * time.Millisecond

	return &tide, nil
}

// GetActiveTides 獲取所有啟用的魚潮配置
func (r *fishTideRepo) GetActiveTides(ctx context.Context) ([]*game.FishTide, error) {
	query := `
		SELECT id, name, fish_type_id, fish_count, duration_seconds,
		       spawn_interval_ms, speed_multiplier, trigger_rule, is_active
		FROM fish_tide_config
		WHERE is_active = TRUE
		ORDER BY id ASC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active tides: %w", err)
	}
	defer rows.Close()

	var tides []*game.FishTide
	for rows.Next() {
		var tide game.FishTide
		var durationSeconds int
		var spawnIntervalMs int

		err := rows.Scan(
			&tide.ID,
			&tide.Name,
			&tide.FishTypeID,
			&tide.FishCount,
			&durationSeconds,
			&spawnIntervalMs,
			&tide.SpeedMultiplier,
			&tide.TriggerRule,
			&tide.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan fish tide: %w", err)
		}

		// 轉換時間單位
		tide.Duration = time.Duration(durationSeconds) * time.Second
		tide.SpawnInterval = time.Duration(spawnIntervalMs) * time.Millisecond

		tides = append(tides, &tide)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating fish tides: %w", err)
	}

	return tides, nil
}

// CreateTide 建立新的魚潮配置
func (r *fishTideRepo) CreateTide(ctx context.Context, tide *game.FishTide) error {
	query := `
		INSERT INTO fish_tide_config
		(name, fish_type_id, fish_count, duration_seconds, spawn_interval_ms,
		 speed_multiplier, trigger_rule, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	// 轉換時間單位為整數
	durationSeconds := int(tide.Duration.Seconds())
	spawnIntervalMs := int(tide.SpawnInterval.Milliseconds())

	err := r.db.Pool.QueryRow(ctx, query,
		tide.Name,
		tide.FishTypeID,
		tide.FishCount,
		durationSeconds,
		spawnIntervalMs,
		tide.SpeedMultiplier,
		tide.TriggerRule,
		tide.IsActive,
	).Scan(&tide.ID)

	if err != nil {
		return fmt.Errorf("failed to create fish tide: %w", err)
	}

	return nil
}

// UpdateTide 更新魚潮配置
func (r *fishTideRepo) UpdateTide(ctx context.Context, tide *game.FishTide) error {
	query := `
		UPDATE fish_tide_config
		SET name = $2,
		    fish_type_id = $3,
		    fish_count = $4,
		    duration_seconds = $5,
		    spawn_interval_ms = $6,
		    speed_multiplier = $7,
		    trigger_rule = $8,
		    is_active = $9,
		    updated_at = NOW()
		WHERE id = $1
	`

	// 轉換時間單位為整數
	durationSeconds := int(tide.Duration.Seconds())
	spawnIntervalMs := int(tide.SpawnInterval.Milliseconds())

	result, err := r.db.Pool.Exec(ctx, query,
		tide.ID,
		tide.Name,
		tide.FishTypeID,
		tide.FishCount,
		durationSeconds,
		spawnIntervalMs,
		tide.SpeedMultiplier,
		tide.TriggerRule,
		tide.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update fish tide: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("fish tide with id %d not found", tide.ID)
	}

	return nil
}

// DeleteTide 刪除魚潮配置
func (r *fishTideRepo) DeleteTide(ctx context.Context, id int64) error {
	query := `DELETE FROM fish_tide_config WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete fish tide: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("fish tide with id %d not found", id)
	}

	return nil
}
