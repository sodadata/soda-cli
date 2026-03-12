package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Org-level data quality overview",
	RunE: func(cmd *cobra.Command, args []string) error {
		d := mock.DashboardSummary

		if output.EffectiveFmt(GCtx) == "json" {
			output.Render([]map[string]string{d}, []string{
				"organization", "profile", "datasources", "datasets",
				"passing_datasets", "failing_datasets", "open_incidents", "jobs_today",
			}, nil, GCtx)
			return nil
		}

		fmt.Printf("  %-22s %s\n", output.Bold.Render("Organization"), d["organization"])
		fmt.Printf("  %-22s %s\n", output.Bold.Render("Profile"), d["profile"])
		fmt.Println()
		fmt.Printf("  %-22s %s\n", "Datasources", d["datasources"])
		fmt.Printf("  %-22s %s\n", "Datasets", d["datasets"])
		fmt.Printf("  %-22s %s  /  %s failing\n",
			"Dataset health",
			output.Green.Render(d["passing_datasets"]+" passing"),
			output.Red.Render(d["failing_datasets"]),
		)
		fmt.Printf("  %-22s %s\n", "Open incidents", output.Red.Render(d["open_incidents"]))
		fmt.Printf("  %-22s %s\n", "Jobs today", d["jobs_today"])
		return nil
	},
}

func init() {}
