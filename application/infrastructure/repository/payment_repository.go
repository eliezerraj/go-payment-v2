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
	PaymentCardGet(ctx context.Context, payment entity.Payment) ([]*entity.PaymentDetail, error)
	PaymentListByOrderID(ctx context.Context, order entity.Order) ([]*entity.Payment, error)
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

// PaymentGet retrieves a payment record from the database based on the provided payment number.
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
				     p.fk_order_id,
					 p.type,
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
		err = rows.Scan(&payment.ID, &payment.PaymentNumber, &payment.TransactionID, &payment.Order.ID, &payment.Type, &payment.CreatedAt, &payment.UpdatedAt)
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

// PaymentAdd inserts a new payment record into the database and returns the inserted payment with its generated ID.
func (p *PaymentRepository) PaymentAdd(ctx context.Context, tx pgx.Tx, payment entity.Payment) (res_payment *entity.Payment, err error) {
	logger.Info(ctx, "payment repository PaymentAdd called" , zap.Any("payment", payment))

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
										 type,
										 fk_order_id,
										 created_at)
				VALUES ($1, $2, $3, $4, $5) RETURNING id`

	rows := tx.QueryRow(ctx, query, payment.PaymentNumber, payment.TransactionID, payment.Type, payment.Order.ID, payment.CreatedAt)

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
	query := `INSERT INTO public.payment_detail (
										 fk_payment_id,
										 fk_pan,
										 payment_date,
										 status,
										 currency,
										 amount,
										 created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	rows := tx.QueryRow(ctx, query, payment.ID, payment.PaymentDetail[0].CreditCard.Pan, payment.PaymentDetail[0].DetailDate, payment.PaymentDetail[0].Status, payment.PaymentDetail[0].Currency, payment.PaymentDetail[0].Amount, payment.PaymentDetail[0].CreatedAt)

	var id int
	if err := rows.Scan(&id); err != nil {
		return nil, err
	}

	//payment.Card.ID = id
	return &payment, nil
}

func (p *PaymentRepository) PaymentCardGet(ctx context.Context, payment entity.Payment) (res_payment_detail []*entity.PaymentDetail, err error) {
	logger.Info(ctx, "payment repository PaymentCardGet called", zap.Any("payment", payment))

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentRepository.PaymentCardGet", trace.SpanKindInternal)
	defer span.End()

	meter := otel.Meter("go-payment-v2.repository")
	counter, _ := meter.Int64Counter("db_payment_card_get_requests_total")
	histogram, _ := meter.Float64Histogram("db_payment_card_get_duration_seconds")
	start := time.Now()

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", "PaymentCardGet"),
	))

	// Defer function to handle error logging and metrics recording
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "payment repository PaymentCardGet failed", zap.Error(err))
		}
		histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
			attribute.String("operation", "PaymentCardGet"),
		))
	}()
	
	query := `select pd.id,
					 pd.payment_date,
					 pd.status,
					 pd.currency, 
					 pd.amount,
					 pd.created_at,
					 c.pan,
					 c.holder
				from payment_detail pd,
					card c
				where pd.fk_pan = c.pan
				and pd.fk_payment_id = $1`

	connectorReader := p.dbConnector.Reader()
	rows, err := connectorReader.Query(ctx, query, payment.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paymentDetails []*entity.PaymentDetail

	for rows.Next() {
		paymentDetail := entity.PaymentDetail{
			CreditCard: &entity.CreditCard{},
		}
		err = rows.Scan(&paymentDetail.ID, &paymentDetail.DetailDate, &paymentDetail.Status, &paymentDetail.Currency, &paymentDetail.Amount, &paymentDetail.CreatedAt, &paymentDetail.CreditCard.Pan, &paymentDetail.CreditCard.Holder)
		if err != nil {
			return nil, err
		}
		paymentDetails = append(paymentDetails, &paymentDetail)
	}

	if len(paymentDetails) == 0 {
		logger.Warn(ctx, "not found", zap.Int("payment_id", payment.ID))
		err = errors.New("payment card not found")
		return nil, err
	}

	return paymentDetails, nil
}

func (p *PaymentRepository) PaymentListByOrderID(ctx context.Context, order entity.Order) (res_payments []*entity.Payment, err error) {
	logger.Info(ctx, "payment repository PaymentListByOrderID called", zap.Any("order", order))

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentRepository.PaymentListByOrderID", trace.SpanKindInternal)
	defer span.End()

	meter := otel.Meter("go-payment-v2.repository")
	counter, _ := meter.Int64Counter("db_payment_list_by_order_id_requests_total")
	histogram, _ := meter.Float64Histogram("db_payment_list_by_order_id_duration_seconds")
	start := time.Now()

	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", "PaymentListByOrderID"),
	))

	// Defer function to handle error logging and metrics recording
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "payment repository PaymentListByOrderID failed", zap.Error(err))
		}
		histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
			attribute.String("operation", "PaymentListByOrderID"),
		))
	}()
	
	query := `select p.id,
					p.transaction_id,
					p.payment_number,
					p.type,
					p.created_at,
					pd.fk_pan, 
					pd.currency,
					pd.amount,
					pd.created_at,
					c.holder 
			from payment p,
				payment_detail pd,
				card c
			where p.id = pd.fk_payment_id 
			and c.pan = pd.fk_pan
			and p.fk_order_id = $1`

	connectorReader := p.dbConnector.Reader()
	rows, err := connectorReader.Query(ctx, query, order.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*entity.Payment	
	for rows.Next() {
		payment := &entity.Payment{}
		paymentDetails := []*entity.PaymentDetail{}
		paymentDetail := entity.PaymentDetail{
			CreditCard: &entity.CreditCard{},
		}
		err = rows.Scan(&payment.ID, &payment.TransactionID, &payment.PaymentNumber, &payment.Type, &payment.CreatedAt, &paymentDetail.CreditCard.Pan, &paymentDetail.Currency, &paymentDetail.Amount, &paymentDetail.CreatedAt, &paymentDetail.CreditCard.Holder)
		if err != nil {
			return nil, err
		}
		paymentDetails = append(paymentDetails, &paymentDetail)
		payment.PaymentDetail = paymentDetails
		payment.Order = order
		res_payments = append(res_payments, payment)
		payments = append(payments, payment)
	}

	return payments, nil
}