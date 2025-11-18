// internal/data/redis/room_counter.go
package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// Room counter keys in Redis
const (
	RoomCountKeyPrefix = "room:count:"        // room:count:{room_type}
	TotalRoomCountKey  = "room:count:total"   // 所有房間總數
)

// IncrementRoomCount 增加指定類型房間的數量
// roomType: "novice", "intermediate", "advanced", "vip"
func (c *Client) IncrementRoomCount(ctx context.Context, roomType string) (int64, error) {
	key := RoomCountKeyPrefix + roomType

	// 使用管道批量執行
	pipe := c.Redis.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	totalIncrCmd := pipe.Incr(ctx, TotalRoomCountKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		c.Logger.Errorf("Failed to increment room count for type %s: %v", roomType, err)
		return 0, err
	}

	count := incrCmd.Val()
	totalCount := totalIncrCmd.Val()

	c.Logger.Infof("Incremented room count: type=%s, count=%d, total=%d", roomType, count, totalCount)
	return count, nil
}

// DecrementRoomCount 減少指定類型房間的數量
func (c *Client) DecrementRoomCount(ctx context.Context, roomType string) (int64, error) {
	key := RoomCountKeyPrefix + roomType

	// 使用管道批量執行
	pipe := c.Redis.Pipeline()
	decrCmd := pipe.Decr(ctx, key)
	totalDecrCmd := pipe.Decr(ctx, TotalRoomCountKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		c.Logger.Errorf("Failed to decrement room count for type %s: %v", roomType, err)
		return 0, err
	}

	count := decrCmd.Val()
	totalCount := totalDecrCmd.Val()

	// 防止計數變成負數
	if count < 0 {
		c.Logger.Warnf("Room count became negative for type %s: %d, resetting to 0", roomType, count)
		pipe = c.Redis.Pipeline()
		pipe.Set(ctx, key, 0, 0)
		pipe.Exec(ctx)
		count = 0
	}

	if totalCount < 0 {
		c.Logger.Warnf("Total room count became negative: %d, resetting to 0", totalCount)
		c.Redis.Set(ctx, TotalRoomCountKey, 0, 0)
		totalCount = 0
	}

	c.Logger.Infof("Decremented room count: type=%s, count=%d, total=%d", roomType, count, totalCount)
	return count, nil
}

// GetRoomCount 獲取指定類型房間的數量
func (c *Client) GetRoomCount(ctx context.Context, roomType string) (int64, error) {
	key := RoomCountKeyPrefix + roomType
	count, err := c.GetInt64(ctx, key)
	if err != nil {
		c.Logger.Errorf("Failed to get room count for type %s: %v", roomType, err)
		return 0, err
	}
	return count, nil
}

// GetTotalRoomCount 獲取所有房間的總數量
func (c *Client) GetTotalRoomCount(ctx context.Context) (int64, error) {
	count, err := c.GetInt64(ctx, TotalRoomCountKey)
	if err != nil {
		c.Logger.Errorf("Failed to get total room count: %v", err)
		return 0, err
	}
	return count, nil
}

// GetAllRoomCounts 獲取所有類型房間的數量（返回 map）
func (c *Client) GetAllRoomCounts(ctx context.Context) (map[string]int64, error) {
	roomTypes := []string{"novice", "intermediate", "advanced", "vip"}
	counts := make(map[string]int64)

	// 使用管道批量查詢
	pipe := c.Redis.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, roomType := range roomTypes {
		key := RoomCountKeyPrefix + roomType
		cmds[roomType] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		c.Logger.Errorf("Failed to get all room counts: %v", err)
		return nil, err
	}

	// 解析結果
	for roomType, cmd := range cmds {
		val, err := cmd.Int64()
		if err == redis.Nil {
			counts[roomType] = 0 // key 不存在時返回 0
		} else if err != nil {
			c.Logger.Warnf("Failed to parse room count for type %s: %v", roomType, err)
			counts[roomType] = 0
		} else {
			counts[roomType] = val
		}
	}

	return counts, nil
}

// ResetRoomCount 重置指定類型房間的數量
func (c *Client) ResetRoomCount(ctx context.Context, roomType string) error {
	key := RoomCountKeyPrefix + roomType
	err := c.Set(ctx, key, 0, 0)
	if err != nil {
		c.Logger.Errorf("Failed to reset room count for type %s: %v", roomType, err)
		return err
	}
	c.Logger.Infof("Reset room count for type: %s", roomType)
	return nil
}

// ResetAllRoomCounts 重置所有房間計數
func (c *Client) ResetAllRoomCounts(ctx context.Context) error {
	roomTypes := []string{"novice", "intermediate", "advanced", "vip"}

	pipe := c.Redis.Pipeline()
	for _, roomType := range roomTypes {
		key := RoomCountKeyPrefix + roomType
		pipe.Set(ctx, key, 0, 0)
	}
	pipe.Set(ctx, TotalRoomCountKey, 0, 0)

	_, err := pipe.Exec(ctx)
	if err != nil {
		c.Logger.Errorf("Failed to reset all room counts: %v", err)
		return err
	}

	c.Logger.Info("Reset all room counts")
	return nil
}

// SaveRoomInfo 保存房間基本信息到 Redis
func (c *Client) SaveRoomInfo(ctx context.Context, roomID string, roomData map[string]interface{}) error {
	key := fmt.Sprintf("room:info:%s", roomID)

	// 使用 HSET 保存房間信息
	for field, value := range roomData {
		if err := c.Redis.HSet(ctx, key, field, value).Err(); err != nil {
			c.Logger.Errorf("Failed to save room info field %s for room %s: %v", field, roomID, err)
			return err
		}
	}

	c.Logger.Debugf("Saved room info to Redis: %s", roomID)
	return nil
}

// GetRoomInfo 從 Redis 獲取房間基本信息
func (c *Client) GetRoomInfo(ctx context.Context, roomID string) (map[string]string, error) {
	key := fmt.Sprintf("room:info:%s", roomID)

	data, err := c.Redis.HGetAll(ctx, key).Result()
	if err != nil {
		c.Logger.Errorf("Failed to get room info for room %s: %v", roomID, err)
		return nil, err
	}

	return data, nil
}

// DeleteRoomInfo 從 Redis 刪除房間信息
func (c *Client) DeleteRoomInfo(ctx context.Context, roomID string) error {
	key := fmt.Sprintf("room:info:%s", roomID)

	err := c.Del(ctx, key)
	if err != nil {
		c.Logger.Errorf("Failed to delete room info for room %s: %v", roomID, err)
		return err
	}

	c.Logger.Debugf("Deleted room info from Redis: %s", roomID)
	return nil
}
