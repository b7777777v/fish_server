package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/go-redis/redis/v8"
)

// RoomConfigCache 房间配置Redis缓存
type RoomConfigCache struct {
	client *redis.Client
}

// NewRoomConfigCache 创建新的 RoomConfigCache 实例
func NewRoomConfigCache(client *redis.Client) *RoomConfigCache {
	return &RoomConfigCache{
		client: client,
	}
}

// GetRoomConfig 从缓存获取房间配置
func (c *RoomConfigCache) GetRoomConfig(ctx context.Context, roomType string) (*game.RoomConfig, error) {
	key := fmt.Sprintf("room_config:%s", roomType)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, err
	}

	var config game.RoomConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SetRoomConfig 设置房间配置到缓存
func (c *RoomConfigCache) SetRoomConfig(ctx context.Context, roomType string, config *game.RoomConfig) error {
	key := fmt.Sprintf("room_config:%s", roomType)

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	// 缓存1小时
	return c.client.Set(ctx, key, data, 1*time.Hour).Err()
}

// DeleteRoomConfig 删除缓存中的房间配置
func (c *RoomConfigCache) DeleteRoomConfig(ctx context.Context, roomType string) error {
	key := fmt.Sprintf("room_config:%s", roomType)
	return c.client.Del(ctx, key).Err()
}

// GetAllRoomConfigs 获取所有房间配置（暂不实现批量缓存，按需从DB加载）
func (c *RoomConfigCache) GetAllRoomConfigs(ctx context.Context) (map[string]*game.RoomConfig, error) {
	// 这个方法可以根据需要实现，但通常我们会逐个房间类型获取缓存
	return nil, fmt.Errorf("not implemented: use GetRoomConfig for each room type")
}
