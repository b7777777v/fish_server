package game_test


import "github.com/b7777777v/fish_server/internal/biz/game"
import (
	"testing"

	"github.com/b7777777v/fish_server/internal/testing/testhelper"
	"github.com/stretchr/testify/assert"
)

// TestFishSpawner_TrySpawnFish tests single fish spawning
func TestFishSpawner_TrySpawnFish(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	// Use high spawn rate config to ensure fish spawns
	highRateConfig := env.RoomConfig
	highRateConfig.FishSpawnRate = 10.0 // Very high rate for testing

	t.Run("spawn fish successfully", func(t *testing.T) {
		// Use SpawnSpecificFish for deterministic testing
		fishTypeID := int32(1) // Small fish
		fish := env.Spawner.SpawnSpecificFish(fishTypeID, highRateConfig)

		// Should spawn fish successfully
		assert.NotNil(t, fish, "Should spawn specific fish type")
		if fish != nil {
			assert.NotZero(t, fish.ID)
			assert.NotNil(t, fish.Type)
			assert.Equal(t, fishTypeID, fish.Type.ID)
			assert.Greater(t, fish.Health, int32(0))
			assert.Greater(t, fish.Value, int64(0))
			assert.Equal(t, game.FishStatusAlive, fish.Status)
		}
	})

	t.Run("spawned fish has valid properties", func(t *testing.T) {
		fish := env.Spawner.TrySpawnFish(highRateConfig)

		if fish != nil {
			// Check position is within room bounds
			assert.GreaterOrEqual(t, fish.Position.X, 0.0)
			assert.LessOrEqual(t, fish.Position.X, env.RoomConfig.RoomWidth)
			assert.GreaterOrEqual(t, fish.Position.Y, 0.0)
			assert.LessOrEqual(t, fish.Position.Y, env.RoomConfig.RoomHeight)

			// Check speed is positive
			assert.Greater(t, fish.Speed, 0.0)

			// Check health equals max health
			assert.Equal(t, fish.MaxHealth, fish.Health)
		}
	})
}

// TestFishSpawner_BatchSpawnFish tests batch fish spawning
func TestFishSpawner_BatchSpawnFish(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	tests := []struct {
		name  string
		count int
	}{
		{"spawn 5 fish", 5},
		{"spawn 10 fish", 10},
		{"spawn 20 fish", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fishes := env.Spawner.BatchSpawnFish(tt.count, env.RoomConfig)

			assert.Len(t, fishes, tt.count)
			for _, fish := range fishes {
				assert.NotNil(t, fish)
				assert.NotZero(t, fish.ID)
				assert.Greater(t, fish.Value, int64(0))
			}

			// Check all fish have unique IDs
			idSet := make(map[int64]bool)
			for _, fish := range fishes {
				assert.False(t, idSet[fish.ID], "Fish IDs should be unique")
				idSet[fish.ID] = true
			}
		})
	}
}

// TestFishSpawner_GetFishTypes tests fish type retrieval
func TestFishSpawner_GetFishTypes(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("get all fish types", func(t *testing.T) {
		fishTypes := env.Spawner.GetFishTypes()

		assert.NotEmpty(t, fishTypes)
		for _, fishType := range fishTypes {
			assert.NotZero(t, fishType.ID)
			assert.NotEmpty(t, fishType.Name)
			assert.NotEmpty(t, fishType.Size)
			assert.Greater(t, fishType.BaseValue, int64(0))
			assert.Greater(t, fishType.BaseSpeed, 0.0)
		}
	})
}

// TestFishSpawner_FishDistribution tests fish type distribution
func TestFishSpawner_FishDistribution(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("fish type distribution follows rarity", func(t *testing.T) {
		sampleSize := 1000
		fishes := env.Spawner.BatchSpawnFish(sampleSize, env.RoomConfig)

		// Count fish by size
		sizeCount := make(map[string]int)
		for _, fish := range fishes {
			sizeCount[fish.Type.Size]++
		}

		// Small fish should be most common
		assert.Greater(t, sizeCount["small"], sizeCount["medium"],
			"Small fish should be more common than medium")
		assert.Greater(t, sizeCount["medium"], sizeCount["large"],
			"Medium fish should be more common than large")

		// Boss fish should be rarest
		if sizeCount["boss"] > 0 {
			assert.Less(t, sizeCount["boss"], sizeCount["small"]/10,
				"Boss fish should be rare")
		}
	})
}

