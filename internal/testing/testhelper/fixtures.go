package testhelper

import (
	"time"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/wallet"
)

// ========================================
// Test Data Fixtures
// ========================================

// NewTestPlayer creates a test player with default values
func NewTestPlayer(playerID int64) *game.Player {
	return &game.Player{
		ID:       playerID,
		UserID:   playerID,
		Nickname: "TestPlayer",
		Balance:  100000, // 1000.00 CNY in cents
		WalletID: 1,
		Status:   game.PlayerStatusIdle,
		JoinTime: time.Now(),
	}
}

// NewTestPlayerWithBalance creates a test player with custom balance
func NewTestPlayerWithBalance(playerID int64, balance int64) *game.Player {
	player := NewTestPlayer(playerID)
	player.Balance = balance
	return player
}

// NewTestWallet creates a test wallet with default values
func NewTestWallet(walletID uint, userID uint) *wallet.Wallet {
	return &wallet.Wallet{
		ID:       walletID,
		UserID:   userID,
		Balance:  1000.00,
		Currency: "CNY",
		Status:   1,
	}
}

// NewTestWalletWithBalance creates a test wallet with custom balance
func NewTestWalletWithBalance(walletID uint, userID uint, balance float64) *wallet.Wallet {
	w := NewTestWallet(walletID, userID)
	w.Balance = balance
	return w
}

// NewTestFish creates a test fish with default values
func NewTestFish(fishID int64, fishType *game.FishType) *game.Fish {
	return &game.Fish{
		ID:        fishID,
		Type:      *fishType,
		Position:  game.Position{X: 500, Y: 400},
		Direction: 0,
		Speed:     fishType.BaseSpeed,
		Health:    fishType.BaseHealth,
		MaxHealth: fishType.BaseHealth,
		Value:     fishType.BaseValue,
		SpawnTime: time.Now(),
		Status:    game.FishStatusAlive,
	}
}

// NewTestBullet creates a test bullet with default values
func NewTestBullet(bulletID int64, playerID int64, power int32, cost int64) *game.Bullet {
	return &game.Bullet{
		ID:        bulletID,
		PlayerID:  playerID,
		Position:  game.Position{X: 600, Y: 750},
		Direction: -1.57, // Pointing up (in radians)
		Speed:     500,
		Power:     power,
		Cost:      cost,
		CreatedAt: time.Now(),
		Status:    game.BulletStatusFlying,
	}
}

// NewTestInventory creates a test inventory with custom RTP values
func NewTestInventory(inventoryID string, totalIn int64, totalOut int64) *game.Inventory {
	currentRTP := 0.0
	if totalIn > 0 {
		currentRTP = float64(totalOut) / float64(totalIn)
	}

	return &game.Inventory{
		ID:         inventoryID,
		TotalIn:    totalIn,
		TotalOut:   totalOut,
		CurrentRTP: currentRTP,
		UpdatedAt:  time.Now(),
	}
}

// FishTypeFixtures provides common fish type configurations for testing
type FishTypeFixtures struct {
	SmallFish  *game.FishType
	MediumFish *game.FishType
	LargeFish  *game.FishType
	BossFish   *game.FishType
}

// NewFishTypeFixtures creates standard fish type fixtures
func NewFishTypeFixtures() *FishTypeFixtures {
	return &FishTypeFixtures{
		SmallFish: &game.FishType{
			ID:          1,
			Name:        "Small Fish",
			Size:        "small",
			BaseHealth:  1,
			BaseValue:   10,
			BaseSpeed:   50,
			Rarity:      0.6,
			HitRate:     0.8,
			Description: "Common small fish",
		},
		MediumFish: &game.FishType{
			ID:          2,
			Name:        "Medium Fish",
			Size:        "medium",
			BaseHealth:  3,
			BaseValue:   50,
			BaseSpeed:   40,
			Rarity:      0.3,
			HitRate:     0.6,
			Description: "Medium-sized fish",
		},
		LargeFish: &game.FishType{
			ID:          3,
			Name:        "Large Fish",
			Size:        "large",
			BaseHealth:  10,
			BaseValue:   200,
			BaseSpeed:   30,
			Rarity:      0.09,
			HitRate:     0.4,
			Description: "Rare large fish",
		},
		BossFish: &game.FishType{
			ID:          4,
			Name:        "Boss Fish",
			Size:        "boss",
			BaseHealth:  50,
			BaseValue:   1000,
			BaseSpeed:   20,
			Rarity:      0.01,
			HitRate:     0.2,
			Description: "Epic boss fish",
		},
	}
}

// AllFishTypes returns all fish types as a slice
func (f *FishTypeFixtures) AllFishTypes() []*game.FishType {
	return []*game.FishType{
		f.SmallFish,
		f.MediumFish,
		f.LargeFish,
		f.BossFish,
	}
}
