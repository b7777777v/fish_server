package game

import (
	"math/rand"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// ========================================
// FishSpawner 魚類生成器
// ========================================

// FishSpawner 魚類生成器
type FishSpawner struct {
	fishTypes    []FishType
	logger       logger.Logger
	lastSpawnTime time.Time
	rng          *rand.Rand
}

// NewFishSpawner 創建魚類生成器
func NewFishSpawner(logger logger.Logger) *FishSpawner {
	spawner := &FishSpawner{
		fishTypes:     getDefaultFishTypes(),
		logger:        logger.With("component", "fish_spawner"),
		lastSpawnTime: time.Now(),
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	
	return spawner
}

// TrySpawnFish 嘗試生成魚
func (fs *FishSpawner) TrySpawnFish(config RoomConfig) *Fish {
	now := time.Now()
	
	// 檢查生成間隔（防止生成過於頻繁）
	if now.Sub(fs.lastSpawnTime) < time.Duration(1000/config.FishSpawnRate)*time.Millisecond {
		return nil
	}
	
	// 隨機決定是否生成魚
	if fs.rng.Float64() > config.FishSpawnRate {
		return nil
	}
	
	// 隨機選擇魚類型
	fishType := fs.selectRandomFishType()
	if fishType == nil {
		return nil
	}
	
	// 創建魚實例
	fish := fs.createFish(fishType, config)
	fs.lastSpawnTime = now
	
	fs.logger.Debugf("Spawned fish: type=%s, id=%d", fishType.Name, fish.ID)
	return fish
}

// SpawnSpecificFish 生成指定類型的魚
func (fs *FishSpawner) SpawnSpecificFish(fishTypeID int32, config RoomConfig) *Fish {
	fishType := fs.getFishTypeByID(fishTypeID)
	if fishType == nil {
		fs.logger.Warnf("Fish type not found: %d", fishTypeID)
		return nil
	}
	
	return fs.createFish(fishType, config)
}

// GetFishTypes 獲取所有魚類型
func (fs *FishSpawner) GetFishTypes() []FishType {
	return fs.fishTypes
}

// selectRandomFishType 隨機選擇魚類型（基於稀有度）
func (fs *FishSpawner) selectRandomFishType() *FishType {
	// 計算總權重
	totalWeight := 0.0
	for _, fishType := range fs.fishTypes {
		totalWeight += (1.0 - fishType.Rarity) // 稀有度越低，權重越高
	}
	
	// 隨機選擇
	randomValue := fs.rng.Float64() * totalWeight
	currentWeight := 0.0
	
	for _, fishType := range fs.fishTypes {
		currentWeight += (1.0 - fishType.Rarity)
		if randomValue <= currentWeight {
			return &fishType
		}
	}
	
	// 如果沒有選中，返回第一個
	if len(fs.fishTypes) > 0 {
		return &fs.fishTypes[0]
	}
	
	return nil
}

// createFish 創建魚實例
func (fs *FishSpawner) createFish(fishType *FishType, config RoomConfig) *Fish {
	// 隨機生成位置和屬性
	spawnSide := fs.rng.Intn(4) // 0=左, 1=右, 2=上, 3=下
	var position Position
	var direction float64
	
	switch spawnSide {
	case 0: // 從左側進入
		position = Position{X: -50, Y: fs.rng.Float64() * config.RoomHeight}
		direction = 1.0 // 向右
	case 1: // 從右側進入
		position = Position{X: config.RoomWidth + 50, Y: fs.rng.Float64() * config.RoomHeight}
		direction = -1.0 // 向左
	case 2: // 從上方進入
		position = Position{X: fs.rng.Float64() * config.RoomWidth, Y: -50}
		direction = 0.5 // 向右下
	case 3: // 從下方進入
		position = Position{X: fs.rng.Float64() * config.RoomWidth, Y: config.RoomHeight + 50}
		direction = 0.5 // 向右上
	}
	
	// 添加隨機變化
	healthVariation := 0.8 + fs.rng.Float64()*0.4 // 80%-120%
	valueVariation := 0.9 + fs.rng.Float64()*0.2  // 90%-110%
	speedVariation := 0.8 + fs.rng.Float64()*0.4  // 80%-120%
	
	health := int32(float64(fishType.BaseHealth) * healthVariation)
	value := int64(float64(fishType.BaseValue) * valueVariation)
	speed := fishType.BaseSpeed * speedVariation
	
	fish := &Fish{
		ID:        time.Now().UnixNano() + int64(fs.rng.Intn(1000)),
		Type:      *fishType,
		Position:  position,
		Direction: direction,
		Speed:     speed,
		Health:    health,
		MaxHealth: health,
		Value:     value,
		SpawnTime: time.Now(),
		Status:    FishStatusAlive,
	}
	
	return fish
}

// getFishTypeByID 根據ID獲取魚類型
func (fs *FishSpawner) getFishTypeByID(id int32) *FishType {
	for _, fishType := range fs.fishTypes {
		if fishType.ID == id {
			return &fishType
		}
	}
	return nil
}

// getDefaultFishTypes 獲取默認魚類型配置
func getDefaultFishTypes() []FishType {
	return []FishType{
		// 小型魚類 - 高頻率出現
		{
			ID:          1,
			Name:        "小丑魚",
			Size:        "small",
			BaseHealth:  1,
			BaseValue:   5,  // 0.05元
			BaseSpeed:   100.0,
			Rarity:      0.1, // 10%稀有度，90%出現率
			HitRate:     0.9,
			Description: "最常見的小魚，容易捕捉",
		},
		{
			ID:          2,
			Name:        "熱帶魚",
			Size:        "small",
			BaseHealth:  1,
			BaseValue:   8,
			BaseSpeed:   120.0,
			Rarity:      0.15,
			HitRate:     0.85,
			Description: "色彩鮮豔的小魚",
		},
		{
			ID:          3,
			Name:        "銀魚",
			Size:        "small",
			BaseHealth:  1,
			BaseValue:   10,
			BaseSpeed:   150.0,
			Rarity:      0.2,
			HitRate:     0.8,
			Description: "游速較快的小魚",
		},
		
		// 中型魚類 - 中等頻率
		{
			ID:          11,
			Name:        "石斑魚",
			Size:        "medium",
			BaseHealth:  3,
			BaseValue:   25, // 0.25元
			BaseSpeed:   80.0,
			Rarity:      0.4,
			HitRate:     0.7,
			Description: "中等大小的魚類，需要多發子彈",
		},
		{
			ID:          12,
			Name:        "鯛魚",
			Size:        "medium",
			BaseHealth:  4,
			BaseValue:   35,
			BaseSpeed:   90.0,
			Rarity:      0.45,
			HitRate:     0.65,
			Description: "較為堅韌的中型魚",
		},
		{
			ID:          13,
			Name:        "比目魚",
			Size:        "medium",
			BaseHealth:  2,
			BaseValue:   40,
			BaseSpeed:   60.0,
			Rarity:      0.5,
			HitRate:     0.6,
			Description: "游速慢但獎勵豐厚",
		},
		
		// 大型魚類 - 低頻率出現
		{
			ID:          21,
			Name:        "鯊魚",
			Size:        "large",
			BaseHealth:  10,
			BaseValue:   100, // 1元
			BaseSpeed:   70.0,
			Rarity:      0.7,
			HitRate:     0.5,
			Description: "大型掠食者，獎勵豐厚但難以捕捉",
		},
		{
			ID:          22,
			Name:        "鮪魚",
			Size:        "large",
			BaseHealth:  8,
			BaseValue:   120,
			BaseSpeed:   110.0,
			Rarity:      0.75,
			HitRate:     0.45,
			Description: "速度很快的大型魚類",
		},
		{
			ID:          23,
			Name:        "魔鬼魚",
			Size:        "large",
			BaseHealth:  12,
			BaseValue:   150,
			BaseSpeed:   50.0,
			Rarity:      0.8,
			HitRate:     0.4,
			Description: "血量極高的大型魚類",
		},
		
		// Boss級魚類 - 極低頻率
		{
			ID:          31,
			Name:        "龍王魚",
			Size:        "boss",
			BaseHealth:  50,
			BaseValue:   500, // 5元
			BaseSpeed:   40.0,
			Rarity:      0.95,
			HitRate:     0.2,
			Description: "傳說中的龍王，極難捕捉但獎勵巨大",
		},
		{
			ID:          32,
			Name:        "金龍魚",
			Size:        "boss",
			BaseHealth:  30,
			BaseValue:   800,
			BaseSpeed:   30.0,
			Rarity:      0.97,
			HitRate:     0.15,
			Description: "黃金之魚，擁有最高的獎勵",
		},
		{
			ID:          33,
			Name:        "海王魚",
			Size:        "boss",
			BaseHealth:  80,
			BaseValue:   1000, // 10元
			BaseSpeed:   25.0,
			Rarity:      0.99,
			HitRate:     0.1,
			Description: "海洋之王，最終Boss級別的魚類",
		},
	}
}

// BatchSpawnFish 批量生成魚（用於房間初始化）
func (fs *FishSpawner) BatchSpawnFish(count int, config RoomConfig) []*Fish {
	fishes := make([]*Fish, 0, count)
	
	for i := 0; i < count; i++ {
		if fish := fs.TrySpawnFish(config); fish != nil {
			fishes = append(fishes, fish)
		}
		
		// 添加小延遲避免ID衝突
		time.Sleep(1 * time.Millisecond)
	}
	
	fs.logger.Infof("Batch spawned %d fishes", len(fishes))
	return fishes
}

// GetFishTypeBySize 根據大小獲取魚類型
func (fs *FishSpawner) GetFishTypesBySize(size string) []FishType {
	var result []FishType
	for _, fishType := range fs.fishTypes {
		if fishType.Size == size {
			result = append(result, fishType)
		}
	}
	return result
}

// UpdateSpawnRate 更新生成率（可用於動態調整）
func (fs *FishSpawner) UpdateSpawnRate(roomID string, newRate float64) {
	fs.logger.Infof("Updated spawn rate for room %s: %f", roomID, newRate)
	// 這裡可以添加房間特定的生成率邏輯
}