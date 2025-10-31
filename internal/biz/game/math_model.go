package game

import (
	"math"
	"math/rand"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// ========================================
// MathModel 遊戲數學模型（簡化版）
// ========================================

// MathModel 數學模型
type MathModel struct {
	logger logger.Logger
	rng    *rand.Rand
	config ModelConfig
}

// ModelConfig 模型配置
type ModelConfig struct {
	// 基礎參數
	BaseHitRate      float64 `json:"base_hit_rate"`      // 基礎命中率
	CriticalRate     float64 `json:"critical_rate"`      // 暴擊率
	CriticalMultiplier float64 `json:"critical_multiplier"` // 暴擊倍數
	
	// 平衡參數
	HouseEdge        float64 `json:"house_edge"`         // 莊家優勢 (5-10%)
	MaxPayout        float64 `json:"max_payout"`         // 最大賠付倍數
	MinHitRate       float64 `json:"min_hit_rate"`       // 最小命中率
	MaxHitRate       float64 `json:"max_hit_rate"`       // 最大命中率
	
	// 動態調整參數
	WinStreakPenalty   float64 `json:"win_streak_penalty"`   // 連勝懲罰
	LoseStreakBonus    float64 `json:"lose_streak_bonus"`    // 連敗獎勵
	BalanceInfluence   float64 `json:"balance_influence"`    // 餘額影響因子
}

// NewMathModel 創建數學模型
func NewMathModel(logger logger.Logger) *MathModel {
	return &MathModel{
		logger: logger.With("component", "math_model"),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		config: getDefaultModelConfig(),
	}
}

// CalculateHit 計算命中結果
func (mm *MathModel) CalculateHit(bullet *Bullet, fish *Fish) *HitResult {
	// 1. 計算基礎命中率
	baseHitRate := mm.calculateBaseHitRate(bullet, fish)
	
	// 2. 應用各種修正因子
	finalHitRate := mm.applyHitRateModifiers(baseHitRate, bullet, fish)
	
	// 3. 判斷是否命中
	isHit := mm.rng.Float64() < finalHitRate
	
	if !isHit {
		return &HitResult{
			Success:    false,
			Damage:     0,
			Reward:     0,
			IsCritical: false,
			Multiplier: 0,
		}
	}
	
	// 4. 計算傷害
	damage := mm.calculateDamage(bullet, fish)
	
	// 5. 判斷是否暴擊
	isCritical := mm.rng.Float64() < mm.config.CriticalRate
	if isCritical {
		damage = int32(float64(damage) * mm.config.CriticalMultiplier)
	}
	
	// 6. 計算獎勵（如果魚死亡）
	reward := int64(0)
	multiplier := 1.0
	
	if damage >= fish.Health {
		// 魚被殺死，計算獎勵
		reward, multiplier = mm.calculateReward(bullet, fish, isCritical)
	}
	
	mm.logger.Debugf("Hit result: hit=%t, damage=%d, reward=%d, critical=%t", 
		isHit, damage, reward, isCritical)
	
	return &HitResult{
		Success:    true,
		Damage:     damage,
		Reward:     reward,
		IsCritical: isCritical,
		Multiplier: multiplier,
	}
}

// calculateBaseHitRate 計算基礎命中率
func (mm *MathModel) calculateBaseHitRate(bullet *Bullet, fish *Fish) float64 {
	// 基於魚的類型命中率
	fishHitRate := fish.Type.HitRate
	
	// 基於子彈威力的修正
	powerModifier := math.Min(float64(bullet.Power)/100.0, 2.0) // 最多2倍修正
	
	// 基於魚的大小修正
	sizeModifier := mm.getSizeModifier(fish.Type.Size)
	
	baseRate := fishHitRate * powerModifier * sizeModifier
	
	// 限制在合理範圍內
	return math.Max(mm.config.MinHitRate, math.Min(baseRate, mm.config.MaxHitRate))
}

// applyHitRateModifiers 應用命中率修正因子
func (mm *MathModel) applyHitRateModifiers(baseRate float64, bullet *Bullet, fish *Fish) float64 {
	finalRate := baseRate
	
	// 魚的速度影響（速度越快越難命中）
	speedPenalty := math.Min(fish.Speed/200.0, 0.3) // 最多30%懲罰
	finalRate *= (1.0 - speedPenalty)
	
	// 魚的稀有度影響（越稀有越難命中）
	rarityPenalty := fish.Type.Rarity * 0.5 // 最多50%懲罰
	finalRate *= (1.0 - rarityPenalty)
	
	// 應用莊家優勢
	finalRate *= (1.0 - mm.config.HouseEdge)
	
	// 確保在合理範圍內
	return math.Max(mm.config.MinHitRate, math.Min(finalRate, mm.config.MaxHitRate))
}

// calculateDamage 計算傷害
func (mm *MathModel) calculateDamage(bullet *Bullet, fish *Fish) int32 {
	// 基礎傷害等於子彈威力
	baseDamage := bullet.Power
	
	// 隨機變化 ±20%
	randomFactor := 0.8 + mm.rng.Float64()*0.4
	
	damage := int32(float64(baseDamage) * randomFactor)
	
	// 至少造成1點傷害
	if damage < 1 {
		damage = 1
	}
	
	return damage
}

// calculateReward 計算獎勵
func (mm *MathModel) calculateReward(bullet *Bullet, fish *Fish, isCritical bool) (int64, float64) {
	// 基礎獎勵
	baseReward := fish.Value
	
	// 子彈威力影響獎勵
	powerMultiplier := 1.0 + float64(bullet.Power)/1000.0 // 威力越高獎勵略微增加
	
	// 暴擊獎勵
	criticalMultiplier := 1.0
	if isCritical {
		criticalMultiplier = mm.config.CriticalMultiplier
	}
	
	// 稀有度獎勵
	rarityMultiplier := 1.0 + fish.Type.Rarity // 稀有度越高獎勵越高
	
	// 計算總倍數
	totalMultiplier := powerMultiplier * criticalMultiplier * rarityMultiplier
	
	// 限制最大賠付
	totalMultiplier = math.Min(totalMultiplier, mm.config.MaxPayout)
	
	// 應用莊家優勢
	totalMultiplier *= (1.0 - mm.config.HouseEdge)
	
	// 計算最終獎勵
	finalReward := int64(float64(baseReward) * totalMultiplier)
	
	// 添加隨機變化 ±10%
	randomFactor := 0.9 + mm.rng.Float64()*0.2
	finalReward = int64(float64(finalReward) * randomFactor)
	
	// 確保獎勵不為負數
	if finalReward < 0 {
		finalReward = 1
	}
	
	return finalReward, totalMultiplier
}

// getSizeModifier 獲取大小修正係數
func (mm *MathModel) getSizeModifier(size string) float64 {
	switch size {
	case "small":
		return 1.2 // 小魚容易命中
	case "medium":
		return 1.0 // 中型魚正常
	case "large":
		return 0.8 // 大魚較難命中
	case "boss":
		return 0.5 // Boss級魚類很難命中
	default:
		return 1.0
	}
}

// CalculateExpectedReturn 計算期望回報率
func (mm *MathModel) CalculateExpectedReturn(bulletCost int64, fish *Fish) float64 {
	// 模擬多次射擊的期望回報
	totalCost := float64(bulletCost)
	expectedReward := 0.0
	
	// 簡化計算：基於魚的基礎獎勵和命中率
	hitProbability := fish.Type.HitRate * (1.0 - mm.config.HouseEdge)
	averageReward := float64(fish.Value) * (1.0 - mm.config.HouseEdge)
	
	expectedReward = hitProbability * averageReward
	
	return expectedReward / totalCost
}

// AdjustDifficulty 動態調整難度（基於玩家表現）
func (mm *MathModel) AdjustDifficulty(playerStats *GameStatistics) {
	// 如果玩家連續獲勝，增加難度
	if playerStats.HitRate > 0.8 {
		mm.config.HouseEdge = math.Min(mm.config.HouseEdge*1.1, 0.15)
		mm.logger.Infof("Increased difficulty due to high hit rate: %f", mm.config.HouseEdge)
	}
	
	// 如果玩家連續失敗，降低難度
	if playerStats.HitRate < 0.3 {
		mm.config.HouseEdge = math.Max(mm.config.HouseEdge*0.9, 0.05)
		mm.logger.Infof("Decreased difficulty due to low hit rate: %f", mm.config.HouseEdge)
	}
	
	// 重置統計以避免過度調整
	if playerStats.TotalShots > 100 {
		mm.logger.Info("Resetting player statistics for difficulty adjustment")
	}
}

// GetModelStats 獲取模型統計信息
func (mm *MathModel) GetModelStats() map[string]interface{} {
	return map[string]interface{}{
		"house_edge":         mm.config.HouseEdge,
		"base_hit_rate":      mm.config.BaseHitRate,
		"critical_rate":      mm.config.CriticalRate,
		"critical_multiplier": mm.config.CriticalMultiplier,
		"max_payout":         mm.config.MaxPayout,
	}
}

// getDefaultModelConfig 獲取默認模型配置
func getDefaultModelConfig() ModelConfig {
	return ModelConfig{
		BaseHitRate:        0.7,  // 70%基礎命中率
		CriticalRate:       0.1,  // 10%暴擊率
		CriticalMultiplier: 2.0,  // 2倍暴擊傷害
		HouseEdge:          0.08, // 8%莊家優勢
		MaxPayout:          10.0, // 最大10倍賠付
		MinHitRate:         0.1,  // 最小10%命中率
		MaxHitRate:         0.95, // 最大95%命中率
		WinStreakPenalty:   0.02, // 2%連勝懲罰
		LoseStreakBonus:    0.02, // 2%連敗獎勵
		BalanceInfluence:   0.01, // 1%餘額影響
	}
}