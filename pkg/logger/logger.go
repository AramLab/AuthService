package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const tsKey = "timestamp"

func NewLogger(level string) (*zap.SugaredLogger, error) {
	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level %q, err: %w", logLevel, err)
	}
	logger, err := zap.Config{
		Level:       logLevel,
		Encoding:    "json",
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",
			TimeKey:    tsKey,
			EncodeTime: zapcore.RFC3339TimeEncoder,
		},
		DisableStacktrace: true,
	}.Build()

	if err != nil {
		return nil, fmt.Errorf("error building logger config: %w", err)
	}

	return logger.Sugar(), nil
}
