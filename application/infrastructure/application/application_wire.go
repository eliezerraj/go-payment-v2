package application

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/httpclient"
	"github.com/eliezerraj/go-core/v3/database/connector"

	//"github.com/go-payment-v2/application/infrastructure/module"
	"github.com/go-payment-v2/application/config"
	"github.com/go-payment-v2/application/controller"
	"github.com/go-payment-v2/application/domain/usecase"
	"github.com/go-payment-v2/application/infrastructure/repository"
)

type Application struct {
	PaymentController *controller.PaymentController
}

type UseCase struct {
	PaymentUsecase usecase.IPaymentUseCase
}

type Repository struct {
	PaymentRepository repository.IPaymentRepository
	//CheckoutRepository repository.ICheckoutRepository
}

func NewApplication(cfg *config.Config) (*Application, error) {
	logger.InfoOutCtx("initializing application SUCCESSFULLY")

	// Initialize database connector Reader.
	readerConfig := connector.ConnectorConfig{
		DSN:              "postgres://" + cfg.Database.Username + ":" + cfg.Database.Password + "@" + cfg.Database.Host + ":" + cfg.Database.Port + "/" + cfg.Database.Name,
		MaxConnIdleTime:  cfg.Database.ConnIdleTime * time.Minute,
		MaxConnLifeTime:  cfg.Database.ConnLifetime * time.Minute,
		MaxConns:         cfg.Database.MaxConnections,
		MinConns:         cfg.Database.MinConnections,
		DBConnTimeout:    cfg.Database.ConnTimeout,
		HealthCheckPeriod: cfg.Database.ConnIdleTime * time.Minute / 2,
	}

	// Initialize database connector Writer.
	writerConfig := connector.ConnectorConfig{
		DSN:              "postgres://" + cfg.Database.Username + ":" + cfg.Database.Password + "@" + cfg.Database.Host + ":" + cfg.Database.Port + "/" + cfg.Database.Name,
		MaxConnIdleTime:  cfg.Database.ConnIdleTime * time.Minute,
		MaxConnLifeTime:  cfg.Database.ConnLifetime * time.Minute,
		MaxConns:         cfg.Database.MaxConnections,
		MinConns:         cfg.Database.MinConnections,
		DBConnTimeout:    cfg.Database.ConnTimeout,
		HealthCheckPeriod: cfg.Database.ConnIdleTime * time.Minute / 2,
	}

	logger.InfoOutCtx("readerConfig initialized SUCCESSFULLY", zap.Any("readerConfig", readerConfig), zap.Any("writerConfig", writerConfig))

	// Initialize database connector
	dbConnector, err := connector.NewDatabaseConnector(cfg.App.Name, readerConfig, writerConfig)
	if err != nil {
		logger.FatalOutCtx("failed to initialize database connector")
		return nil, err
	}
	
	logger.InfoOutCtx("dbConnector initialized SUCCESSFULLY", zap.Any("dbConnector", dbConnector))
	
	pgConnection := &connector.PgConnection{}
	_, err = pgConnection.NewPool(context.Background(), readerConfig)
	if err != nil {
		logger.FatalOutCtx("failed to create database pool")
		return nil, err
	}
	err = pgConnection.Ping(context.Background())
	if err != nil {
		logger.FatalOutCtx("failed to ping pg connection")
		return nil, err
	}

	// Repository initialization
	paymentRepository := repository.NewPaymentRepository(dbConnector)

	// Create the forwards modules.
	httpConfig := &httpclient.HttpConfig{
		Timeout:             cfg.HTTP.Timeout * time.Second,
		KeepAlive:           cfg.HTTP.KeepAlive * time.Second,
		IdleConnTimeout:     cfg.HTTP.IdleConnTimeout * time.Second,
		MaxIdleConns:        cfg.HTTP.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.HTTP.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.HTTP.MaxConnsPerHost,
		ServiceName:         "go-payment-v2",
	}

	_ = httpConfig
	//invHttpClient := httpclient.NewHttpClient(httpConfig)

	// Create the inventory module.
	//orderModule := module.NewOrderModule(cfg, invHttpClient)

	// UseCase initialization
	paymentUsecase := usecase.NewPaymentUseCase(paymentRepository)
	paymentUsecaseDecorator := usecase.NewPaymentUsecaseEventDecorator(paymentUsecase, true)

	// Controller initialization
	paymentController := controller.NewPaymentController(paymentUsecaseDecorator)

	return &Application{
		PaymentController: paymentController,
	}, nil
}
