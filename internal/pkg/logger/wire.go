package logger

import (
	"github.com/google/wire"
)

// ProviderSet is a Wire provider set for the logger package.
// It tells Wire how to create a Config object.
// We export this so that other packages can use it.
var ProviderSet = wire.NewSet(NewLogger)
