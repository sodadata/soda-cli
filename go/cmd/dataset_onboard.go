package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var datasetOnboardCmd = &cobra.Command{
	Use:   "onboard <dataset-id>",
	Short: "Guided setup: enable monitors and contracts for a dataset",
	Long: `Set up a dataset with default monitors and optionally generate a contract.

Interactive mode walks through each step. Use flags for CI/CD or AI agents:

  soda dataset onboard <id> --monitoring --contracts skeleton`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		datasetID := args[0]
		monitoringFlag := cmd.Flags().Changed("monitoring") || cmd.Flags().Changed("no-monitoring")
		contractsFlag := cmd.Flags().Changed("contracts")

		enableMonitoring, _ := cmd.Flags().GetBool("monitoring")
		noMonitoring, _ := cmd.Flags().GetBool("no-monitoring")
		contractsMode, _ := cmd.Flags().GetString("contracts")

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		// Validate dataset exists
		fmt.Println(output.Dim.Render("  Checking dataset..."))
		datasets, err := client.ListDatasets(api.ListDatasetsParams{Size: 500})
		if err != nil {
			return err
		}
		var datasetName string
		for _, d := range datasets.Content {
			if d.ID == datasetID {
				datasetName = d.Name
				break
			}
		}
		if datasetName == "" {
			return output.Errorf(2, "dataset '%s' not found", datasetID)
		}
		fmt.Printf("  Dataset: %s\n\n", output.Bold.Render(datasetName))

		// ── Determine settings ──────────────────────────────────────────────

		if !monitoringFlag && !contractsFlag {
			// Interactive mode
			if GCtx.NoInteractive {
				return output.Errorf(2, "flags required in non-interactive mode: --monitoring/--no-monitoring and --contracts ai|skeleton|none")
			}

			monitoringChoice := "yes"
			contractChoice := "none"

			form := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().
					Title("Enable default metric monitoring?").
					Description("Row count, row count change, freshness, schema changes,\npartition row count, most recent timestamp.\nYou can always add or remove monitors afterwards.").
					Options(
						huh.NewOption("Yes", "yes"),
						huh.NewOption("No", "no"),
					).
					Value(&monitoringChoice),
				huh.NewSelect[string]().
					Title("Set up a data contract?").
					Description("You can always add, remove or modify contracts afterwards.").
					Options(
						huh.NewOption("AI-generated contract (Copilot)", "ai"),
						huh.NewOption("Skeleton contract (empty template)", "skeleton"),
						huh.NewOption("No contract", "none"),
					).
					Value(&contractChoice),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "onboarding cancelled")
			}
			enableMonitoring = monitoringChoice == "yes"
			contractsMode = contractChoice
		} else {
			// Flags mode
			if noMonitoring {
				enableMonitoring = false
			}
			if contractsMode == "" {
				contractsMode = "none"
			}
		}

		// ── Execute ─────────────────────────────────────────────────────────

		// Step 1: Monitoring
		if enableMonitoring {
			fmt.Printf("  %s Enabling default dataset monitors is not yet available in the public API.\n", output.Yellow.Render("⚠"))
			fmt.Println(output.Dim.Render("    Enable monitors from the Soda Cloud UI, or use:"))
			fmt.Println(output.Dim.Render("    soda monitor add --dataset <id> --type column --column <col> --metric <metric>"))
		} else {
			fmt.Println(output.Dim.Render("  Skipping monitoring setup."))
		}

		// Step 2: Contracts
		switch contractsMode {
		case "ai":
			fmt.Println(output.Dim.Render("  AI-generated contracts are not yet available. Coming soon."))
		case "skeleton":
			fmt.Println(output.Dim.Render("  Skeleton contract generation is not yet available. Coming soon."))
		case "none":
			fmt.Println(output.Dim.Render("  Skipping contract setup."))
		default:
			return output.Errorf(2, "unknown contracts mode '%s' — use ai, skeleton, or none", contractsMode)
		}

		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Dataset '%s' onboarding complete.", datasetName), GCtx)
		return nil
	},
}

func init() {
	datasetOnboardCmd.Flags().Bool("monitoring", false, "Enable default metric monitors")
	datasetOnboardCmd.Flags().Bool("no-monitoring", false, "Skip monitoring setup")
	datasetOnboardCmd.Flags().String("contracts", "", "Generate contract: ai|skeleton|none")

	datasetCmd.AddCommand(datasetOnboardCmd)
}
