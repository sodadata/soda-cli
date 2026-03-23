package main

import (
	"fmt"
	"os"
	"time"

	"github.com/soda-data-inc/soda-cli/cmd"
	"github.com/soda-data-inc/soda-cli/internal/output"
	"github.com/soda-data-inc/soda-cli/internal/telemetry"
)

func main() {
	start := time.Now()

	err := cmd.Execute()

	// Send telemetry (runs in background goroutine)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*output.ExitError); ok {
			exitCode = exitErr.Code
		} else {
			exitCode = 2
		}
	}
	waitForTelemetry := telemetry.Send(telemetry.Event{
		Command:    telemetry.CommandName(os.Args),
		Flags:      telemetry.FlagNames(os.Args),
		ExitCode:   exitCode,
		DurationMs: time.Since(start).Milliseconds(),
		CLIVersion: cmd.Version,
	})

	if err != nil {
		if exitErr, ok := err.(*output.ExitError); ok {
			if exitErr.Msg != "" {
				fmt.Fprintln(os.Stderr, output.Red.Render("[Error]")+" "+exitErr.Msg)
			}
			waitForTelemetry()
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, output.Red.Render("[Error]")+" "+err.Error())
		waitForTelemetry()
		os.Exit(2)
	}

	waitForTelemetry()
}
