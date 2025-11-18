package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/b7777777v/fish_server/internal/conf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
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

	encoder := buildEncoder(c.Format, encoderConfig)

	// 建立多個輸出目標
	var cores []zapcore.Core
	var filesToClose []io.Closer

	// 1. 總是輸出到控制台
	cores = append(cores, zapcore.NewCore(
		encoder,
		zapcore.AddSync(zapcore.Lock(os.Stdout)),
		level,
	))

	// 2. 如果配置了檔案路徑，也輸出到檔案（使用 lumberjack 實現按日期分割）
	if c.FilePath != "" {
		// 自動創建日誌文件所在的目錄
		dir := filepath.Dir(c.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// 使用 lumberjack 實現日誌輪轉
		// 按日期分割：每天自動創建新的日誌文件
		lumberjackLogger := &lumberjack.Logger{
			Filename:   c.FilePath,
			MaxSize:    100,  // 單個日誌文件最大 100 MB
			MaxBackups: 30,   // 保留最近 30 個日誌文件
			MaxAge:     30,   // 保留 30 天內的日誌
			Compress:   true, // 壓縮舊日誌文件
			LocalTime:  true, // 使用本地時間（而非 UTC）
		}
		filesToClose = append(filesToClose, lumberjackLogger)

		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(lumberjackLogger),
			level,
		))
	}

	// 組合多個 core
	core := zapcore.NewTee(cores...)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugaredLogger := logger.Sugar()
	cleanup := func() {
		_ = sugaredLogger.Sync()
		for _, f := range filesToClose {
			_ = f.Close()
		}
	}
	return sugaredLogger, cleanup, nil
}

func buildEncoder(format string, config zapcore.EncoderConfig) zapcore.Encoder {
	if format == "json" {
		return zapcore.NewJSONEncoder(config)
	}
	return zapcore.NewConsoleEncoder(config)
}
