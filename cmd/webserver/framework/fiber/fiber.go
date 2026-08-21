package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/json-iterator/go"

	"github.com/gofiber/fiber/v2/middleware/compress"

	"github.com/go-payment-v2/cmd/webserver/framework/fiber/adapter"
	"github.com/go-payment-v2/application/config"
	"github.com/go-payment-v2/application/infrastructure/application"
	"github.com/go-payment-v2/cmd/webserver/framework/fiber/middleware"

	"github.com/eliezerraj/go-core/v3/logger"
)

// Create a new Server configuration.
type FiberServerConfig struct {
	fiber.Config
	Port string
}

func NewServerConfig(cfg *config.HTTP) FiberServerConfig {
	logger.InfoOutCtx("initializing fiber server config SUCCESSFULLY")

	return FiberServerConfig{
		Config: fiber.Config{
			JSONEncoder:           jsoniter.Marshal,
			JSONDecoder:           jsoniter.Unmarshal,
			DisableStartupMessage: cfg.DisableStartupMessage,
			ReadBufferSize:        cfg.ReadBufferSize,
			ReadTimeout:           cfg.ReadTimeout,
			WriteTimeout:          cfg.WriteTimeout,
			IdleTimeout:           cfg.IdleConnTimeout,
		},
		Port: cfg.Port,
	}
}

// httpAdapter is a struct that holds the metadata and application adapters for the Fiber server.
type httpAdapter struct {
	metadataAdp  	*adapter.MetadataAdapter
	applicationAdp   *adapter.ApplicationAdapter
}

// Create a new httpAdapter with the provided configuration and application.
func newAdapters(cfg *config.Config, application *application.Application) *httpAdapter {
	logger.InfoOutCtx("initializing fiber adapters SUCCESSFULLY")

	return &httpAdapter{
		metadataAdp:   	adapter.NewMetadataAdapter(cfg),
		applicationAdp:  adapter.NewApplicationAdapter(cfg, application),
	}
}

// FiberServer represents the Fiber server with its configuration and application instance.
type FiberServer struct {
	cfg *config.Config
	FiberApp    *fiber.App
	fiberConfig FiberServerConfig
}

func NewFiberServer(cfg *config.Config) *FiberServer {
	logger.InfoOutCtx("initializing fiber server SUCCESSFULLY")

	// Create Fiber server configuration
	fiberConfig := NewServerConfig(&cfg.HTTP)
	fiberApp := fiber.New(fiberConfig.Config)

	// Setup middleware for the Fiber server
	setupMiddleware(cfg, fiberApp)

	return &FiberServer{
		cfg:    cfg,
		FiberApp:    fiberApp,
		fiberConfig: fiberConfig,
	}
}

func setupMiddleware(cfg *config.Config, fiberApp *fiber.App) {
	logger.InfoOutCtx("setting up middleware for fiber server")
	
	fiberApp.Use(middleware.HeaderMiddleware())
	fiberApp.Use(middleware.RequestIDMiddleware())
	fiberApp.Use(middleware.TraceExtractionMiddleware())
	
	fiberApp.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
}

// SetupRoutes sets up the routes for the Fiber server using the provided application instance.
func (s *FiberServer) SetupRoutes(application *application.Application) {
	logger.InfoOutCtx("setting up routes for fiber server SUCCESSFULLY")

	root := s.FiberApp.Group("/")

	// Create adapters for controllers						
	adapters := newAdapters(s.cfg, application)
	root.Get("/health", adapters.metadataAdp.HealthGet)

	appRoutes := root.Group("/v1")
	appRoutes.Get("/info", adapters.metadataAdp.InfoGet)
	appRoutes.Get("/echo-header", adapters.metadataAdp.HeadersGet)
	appRoutes.Get("/echo-context", adapters.metadataAdp.ContextGet)
	appRoutes.Get("/payment/:payment_number", middleware.MetricsMiddleware(adapters.applicationAdp.PaymentGet))
	appRoutes.Post("/payment", middleware.MetricsMiddleware(adapters.applicationAdp.PaymentAdd))
}