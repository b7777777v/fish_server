package redis

import (
	"context"
	"encoding/json"

	"github.com/b7777777v/fish_server/internal/biz/lobby"
	"github.com/redis/go-redis/v9"
)

// TODO: 實現大廳 Redis 快取層
// 此檔案實現 RoomCache 介面，用於房間列表的快取管理

// roomCache 實現 lobby.RoomCache 介面
type roomCache struct {
	client *redis.Client
}

// NewRoomCache 建立新的 RoomCache 實例
func NewRoomCache(client *redis.Client) lobby.RoomCache {
	return &roomCache{
		client: client,
	}
}

// GetAllRooms 獲取所有房間資訊
func (c *roomCache) GetAllRooms(ctx context.Context) ([]*lobby.RoomInfo, error) {
	// TODO: 實現從 Redis 獲取所有房間資訊
	// 1. 使用 KEYS 或 SCAN 命令獲取所有 "room:server:*" 鍵
	// 2. 使用 MGET 批量獲取所有房間資料
	// 3. 解析 JSON 並返回房間列表
	// 建議的 Redis key 格式: "room:server:{game_server_id}"
	// 建議的值格式: JSON array of RoomInfo
	panic("not implemented")
}

// UpdateRoomInfo 更新房間資訊
func (c *roomCache) UpdateRoomInfo(ctx context.Context, gameServerID string, rooms []*lobby.RoomInfo) error {
	// TODO: 實現更新房間資訊到 Redis
	// 1. 將 rooms 序列化為 JSON
	// 2. 使用 SET 命令存儲到 Redis
	// 3. 設定適當的過期時間（例如 15 秒，比上報週期略長）
	// Key: "room:server:{game_server_id}"
	// Value: JSON array of RoomInfo

	data, err := json.Marshal(rooms)
	if err != nil {
		return err
	}

	key := "room:server:" + gameServerID
	// TODO: 使用 client.Set() 存儲資料，並設定過期時間
	_ = data
	_ = key
	panic("not implemented")
}
