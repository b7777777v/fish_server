// internal/biz/wire.go
package biz

import (
	"github.com/b7777777v/fish_server/internal/biz/account"
	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/biz/lobby"
	"github.com/b7777777v/fish_server/internal/biz/player"
	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/pkg/token"

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

	// Lobby module providers
	lobby.NewLobbyUsecase,
	// Bind TokenService interface from account package to token.TokenHelper
	wire.Bind(new(account.TokenService), new(*token.TokenHelper)),
)
