package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

func main() {
	// Define command-line flags
	bulletPower := flag.Int("power", 10, "Power of the bullet.")
	fishHealth := flag.Int("health", 100, "Health of the fish.")
	fishValue := flag.Int64("value", 50, "Base value (reward) of the fish.")
	simulations := flag.Int("sims", 1000, "Number of simulations to run.")
	flag.Parse()

	// --- Setup ---
	log := logger.New(os.Stdout, "info", "console")
	mathModel := game.NewMathModel(log)
	modelConfig := mathModel.GetModelConfig()

	bullet := &game.Bullet{
		Power: int32(*bulletPower),
	}
	fish := &game.Fish{
		Health: int32(*fishHealth),
		Value:  *fishValue,
	}

	// --- Simulation ---
	fmt.Printf("--- Running Math Model Simulation ---\n")
	fmt.Printf("Simulations: %d\n", *simulations)
	fmt.Printf("Bullet Power: %d\n", *bulletPower)
	fmt.Printf("Fish (Health: %d, Value: %d)\n", *fishHealth, *fishValue)
	fmt.Printf("Model Config (CritRate: %.2f%%, CritMultiplier: %.2fx)\n\n", modelConfig.CriticalRate*100, modelConfig.CriticalMultiplier)

	// Stats trackers
	kills := 0
	totalReward := int64(0)
	totalDamage := int64(0)
	criticalHits := 0

	for i := 0; i < *simulations; i++ {
		// Reset fish health for each simulation
		fish.Health = int32(*fishHealth)

		result := mathModel.CalculatePotentialHit(bullet, fish)

		totalDamage += int64(result.Damage)
		if result.IsCritical {
			criticalHits++
		}

		if result.Success { // Success means a potential kill
			kills++
			totalReward += result.Reward
		}
	}

	// --- Results ---
	killRate := float64(kills) / float64(*simulations) * 100
	avgRewardPerKill := float64(0)
	if kills > 0 {
		avgRewardPerKill = float64(totalReward) / float64(kills)
	}
	avgDamage := float64(totalDamage) / float64(*simulations)
	critRate := float64(criticalHits) / float64(*simulations) * 100

	fmt.Printf("--- Simulation Results ---\n")
	fmt.Printf("Kill Rate:         %.2f%% (%d / %d)\n", killRate, kills, *simulations)
	fmt.Printf("Critical Hit Rate: %.2f%% (%d / %d)\n", critRate, criticalHits, *simulations)
	fmt.Printf("Average Damage:    %.2f\n", avgDamage)
	fmt.Printf("Total Reward:      %d\n", totalReward)
	fmt.Printf("Average Reward (per kill): %.2f\n", avgRewardPerKill)
	fmt.Println("------------------------------------")
	fmt.Println("\nUsage Example:")
	fmt.Println("go run ./scripts/math_test -power=50 -health=500 -value=250 -sims=10000")
}
