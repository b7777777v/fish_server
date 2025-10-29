// internal/conf/wire.go
package conf

import "github.com/google/wire"

// ProviderSet is conf providers.
var ProviderSet = wire.NewSet(
	// wire.FieldsOf 告訴 wire，Config 結構中的所有欄位都可以被當作 Provider。
	// 例如，當有地方需要 *Data 時，wire 會知道可以從 *Config 中取得。
	wire.FieldsOf(new(*Config), "Data", "Log", "JWT", "Server"),
)