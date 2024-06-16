package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func LoggerInitializer(level string) (*zap.SugaredLogger, error) {
	cfg := zap.NewProductionConfig()
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("attempt to parse logger level failed - %v", err)
	}
	cfg.Level = lvl
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("attempt to initialized logger failed - %v", err)
	}
	defer logger.Sync()
	return logger.Sugar(), nil
}
