package logger

import (
	"context"
	application "github.com/debugger84/modulus-application"
	"go.uber.org/zap"
)

type applicationLogger struct {
	zapLogger *zap.Logger
}

func NewApplicationLogger(zapLogger *zap.Logger) application.Logger {
	return &applicationLogger{zapLogger: zapLogger}
}

func (a *applicationLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	a.zapLogger.Sugar().Warnw(s, i...)
}

func (a *applicationLogger) Info(ctx context.Context, s string, i ...interface{}) {
	a.zapLogger.Sugar().Infow(s, i...)
}

func (a *applicationLogger) Error(ctx context.Context, s string, i ...interface{}) {
	a.zapLogger.Sugar().Errorw(s, i...)
}

func (a *applicationLogger) Debug(ctx context.Context, s string, i ...interface{}) {
	a.zapLogger.Sugar().Debugw(s, i...)
}

func (a *applicationLogger) Panic(ctx context.Context, s string, i ...interface{}) {
	a.zapLogger.Sugar().Panicw(s, i...)
}
