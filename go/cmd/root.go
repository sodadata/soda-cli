// Package cmd contains all CLI commands.
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/ctx"
)

// GCtx is the global context, populated from persistent flags before any RunE.
var GCtx = &ctx.GlobalCtx{}

var rootCmd = &cobra.Command{
	Use:           "soda",
	Short:         "Soda — data quality CLI",
	Long:          "Run checks, manage contracts, and monitor your data.\n\nDocs: https://docs.soda.io/soda-cli",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&GCtx.Output, "output", "o", "auto", "Output format: table|json|csv")
	rootCmd.PersistentFlags().StringVar(&GCtx.Profile, "profile", "", "Override active auth profile")
	rootCmd.PersistentFlags().BoolVar(&GCtx.NoColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVarP(&GCtx.Quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&GCtx.Verbose, "verbose", "v", false, "Show detailed output")
	rootCmd.PersistentFlags().BoolVar(&GCtx.NoInteractive, "no-interactive", false, "Never prompt; fail with clear error if input missing")
	rootCmd.Version = "0.1.0-dev"
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
