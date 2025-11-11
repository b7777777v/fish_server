package game

import (
	"context"
)

// RoomConfigRepo 房间配置仓库接口
type RoomConfigRepo interface {
	// GetRoomConfig 根据房间类型获取配置
	GetRoomConfig(ctx context.Context, roomType string) (*RoomConfig, error)

	// GetAllRoomConfigs 获取所有房间配置
	GetAllRoomConfigs(ctx context.Context) (map[string]*RoomConfig, error)
}
