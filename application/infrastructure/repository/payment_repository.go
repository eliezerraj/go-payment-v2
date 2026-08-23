package repository

import (
	"time"
	"context"
	"errors"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"

	"github.com/go-payment-v2/application/domain/entity"
	"github.com/go-payment-v2/application/tracing"
	
	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/database/connector"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/attribute"
)

type PaymentRepository struct {
	dbConnector connector.IDatabaseConnector
}

type IPaymentRepository interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	PaymentAdd(ctx context.Context, tx pgx.Tx, payment entity.Payment) (*entity.Payment, error)
	PaymentGet(ctx context.Context, payment entity.Payment) (*entity.Payment, error)
	PaymentCardAdd(ctx context.Context, tx pgx.Tx, payment entity.Payment) (*entity.Payment, error)
}

func NewPaymentRepository(dbConnector connector.IDatabaseConnector) IPaymentRepository {
	logger.InfoOutCtx("initializing payment repository SUCCESSFULLY")

	return &PaymentRepository{
		dbConnector: dbConnector,
	}
}

func (p *PaymentRepository) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	logger.InfoOutCtx("payment repository BeginTx called")

	tx, err := p.dbConnector.Writer().BeginTx(ctx, opts)
	if err != nil {
		logger.ErrorOutCtx("payment repository BeginTx failed", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

func (p *PaymentRepository) PaymentGet(ctx context.Context, payment entity.Payment) (res_payment *entity.Payment, err error) {
	logger.Info(ctx, "payment repository PaymentGet called")

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentRepository.PaymentGet", trace.SpanKindInternal)
	defer span.End()

    meter := otel.Meter("go-payment-v2.repository")
    counter, _ := meter.Int64Counter("db_payment_get_requests_total")
    histogram, _ := meter.Float64Histogram("db_payment_get_duration_seconds")
    start := time.Now()

    counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "PaymentGet"),
    ))

	// Defer function to handle error logging and metrics recording	
	defer func() {
		if err != nil {
			span.RecordError(err) 
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "payment repository PaymentGet failed", zap.Error(err))
		}
        histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
            attribute.String("operation", "PaymentGet"),
        ))
	}()

	// Get a reader connection from the database connector
	connectorReader := p.dbConnector.Reader()

	query := `select p.id,
					 p.payment_number,
					 p.transaction_id,
					 p.payment_date,
					 p.status,
					 p.currency,
					 p.amount,
					 p.created_at,
					 p.updated_at
				from public.payment p
				where p.payment_number = $1`

	rows, err := connectorReader.Query(ctx, query, payment.PaymentNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payment = entity.Payment{}
	if rows.Next() {
		err = rows.Scan(&payment.ID, &payment.PaymentNumber, &payment.TransactionID, &payment.PaymentDate, &payment.Status, &payment.Currency, &payment.Amount, &payment.CreatedAt, &payment.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Warn(ctx, "not found", zap.Int("payment_id", payment.ID))
		err = errors.New("payment not found")
		return nil, err
	}

	return &payment, nil
}

func (p *PaymentRepository) PaymentAdd(ctx context.Context, tx pgx.Tx, payment entity.Payment) (res_payment *entity.Payment, err error) {
	logger.Info(ctx, "payment repository PaymentAdd called")

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentRepository.PaymentAdd", trace.SpanKindInternal)
	defer span.End()
	
	meter := otel.Meter("go-payment-v2.repository")
    counter, _ := meter.Int64Counter("db_payment_add_requests_total")
    histogram, _ := meter.Float64Histogram("db_payment_add_duration_seconds")
    start := time.Now()

    counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "PaymentAdd"),
    ))

	// Defer function to handle error logging and metrics recording
	defer func() {
		if err != nil {
			span.RecordError(err) 
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "payment repository PaymentAdd failed", zap.Error(err))
		}
		histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
			attribute.String("operation", "PaymentAdd"),
		))
	}()

	// Insert the payment into the database
	query := `INSERT INTO public.payment (payment_number,
										 transaction_id,
										 payment_date,
										 status,
										 currency,
										 amount,
										 created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	rows := tx.QueryRow(ctx, query, payment.PaymentNumber, payment.TransactionID, payment.PaymentDate, payment.Status, payment.Currency, payment.Amount, payment.CreatedAt)

	var id int
	if err := rows.Scan(&id); err != nil {
		return nil, err
	}

	payment.ID = id
	return &payment, nil
}

func (p *PaymentRepository) PaymentCardAdd(ctx context.Context, tx pgx.Tx, payment entity.Payment) (res_payment *entity.Payment, err error) {
	logger.Info(ctx, "payment repository PaymentCardAdd called")

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentRepository.PaymentCardAdd", trace.SpanKindInternal)
	defer span.End()
	
	meter := otel.Meter("go-payment-v2.repository")
    counter, _ := meter.Int64Counter("db_payment_add_requests_total")
    histogram, _ := meter.Float64Histogram("db_payment_add_duration_seconds")
    start := time.Now()

    counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "PaymentCardAdd"),
    ))

	// Defer function to handle error logging and metrics recording
	defer func() {
		if err != nil {
			span.RecordError(err) 
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "payment repository PaymentCardAdd failed", zap.Error(err))
		}
		histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
			attribute.String("operation", "PaymentCardAdd"),
		))
	}()

	// Insert the payment into the database
	query := `INSERT INTO public.payment_card (
										 fk_order_id,
										 fk_payment_id,
										 fk_pan,
										 currency,
										 amount,
										 created_at)
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	rows := tx.QueryRow(ctx, query, payment.OrderID, payment.ID, payment.CreditCard.Pan, payment.Currency, payment.Amount, payment.CreatedAt)

	var id int
	if err := rows.Scan(&id); err != nil {
		return nil, err
	}

	//payment.Card.ID = id
	return &payment, nil
}
