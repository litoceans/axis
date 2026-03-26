package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Tracer wraps the OpenTelemetry tracer
type Tracer struct {
	tp           *sdktrace.TracerProvider
	tracer       trace.Tracer
	enabled      bool
	grpcConn     *grpc.ClientConn
	shutdownFunc func(context.Context) error
}

// Config holds tracer configuration
type Config struct {
	Enabled      bool
	Endpoint     string
	ServiceName  string
	SampleRatio  float64
}

// New creates a new tracer
func New(ctx context.Context, cfg Config) (*Tracer, error) {
	if !cfg.Enabled {
		return &Tracer{enabled: false}, nil
	}

	// Create gRPC connection
	conn, err := grpc.NewClient(
		cfg.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithGRPCConn(conn),
		otlptracegrpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		exporter.Shutdown(ctx)
		conn.Close()
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider with batch exporter
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithExportTimeout(10*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRatio)),
	)

	// Set global propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	tracer := tp.Tracer(cfg.ServiceName)

	return &Tracer{
		tp:      tp,
		tracer:  tracer,
		enabled: true,
		grpcConn: conn,
		shutdownFunc: func(ctx context.Context) error {
			return tp.Shutdown(ctx)
		},
	}, nil
}

// Tracer returns the underlying tracer
func (t *Tracer) Tracer() trace.Tracer {
	return t.tracer
}

// StartSpan starts a new span
func (t *Tracer) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !t.enabled {
		return ctx, trace.SpanFromContext(ctx)
	}
	return t.tracer.Start(ctx, spanName, opts...)
}

// Shutdown shuts down the tracer
func (t *Tracer) Shutdown(ctx context.Context) error {
	if !t.enabled {
		return nil
	}
	
	err := t.shutdownFunc(ctx)
	if t.grpcConn != nil {
		t.grpcConn.Close()
	}
	return err
}

// IsEnabled returns whether tracing is enabled
func (t *Tracer) IsEnabled() bool {
	return t.enabled
}
