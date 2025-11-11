package admin

import (
	"github.com/google/wire"
)

// ProviderSet 是 admin 模組的 wire provider set
var ProviderSet = wire.NewSet(
	NewAdminService,
	NewServer,
	NewAdminApp,

	// Handlers
	NewAccountHandler,
	NewLobbyHandler,
)