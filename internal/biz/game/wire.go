package game

import (
	"github.com/google/wire"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// ========================================
// Wire 依賴注入配置
// ========================================

// ProviderSet 遊戲業務邏輯提供者集合
var ProviderSet = wire.NewSet(
	// 核心組件
	NewMathModelProvider,
	NewFishSpawnerProvider,
	NewRoomManagerProvider,
	
	// 用例
	NewGameUsecaseProvider,
)

// NewMathModelProvider 創建數學模型提供者
func NewMathModelProvider(logger logger.Logger) *MathModel {
	return NewMathModel(logger)
}

// NewFishSpawnerProvider 創建魚類生成器提供者
func NewFishSpawnerProvider(logger logger.Logger) *FishSpawner {
	return NewFishSpawner(logger)
}

// NewRoomManagerProvider 創建房間管理器提供者
func NewRoomManagerProvider(logger logger.Logger, spawner *FishSpawner, mathModel *MathModel) *RoomManager {
	return NewRoomManager(logger, spawner, mathModel)
}

// NewGameUsecaseProvider 創建遊戲用例提供者
// 注意：GameRepo 和 PlayerRepo 將由數據層提供
func NewGameUsecaseProvider(
	gameRepo GameRepo,
	playerRepo PlayerRepo,
	roomManager *RoomManager,
	spawner *FishSpawner,
	mathModel *MathModel,
	logger logger.Logger,
) *GameUsecase {
	return NewGameUsecase(gameRepo, playerRepo, roomManager, spawner, mathModel, logger)
}