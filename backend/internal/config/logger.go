package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a new Zap logger based on configuration.
func NewLogger(cfg *Config) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Log.Level)); err != nil {
		level = zapcore.DebugLevel
	}

	var config zap.Config
	if cfg.IsDevelopment() {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = zap.NewAtomicLevelAt(level)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// Replace global logger
	zap.ReplaceGlobals(logger)

	return logger, nil
}

// NewLoggerOrPanic creates a logger or panics on failure.
func NewLoggerOrPanic(cfg *Config) *zap.Logger {
	logger, err := NewLogger(cfg)
	if err != nil {
		// Fallback: write to stderr and exit
		_, _ = os.Stderr.WriteString("Failed to initialize logger: " + err.Error() + "\n")
		os.Exit(1)
	}
	return logger
}
