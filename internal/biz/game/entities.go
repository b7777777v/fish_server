package game

import (
	"time"
)

// ========================================
// 遊戲核心實體定義
// ========================================

// Player 遊戲玩家
type Player struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"user_id"`
	Nickname string    `json:"nickname"`
	Balance  int64     `json:"balance"`  // 玩家餘額（以分為單位）
	WalletID uint      `json:"wallet_id"` // 錢包ID，用於交易記錄
	RoomID   string    `json:"room_id"`  // 當前房間ID
	SeatID   int       `json:"seat_id"`  // 座位ID (0-3)，-1 表示未分配
	Status   PlayerStatus `json:"status"`
	JoinTime time.Time `json:"join_time"`
}

// PlayerStatus 玩家狀態
type PlayerStatus string

const (
	PlayerStatusIdle    PlayerStatus = "idle"    // 閒置
	PlayerStatusPlaying PlayerStatus = "playing" // 遊戲中
	PlayerStatusOffline PlayerStatus = "offline" // 離線
)

// Fish 魚類實體
type Fish struct {
	ID         int64     `json:"id"`
	Type       FishType  `json:"type"`
	Position   Position  `json:"position"`
	Direction  float64   `json:"direction"`  // 移動方向（弧度）
	Speed      float64   `json:"speed"`      // 移動速度
	Health     int32     `json:"health"`     // 血量
	MaxHealth  int32     `json:"max_health"` // 最大血量
	Value      int64     `json:"value"`      // 擊殺獎勵
	SpawnTime  time.Time `json:"spawn_time"`
	Status     FishStatus `json:"status"`
}

// FishType 魚類型
type FishType struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	Size        string  `json:"size"`        // small, medium, large, boss
	BaseHealth  int32   `json:"base_health"`
	BaseValue   int64   `json:"base_value"`
	BaseSpeed   float64 `json:"base_speed"`
	Rarity      float64 `json:"rarity"`      // 稀有度 0.0-1.0
	HitRate     float64 `json:"hit_rate"`    // 命中率 0.0-1.0
	Description string  `json:"description"`
}

// FishStatus 魚的狀態
type FishStatus string

const (
	FishStatusAlive  FishStatus = "alive"  // 存活
	FishStatusDying  FishStatus = "dying"  // 死亡中
	FishStatusDead   FishStatus = "dead"   // 已死亡
)

// Position 位置信息
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Bullet 子彈實體
type Bullet struct {
	ID        int64    `json:"id"`
	PlayerID  int64    `json:"player_id"`
	Position  Position `json:"position"`
	Direction float64  `json:"direction"`
	Speed     float64  `json:"speed"`
	Power     int32    `json:"power"`     // 攻擊力
	Cost      int64    `json:"cost"`      // 子彈成本
	CreatedAt time.Time `json:"created_at"`
	Status    BulletStatus `json:"status"`
}

// BulletStatus 子彈狀態
type BulletStatus string

const (
	BulletStatusFlying BulletStatus = "flying" // 飛行中
	BulletStatusHit    BulletStatus = "hit"    // 命中目標
	BulletStatusMissed BulletStatus = "missed" // 未命中
)

// Room 遊戲房間
type Room struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Type        RoomType         `json:"type"`
	MaxPlayers  int32            `json:"max_players"`
	Players     map[int64]*Player `json:"players"`
	Seats       []int64          `json:"seats"`        // 座位切片，存储玩家ID，0表示空座位，长度由配置决定
	Fishes      map[int64]*Fish   `json:"fishes"`
	Bullets     map[int64]*Bullet `json:"bullets"`
	Status      RoomStatus       `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Config      RoomConfig       `json:"config"`
}

// RoomType 房間類型
type RoomType string

const (
	RoomTypeNovice      RoomType = "novice"       // 新手房
	RoomTypeIntermediate RoomType = "intermediate" // 中級房
	RoomTypeAdvanced    RoomType = "advanced"     // 高級房
	RoomTypeVIP         RoomType = "vip"          // VIP房
)

// RoomStatus 房間狀態
type RoomStatus string

const (
	RoomStatusWaiting RoomStatus = "waiting" // 等待中
	RoomStatusPlaying RoomStatus = "playing" // 遊戲中
	RoomStatusClosed  RoomStatus = "closed"  // 已關閉
)

// RoomConfig 房間配置
type RoomConfig struct {
	MaxPlayers           int32   `json:"max_players"`       // 最大玩家數（座位數）
	MinBet               int64   `json:"min_bet"`           // 最小下注
	MaxBet               int64   `json:"max_bet"`           // 最大下注
	BulletCostMultiplier float64 `json:"bullet_cost_multiplier"` // 子彈成本倍數
	FishSpawnRate        float64 `json:"fish_spawn_rate"`       // 魚類生成率
	MaxFishCount         int32   `json:"max_fish_count"`    // 最大魚數量
	RoomWidth            float64 `json:"room_width"`        // 房間寬度
	RoomHeight           float64 `json:"room_height"`       // 房間高度
	TargetRTP            float64 `json:"target_rtp"`           // 目標RTP, e.g., 0.96 for 96%
}

// Inventory 遊戲庫存系統
type Inventory struct {
	ID         string    `json:"id"`         // 唯一標識, e.g., room_type_novice
	TotalIn    int64     `json:"total_in"`    // 總投入 (所有玩家的總花費)
	TotalOut   int64     `json:t:"total_out"`   // 總產出 (所有玩家的總贏得)
	CurrentRTP float64   `json:"current_rtp"` // 當前實際RTP (TotalOut / TotalIn)
	UpdatedAt  time.Time `json:"updated_at"`  // 最後更新時間
}

// GameEvent 遊戲事件
type GameEvent struct {
	ID        int64           `json:"id"`
	Type      GameEventType   `json:"type"`
	RoomID    string          `json:"room_id"`
	PlayerID  int64           `json:"player_id,omitempty"`
	Data      interface{}     `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// GameEventType 遊戲事件類型
type GameEventType string

const (
	EventPlayerJoin   GameEventType = "player_join"   // 玩家加入
	EventPlayerLeave  GameEventType = "player_leave"  // 玩家離開
	EventFishSpawn    GameEventType = "fish_spawn"    // 魚類生成
	EventFishDie      GameEventType = "fish_die"      // 魚類死亡
	EventBulletFire   GameEventType = "bullet_fire"   // 開火
	EventBulletHit    GameEventType = "bullet_hit"    // 子彈命中
	EventPlayerReward GameEventType = "player_reward" // 玩家獲得獎勵
)

// HitResult 命中結果
type HitResult struct {
	Success   bool    `json:"success"`   // 是否命中
	Damage    int32   `json:"damage"`    // 造成傷害
	Reward    int64   `json:"reward"`    // 獲得獎勵
	IsCritical bool   `json:"is_critical"` // 是否暴擊
	Multiplier float64 `json:"multiplier"`  // 獎勵倍數
}

// GameStatistics 遊戲統計
type GameStatistics struct {
	TotalShots     int64 `json:"total_shots"`     // 總射擊次數
	TotalHits      int64 `json:"total_hits"`      // 總命中次數
	TotalRewards   int64 `json:"total_rewards"`   // 總獎勵
	TotalCosts     int64 `json:"total_costs"`     // 總花費
	FishKilled     int64 `json:"fish_killed"`     // 殺死魚數量
	PlayTime       int64 `json:"play_time"`       // 遊戲時間（秒）
	HitRate        float64 `json:"hit_rate"`      // 命中率
	ProfitRate     float64 `json:"profit_rate"`   // 盈利率
}