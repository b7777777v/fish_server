// internal/biz/wire.go
package biz

import (
	"github.com/b7777777v/fish_server/internal/biz/account"
	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/lobby"
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/biz/wallet"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	game.ProviderSet,
	player.NewPlayerUsecase,
	wallet.NewWalletUsecase,

	// Account module providers
	account.NewAccountUsecase,
	account.NewOAuthService,
	ProvideWalletCreator, // 提供 WalletCreator 給 AccountUsecase

	// Lobby module providers
	lobby.NewLobbyUsecase,
)

// ProvideWalletCreator 將 WalletUsecase 轉換為 WalletCreator 介面
func ProvideWalletCreator(uc *wallet.WalletUsecase) account.WalletCreator {
	return account.NewWalletCreatorFromUsecase(uc)
}
