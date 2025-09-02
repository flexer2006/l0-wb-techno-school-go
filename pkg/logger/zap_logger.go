package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const TraceIDKey = "trace_id"

type ZapLogger struct {
	Logger  *zap.Logger
	Sugared *zap.SugaredLogger
}

type Logger interface {
	Debug(msg string, fields ...any)
	Info(msg string, fields ...any)
	Warn(msg string, fields ...any)
	Error(msg string, fields ...any)
	Fatal(msg string, fields ...any)
	WithField(key string, value any) Logger
	WithContext(ctx context.Context) Logger
}

func ParseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func ParseTimeEncoder(enc string) zapcore.TimeEncoder {
	if enc == "iso8601" {
		return zapcore.ISO8601TimeEncoder
	}
	return zapcore.EpochTimeEncoder
}

func ParseLevelEncoder(enc string) zapcore.LevelEncoder {
	if enc == "lower" {
		return zapcore.LowercaseLevelEncoder
	}
	return zapcore.CapitalLevelEncoder
}

func (z *ZapLogger) Debug(msg string, fields ...any) {
	z.Sugared.Debugw(msg, fields...)
}

func (z *ZapLogger) Info(msg string, fields ...any) {
	z.Sugared.Infow(msg, fields...)
}

func (z *ZapLogger) Warn(msg string, fields ...any) {
	z.Sugared.Warnw(msg, fields...)
}

func (z *ZapLogger) Error(msg string, fields ...any) {
	z.Sugared.Errorw(msg, fields...)
}

func (z *ZapLogger) Fatal(msg string, fields ...any) {
	z.Sugared.Fatalw(msg, fields...)
}

func (z *ZapLogger) WithField(key string, value any) Logger {
	newLogger := z.Logger.With(zap.Any(key, value))
	return &ZapLogger{
		Logger:  newLogger,
		Sugared: newLogger.Sugar(),
	}
}

func (z *ZapLogger) WithContext(ctx context.Context) Logger {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		newLogger := z.Logger.With(zap.String(TraceIDKey, traceID))
		return &ZapLogger{
			Logger:  newLogger,
			Sugared: newLogger.Sugar(),
		}
	}
	return z
}
