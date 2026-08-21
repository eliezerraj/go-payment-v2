package adapter

import (
	"context"
	"github.com/gofiber/fiber/v2"

	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-payment-v2/application/config"
)
type MetadataAdapter struct {
	cfg *config.Config
}

func NewMetadataAdapter(cfg *config.Config) *MetadataAdapter {
	logger.InfoOutCtx("initializing metadata adapter SUCCESSFULLY")

	return &MetadataAdapter{
		cfg: cfg,
	}
}

// HealthGet handles the health check endpoint. It responds with a JSON object indicating the service status.
func (c *MetadataAdapter) HealthGet(ctxFiber *fiber.Ctx) error {
	_, cancel := context.WithTimeout(ctxFiber.UserContext(), c.cfg.HTTP.Timeout)
	defer cancel()

	return ctxFiber.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}

// ContextGet handles the context echo endpoint. It responds with a JSON object containing request headers, method, path, and context information.
func (c *MetadataAdapter) ContextGet(ctxFiber *fiber.Ctx) error {
	_, cancel := context.WithTimeout(ctxFiber.UserContext(), c.cfg.HTTP.Timeout)
	defer cancel()

	return ctxFiber.Status(fiber.StatusOK).JSON(fiber.Map{
		"headers": ctxFiber.GetReqHeaders(),
		"method": ctxFiber.Method(),
		"path": ctxFiber.Path(),
		"context": ctxFiber.Context(),
	})
}

// HeadersGet handles the headers echo endpoint. It responds with a JSON object containing the request headers.
func (c *MetadataAdapter) HeadersGet(ctxFiber *fiber.Ctx) error {
	_, cancel := context.WithTimeout(ctxFiber.UserContext(), c.cfg.HTTP.Timeout)
	defer cancel()

	return ctxFiber.Status(fiber.StatusOK).JSON(fiber.Map{
		"headers": ctxFiber.GetReqHeaders(),
	})
}

func (c *MetadataAdapter) InfoGet(ctxFiber *fiber.Ctx) error {
	return ctxFiber.Status(fiber.StatusOK).JSON(fiber.Map{"info": c.cfg})
}