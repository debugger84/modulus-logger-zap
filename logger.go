package logger

import (
	"github.com/debugger84/modulus-logger-zap/encoder"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(cfg *ModuleConfig) *zap.Logger {
	sampling := &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}

	var loggerConfig ConfigBuilder
	if cfg.loggerType == ConsoleType {
		loggerConfig = consoleConfig(cfg, sampling)
	} else {
		loggerConfig = jsonConfig(cfg, sampling)
	}

	logger, _ := loggerConfig.Build()
	return logger
}

func jsonConfig(cfg *ModuleConfig, sampling *zap.SamplingConfig) Config {

	initialFields := []zapcore.Field{
		zap.String("app", cfg.app),
	}

	return Config{
		Level:            GetAtomicLevel(cfg.level),
		Sampling:         sampling,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    initialFields,
		EncoderConfig:    cfg.EncoderConfig(),
		EncoderConstructor: func(encoderConfig interface{}) (zapcore.Encoder, error) {
			enc := encoder.NewEncoder(encoderConfig)
			return enc, nil
		},
	}
}

func consoleConfig(cfg *ModuleConfig, sampling *zap.SamplingConfig) zap.Config {

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoderConfig.LineEnding = "\n\n"

	return zap.Config{
		Level:            GetAtomicLevel(cfg.level),
		Development:      true,
		Sampling:         sampling,
		Encoding:         "console",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func GetAtomicLevel(level string) zap.AtomicLevel {

	var atomicLevel zap.AtomicLevel
	if err := atomicLevel.UnmarshalText([]byte(level)); err != nil {
		atomicLevel.SetLevel(zap.DebugLevel)
	}

	return atomicLevel
}
