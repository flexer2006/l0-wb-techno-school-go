package di

import (
	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(cfg config.LoggerConfig) logger.Logger {
	zapConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(logger.ParseLevel(cfg.Level)),
		Development: cfg.Development,
		Encoding:    cfg.Encoding,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      cfg.Encoder.TimeKey,
			LevelKey:     cfg.Encoder.LevelKey,
			MessageKey:   cfg.Encoder.MessageKey,
			CallerKey:    cfg.Encoder.CallerKey,
			EncodeTime:   logger.ParseTimeEncoder(cfg.Encoder.TimeEncoder),
			EncodeLevel:  logger.ParseLevelEncoder(cfg.Encoder.LevelEncoder),
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
		OutputPaths:      cfg.OutputPaths,
		ErrorOutputPaths: cfg.ErrorPaths,
	}

	zapLogger := zap.Must(zapConfig.Build())

	return &logger.ZapLogger{
		Logger:  zapLogger,
		Sugared: zapLogger.Sugar(),
	}
}
