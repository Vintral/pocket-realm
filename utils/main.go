package utils

import (
	"context"
	"math"
	"os"

	"go.opentelemetry.io/otel/sdk/trace"
	span "go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type KeyTraceProvider struct{}
type KeyUser struct{}
type KeyPayload struct{}
type KeyDB struct{}

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

func StartSpanWithOpts(baseContext context.Context, spanName string, opts *StartSpanOpts) (context.Context, span.Span) {
	provider := baseContext.Value(KeyTraceProvider{}).(*trace.TracerProvider)
	ctx, span := provider.Tracer(opts.TracerName).Start(baseContext, spanName)
	ctx = context.WithValue(ctx, KeyTraceProvider{}, provider)

	if baseContext.Value(KeyUser{}) != nil {
		ctx = context.WithValue(ctx, KeyUser{}, baseContext.Value(KeyUser{}).(any))
	}

	if baseContext.Value(KeyDB{}) != nil {
		ctx = context.WithValue(ctx, KeyDB{}, baseContext.Value(KeyDB{}).(*gorm.DB))
	}

	return ctx, span
}

func StartSpan(ctx context.Context, spanName string) (context.Context, span.Span) {
	return StartSpanWithOpts(ctx, spanName, &StartSpanOpts{TracerName: "realm-game"})
}

func StartCronSpan(ctx context.Context, spanName string) (context.Context, span.Span) {
	return StartSpanWithOpts(ctx, spanName, &StartSpanOpts{TracerName: "realm-cron"})
}

func GetSpan(ctx context.Context) span.Span {
	return span.SpanFromContext(ctx)
}
