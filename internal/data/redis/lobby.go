package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/lobby"
	"github.com/go-redis/redis/v8"
)

// LobbyRedisCache implements lobby Redis caching layer
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
	// 使用 SCAN 命令獲取所有 "room:server:*" 鍵（比 KEYS 更安全）
	pattern := "room:server:*"
	var allRooms []*lobby.RoomInfo

	// 使用 SCAN 遍歷所有匹配的鍵
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// 獲取鍵對應的值
		data, err := c.client.Get(ctx, key).Result()
		if err != nil {
			if err == redis.Nil {
				// 鍵已過期，跳過
				continue
			}
			return nil, err
		}

		// 解析 JSON
		var rooms []*lobby.RoomInfo
		err = json.Unmarshal([]byte(data), &rooms)
		if err != nil {
			// 跳過無效的資料
			continue
		}

		allRooms = append(allRooms, rooms...)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return allRooms, nil
}

// UpdateRoomInfo 更新房間資訊
func (c *roomCache) UpdateRoomInfo(ctx context.Context, gameServerID string, rooms []*lobby.RoomInfo) error {
	// 將 rooms 序列化為 JSON
	data, err := json.Marshal(rooms)
	if err != nil {
		return err
	}

	key := "room:server:" + gameServerID

	// 使用 SET 命令存儲資料，並設定過期時間為 15 秒
	err = c.client.Set(ctx, key, data, 15*time.Second).Err()
	return err
}
