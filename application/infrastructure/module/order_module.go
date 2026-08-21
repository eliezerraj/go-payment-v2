package module

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel"

	"github.com/eliezerraj/go-core/v3/httpclient"
	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-payment-v2/application/config"
	"github.com/go-payment-v2/application/domain/entity"
	"github.com/go-payment-v2/application/domain/external"
)

type OrderModule struct {
	cfg *config.Config
	client	httpclient.IHTTPClient
}

func NewOrderModule(cfg *config.Config, client	httpclient.IHTTPClient) OrderModule {
	logger.InfoOutCtx("NewOrderModule called")

	return OrderModule{
		cfg: cfg,
		client: client,
	}
}

func (im *OrderModule) OrderPut(ctx context.Context, order entity.Order) (*entity.Order, error) {
	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Inventory.Timeout)
	defer cancel()

	tracer := otel.Tracer("order.module")
	ctxSpan, span := tracer.Start(ctxHttpTimeout, "OrderModule.OrderPut")
	defer span.End()

	logger.Info(ctx, "order module OrderPut called")

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
	
	req, err := http.NewRequestWithContext(ctxSpan, method, endpoint, nil)
	if err != nil {
		logger.Error(ctx, "order module OrderPut failed to create request", zap.Error(err))
		return nil, err
	}
	resp, err := im.client.Do(req.WithContext(ctxSpan))
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
