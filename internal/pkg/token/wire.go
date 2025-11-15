// internal/pkg/token/wire.go
package token

import (
	"github.com/b7777777v/fish_server/internal/biz/account"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/google/wire"
)

// ProviderSet is a provider set for token helper.
var ProviderSet = wire.NewSet(
	ProvideTokenHelper,
	wire.Bind(new(account.TokenService), new(*TokenHelper)),
)

// ProvideTokenHelper 提供一個帶 Redis cache 的 TokenHelper
// 如果 TokenCache 可用，會自動注入
func ProvideTokenHelper(c *conf.JWT, cache TokenCache) *TokenHelper {
	return NewTokenHelperWithCache(c, cache)
}
