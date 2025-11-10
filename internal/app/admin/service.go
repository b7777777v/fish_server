package admin

import (
	"github.com/b7777777v/fish_server/internal/app/game"
	gamebiz "github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/b7777777v/fish_server/internal/pkg/token"
)

// AdminService 管理後台服務
type AdminService struct {
	playerUC           *player.PlayerUsecase
	walletUC           *wallet.WalletUsecase
	gameApp            *game.GameApp
	formationConfigSvc *gamebiz.FormationConfigService // 陣型配置服務
	tokenHelper        *token.TokenHelper
	config             *conf.Config
	logger             logger.Logger

	// New handlers
	accountHandler *AccountHandler
	lobbyHandler   *LobbyHandler
}

// NewAdminService 創建一個新的 AdminService 實例
func NewAdminService(
	playerUC *player.PlayerUsecase,
	walletUC *wallet.WalletUsecase,
	gameApp *game.GameApp,
	formationConfigSvc *gamebiz.FormationConfigService, // 修正：使用正確的套件別名
	tokenHelper *token.TokenHelper,
	config *conf.Config,
	logger logger.Logger,
	accountHandler *AccountHandler,
	lobbyHandler *LobbyHandler,
) *AdminService {
	return &AdminService{
		playerUC:           playerUC,
		walletUC:           walletUC,
		gameApp:            gameApp,
		formationConfigSvc: formationConfigSvc, // 保存服務引用
		tokenHelper:        tokenHelper,
		config:             config,
		logger:             logger.With("module", "app/admin"),
		accountHandler:     accountHandler,
		lobbyHandler:       lobbyHandler,
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
