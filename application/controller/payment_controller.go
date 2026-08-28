package controller

import (
	"context"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel/trace"

	"github.com/eliezerraj/go-core/v3/logger"
	
	"github.com/go-payment-v2/application/tracing"
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
	logger.Info(ctx, "payment controller PaymentAdd called")

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentController.PaymentAdd", trace.SpanKindInternal)
	defer span.End()

	// Create the payment entity
	order := entity.Order{
		ID:          req.Order.ID,
		OrderNumber: req.Order.OrderNumber,
	}

	paymentDetails := make([]*entity.PaymentDetail, len(req.PaymentDetail))

	for i, paymentDetailReq := range req.PaymentDetail {
		creditCard := entity.CreditCard{
			Pan:            paymentDetailReq.CreditCard.Pan,
			Holder:         paymentDetailReq.CreditCard.Holder,
			Password:       paymentDetailReq.CreditCard.Password,
			CVV:            paymentDetailReq.CreditCard.CVV,
		}
		paymentDetail := entity.PaymentDetail{
			DetailDate: paymentDetailReq.DetailDate,
			Status:     paymentDetailReq.Status,
			Currency:   paymentDetailReq.Currency,
			Amount:     paymentDetailReq.Amount,
			CreditCard: &creditCard,
		}
		paymentDetails[i] = &paymentDetail
	}

	payment := entity.Payment{
		TransactionID: req.TransactionID,
		Type:          req.Type,
		Order:         order,
		PaymentDetail: paymentDetails,
	}

	logger.Info(ctx, "payment controller PaymentAdd: calling payment usecase PaymentAdd", zap.Any("payment", payment))

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
	logger.Info(ctx, "payment controller PaymentGet called", zap.String("payment_number", req.PaymentNumber))

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentController.PaymentGet", trace.SpanKindInternal)
	defer span.End()

	payment := entity.Payment{
		PaymentNumber: req.PaymentNumber,
	}

	// Call the use case to get the order
	res, err := p.paymentUseCase.PaymentGet(ctx, payment)
	if err != nil {
		logger.Error(ctx, "payment controller PaymentGet failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}
