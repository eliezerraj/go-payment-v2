package webserver

import (
	"os"
	"context"
	"time"
	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-payment-v2/application/infrastructure/application"
	"github.com/go-payment-v2/cmd/webserver/framework/fiber"
	"github.com/go-payment-v2/application/config"
)

type WebServer struct {
	cfg	*config.Config
	fiberServer *fiber.FiberServer
}

func NewWebServer(cfg *config.Config) *WebServer {
	logger.InfoOutCtx("initializing webserver SUCCESSFULLY")

	_, cancel := context.WithTimeout(context.Background(), cfg.Database.ConnTimeout)
	defer cancel()

	application, err := application.NewApplication(cfg)
	if err != nil {
		logger.FatalOutCtx("failed to initialize application", zap.Error(err))
		os.Exit(1)
	}

	fiberServer := fiber.NewFiberServer(cfg)
	fiberServer.SetupRoutes(application)
	return &WebServer{
		cfg: cfg,
		fiberServer: fiberServer,
	}
}

func (s *WebServer) Run() {
	logger.InfoOutCtx("starting fiber server on port: " + s.cfg.HTTP.Port)

	if err := s.fiberServer.FiberApp.Listen(":" + s.cfg.HTTP.Port); err != nil {
		logger.FatalOutCtx("failed to start HTTP server", zap.Error(err))
	}
}

func (s *WebServer) Shutdown() {
	logger.InfoOutCtx("webserver is shutting down SUCCESSFULLY")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.fiberServer.FiberApp.Shutdown(); err != nil {
		logger.Error(ctx, "failed to shutdown HTTP server", zap.Error(err))
	}
}
