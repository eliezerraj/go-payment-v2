package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const RequestIDHeaderName = "x-request-id"

func CustomStartSpanCtx(ctx context.Context, spanName string, spanKind trace.SpanKind) (context.Context, trace.Span) {
	
	tracer := otel.Tracer("go-core.v3.tracing")

	id, ok := ctx.Value(RequestIDHeaderName).(string)
	if !ok {
		id = "not-informed"
	}

	ctx, span := tracer.Start(ctx, spanName,
							  	trace.WithSpanKind(spanKind),
									trace.WithAttributes(
									attribute.String(RequestIDHeaderName, id),
								),
	)

	return ctx, span
}