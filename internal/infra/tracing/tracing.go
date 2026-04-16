package tracing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/iVampireSP/go-template/internal/infra/config"
	"github.com/iVampireSP/go-template/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

type Tracing struct {
	exporter       sdktrace.SpanExporter
	sampleRatio    float64
	providers      map[string]*sdktrace.TracerProvider
	providersMutex sync.Mutex

	queryTraceURL *url.URL
	queryUsername string
	queryPassword string
	queryClient   *http.Client
}

var (
	globalTracing     *Tracing
	globalTracingErr  error
	globalTracingOnce sync.Once
)

func NewTracing() (*Tracing, error) {
	globalTracingOnce.Do(func() {
		globalTracing, globalTracingErr = newTracing()
	})
	if globalTracingErr != nil {
		return nil, globalTracingErr
	}
	if globalTracing == nil {
		return nil, ErrTracingNotInitialized
	}
	return globalTracing, nil
}

func newTracing() (*Tracing, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	promExporter, err := otelprom.New()
	if err != nil {
		return nil, err
	}
	mp := metric.NewMeterProvider(metric.WithReader(promExporter))
	otel.SetMeterProvider(mp)

	if !config.Bool("tracing.enabled", false) {
		return nil, ErrTracingDisabled
	}

	endpoint := config.String("tracing.endpoint", "localhost:4317")
	sampleRatio := config.GetFloat("tracing.sample_ratio", 1.0)

	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	queryBaseURL, err := url.Parse(config.String("tracing.query_url", "http://jaeger-query:16686"))
	if err != nil {
		return nil, fmt.Errorf("invalid tracing query url: %w", err)
	}
	if queryBaseURL.Scheme != "http" && queryBaseURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid tracing query url scheme: %s", queryBaseURL.Scheme)
	}
	if queryBaseURL.Host == "" {
		return nil, fmt.Errorf("invalid tracing query url host")
	}

	queryTimeout := config.Duration("tracing.query_timeout", 15*time.Second)
	queryTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	t := &Tracing{
		exporter:      exporter,
		sampleRatio:   sampleRatio,
		providers:     make(map[string]*sdktrace.TracerProvider),
		queryTraceURL: queryBaseURL.ResolveReference(&url.URL{Path: "/api/traces/"}),
		queryUsername: config.String("tracing.query_username"),
		queryPassword: config.String("tracing.query_password"),
		queryClient: &http.Client{
			Timeout:   queryTimeout,
			Transport: queryTransport,
		},
	}

	logger.Info("Tracing enabled", "endpoint", endpoint, "sample_ratio", sampleRatio)

	return t, nil
}

func GetService(serviceName string) (*sdktrace.TracerProvider, error) {
	t, err := NewTracing()
	if err != nil {
		if errors.Is(err, ErrTracingDisabled) {
			return nil, nil
		}
		return nil, err
	}
	return t.GetService(serviceName)
}

func (t *Tracing) GetService(serviceName string) (*sdktrace.TracerProvider, error) {
	t.providersMutex.Lock()
	defer t.providersMutex.Unlock()

	if tp, ok := t.providers[serviceName]; ok {
		return tp, nil
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(t.exporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(t.sampleRatio))),
	)
	otel.SetTracerProvider(tp)
	t.providers[serviceName] = tp

	return tp, nil
}

func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) {
	if tp == nil {
		return
	}
	if err := tp.Shutdown(ctx); err != nil {
		logger.Error("Tracing shutdown", "error", err)
	}
}

// ShutdownWithTimeout 带 10 秒超时的关闭，防止无限阻塞导致进程无法退出。
func ShutdownWithTimeout(tp *sdktrace.TracerProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	Shutdown(ctx, tp)
}
