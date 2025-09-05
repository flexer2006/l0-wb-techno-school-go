package ports

import (
	"context"
)

type OrderService interface {
	ProcessMessage(ctx context.Context, payload []byte) error
}
