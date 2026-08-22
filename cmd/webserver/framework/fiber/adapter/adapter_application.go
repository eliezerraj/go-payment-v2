package adapter

import (
	"context"

	"go.uber.org/zap"

	"github.com/go-payment-v2/application/config"
	"github.com/go-payment-v2/application/infrastructure/application"
	"github.com/go-payment-v2/application/domain/external"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/http/utils"

	"github.com/gofiber/fiber/v2"

	"go.opentelemetry.io/otel"
)

type ApplicationAdapter struct {
	cfg *config.Config
	application *application.Application
}

func NewApplicationAdapter(cfg *config.Config, application *application.Application) *ApplicationAdapter {
	logger.InfoOutCtx("initializing application adapter SUCCESSFULLY")

	return &ApplicationAdapter{
		cfg:         cfg,
		application: application,
	}
}

// Adapter methods for ProductController 
func (a *ApplicationAdapter) PaymentGet(ctxFiber *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	tracer := otel.Tracer("payment.adapter")
	ctx, span := tracer.Start(ctxWithTimeout, "ApplicationAdapter.PaymentGet")
	defer span.End()

	logger.Info(ctx, "PaymentGet called")

	logger.Debug(
		ctxWithTimeout,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.String("host", ctxFiber.Hostname()),
		zap.String("path", ctxFiber.Path()),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)

	payment_number := ctxFiber.Params("payment_number")
	if payment_number == "" {
		payment_number = ctxFiber.Query("payment_number")
	}

	payment := external.PaymentRequest{
		PaymentNumber: payment_number,
	}

	res, err := a.application.PaymentController.PaymentGet(ctxWithTimeout, payment)
	if err != nil {
		logger.Error(ctxWithTimeout, "failed to get payment ", zap.Error(err))
		errorResponse := external.NewResponseError(ctxWithTimeout,
			fiber.StatusNotFound,
			fiber.ErrNotFound,
			fiber.ErrNotFound.Message,
			"failed to get payment",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.PaymentResponse{
		Response: "Payment retrieved successfully",
		Payment: res,
	}

	return ctxFiber.Status(fiber.StatusOK).JSON(resp)
}

// Adapter methods for PaymentController
func (a *ApplicationAdapter) PaymentAdd(ctxFiber *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	logger.Info(ctxWithTimeout, "PaymentAdd called")

	logger.Debug(
		ctxWithTimeout,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.String("host", ctxFiber.Hostname()),
		zap.String("path", ctxFiber.Path()),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)
	
	payment := external.PaymentRequest{}
	if err := ctxFiber.BodyParser(&payment); err != nil {
		logger.Error(ctxWithTimeout, "failed to parse request body", zap.Error(err))
		errorResponse := external.NewResponseError(ctxWithTimeout,
			fiber.StatusBadRequest,
			fiber.ErrBadRequest,
			fiber.ErrBadRequest.Message,
			"failed to parse request body",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	res, err := a.application.PaymentController.PaymentAdd(ctxWithTimeout, payment)
	if err != nil {
		logger.Error(ctxWithTimeout, "failed to add payment", zap.Error(err))
		errorResponse := external.NewResponseError(ctxWithTimeout,
			fiber.StatusInternalServerError,
			fiber.ErrInternalServerError,
			fiber.ErrInternalServerError.Message,
			"failed to add payment",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.PaymentResponse{
		Response: "Payment added successfully",
		Payment: res,
	}

	return ctxFiber.Status(fiber.StatusCreated).JSON(resp)
}
