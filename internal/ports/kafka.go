package ports

import "context"

type KafkaConsumer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
