package main

import (
	"fmt"
	"os"

	"github.com/soda-data-inc/soda-cli/cmd"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if exitErr, ok := err.(*output.ExitError); ok {
			if exitErr.Msg != "" {
				fmt.Fprintln(os.Stderr, output.Red.Render("[Error]")+" "+exitErr.Msg)
			}
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, output.Red.Render("[Error]")+" "+err.Error())
		os.Exit(2)
	}
}
