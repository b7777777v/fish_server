package game

import (
	"math"
	"math/rand"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// ========================================
// MathModel 遊戲數學模型（完整版）
// ========================================

// MathModel 數學模型
type MathModel struct {
	logger logger.Logger
	rng    *rand.Rand
	config ModelConfig
}

// ModelConfig 模型配置
type ModelConfig struct {
	CriticalRate       float64 `json:"critical_rate"`      // 暴擊率
	CriticalMultiplier float64 `json:"critical_multiplier"` // 暴擊倍數
	MaxPayoutMultiplier  float64 `json:"max_payout_multiplier"` // 最大賠付倍數
}

// NewMathModel 創建數學模型
func NewMathModel(logger logger.Logger) *MathModel {
	return &MathModel{
		logger: logger.With("component", "math_model"),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		config: getDefaultModelConfig(),
	}
}

// CalculatePotentialHit calculates the potential outcome of a bullet hitting a fish.
// It determines the potential damage and reward, but does not make the final decision.
// The decision to grant the reward is left to the RTPController.
func (mm *MathModel) CalculatePotentialHit(bullet *Bullet, fish *Fish) *HitResult {
	// 1. Calculate base damage
	damage := mm.calculateDamage(bullet)

	// 2. Determine if it's a critical hit
	isCritical := mm.rng.Float64() < mm.config.CriticalRate
	if isCritical {
		damage = int32(float64(damage) * mm.config.CriticalMultiplier)
	}

	// 3. Check if the damage is enough to kill the fish
	kill := damage >= fish.Health

	// 4. If it's a kill, calculate the potential reward
	potentialReward := int64(0)
	multiplier := 1.0
	if kill {
		potentialReward, multiplier = mm.calculateReward(bullet, fish, isCritical)
	}

	return &HitResult{
		Success:    kill, // Success now means a potential kill
		Damage:     damage,
		Reward:     potentialReward,
		IsCritical: isCritical,
		Multiplier: multiplier,
	}
}

// calculateDamage calculates the damage a bullet deals.
func (mm *MathModel) calculateDamage(bullet *Bullet) int32 {
	// Base damage is the bullet's power
	baseDamage := bullet.Power

	// Add some randomness (+/- 20%)
	randomFactor := 0.8 + mm.rng.Float64()*0.4
	damage := int32(float64(baseDamage) * randomFactor)

	return damage
}

// calculateReward calculates the reward for killing a fish.
func (mm *MathModel) calculateReward(bullet *Bullet, fish *Fish, isCritical bool) (int64, float64) {
	baseReward := fish.Value

	// Multiplier based on bullet power (higher power, slightly better reward ratio)
	powerMultiplier := 1.0 + math.Log1p(float64(bullet.Power)/10.0)*0.1

	criticalMultiplier := 1.0
	if isCritical {
		criticalMultiplier = mm.config.CriticalMultiplier
	}

	// Add a small random multiplier for variance
	randomMultiplier := 1.0 + (mm.rng.Float64()-0.5)*0.2 // +/- 10%

	totalMultiplier := powerMultiplier * criticalMultiplier * randomMultiplier

	// Cap the multiplier to avoid extreme payouts
	if totalMultiplier > mm.config.MaxPayoutMultiplier {
		totalMultiplier = mm.config.MaxPayoutMultiplier
	}

	finalReward := int64(float64(baseReward) * totalMultiplier)

	// Ensure reward is at least the fish's base value
	if finalReward < baseReward {
		finalReward = baseReward
	}

	return finalReward, totalMultiplier
}

// GetModelConfig returns the current configuration of the math model.
func (mm *MathModel) GetModelConfig() ModelConfig {
	return mm.config
}

// getDefaultModelConfig returns the default model configuration.
func getDefaultModelConfig() ModelConfig {
	return ModelConfig{
		CriticalRate:       0.05,  // 5% chance of a critical hit
		CriticalMultiplier: 2.5,   // 2.5x reward on critical
		MaxPayoutMultiplier:  50.0,  // Max reward is 50x the fish's base value
	}
}