package logger

import (
	"errors"
	application "github.com/debugger84/modulus-application"
	"github.com/debugger84/modulus-logger-zap/encoder"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

const (
	DebugLevel   = "debug"
	InfoLevel    = "info"
	WarningLevel = "warn"
	ErrorLevel   = "error"
	FatalLevel   = "fatal"

	defaultLevel = WarningLevel
)

const (
	ConsoleType = "console"
	JsonType    = "json"

	defaultType = JsonType
)

type ModuleConfig struct {
	app           string
	level         string
	loggerType    string
	encoderConfig *encoder.Config
}

// NewModuleConfig set nil to encoderConfig to use default encoding settings
func NewModuleConfig(encoderConfig *encoder.Config) *ModuleConfig {
	return &ModuleConfig{encoderConfig: encoderConfig}
}

func (l *ModuleConfig) EncoderConfig() encoder.Config {
	if l.encoderConfig == nil {
		l.encoderConfig = &encoder.Config{
			TimeKey:        "datetime",
			LevelKey:       "level_name",
			LevelIntKey:    "level",
			EnvKey:         "env",
			CallerKey:      "script_name",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			FieldsGroupKey: "context",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
		}
	}
	return *l.encoderConfig
}

func (l *ModuleConfig) SetAppNam(appName string) error {
	l.app = appName

	return nil
}

func (l *ModuleConfig) SetLevel(level string) error {
	l.level = level
	if level != DebugLevel && level != InfoLevel && level != WarningLevel &&
		level != ErrorLevel && level != FatalLevel {
		return errors.New("wrong level value")
	}

	return nil
}

func (l *ModuleConfig) SetLoggerType(loggerType string) error {
	l.loggerType = loggerType
	if loggerType != ConsoleType && loggerType != JsonType {
		return errors.New("wrong logger type")
	}

	return nil
}

func (l *ModuleConfig) IsDebug() bool {
	return l.level == DebugLevel
}

func (s *ModuleConfig) ProvidedServices() []interface{} {
	return []interface{}{
		NewZapLogger,
		NewApplicationLogger,
		func() *ModuleConfig {
			return s
		},
	}
}

func (s *ModuleConfig) InitConfig(config application.Config) error {
	if s.app == "" {
		s.app = config.GetEnv("LOGGER_APP")
	}
	if s.level == "" {
		val := config.GetEnv("LOGGER_LEVEL")
		err := s.SetLevel(val)
		if err != nil {
			s.level = defaultLevel
		}
	}

	if s.loggerType == "" {
		val := config.GetEnv("LOGGER_TYPE")
		err := s.SetLoggerType(val)
		if err != nil {
			s.loggerType = defaultType
		}
	}
	s.EncoderConfig()

	return nil
}

type ConfigBuilder interface {
	Build(opts ...zap.Option) (*zap.Logger, error)
}

type Config struct {
	// Level is the minimum enabled logging level. Note that this is a dynamic
	Level zap.AtomicLevel `json:"level" yaml:"level"`

	// Sampling sets a sampling policy. A nil SamplingConfig disables sampling.
	Sampling *zap.SamplingConfig `json:"sampling" yaml:"sampling"`

	// EncoderConfig sets options for the chosen encoder.
	EncoderConfig encoder.Config `json:"encoderConfig" yaml:"encoderConfig"`

	// OutputPaths is a list of URLs or file paths to write logging output to.
	OutputPaths []string `json:"outputPaths" yaml:"outputPaths"`

	// ErrorOutputPaths is a list of URLs to write internal logger errors to.
	ErrorOutputPaths []string `json:"errorOutputPaths" yaml:"errorOutputPaths"`

	// InitialFields is a collection of fields to add to the root logger.
	InitialFields []zapcore.Field `json:"initialFields" yaml:"initialFields"`

	EncoderConstructor func(interface{}) (zapcore.Encoder, error)
}

// Build constructs a logger from the Config and Options.
func (cfg Config) Build(opts ...zap.Option) (*zap.Logger, error) {
	enc, err := cfg.EncoderConstructor(&cfg.EncoderConfig)
	if err != nil {
		return nil, err
	}

	sink, errSink, err := cfg.openSinks()
	if err != nil {
		return nil, err
	}

	log := zap.New(
		zapcore.NewCore(enc, sink, cfg.Level),
		cfg.buildOptions(errSink)...,
	)
	if len(opts) > 0 {
		log = log.WithOptions(opts...)
	}
	return log, nil
}

func (cfg Config) buildOptions(errSink zapcore.WriteSyncer) []zap.Option {
	opts := []zap.Option{zap.ErrorOutput(errSink)}

	opts = append(opts, zap.AddCaller())

	stackLevel := zap.ErrorLevel
	opts = append(opts, zap.AddStacktrace(stackLevel))

	if cfg.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, int(cfg.Sampling.Initial), int(cfg.Sampling.Thereafter))
		}))
	}

	if len(cfg.InitialFields) > 0 {
		opts = append(opts, zap.Fields(cfg.InitialFields...))
	}

	return opts
}

func (cfg Config) openSinks() (zapcore.WriteSyncer, zapcore.WriteSyncer, error) {
	sink, closeOut, err := zap.Open(cfg.OutputPaths...)
	if err != nil {
		return nil, nil, err
	}
	errSink, _, err := zap.Open(cfg.ErrorOutputPaths...)
	if err != nil {
		closeOut()
		return nil, nil, err
	}
	return sink, errSink, nil
}
