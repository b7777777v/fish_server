package admin

import (
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// AdminService 管理後台服務
type AdminService struct {
	playerUC *player.PlayerUsecase
	walletUC *wallet.WalletUsecase
	logger   logger.Logger
}

// NewAdminService 創建一個新的 AdminService 實例
func NewAdminService(
	playerUC *player.PlayerUsecase,
	walletUC *wallet.WalletUsecase,
	logger logger.Logger,
) *AdminService {
	return &AdminService{
		playerUC: playerUC,
		walletUC: walletUC,
		logger:   logger.With("module", "app/admin"),
	}
}

// GetPlayerUsecase 獲取玩家用例
func (s *AdminService) GetPlayerUsecase() *player.PlayerUsecase {
	return s.playerUC
}

// GetWalletUsecase 獲取錢包用例
func (s *AdminService) GetWalletUsecase() *wallet.WalletUsecase {
	return s.walletUC
}

// GetLogger 獲取日誌記錄器
func (s *AdminService) GetLogger() logger.Logger {
	return s.logger
}