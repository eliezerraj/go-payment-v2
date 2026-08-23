package module

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel/trace"
	
	"github.com/eliezerraj/go-core/v3/httpclient"
	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-payment-v2/application/tracing"
	"github.com/go-payment-v2/application/config"
	"github.com/go-payment-v2/application/domain/entity"
	"github.com/go-payment-v2/application/domain/external"
)

type OrderModule struct {
	cfg *config.Config
	client	httpclient.IHTTPClient
}

const (
	AcceptHeader      = "Accept"
	ContentTypeHeader = "Content-Type"
	ConnectionHeader  = "Connection"
	KeepAlive         = "keep-alive"
	XResquestID		 = "X-Request-ID"
)

func NewOrderModule(cfg *config.Config, client	httpclient.IHTTPClient) OrderModule {
	logger.InfoOutCtx("NewOrderModule called")

	return OrderModule{
		cfg: cfg,
		client: client,
	}
}

func (im *OrderModule) OrderPut(ctx context.Context, order entity.Order) (*entity.Order, error) {
	logger.Info(ctx, "order module OrderPut called")

	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Inventory.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxHttpTimeout, "orderModule.OrderPut", trace.SpanKindInternal)
	defer span.End()

	var endpoint string
	method := "GET"
	
	if order.OrderNumber != "" {
		endpoint = fmt.Sprintf("%s%s/%s", im.cfg.Inventory.Endpoint, im.cfg.Inventory.UrlPath, order.OrderNumber)
	} else if order.ID != 0 {
		endpoint = fmt.Sprintf("%s%s/%d", im.cfg.Inventory.Endpoint, im.cfg.Inventory.UrlPath, order.ID)
	} else {
		err := fmt.Errorf("order must have either OrderNumber or ID")
		logger.Error(ctx, "order module OrderPut failed", zap.Error(err))
		return nil, err
	}

	logger.Info(ctx, "order module OrderPut request", zap.String("method", method), zap.String("endpoint", endpoint))
	
	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		logger.Error(ctx, "order module OrderPut failed to create request", zap.Error(err))
		return nil, err
	}

	// Set headers for the request. the const are in payment_module.go file
	xrequestid, ok := ctx.Value(RequestIDHeaderName).(string)
	if !ok {
		xrequestid = "not-informed"
	}

	headers := map[string]string{
		ConnectionHeader:  KeepAlive,
		AcceptHeader:      "application/json",
		ContentTypeHeader: "application/json",
		KeepAlive: "timeout=5, max=1000",
		XResquestID: xrequestid,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := im.client.Do(req.WithContext(ctx))
	if err != nil {
		logger.Error(ctx, "order module OrderPut failed to perform request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, "order module OrderPut failed, inventory service returned non-OK status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	}

	// Decode the response body into an Order struct
	var res external.OrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.Error(ctx, "order module OrderPut failed to decode response", zap.Error(err))
		return nil, err
	}

	order, ok := res.Order.(entity.Order)
	if !ok {
		err := fmt.Errorf("invalid order response")
		logger.Error(ctx, "order module OrderPut failed", zap.Error(err))
		return nil, err
	}

	return &order, nil
}
