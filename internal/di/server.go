package di

import (
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/server"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/adapters/server/handlers"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/config"
	"github.com/flexer2006/l0-wb-techno-school-go/internal/ports"
	"github.com/flexer2006/l0-wb-techno-school-go/pkg/logger"
)

func NewHTTPServer(
	cache ports.Cache,
	repo ports.OrderRepository,
	log logger.Logger,
	cfg config.ServerConfig,
) ports.HTTPServer {
	httpSrv := server.NewHTTPServer(log, cfg)
	orderHandler := handlers.OrderHandler(cache, repo, log)
	httpSrv.RegisterRoutes(orderHandler)
	return httpSrv
}
