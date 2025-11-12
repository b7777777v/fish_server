// internal/pkg/token/wire.go
package token

import (
	"github.com/b7777777v/fish_server/internal/biz/account"
	"github.com/google/wire"
)

// ProviderSet is a provider set for token helper.
var ProviderSet = wire.NewSet(
	NewTokenHelper,
	wire.Bind(new(account.TokenService), new(*TokenHelper)),
)
