// internal/data/wire.go
package data

import "github.com/google/wire"

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewPlayerRepo)
