package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Vintral/pocket-realm/utils"
	"github.com/rs/zerolog/log"
)

const timeOut = 10

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, tracerProvider *trace.TracerProvider, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.go ad
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		fmt.Println("in shutdown")
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up resource.
	OTEL_SERVICE := utils.GetEnv("OTEL_SERVICE", "default-service")
	OTEL_VERSION := utils.GetEnv("OTEL_VERSION", "0.0.1")

	fmt.Println("OTEL_SERVICE:", OTEL_SERVICE)
	res, err := newResource(OTEL_SERVICE, OTEL_VERSION)
	if err != nil {
		handleErr(err)
		return
	}

	log.Info().Msg("Setup Resource")

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)
	log.Info().Msg("Set Propagator")

	// Set up trace provider.
	tracerProvider, err = newTraceProvider(res, ctx)
	if err != nil {
		log.Error().AnErr("error", err).Msg("Error Creating trace provider")
		handleErr(err)
		return
	}
	log.Info().Msg("Created trace provider")
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	log.Info().Msg("Setup shutdown funcs")
	otel.SetTracerProvider(tracerProvider)
	log.Info().Msg("Set Tracer Provider")

	return
}

func newResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.New(
		context.Background(),
		resource.WithContainer(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("service.version", serviceVersion),
		),
	)
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTraceProvider(res *resource.Resource, ctx context.Context) (*trace.TracerProvider, error) {
	// traceExporter, err := stdouttrace.New(
	// 	stdouttrace.WithPrettyPrint())
	// if err != nil {
	// 	return nil, err
	// }

	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*timeOut))
	defer cancel()

	OTEL_EXPORTER_HOST := utils.GetEnv("OTEL_EXPORTER_HOST", "127.0.0.1")
	OTEL_EXPORTER_PORT := utils.GetEnv("OTEL_EXPORTER_PORT", "4317")

	log.Info().Msg("Connection Collector: " + OTEL_EXPORTER_HOST + ":" + OTEL_EXPORTER_PORT)

	conn, err := grpc.NewClient(
		OTEL_EXPORTER_HOST+":"+OTEL_EXPORTER_PORT,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	//traceExporter, err := stdouttrace.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// traceProvider := trace.NewTracerProvider(
	// 	trace.WithBatcher(traceExporter,
	// 		// Default is 5s. Set to 1s for demonstrative purposes.
	// 		trace.WithBatchTimeout(time.Second)),
	// 	trace.WithResource(res),
	// )

	traceProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(
			sdktrace.NewBatchSpanProcessor(traceExporter),
		),
	)
	return traceProvider, nil
}

// func newMeterProvider(res *resource.Resource, ctx context.Context) (*metric.MeterProvider, error) {
// 	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*timeOut))
// 	defer cancel()

// 	OTEL_EXPORTER_HOST := utils.GetEnv("OTEL_EXPORTER_HOST", "127.0.0.1")
// 	OTEL_EXPORTER_PORT := utils.GetEnv("OTEL_EXPORTER_PORT", "4317")
// 	conn, err := grpc.DialContext(ctx, OTEL_EXPORTER_HOST+":"+OTEL_EXPORTER_PORT,
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 		grpc.WithBlock(),
// 	)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// metricExporter, err := stdoutmetric.New()
// 	metricExporter, err := otlpmetricgrpc.New(
// 		context.Background(),
// 		otlpmetricgrpc.WithGRPCConn(conn),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	meterProvider := metric.NewMeterProvider(
// 		metric.WithResource(res),
// 		metric.WithReader(metric.NewPeriodicReader(
// 			metricExporter,
// 			metric.WithInterval(time.Second*15),
// 		)),
// 	)
// 	return meterProvider, nil
// }
