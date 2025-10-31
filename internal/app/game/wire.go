package game

import (
	"github.com/google/wire"
)

// ========================================
// Wire 依賴注入配置
// ========================================

// ProviderSet 遊戲應用層提供者集合
var ProviderSet = wire.NewSet(
	// WebSocket 相關組件
	NewHub,
	NewWebSocketHandler,
	NewMessageHandler,
	
	// 遊戲應用
	NewGameApp,
)

