package di

import (
	kafkalocal "github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/kafka"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
	"github.com/segmentio/kafka-go"
)

func NewKafkaConsumer(cfg config.KafkaConfig, service ports.OrderService, log logger.Logger) ports.KafkaConsumer {
	var startOffset int64
	switch cfg.AutoOffsetReset {
	case "earliest":
		startOffset = kafka.FirstOffset
	case "latest":
		startOffset = kafka.LastOffset
	default:
		startOffset = kafka.LastOffset
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		Topic:          cfg.Topic,
		GroupID:        cfg.GroupID,
		StartOffset:    startOffset,
		MinBytes:       cfg.MinBytes,
		MaxBytes:       cfg.MaxBytes,
		MaxWait:        cfg.MaxWait,
		SessionTimeout: cfg.SessionTimeout,
		CommitInterval: cfg.CommitInterval,
	}

	reader := kafka.NewReader(readerConfig)
	return kafkalocal.NewKafkaConsumer(reader, service, log)
}
