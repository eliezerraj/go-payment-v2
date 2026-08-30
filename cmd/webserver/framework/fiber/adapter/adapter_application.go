package adapter

import (
	"context"
	"strconv"

	"go.uber.org/zap"

	"github.com/go-payment-v2/application/tracing"
	"github.com/go-payment-v2/application/config"
	"github.com/go-payment-v2/application/infrastructure/application"
	"github.com/go-payment-v2/application/domain/external"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/http/utils"

	"github.com/gofiber/fiber/v2"

	"go.opentelemetry.io/otel/trace"
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
	logger.InfoOutCtx("PaymentGet called")

	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.paymentGet", trace.SpanKindInternal)
	defer span.End()

	logger.Debug(
		ctx,
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

	res, err := a.application.PaymentController.PaymentGet(ctx, payment)
	if err != nil {
		logger.Error(ctx, "failed to get payment ", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
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
	logger.InfoOutCtx("PaymentAdd called")

	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.paymentAdd", trace.SpanKindInternal)
	defer span.End()

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.String("host", ctxFiber.Hostname()),
		zap.String("path", ctxFiber.Path()),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)
	
	payment := external.PaymentRequest{}
	if err := ctxFiber.BodyParser(&payment); err != nil {
		logger.Error(ctx, "failed to parse request body", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusBadRequest,
			fiber.ErrBadRequest,
			fiber.ErrBadRequest.Message,
			"failed to parse request body",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	res, err := a.application.PaymentController.PaymentAdd(ctx, payment)
	if err != nil {
		logger.Error(ctx, "failed to add payment", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
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

func (a *ApplicationAdapter) PaymentListByOrderID(ctxFiber *fiber.Ctx) error {
	logger.InfoOutCtx("PaymentListByOrderID called")

	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.PaymentListByOrderID", trace.SpanKindInternal)
	defer span.End()

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.String("host", ctxFiber.Hostname()),
		zap.String("path", ctxFiber.Path()),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)

	number := ctxFiber.Params("order_id")
	if number == "" {
		number = ctxFiber.Query("order_id")
	}

	order := external.OrderRequest{
		OrderNumber: number,
	}

	if id, err := strconv.Atoi(number); err == nil {
		order.ID = id
	}

	paymentRequest := external.PaymentRequest{
		Order: &order,
	}

	res, err := a.application.PaymentController.PaymentListByOrderID(ctx, paymentRequest)
	if err != nil {
		logger.Error(ctx, "failed to list payments by order ID", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusInternalServerError,
			fiber.ErrInternalServerError,
			fiber.ErrInternalServerError.Message,
			"failed to list payments by order ID",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.PaymentResponse{
		Response: "Payments retrieved successfully",
		Payment: res,
	}

	return ctxFiber.Status(fiber.StatusOK).JSON(resp)
}