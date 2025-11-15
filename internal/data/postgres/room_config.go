package postgres

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/game"
)

// RoomConfigRepo 实现房间配置的数据访问
type RoomConfigRepo struct {
	dbManager *DBManager
}

// NewRoomConfigRepo 创建新的 RoomConfigRepo 实例
func NewRoomConfigRepo(dbManager *DBManager) *RoomConfigRepo {
	return &RoomConfigRepo{
		dbManager: dbManager,
	}
}

// RoomConfigPO 房间配置持久化对象
type RoomConfigPO struct {
	ID                   int64
	RoomType             string
	RoomName             string
	MaxPlayers           int
	MinBet               int64
	MaxBet               int64
	EntryFee             int64
	BulletCostMultiplier float64
	FishSpawnRate        float64
	MinFishCount         int
	MaxFishCount         int
	RoomWidth            float64
	RoomHeight           float64
	TargetRTP            float64
	IsActive             bool
	Description          string
}

// GetRoomConfig 根据房间类型获取配置
func (r *RoomConfigRepo) GetRoomConfig(ctx context.Context, roomType string) (*game.RoomConfig, error) {
	query := `
		SELECT id, room_type, room_name, max_players, min_bet, max_bet, entry_fee,
		       bullet_cost_multiplier, fish_spawn_rate, min_fish_count, max_fish_count,
		       room_width, room_height, target_rtp, is_active, description
		FROM room_configs
		WHERE room_type = $1 AND is_active = true
	`

	var po RoomConfigPO
	// 讀操作使用 Read DB
	err := r.dbManager.Read().QueryRow(ctx, query, roomType).Scan(
		&po.ID,
		&po.RoomType,
		&po.RoomName,
		&po.MaxPlayers,
		&po.MinBet,
		&po.MaxBet,
		&po.EntryFee,
		&po.BulletCostMultiplier,
		&po.FishSpawnRate,
		&po.MinFishCount,
		&po.MaxFishCount,
		&po.RoomWidth,
		&po.RoomHeight,
		&po.TargetRTP,
		&po.IsActive,
		&po.Description,
	)

	if err != nil {
		return nil, err
	}

	// 转换为业务实体
	config := &game.RoomConfig{
		MaxPlayers:           int32(po.MaxPlayers),
		MinBet:               po.MinBet,
		MaxBet:               po.MaxBet,
		BulletCostMultiplier: po.BulletCostMultiplier,
		FishSpawnRate:        po.FishSpawnRate,
		MinFishCount:         int32(po.MinFishCount),
		MaxFishCount:         int32(po.MaxFishCount),
		RoomWidth:            po.RoomWidth,
		RoomHeight:           po.RoomHeight,
		TargetRTP:            po.TargetRTP,
	}

	return config, nil
}

// GetAllRoomConfigs 获取所有活跃的房间配置
func (r *RoomConfigRepo) GetAllRoomConfigs(ctx context.Context) (map[string]*game.RoomConfig, error) {
	query := `
		SELECT id, room_type, room_name, max_players, min_bet, max_bet, entry_fee,
		       bullet_cost_multiplier, fish_spawn_rate, min_fish_count, max_fish_count,
		       room_width, room_height, target_rtp, is_active, description
		FROM room_configs
		WHERE is_active = true
		ORDER BY room_type
	`

	// 讀操作使用 Read DB
	rows, err := r.dbManager.Read().Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	configs := make(map[string]*game.RoomConfig)
	for rows.Next() {
		var po RoomConfigPO
		err := rows.Scan(
			&po.ID,
			&po.RoomType,
			&po.RoomName,
			&po.MaxPlayers,
			&po.MinBet,
			&po.MaxBet,
			&po.EntryFee,
			&po.BulletCostMultiplier,
			&po.FishSpawnRate,
			&po.MinFishCount,
			&po.MaxFishCount,
			&po.RoomWidth,
			&po.RoomHeight,
			&po.TargetRTP,
			&po.IsActive,
			&po.Description,
		)
		if err != nil {
			return nil, err
		}

		configs[po.RoomType] = &game.RoomConfig{
			MaxPlayers:           int32(po.MaxPlayers),
			MinBet:               po.MinBet,
			MaxBet:               po.MaxBet,
			BulletCostMultiplier: po.BulletCostMultiplier,
			FishSpawnRate:        po.FishSpawnRate,
			MinFishCount:         int32(po.MinFishCount),
			MaxFishCount:         int32(po.MaxFishCount),
			RoomWidth:            po.RoomWidth,
			RoomHeight:           po.RoomHeight,
			TargetRTP:            po.TargetRTP,
		}
	}

	return configs, rows.Err()
}