// TestFishSpawner_MinFishCountReplenishment tests automatic fish replenishment
func TestFishSpawner_MinFishCountReplenishment(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("batch spawn for replenishment", func(t *testing.T) {
		// Simulate low fish count scenario
		minFish := int(env.RoomConfig.MinFishCount)
		maxFish := int(env.RoomConfig.MaxFishCount)
		currentFish := minFish - 5 // Below minimum

		// Calculate replenishment needed
		targetFish := int(float64(maxFish) * 0.75)
		replenishCount := targetFish - currentFish

		// Spawn fish to replenish
		fishes := env.Spawner.BatchSpawnFish(replenishCount, env.RoomConfig)

		assert.Len(t, fishes, replenishCount)
		assert.Equal(t, targetFish, currentFish+len(fishes),
			"Should replenish to 75% of max")
	})
}

// TestFishSpawner_FishValueRange tests fish value ranges by type
func TestFishSpawner_FishValueRange(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("fish values scale with size", func(t *testing.T) {
		fishes := env.Spawner.BatchSpawnFish(100, env.RoomConfig)

		valueBySize := make(map[string][]int64)
		for _, fish := range fishes {
			valueBySize[fish.Type.Size] = append(valueBySize[fish.Type.Size], fish.Value)
		}

		// Calculate average values
		avgValue := func(values []int64) float64 {
			if len(values) == 0 {
				return 0
			}
			sum := int64(0)
			for _, v := range values {
				sum += v
			}
			return float64(sum) / float64(len(values))
		}

		// Boss fish should have highest average value
		if len(valueBySize["boss"]) > 0 {
			assert.Greater(t, avgValue(valueBySize["boss"]), avgValue(valueBySize["large"]))
		}
		if len(valueBySize["large"]) > 0 {
			assert.Greater(t, avgValue(valueBySize["large"]), avgValue(valueBySize["medium"]))
		}
		if len(valueBySize["medium"]) > 0 {
			assert.Greater(t, avgValue(valueBySize["medium"]), avgValue(valueBySize["small"]))
		}
	})
}

// TestFishSpawner_EdgeCases tests edge cases
func TestFishSpawner_EdgeCases(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("batch spawn with zero count", func(t *testing.T) {
		fishes := env.Spawner.BatchSpawnFish(0, env.RoomConfig)
		assert.Empty(t, fishes)
	})

	t.Run("batch spawn with negative count", func(t *testing.T) {
		// Should handle gracefully
		assert.NotPanics(t, func() {
			env.Spawner.BatchSpawnFish(-5, env.RoomConfig)
		})
	})

	t.Run("batch spawn large count", func(t *testing.T) {
		// Should handle large batches
		largeCount := 100
		fishes := env.Spawner.BatchSpawnFish(largeCount, env.RoomConfig)
		assert.Len(t, fishes, largeCount)
	})
}

// TestFishSpawner_Configuration tests spawner with different configurations
func TestFishSpawner_Configuration(t *testing.T) {
	tests := []struct {
		name       string
		spawnRate  float64
		batchCount int
		minExpect  int
	}{
		{
			name:       "high spawn rate",
			spawnRate:  10.0, // High rate for reliable spawning
			batchCount: 10,
			minExpect:  5, // Expect at least half to spawn
		},
		{
			name:       "medium spawn rate",
			spawnRate:  5.0,
			batchCount: 10,
			minExpect:  3,
		},
		{
			name:       "low spawn rate",
			spawnRate:  1.0,
			batchCount: 10,
			minExpect:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customConfig := testhelper.DefaultRoomConfig()
			customConfig.FishSpawnRate = tt.spawnRate

			env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
				RoomConfig: &customConfig,
			})
			defer env.AssertExpectations(t)

			// Use BatchSpawnFish which handles timing internally
			fishes := env.Spawner.BatchSpawnFish(tt.batchCount, customConfig)

			assert.GreaterOrEqual(t, len(fishes), tt.minExpect,
				"Should spawn at least %d fish with spawn rate %.1f", tt.minExpect, tt.spawnRate)
		})
	}
}

// TestFishSpawner_FishMovement tests fish movement properties
func TestFishSpawner_FishMovement(t *testing.T) {
	env := testhelper.NewGameTestEnv(t, nil)
	defer env.AssertExpectations(t)

	t.Run("fish have valid movement direction", func(t *testing.T) {
		fishes := env.Spawner.BatchSpawnFish(50, env.RoomConfig)

		for _, fish := range fishes {
			// Direction should be in radians (0 to 2π or -π to π)
			assert.GreaterOrEqual(t, fish.Direction, -3.15)
			assert.LessOrEqual(t, fish.Direction, 6.29)

			// Speed should be based on fish type
			assert.Greater(t, fish.Speed, 0.0)
			assert.LessOrEqual(t, fish.Speed, fish.Type.BaseSpeed*2.0,
				"Fish speed should not exceed 2x base speed")
		}
	})
}
