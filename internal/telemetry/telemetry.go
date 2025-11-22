package telemetry

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// Setup initializes OpenTelemetry tracing.
// If OTLP_SERVER is unset, tracing is disabled and the returned shutdown is a no-op.
func Setup(ctx context.Context, l *zap.Logger) (func(context.Context) error, error) {
	server := os.Getenv("OTLP_SERVER")
	if server == "" {
		if l != nil {
			l.Info("OTLP_SERVER not set; tracing disabled")
		}
		otel.SetTracerProvider(otel.GetTracerProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		return func(context.Context) error { return nil }, nil
	}

	exp, err := otlptracehttp.New(
		ctx,
		otlptracehttp.WithEndpoint(server),
		otlptracehttp.WithURLPath("/v1/traces"),
		otlptracehttp.WithInsecure(), // OTLP HTTP (4318) is typically plain HTTP inside cluster
	)
	if err != nil {
		return func(context.Context) error { return nil }, err
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithAttributes(
			attribute.String("service.name", "mf-mawinter-maw"),
		),
	)
	if err != nil {
		return func(context.Context) error { return nil }, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exp),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}
