package usecase

import (
	"github.com/google/uuid"
	"fmt"
	"time"
	"context"
	"errors"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"
	gocore_kafka "github.com/eliezerraj/go-core/v3/event/kafka"
	"github.com/eliezerraj/go-core/v3/event/kafka/producer"

	"github.com/go-payment-v2/application/tracing"
	"github.com/go-payment-v2/application/domain/entity"
	"github.com/go-payment-v2/application/infrastructure/repository"
	"github.com/go-payment-v2/application/config"
	"github.com/go-payment-v2/application/shared/otelkafka"
	
	"github.com/jackc/pgx/v5"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel"

)

const (
	PaymentStatusPending = "payment:pending"
	PaymentStatusCompleted = "payment:completed"
	RequestIDHeaderName = "x-request-id"
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

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentUsecase.PaymentAdd", trace.SpanKindInternal)
	defer span.End()

	// Start a new transaction
	tx, err := o.paymentRepository.BeginTx(ctx, pgx.TxOptions{ IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite })
	if err != nil {
		logger.Error(ctx, "payment usecase PaymentAdd failed to begin transaction", zap.Error(err))
		return nil, err
	}

	//Defer a function to handle commit or rollback based on the outcome of the operation
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
	payment.PaymentNumber = "pay:" + payment.Order.OrderNumber
	payment.TransactionID = payment.TransactionID // the transaction ID is provided externally
	payment.Type = payment.Type
	payment.CreatedAt = createAt

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
	for i, paymentDetail := range payment.PaymentDetail {

		paymentDetail.Status = PaymentStatusPending
		paymentDetail.ID = payment.ID
		paymentDetail.CreatedAt = createAt
		if paymentDetail.DetailDate == (time.Time{}) {
			paymentDetail.DetailDate = createAt
		}

		if paymentDetail.CreditCard == nil {
			logger.Error(ctx, "payment usecase PaymentAdd: no credit card provided, skipping payment card addition")
			return nil, errors.New("no credit card provided for payment")
		}
		// Add the payment card to the repository
		payment.PaymentDetail[i] = paymentDetail

		res_payment_detail, err := o.paymentRepository.PaymentCardAdd(ctx, tx, payment)
		if err != nil {
			logger.Error(ctx, "payment usecase PaymentAdd failed to add payment card", zap.Error(err))
			return nil, err
		}
		payment.PaymentDetail[i].ID = res_payment_detail.ID
	}

	logger.Info(ctx, "payment usecase PaymentAdd completed SUCCESSFULLY")
	return res_payment, nil
}

// PaymentGet retrieves a payment from the repository based on the provided payment details.
func (o *PaymentUsecase) PaymentGet(ctx context.Context, payment entity.Payment) (*entity.Payment, error) {
	logger.Info(ctx, "payment usecase PaymentGet called")
	
	// Tracing
	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentUsecase.PaymentGet", trace.SpanKindInternal)
	defer span.End()

	// Get the payment from the repository
	res_payment, err := o.paymentRepository.PaymentGet(ctx, payment)
	if err != nil {
		logger.Error(ctx, "payment usecase PaymentGet failed", zap.Error(err))
		return nil, err
	}

	// Set the payment ID for further processing
	payment.ID = res_payment.ID

	// Get the payment card details from the repository
	res_payment_detail, err := o.paymentRepository.PaymentCardGet(ctx, payment)
	if err != nil {
		logger.Error(ctx, "payment usecase PaymentGet failed to get payment card", zap.Error(err))
		return nil, err
	}

	res_payment.PaymentDetail = res_payment_detail

	logger.Info(ctx, "payment usecase PaymentGet completed SUCCESSFULLY")

	return res_payment, nil
}

// -------------------------------------------------------------------------------------------------------
// PaymentUsecaseEventDecorator is a decorator for the PaymentUsecase that adds event publishing functionality.
type PaymentUsecaseEventDecorator struct {
	next IPaymentUseCase
	producerWorker *producer.ProducerWorker
	enabled	bool
	kafkaProducer *config.KafkaProducer
}

