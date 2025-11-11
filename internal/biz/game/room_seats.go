package game

import (
	"fmt"
)

// ========================================
// 房间座位管理辅助方法
// ========================================

// AllocateSeat 为玩家分配座位
// 返回分配的座位ID (0-3)，如果房间已满则返回错误
func (r *Room) AllocateSeat(playerID int64) (int, error) {
	// 遍历座位数组，找到第一个空座位
	for seatID := 0; seatID < len(r.Seats); seatID++ {
		if r.Seats[seatID] == 0 {
			r.Seats[seatID] = playerID
			return seatID, nil
		}
	}

	return -1, fmt.Errorf("room is full, no available seats")
}

// ReleaseSeat 释放座位
func (r *Room) ReleaseSeat(seatID int) error {
	if seatID < 0 || seatID >= len(r.Seats) {
		return fmt.Errorf("invalid seat ID: %d", seatID)
	}

	r.Seats[seatID] = 0
	return nil
}

// GetPlayerSeat 获取玩家的座位ID
func (r *Room) GetPlayerSeat(playerID int64) int {
	for seatID, occupantID := range r.Seats {
		if occupantID == playerID {
			return seatID
		}
	}
	return -1 // 玩家不在任何座位上
}

// IsSeated 检查玩家是否已入座
func (r *Room) IsSeated(playerID int64) bool {
	return r.GetPlayerSeat(playerID) != -1
}

// GetAvailableSeatsCount 获取可用座位数
func (r *Room) GetAvailableSeatsCount() int {
	count := 0
	for _, occupantID := range r.Seats {
		if occupantID == 0 {
			count++
		}
	}
	return count
}

// IsFull 检查房间是否已满
func (r *Room) IsFull() bool {
	return r.GetAvailableSeatsCount() == 0
}

// GetSeatedPlayers 获取所有已入座的玩家ID列表
func (r *Room) GetSeatedPlayers() []int64 {
	players := make([]int64, 0, len(r.Seats))
	for _, playerID := range r.Seats {
		if playerID != 0 {
			players = append(players, playerID)
		}
	}
	return players
}

// GetSeatInfo 获取座位信息（用于调试和展示）
func (r *Room) GetSeatInfo() map[int]int64 {
	seatInfo := make(map[int]int64)
	for seatID, playerID := range r.Seats {
		seatInfo[seatID] = playerID
	}
	return seatInfo
}
