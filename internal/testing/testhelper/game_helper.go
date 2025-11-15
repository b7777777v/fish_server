// Package testhelper provides helper functions and utilities for testing
package testhelper

import (
	"context"
	"os"
	"testing"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/testing/mocks"
	"github.com/stretchr/testify/mock"
)

// GameTestEnv represents a complete test environment for game testing
type GameTestEnv struct {
	Ctx context.Context
	Log logger.Logger

	// Mocked Repositories
	GameRepo      *mocks.GameRepo
	PlayerRepo    *mocks.PlayerRepo
	WalletRepo    *mocks.WalletRepo
	InventoryRepo *mocks.InventoryRepo

	// Business Logic Components
	WalletUsecase    *wallet.WalletUsecase
	Spawner          *game.FishSpawner
	MathModel        *game.MathModel
	InventoryManager *game.InventoryManager
	RTPController    *game.RTPController
	RoomManager      *game.RoomManager
	GameUsecase      *game.GameUsecase

	// Test Configuration
	RoomConfig game.RoomConfig
}

// GameTestEnvOptions configures the test environment setup
type GameTestEnvOptions struct {
	// LogLevel specifies the log level (default: "debug")
	LogLevel string
	// RoomConfig provides custom room configuration (optional)
	RoomConfig *game.RoomConfig
	// SkipDefaultMocks skips setting up default mock expectations
	SkipDefaultMocks bool
}

// NewGameTestEnv creates a new test environment with all necessary mocks and dependencies
func NewGameTestEnv(t *testing.T, opts *GameTestEnvOptions) *GameTestEnv {
	if opts == nil {
		opts = &GameTestEnvOptions{}
	}

	// Set defaults
	if opts.LogLevel == "" {
		opts.LogLevel = "debug"
	}

	// Create logger
	log := logger.New(os.Stdout, opts.LogLevel, "console")

	// Create mocks
	gameRepo := new(mocks.GameRepo)
	playerRepo := new(mocks.PlayerRepo)
	walletRepo := new(mocks.WalletRepo)
	inventoryRepo := new(mocks.InventoryRepo)

	// Setup default mock behavior if not skipped
	if !opts.SkipDefaultMocks {
		setupDefaultMocks(gameRepo, playerRepo, walletRepo, inventoryRepo)
	}

	// Create wallet usecase
	walletUsecase := wallet.NewWalletUsecase(walletRepo, log)

	// Create room configuration
	roomConfig := DefaultRoomConfig()
	if opts.RoomConfig != nil {
		roomConfig = *opts.RoomConfig
	}

	// Create game components
	spawner := game.NewFishSpawner(log, roomConfig)
	mathModel := game.NewMathModel(log)
	inventoryManager, err := game.NewInventoryManager(inventoryRepo, log)
	if err != nil {
		t.Fatalf("Failed to create inventory manager: %v", err)
	}

	rtpController := game.NewRTPController(inventoryManager, log)
	roomManager := game.NewRoomManager(log, spawner, mathModel, inventoryManager, rtpController)
	gameUsecase := game.NewGameUsecase(
		gameRepo,
		playerRepo,
		walletUsecase,
		roomManager,
		spawner,
		mathModel,
		inventoryManager,
		rtpController,
		log,
	)

	return &GameTestEnv{
		Ctx:              context.Background(),
		Log:              log,
		GameRepo:         gameRepo,
		PlayerRepo:       playerRepo,
		WalletRepo:       walletRepo,
		InventoryRepo:    inventoryRepo,
		WalletUsecase:    walletUsecase,
		Spawner:          spawner,
		MathModel:        mathModel,
		InventoryManager: inventoryManager,
		RTPController:    rtpController,
		RoomManager:      roomManager,
		GameUsecase:      gameUsecase,
		RoomConfig:       roomConfig,
	}
}

// AssertExpectations verifies all mock expectations
func (env *GameTestEnv) AssertExpectations(t *testing.T) {
	env.GameRepo.AssertExpectations(t)
	env.PlayerRepo.AssertExpectations(t)
	env.WalletRepo.AssertExpectations(t)
	env.InventoryRepo.AssertExpectations(t)
}

// DefaultRoomConfig returns a default room configuration for testing
func DefaultRoomConfig() game.RoomConfig {
	return game.RoomConfig{
		MaxPlayers:           4,
		MinBet:               1,
		MaxBet:               100,
		BulletCostMultiplier: 1.0,
		FishSpawnRate:        0.3,
		MinFishCount:         10,
		MaxFishCount:         20,
		RoomWidth:            1200,
		RoomHeight:           800,
		TargetRTP:            0.96,
	}
}

