package controller

import (
	"context"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel"

	"github.com/eliezerraj/go-core/v3/logger"
	
	"github.com/go-payment-v2/application/domain/usecase"
	"github.com/go-payment-v2/application/domain/external"
	"github.com/go-payment-v2/application/domain/entity"
)

type PaymentController struct {
	paymentUseCase usecase.IPaymentUseCase
}

func NewPaymentController(paymentUseCase usecase.IPaymentUseCase) *PaymentController {
	logger.InfoOutCtx("initializing payment controller SUCCESSFULLY")

	return &PaymentController{
		paymentUseCase: paymentUseCase,
	}
}

// PaymentAdd handles the addition of a new payment based on the provided request.
func (p *PaymentController) PaymentAdd(ctx context.Context, req external.PaymentRequest) (*entity.Payment, error) {
	tracer := otel.Tracer("payment.controller")
	ctx, span := tracer.Start(ctx, "PaymentController.PaymentAdd")
	defer span.End()

	logger.Info(ctx, "payment controller PaymentAdd called")

	// Create the payment entity
	payment := entity.Payment{
		OrderID:       req.OrderID,
		OrderNumber:   req.OrderNumber,
		TransactionID: req.TransactionID,
		Type:          req.Type,
		Currency:      req.Currency,
		Amount:        req.Amount,
	}

	// Create CreditCard entity if provided
	if req.CreditCard != nil {
		payment.CreditCard = &entity.CreditCard{
			Pan:      req.CreditCard.Pan,
			Holder:   req.CreditCard.Holder,
			Password: req.CreditCard.Password,
			CVV:      req.CreditCard.CVV,
		}
	}

	// Call the use case to add the order
	res, err := p.paymentUseCase.PaymentAdd(ctx, payment)
	if err != nil {
		logger.Error(ctx, "payment controller PaymentAdd failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}

// PaymentGet handles the retrieval of a payment based on the provided request.
func (p *PaymentController) PaymentGet(ctx context.Context, req external.PaymentRequest) (*entity.Payment, error) {
	tracer := otel.Tracer("payment.controller")
	ctx, span := tracer.Start(ctx, "PaymentController.PaymentGet")
	defer span.End()

	logger.Info(ctx, "payment controller PaymentGet called", zap.String("payment_number", req.PaymentNumber))

	payment := entity.Payment{
		PaymentNumber: req.PaymentNumber,
	}

	res, err := p.paymentUseCase.PaymentGet(ctx, payment)
	if err != nil {
		logger.Error(ctx, "payment controller PaymentGet failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}
