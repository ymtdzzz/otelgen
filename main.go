package main

import (
	"context"
	"fmt"

	prompt "github.com/c-bata/go-prompt"
	"github.com/ymtdzzz/otelgen/completer"
	"github.com/ymtdzzz/otelgen/executor"
	"github.com/ymtdzzz/otelgen/telemetry"
)

func main() {
	defaultProvider, err := telemetry.CreateDefaultTracerProvider()
	if err != nil {
		panic(err)
	}

	telemetry.InitTracerManager(defaultProvider)
	go func() {
		if err := telemetry.GetTracerManager().Shutdown(context.Background()); err != nil {
			fmt.Printf("Error shutting down tracer manager: %v\n", err)
		}
	}()

	telemetry.InitStore()

	fmt.Println("OpenTelemetry Trace CLI (type 'exit' to quit)")
	p := prompt.New(executor.Executor, completer.Completer, prompt.OptionPrefix("otelgen> "))
	p.Run()
}
