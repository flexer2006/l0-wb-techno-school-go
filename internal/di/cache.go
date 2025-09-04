package di

import (
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/cache"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

func NewCache(log logger.Logger) ports.Cache {
	return cache.NewInMemoryCache(log)
}
