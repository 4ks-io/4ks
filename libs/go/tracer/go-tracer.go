package tracing

import (
	"log"
	"strings"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	resource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// Config contains tracing exporter settings loaded by the caller.
type Config struct {
	ExporterType       string
	JaegerEndpoint     string
	GoogleCloudProject string
	ServiceName        string
}

// InitTracerProvider configures the process tracer provider once at startup.
func InitTracerProvider(cfg Config) *sdktrace.TracerProvider {
	var exporter sdktrace.SpanExporter
	var err error

	if cfg.ServiceName == "" {
		cfg.ServiceName = "4ks-api"
	}

	exporterType := strings.ToUpper(cfg.ExporterType)
	if exporterType == "" {
		exporterType = "CONSOLE"
	}
	switch exporterType {
	case "GOOGLE":
		exporter, err = texporter.New(texporter.WithProjectID(cfg.GoogleCloudProject))
	case "JAEGER":
		jaegerEndpoint := cfg.JaegerEndpoint
		if jaegerEndpoint == "" {
			jaegerEndpoint = "http://jaeger:14268/api/traces"
		}
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	case "CONSOLE":
		exporter, err = stdout.New(stdout.WithPrettyPrint())
	default:
		exporter, err = stdout.New(stdout.WithPrettyPrint())
	}

	if err != nil {
		log.Fatal(err)
	}

	// r, err := resource.Merge(
	// 	resource.Default(),
	// 	resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceNameKey.String("ExampleService")),
	// )
	// if err != nil {
	// 	panic(err)
	// }

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func NewTracer(tracerName string) trace.Tracer {
	return otel.Tracer(tracerName)
}