func NewPaymentUsecaseEventDecorator(next IPaymentUseCase, enabled bool, kafkaProducer config.KafkaProducer) *PaymentUsecaseEventDecorator {
	logger.InfoOutCtx("initializing payment usecase event decorator SUCCESSFULLY")

	dialerConfig := gocore_kafka.DialerConfig{
		Username:   kafkaProducer.Username,
		Password:   kafkaProducer.Password,
		Protocol:   kafkaProducer.Protocol,
		Mechanisms: kafkaProducer.Mechanism,
		Brokers:    kafkaProducer.BrokerList,
	}

	kafkaDialer := gocore_kafka.NewKafkaDialer(dialerConfig)
	producerConfig := kafkaDialer.ProducerConfig(kafkaProducer.Name)
	producerWorker, err := producer.NewProducerWorker(producerConfig)
	if err != nil {
		logger.FatalOutCtx("failed to create ProducerWorker", zap.Error(err))
		return nil
	}

	logger.InfoOutCtx("ProducerWorker created successfully", zap.Any("producerConfig", producerConfig))
	return &PaymentUsecaseEventDecorator{
		next:          next,
		producerWorker: producerWorker,
		enabled:       enabled,
		kafkaProducer: &kafkaProducer,
	}
}

// PaymentAdd adds a new payment and publishes an event to Kafka if the decorator is enabled.
func (d *PaymentUsecaseEventDecorator) PaymentAdd(ctx context.Context, payment entity.Payment) (*entity.Payment, error) {
	logger.Info(ctx, "PaymentUsecaseEventDecorator PaymentAdd called")

	ctx, span := tracing.CustomStartSpanCtx(ctx, "paymentUsecaseEventDecorator.PaymentAdd", trace.SpanKindInternal)
	defer span.End()

	if !d.enabled {
		logger.Info(ctx, "PaymentUsecaseEventDecorator is disabled, proceeding without event publishing")
		return d.next.PaymentAdd(ctx, payment)
	}

	logger.Info(ctx, "PaymentUsecaseEventDecorator is enabled, proceeding with event publishing")

	// Call the next use case in the chain
	res_payment, err := d.next.PaymentAdd(ctx, payment)
	if err != nil {
		logger.Error(ctx, "PaymentUsecaseEventDecorator: failed to add payment", zap.Error(err))
		return nil, err
	}

	// ------------------------------------------
	// Publish event to Kafka
	// ------------------------------------------
	logger.Info(ctx, "KAFKA PaymentUsecaseEventDecorator KAFKA ======>>>>>>> publishing")

	key := fmt.Sprintf("payment:%v", res_payment.PaymentNumber)
	topic := d.kafkaProducer.Topic
	
	// Set headers for the request. the const are in payment_module.go file
	xrequestid, ok := ctx.Value(RequestIDHeaderName).(string)
	if !ok {
		xrequestid = "not-informed"
	}

	// Create the event to be published to Kafka
	event := entity.Event{
		ID:   uuid.New().String(),
		Date: time.Now(),
		Type: topic,
		Metadata: map[string]interface{}{
			"x-request-id": xrequestid,
		},
		Data: res_payment,
	}
	
	// Convert the event to JSON payload for Kafka
	payload_bytes, err := json.Marshal(event)
	if err != nil {
		logger.Error(ctx, "PaymentUsecaseEventDecorator: failed to marshal payment", zap.Error(err))
		return nil, err
	}

	// Prepare Kafka headers for tracing and request ID
	kafkaHeaders := otelkafka.KafkaHeaderCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, &kafkaHeaders)
	// Inject tracing context into Kafka headers
	kafkaHeaders.Set("x-request-id", xrequestid)
	err = d.producerWorker.ProduceMessage(ctx,topic, key, kafkaHeaders, payload_bytes)
	if err != nil {
		logger.Error(ctx, "PaymentUsecaseEventDecorator: failed to produce message", zap.Error(err))

		return nil, err
	}

	logger.Info(ctx, "PaymentUsecaseEventDecorator: payment added and event published successfully")
	return res_payment, nil
}

// PaymentGet retrieves a payment and publishes an event to Kafka if the decorator is enabled.
func (d *PaymentUsecaseEventDecorator) PaymentGet(ctx context.Context, payment entity.Payment) (*entity.Payment, error) {
	logger.Info(ctx, "PaymentUsecaseEventDecorator PaymentGet called")

	return d.next.PaymentGet(ctx, payment)
}

// BeginTx starts a new database transaction with the specified options.
func (d *PaymentUsecaseEventDecorator) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	logger.Info(ctx, "PaymentUsecaseEventDecorator BeginTx called")

	return d.next.BeginTx(ctx, opts)
}