package cmd

import (
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Org-level data quality overview",
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "dashboard is not yet available in the public API")
	},
}

func init() {}
