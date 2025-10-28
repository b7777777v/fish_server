// internal/pkg/token/wire.go
package token

import "github.com/google/wire"

// ProviderSet is a provider set for token helper.
var ProviderSet = wire.NewSet(NewTokenHelper)
