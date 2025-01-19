package utilities

import (
	"context"
	"math"
	"os"

	"go.opentelemetry.io/otel/sdk/trace"
	span "go.opentelemetry.io/otel/trace"
)

type KeyTraceProvider struct{}
type KeyUser struct{}
type KeyPayload struct{}

func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func RoundFloat(val float64, precision uint) float64 {
	if val == 0 {
		return 0.0
	}

	ratio := math.Pow(10, float64(precision))
	if val > 0 {
		return math.Round(val*ratio) / ratio
	}

	return math.Round(val*ratio-0.5) / ratio
}

type StartSpanOpts struct {
	TracerName string
}

func StartSpanWithOpts(ctx context.Context, spanName string, opts *StartSpanOpts) (context.Context, span.Span) {
	return ctx.Value(KeyTraceProvider{}).(*trace.TracerProvider).Tracer(opts.TracerName).Start(context.Background(), spanName)
}

func StartSpan(ctx context.Context, spanName string) (context.Context, span.Span) {
	return StartSpanWithOpts(ctx, spanName, &StartSpanOpts{TracerName: "realm-game"})
}

func GetSpan(ctx context.Context) span.Span {
	return span.SpanFromContext(ctx)
}
