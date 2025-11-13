// internal/biz/wire.go
package biz

import (
	"context"

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
	// 創建一個包裝器，將 CreateWallet 方法適配為 account.WalletCreator 介面
	wrapper := &walletUsecaseWrapper{uc: uc}
	return account.NewWalletCreatorFromUsecase(wrapper)
}

// walletUsecaseWrapper 包裝 WalletUsecase 以符合 account.WalletCreator 介面
type walletUsecaseWrapper struct {
	uc *wallet.WalletUsecase
}

func (w *walletUsecaseWrapper) CreateWallet(ctx context.Context, userID uint, currency string) error {
	_, err := w.uc.CreateWallet(ctx, userID, currency)
	return err
}