// setupDefaultMocks sets up default behavior for all mocks
func setupDefaultMocks(gameRepo *mocks.GameRepo, playerRepo *mocks.PlayerRepo, walletRepo *mocks.WalletRepo, inventoryRepo *mocks.InventoryRepo) {
	// GameRepo defaults
	gameRepo.On("SaveRoom", mock.Anything, mock.Anything).Return(nil).Maybe()
	gameRepo.On("DeleteRoom", mock.Anything, mock.Anything).Return(nil).Maybe()
	gameRepo.On("SaveGameStatistics", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	gameRepo.On("SaveGameEvent", mock.Anything, mock.Anything).Return(nil).Maybe()
	gameRepo.On("SaveFishTypeCache", mock.Anything, mock.Anything).Return(nil).Maybe()
	gameRepo.On("GetAllFishTypes", mock.Anything).Return([]*game.FishType{
		{ID: 1, Name: "Small Fish", Size: "small", BaseHealth: 1, BaseValue: 10, BaseSpeed: 50, Rarity: 0.6, HitRate: 0.8},
		{ID: 2, Name: "Medium Fish", Size: "medium", BaseHealth: 3, BaseValue: 50, BaseSpeed: 40, Rarity: 0.3, HitRate: 0.6},
		{ID: 3, Name: "Large Fish", Size: "large", BaseHealth: 10, BaseValue: 200, BaseSpeed: 30, Rarity: 0.09, HitRate: 0.4},
		{ID: 4, Name: "Boss Fish", Size: "boss", BaseHealth: 50, BaseValue: 1000, BaseSpeed: 20, Rarity: 0.01, HitRate: 0.2},
	}, nil).Maybe()
	gameRepo.On("ListRooms", mock.Anything, mock.Anything).Return([]*game.Room{}, nil).Maybe()
	gameRepo.On("GetGameStatistics", mock.Anything, mock.Anything).Return(&game.GameStatistics{}, nil).Maybe()
	gameRepo.On("GetGameEvents", mock.Anything, mock.Anything, mock.Anything).Return([]*game.GameEvent{}, nil).Maybe()

	// PlayerRepo defaults
	playerRepo.On("GetPlayer", mock.Anything, mock.Anything).Return(func(ctx context.Context, playerID int64) *game.Player {
		return &game.Player{
			ID:       playerID,
			UserID:   playerID,
			Nickname: "TestPlayer",
			Balance:  100000,
			WalletID: 1,
			Status:   game.PlayerStatusIdle,
		}
	}, nil).Maybe()
	playerRepo.On("UpdatePlayerBalance", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	playerRepo.On("UpdatePlayerStatus", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	// WalletRepo defaults
	walletRepo.On("FindByID", mock.Anything, mock.Anything).Return(func(ctx context.Context, id uint) *wallet.Wallet {
		return &wallet.Wallet{ID: id, UserID: id, Balance: 1000.00, Currency: "CNY", Status: 1}
	}, nil).Maybe()
	walletRepo.On("FindByUserID", mock.Anything, mock.Anything, mock.Anything).Return(func(ctx context.Context, userID uint, currency string) *wallet.Wallet {
		return &wallet.Wallet{ID: 1, UserID: userID, Balance: 1000.00, Currency: currency, Status: 1}
	}, nil).Maybe()
	walletRepo.On("FindAllByUserID", mock.Anything, mock.Anything).Return([]*wallet.Wallet{
		{ID: 1, UserID: 1, Balance: 1000.00, Currency: "CNY", Status: 1},
	}, nil).Maybe()
	walletRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Maybe()
	walletRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Maybe()
	walletRepo.On("Deposit", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	walletRepo.On("Withdraw", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	walletRepo.On("CreateTransaction", mock.Anything, mock.Anything).Return(nil).Maybe()
	walletRepo.On("FindTransactionsByWalletID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]*wallet.Transaction{}, nil).Maybe()

	// InventoryRepo defaults - return a default inventory
	inventoryRepo.On("GetInventory", mock.Anything, mock.Anything).Return(func(ctx context.Context, inventoryID string) *game.Inventory {
		return &game.Inventory{
			ID:         inventoryID,
			TotalIn:    0,
			TotalOut:   0,
			CurrentRTP: 0.0,
		}
	}, nil).Maybe()
	inventoryRepo.On("SaveInventory", mock.Anything, mock.Anything).Return(nil).Maybe()
	inventoryRepo.On("GetAllInventories", mock.Anything).Return(map[string]*game.Inventory{}, nil).Maybe()
}
