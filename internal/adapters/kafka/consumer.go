package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
	"golang.org/x/sync/errgroup"

	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

var (
	ErrConsumerAlreadyStarted = errors.New("consumer already started")
	ErrConsumerNotStarted     = errors.New("consumer not started")
	ErrInvalidConfig          = errors.New("invalid kafka config")
)

type kafkaConsumer struct {
	reader  *kafka.Reader
	service ports.OrderService
	log     logger.Logger
	g       errgroup.Group
	started bool
	mu      sync.Mutex
	cancel  context.CancelFunc
}

func NewKafkaConsumer(reader *kafka.Reader, service ports.OrderService, log logger.Logger) ports.KafkaConsumer {
	return &kafkaConsumer{
		reader:  reader,
		service: service,
		log:     log,
		started: false,
	}
}

func (c *kafkaConsumer) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.started {
		c.mu.Unlock()
		return ErrConsumerAlreadyStarted
	}
	c.started = true

	consumerCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel
	c.mu.Unlock()

	c.g.Go(func() error {
		defer func() {
			if r := recover(); r != nil {
				c.log.Error("panic in consumer goroutine", "panic", r)
			}
		}()

		c.log.Info("starting kafka consumer", "topic", c.reader.Config().Topic, "group_id", c.reader.Config().GroupID)

		for {
			select {
			case <-consumerCtx.Done():
				c.log.Info("stopping kafka consumer due to context cancellation")
				return nil
			default:
				msg, err := c.reader.ReadMessage(consumerCtx)
				if err != nil {
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						c.log.Info("consumer stopped due to context", "error", err)
						return nil
					}
					c.log.Error("failed to read message from kafka", "error", err)
					continue
				}

				keyHex := fmt.Sprintf("%x", msg.Key)
				c.log.Debug("received message from kafka", "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset, "key", keyHex)

				if err := c.service.ProcessMessage(consumerCtx, msg.Value); err != nil {
					c.log.Error("failed to process message", "offset", msg.Offset, "error", err)
					continue
				}

				if err := c.reader.CommitMessages(consumerCtx, msg); err != nil {
					c.log.Error("failed to commit message", "offset", msg.Offset, "error", err)
				} else {
					c.log.Debug("message committed", "offset", msg.Offset)
				}
			}
		}
	})

	return nil
}

func (c *kafkaConsumer) Stop(ctx context.Context) error {
	c.mu.Lock()
	if !c.started {
		c.mu.Unlock()
		return ErrConsumerNotStarted
	}
	c.started = false

	if c.cancel != nil {
		c.cancel()
	}
	c.mu.Unlock()

	c.log.Info("stopping kafka consumer")

	if err := c.reader.Close(); err != nil {
		c.log.Error("failed to close kafka reader", "error", err)
		return fmt.Errorf("close reader: %w", err)
	}

	if err := c.g.Wait(); err != nil {
		c.log.Error("consumer goroutine failed", "error", err)
		return fmt.Errorf("consumer goroutine error: %w", err)
	}

	c.log.Info("kafka consumer stopped successfully")
	return nil
}
