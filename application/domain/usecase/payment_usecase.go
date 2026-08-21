package usecase

import (
	"time"
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-payment-v2/application/domain/entity"
	"github.com/go-payment-v2/application/infrastructure/repository"
	//"github.com/go-payment-v2/application/infrastructure/module"

	"github.com/jackc/pgx/v5"

	"go.opentelemetry.io/otel"
)

const (
	// PaymentStatusPending represents the pending status of an order.
	PaymentStatusPending = "pending"
	// PaymentStatusCompleted represents the completed status of a payment.
	PaymentStatusCompleted = "completed"
)

type PaymentUsecase struct {
	paymentRepository repository.IPaymentRepository
}

type IPaymentUseCase interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	PaymentAdd(ctx context.Context, payment entity.Payment) (*entity.Payment, error)
	PaymentGet(ctx context.Context, payment entity.Payment) (*entity.Payment, error)

}

func NewPaymentUseCase(paymentRepository repository.IPaymentRepository) *PaymentUsecase {
	logger.InfoOutCtx("initializing payment usecase SUCCESSFULLY")

	return &PaymentUsecase{
		paymentRepository: paymentRepository,
	}
}

// BeginTx starts a new database transaction with the specified options.
func (o *PaymentUsecase) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	logger.Info(ctx, "payment usecase BeginTx called")

	tx, err := o.paymentRepository.BeginTx(ctx, opts)
	if err != nil {
		logger.Error(ctx, "payment usecase BeginTx failed", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

// PaymentAdd adds a new payment to the repository.
func (o *PaymentUsecase) PaymentAdd(ctx context.Context, payment entity.Payment) (res_payment *entity.Payment, err error) {
	logger.Info(ctx, "payment usecase PaymentAdd called")

	tracer := otel.Tracer("payment.repository")
	ctx, span := tracer.Start(ctx, "PaymentUsecase.PaymentAdd")
	defer span.End()

	tx, err := o.paymentRepository.BeginTx(ctx, pgx.TxOptions{ IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite })
	if err != nil {
		logger.Error(ctx, "payment usecase PaymentAdd failed to begin transaction", zap.Error(err))
		return nil, err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
				logger.Error(ctx, "payment usecase PaymentAdd failed to rollback transaction", zap.Error(rollbackErr))
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				logger.Error(ctx, "payment usecase PaymentAdd failed to commit transaction", zap.Error(commitErr))
				err = commitErr
			}
		}
	}()

	//-----------------------------------------------------------
	// Payment SECTION
	//-----------------------------------------------------------

	// Business logic: Set default values for payment
	createAt := time.Now().UTC()
	payment.CreatedAt = createAt 
	if payment.PaymentDate == (time.Time{}) {
		payment.PaymentDate = createAt
	}
	payment.PaymentNumber = "pay:" + payment.OrderNumber
	payment.Status = PaymentStatusPending
	payment.TransactionID = "txn_" + payment.TransactionID

	// Add the payment to the repository
	res_payment, err = o.paymentRepository.PaymentAdd(ctx, tx, payment)
	if err != nil {
		logger.Error(ctx, "payment usecase PaymentAdd failed", zap.Error(err))
		return nil, err
	}

	payment.ID = res_payment.ID

	//-----------------------------------------------------------
	// PaymentCard SECTION
	//-----------------------------------------------------------

	if payment.CreditCard == nil {
		logger.Error(ctx, "payment usecase PaymentAdd: no credit card provided, skipping payment card addition")
		return nil, errors.New("no credit card provided for payment")
	}

	// Add the payment card to the repository
	res_payment_card, err := o.paymentRepository.PaymentCardAdd(ctx, tx, payment)
	if err != nil {
		logger.Error(ctx, "payment usecase PaymentAdd failed to add payment card", zap.Error(err))
		return nil, err
	}

	res_payment.CreditCard = res_payment_card.CreditCard

	logger.Info(ctx, "payment usecase PaymentAdd completed SUCCESSFULLY")
	return res_payment, nil
}

// PaymentGet retrieves a payment from the repository based on the provided payment details.
func (o *PaymentUsecase) PaymentGet(ctx context.Context, payment entity.Payment) (*entity.Payment, error) {
	tracer := otel.Tracer("payment.repository")
	ctx, span := tracer.Start(ctx, "PaymentUsecase.PaymentGet")
	defer span.End()

	logger.Info(ctx, "payment usecase PaymentGet called")

	// Get the payment from the repository
	res_payment, err := o.paymentRepository.PaymentGet(ctx, payment)
	if err != nil {
		logger.Error(ctx, "payment usecase PaymentGet failed", zap.Error(err))
		return nil, err
	}

	return res_payment, nil
}
