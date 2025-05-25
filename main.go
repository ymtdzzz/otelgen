package main

import (
	"context"
	"fmt"

	prompt "github.com/c-bata/go-prompt"
	"github.com/ymtdzzz/otelgen/completer"
	"github.com/ymtdzzz/otelgen/executor"
	"github.com/ymtdzzz/otelgen/telemetry"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var tracerProvider *sdktrace.TracerProvider

func initExporter() error {
	ctx := context.Background()
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		return err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.Default()),
	)
	otel.SetTracerProvider(tp)
	tracerProvider = tp
	return nil
}

func main() {
	if err := initExporter(); err != nil {
		panic(err)
	}
	telemetry.InitStore()
	fmt.Println("OpenTelemetry Trace CLI (type 'exit' to quit)")
	p := prompt.New(executor.Executor, completer.Completer, prompt.OptionPrefix("otelgen> "))
	p.Run()
}
