// internal/pkg/logger/logger.go
package logger

import (
	"os"

	"github.com/b7777777v/fish_server/internal/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 是一個接口，定義了我們需要的日誌方法
// 業務程式碼應該依賴這個接口，而不是具體的 Zap Logger
type Logger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
}

// NewLogger 是一個構造函數，根據配置創建 Logger 實例
// 這是 Wire 將會調用的 Provider
func NewLogger(c *conf.Log) (Logger, error) {
	var level zapcore.Level
	// 解析配置中的日誌級別
	if err := level.UnmarshalText([]byte(c.Level)); err != nil {
		return nil, err
	}

	// 創建 Zap 的配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder, // ISO8601 格式
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	// 創建一個 Zap Core
	core := zapcore.NewCore(
		// 根據配置選擇 Encoder
		buildEncoder(c.Format, encoderConfig),
		// 寫入到標準輸出
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		// 設置日誌級別
		level,
	)

	// 構建 Logger，並添加 caller 信息
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// 返回 SugaredLogger，它提供了更方便的 key-value 風格的日誌方法 (例如 Infow)
	return logger.Sugar(), nil
}

// buildEncoder 根據配置格式返回對應的 Encoder
func buildEncoder(format string, config zapcore.EncoderConfig) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	// 默認返回 console 格式
	return zapcore.NewConsoleEncoder(config)
}
