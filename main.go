package main

import (
	"context"
	"fmt"

	prompt "github.com/c-bata/go-prompt"
	"github.com/ymtdzzz/otelgen/completer"
	"github.com/ymtdzzz/otelgen/executor"
	"github.com/ymtdzzz/otelgen/telemetry"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	exporterFn := func() (sdktrace.SpanExporter, error) {
		return otlptracegrpc.New(context.Background(),
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint("localhost:4317"),
		)
	}

	telemetry.InitTracerManager(exporterFn, nil)
	defer func() {
		if err := telemetry.GetTracerManager().Shutdown(context.Background()); err != nil {
			fmt.Printf("Error shutting down tracer manager: %v\n", err)
		}
	}()

	telemetry.InitStore()

	fmt.Println("OpenTelemetry CLI generator (type 'exit' to quit)")
	p := prompt.New(executor.Executor, completer.Completer, prompt.OptionPrefix("otelgen> "))
	p.Run()
}
