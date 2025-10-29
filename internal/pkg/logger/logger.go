package logger

import (
	"io"
	"os"

	"github.com/b7777777v/fish_server/internal/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 為了方便，我們直接使用 SugaredLogger
type Logger = *zap.SugaredLogger

// New 是一個簡易的構造函數，用於在 App 初始化前創建臨時 Logger
func New(writer io.Writer, level, format string) Logger {
	var lvl zapcore.Level
	_ = lvl.UnmarshalText([]byte(level))

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	var encoder zapcore.Encoder
	if format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(writer),
		lvl,
	)
	return zap.New(core).Sugar()
}

// NewLogger 是 Wire 將會調用的 Provider
func NewLogger(c *conf.Log) (Logger, func(), error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(c.Level)); err != nil {
		return nil, nil, err
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	core := zapcore.NewCore(
		buildEncoder(c.Format, encoderConfig),
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugaredLogger := logger.Sugar()
	cleanup := func() {
		_ = sugaredLogger.Sync()
	}
	return sugaredLogger, cleanup, nil
}

func buildEncoder(format string, config zapcore.EncoderConfig) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	return zapcore.NewConsoleEncoder(config)
}
