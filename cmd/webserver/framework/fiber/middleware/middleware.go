package middleware

import (
	"context"
	"time"
	"strings"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
)

const RequestIDHeaderName = "x-request-id"

// AuthorizationMiddleware is a middleware function that checks for the presence of an Authorization header in the request. If the header is missing, it returns a 401 Unauthorized response.
func AuthorizationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Accepts("application/json")

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header is required",
			})
		}
		return c.Next()
	}
}

// HeaderMiddleware is a middleware function that sets security and CORS headers for the response.
func HeaderMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {

		c.Accepts("application/json")

		securityHeaders := map[string]string{
			"X-XSS-Protection":          "1; mode=block",
			"X-Content-Type-Options":    "nosniff",
			"X-Download-Options":        "noopen",
			"Strict-Transport-Security": "max-age=5184000",
			"X-Frame-Options":           "SAMEORIGIN",
			"X-DNS-Prefetch-Control":    "off",
		}

		corsHeaders := map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Headers":     "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With",
			"Access-Control-Allow-Methods":     "POST, GET, PUT, DELETE",
		}

		for key, value := range securityHeaders {
			c.Set(key, value)
		}
		for key, value := range corsHeaders {
			c.Set(key, value)
		}

		return c.Next()
	}
}

// RequestIDMiddleware is a middleware function that generates a unique request ID for each incoming request and adds it to the request context. If the request already has a request ID in the "x-request-id" header, it uses that value instead of generating a new one.
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(RequestIDHeaderName)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(RequestIDHeaderName, requestID)
		ctx := context.WithValue(c.UserContext(), RequestIDHeaderName, requestID)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// MetricsMiddleware is a middleware function that can be used to collect metrics for each incoming request. It is currently a placeholder and does not implement any metrics collection logic.
func MetricsMiddleware(next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
        start := time.Now()
        err := next(c)

		routePath := c.Route().Path
        if routePath == "" {
            routePath = c.Path()
            if i := strings.Index(routePath, "?"); i >= 0 {
                routePath = routePath[:i]
            }
        }

        meter := otel.Meter("go-payment-v2.http")
        counter, _ := meter.Int64Counter("http_custom_requests_total")
        histogram, _ := meter.Float64Histogram("http_custom_request_duration_seconds")

        counter.Add(c.UserContext(), 1, metric.WithAttributes(
            attribute.String("method", c.Method()),
            attribute.String("path", routePath),
            attribute.Int("status_code", c.Response().StatusCode()),
        ))

        histogram.Record(c.UserContext(), time.Since(start).Seconds(), metric.WithAttributes(
            attribute.String("method", c.Method()),
            attribute.String("path", c.Path()),
        ))

        return err
	}
}


// TraceExtractionMiddleware is a middleware function that extracts the trace context from incoming requests and sets it in the request context. This allows for distributed tracing across services.
func TraceExtractionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		headerMap := make(http.Header)
		for k, values := range c.GetReqHeaders() {
			for _, v := range values {
				headerMap.Add(k, v)
			}
		}

		propagator := otel.GetTextMapPropagator()
		ctx := propagator.Extract(c.UserContext(), propagation.HeaderCarrier(headerMap))
		c.SetUserContext(ctx)

		return c.Next()
	}
}