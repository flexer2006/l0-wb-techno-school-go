package logger

import (
	"context"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const TraceIDKey = "trace_id"

type ZapLogger struct {
	logger  *zap.Logger
	sugared *zap.SugaredLogger
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

func NewZapLogger() *ZapLogger {
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(parseLevel(viper.GetString("logger.level"))),
		Development: viper.GetBool("logger.development"),
		Encoding:    viper.GetString("logger.encoding"),
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      viper.GetString("logger.encoder.time_key"),
			LevelKey:     viper.GetString("logger.encoder.level_key"),
			MessageKey:   viper.GetString("logger.encoder.message_key"),
			CallerKey:    viper.GetString("logger.encoder.caller_key"),
			EncodeTime:   parseTimeEncoder(viper.GetString("logger.encoder.time_encoder")),
			EncodeLevel:  parseLevelEncoder(viper.GetString("logger.encoder.level_encoder")),
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
		OutputPaths:      viper.GetStringSlice("logger.output_paths"),
		ErrorOutputPaths: viper.GetStringSlice("logger.error_output_paths"),
	}
	logger := zap.Must(config.Build())

	return &ZapLogger{
		logger:  logger,
		sugared: logger.Sugar(),
	}
}

func parseLevel(level string) zapcore.Level {
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

func parseTimeEncoder(enc string) zapcore.TimeEncoder {
	if enc == "iso8601" {
		return zapcore.ISO8601TimeEncoder
	}
	return zapcore.EpochTimeEncoder
}

func parseLevelEncoder(enc string) zapcore.LevelEncoder {
	if enc == "lower" {
		return zapcore.LowercaseLevelEncoder
	}
	return zapcore.CapitalLevelEncoder
}

func (z *ZapLogger) Debug(msg string, fields ...any) {
	z.sugared.Debugw(msg, fields...)
}

func (z *ZapLogger) Info(msg string, fields ...any) {
	z.sugared.Infow(msg, fields...)
}

func (z *ZapLogger) Warn(msg string, fields ...any) {
	z.sugared.Warnw(msg, fields...)
}

func (z *ZapLogger) Error(msg string, fields ...any) {
	z.sugared.Errorw(msg, fields...)
}

func (z *ZapLogger) Fatal(msg string, fields ...any) {
	z.sugared.Fatalw(msg, fields...)
}

func (z *ZapLogger) WithField(key string, value any) Logger {
	newLogger := z.logger.With(zap.Any(key, value))
	return &ZapLogger{
		logger:  newLogger,
		sugared: newLogger.Sugar(),
	}
}

func (z *ZapLogger) WithContext(ctx context.Context) Logger {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		newLogger := z.logger.With(zap.String(TraceIDKey, traceID))
		return &ZapLogger{
			logger:  newLogger,
			sugared: newLogger.Sugar(),
		}
	}
	return z
}
