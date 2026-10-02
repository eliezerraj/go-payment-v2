package main

import (
	"os"
	"os/signal"
	"syscall"
	"context"
  	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.uber.org/zap"
	stdLog "log"

	"github.com/go-payment-v2/cmd/webserver"
	"github.com/go-payment-v2/application/config"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/observability/tracing"
	coreMetricLib "github.com/eliezerraj/go-core/v3/observability/metric"

	"go.opentelemetry.io/otel"
)

const RequestIDHeaderName = "x-request-id"

// Setup logging configuration and initialize logger and hook for request ID.
func setupLogging(cfg *config.Config) {
	logger.NewLogger(
		cfg.Log.Level,
		cfg.Log.Mode,
	).WithHook(func(ctx context.Context) []zap.Field {
		zapFields := []zap.Field{}
		zapFields = append(zapFields, zap.String("app_name", cfg.App.Name))

		if id, ok := ctx.Value(RequestIDHeaderName).(string); ok && id != "" {
			zapFields = append(zapFields, zap.String("x-request-id", id))
		}
		return zapFields
	})
}

// Setup observability
func setupObservability(cfg *config.Config){

	var tracerProvider *tracing.TracerProvider
	
	appInfoTrace := tracing.InfoTrace{
		Name:        cfg.App.Name,
		Version:     cfg.App.Version,
		ServiceType: "k8-workload",
		Env:         cfg.App.Env,
		Account:     cfg.App.Account,
	}

	otelEnvTrace := &tracing.EnvTrace{
		OtelExportEndpoint:      cfg.OtelEnv.OtelExportEndpoint,
		UseStdoutTracerExporter: cfg.OtelEnv.UseStdoutTracerExporter,
		UseOtlpCollector:        cfg.OtelEnv.UseOtlpCollector,
		TimeInterval:            1,
		TimeAliveIncrementer:    1,
		TotalHeapSizeUpperBound: 100,
		ThreadsActiveUpperBound: 10,
		CpuUsageUpperBound:      100,
		SampleAppPorts:          []string{},
		AWSCloudWatchLogGroup:   []string{},
	} 
	
	tracerProvider = tracing.NewTracerProvider(
		context.Background(),
		*otelEnvTrace,
		appInfoTrace,
	)

	otel.SetTracerProvider(tracerProvider.TracerProvider)
}

// Setup metrics
func setupMetrics(cfg *config.Config) {
	ctx := context.Background()

	mp, err := coreMetricLib.NewMeterProvider(ctx, coreMetricLib.InfoMetric{
        Name:    cfg.App.Name,
        Version: cfg.App.Version,
        Env:     cfg.App.Env,
        Account: cfg.App.Account,
    })
	if err != nil {
		stdLog.Fatalf("failed to initialize meter provider: %+v", err)
	}

	otel.SetMeterProvider(mp)

	// Start a separate HTTP server for Prometheus metrics
	mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())

    go func() {
		logger.Info(context.Background(), "starting metrics server", zap.String("port", cfg.OtelEnv.OtelMetricsPort))
        if err := http.ListenAndServe(":"+cfg.OtelEnv.OtelMetricsPort, mux); err != nil {
            logger.Error(context.Background(), "metrics server error", zap.Error(err))
        }
    }()
}

func main() {
	// Load environment configurations
	cfg, err := config.Load()
	if err != nil {
		stdLog.Fatalf("load environment configurations error: %+v", err)
		return
	}

	// Setup logging
	setupLogging(cfg)
	defer logger.Close()

	logger.Info(context.Background(), "starting application")
	logger.Info(context.Background(), "application configuration", zap.Any("config", cfg))

	// Setup observability and metrics
	setupObservability(cfg)
	setupMetrics(cfg)

	// Setup signal handling for graceful shutdown
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	// Define the process type webserver or worker.
	switch cfg.App.Type {
	case "worker":
		logger.Info(context.Background(), "worker process NOT implemented")
	case "webserver":
		logger.Info(context.Background(), "starting webserver process")
		
		webServer := webserver.NewWebServer(cfg)
		go webServer.Run()

		<-stopSignal
		webServer.Shutdown()
	}
}