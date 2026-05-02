package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var datasetOnboardCmd = &cobra.Command{
	Use:   "onboard <dataset-id>",
	Short: "Guided setup: enable monitors, profiling and contracts for a dataset",
	Long: `Set up a dataset with default monitors, profiling and optionally generate a contract.

Interactive mode walks through each step. Use flags for CI/CD or AI agents:

  sodacli dataset onboard <id> --monitoring --profiling --contracts skeleton`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		datasetID := args[0]
		hasMonitoring := cmd.Flags().Changed("monitoring") || cmd.Flags().Changed("no-monitoring")
		hasProfiling := cmd.Flags().Changed("profiling") || cmd.Flags().Changed("no-profiling")
		hasContracts := cmd.Flags().Changed("contracts")
		hasFailedRows := cmd.Flags().Changed("collect-failed-rows") || cmd.Flags().Changed("no-collect-failed-rows") || cmd.Flags().Changed("unique-keys")
		noInteractive := GCtx.NoInteractive || (hasMonitoring && hasProfiling && hasContracts)

		enableMonitoring, _ := cmd.Flags().GetBool("monitoring")
		noMonitoring, _ := cmd.Flags().GetBool("no-monitoring")
		enableProfiling, _ := cmd.Flags().GetBool("profiling")
		noProfiling, _ := cmd.Flags().GetBool("no-profiling")
		contractsMode, _ := cmd.Flags().GetString("contracts")
		enableCollectFailedRows, _ := cmd.Flags().GetBool("collect-failed-rows")
		uniqueKeys, _ := cmd.Flags().GetStringSlice("unique-keys")

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
		var qualifiedName string
		for _, d := range datasets.Content {
			if d.ID == datasetID {
				datasetName = d.Name
				qualifiedName = d.Datasource.Name + "/" + strings.ReplaceAll(d.QualifiedName, ".", "/")
				break
			}
		}
		if datasetName == "" {
			// Not onboarded yet — look across discovered datasets and promote on the fly.
			dsPage, dsErr := client.ListDatasources(0, 500)
			if dsErr != nil {
				return dsErr
			}
			var foundDatasourceID string
			var foundDiscovered *api.DiscoveredDataset
			for _, ds := range dsPage.Content {
				discPage, discErr := client.ListDiscoveredDatasets(ds.ID, 0, 500)
				if discErr != nil {
					continue
				}
				for i := range discPage.Content {
					d := &discPage.Content[i]
					if d.ID == datasetID {
						foundDatasourceID = ds.ID
						foundDiscovered = d
						break
					}
				}
				if foundDiscovered != nil {
					break
				}
			}
			if foundDiscovered == nil {
				return output.Errorf(2, "dataset '%s' not found", datasetID)
			}
			fmt.Println(output.Dim.Render("  Onboarding dataset..."))
			if err := client.OnboardDiscoveredDatasets(foundDatasourceID, api.OnboardDatasetsRequest{
				DiscoveredDatasetIDs: []string{datasetID},
			}); err != nil {
				return err
			}
			// Re-fetch via the standard endpoint so qualifiedName matches the
			// format used by the already-onboarded path (DiscoveredDataset
			// includes the datasource prefix; Dataset does not).
			detail, err := client.GetDataset(datasetID)
			if err != nil {
				return err
			}
			datasetName = detail.Name
			qualifiedName = detail.Datasource.Name + "/" + strings.ReplaceAll(detail.QualifiedName, ".", "/")
		}
		fmt.Printf("  Dataset: %s\n\n", output.Bold.Render(datasetName))

		// ── Determine settings ──────────────────────────────────────────────

		if !hasMonitoring && !hasProfiling && !hasContracts && !hasFailedRows {
			if noInteractive {
				return output.Errorf(2, "flags required in non-interactive mode: --monitoring/--no-monitoring, --profiling/--no-profiling, --contracts copilot|skeleton|none")
			}

			monitoringChoice := "yes"
			profilingChoice := "yes"
			contractChoice := "none"
			failedRowsChoice := "no"
			uniqueKeysInput := ""

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Enable default metric monitoring?").
						Description("Row count, row count change, freshness, schema changes,\npartition row count, most recent timestamp.").
						Options(
							huh.NewOption("Yes", "yes"),
							huh.NewOption("No", "no"),
						).
						Value(&monitoringChoice),
					huh.NewSelect[string]().
						Title("Enable dataset profiling?").
						Description("Column stats, row counts, and data type distribution.").
						Options(
							huh.NewOption("Yes", "yes"),
							huh.NewOption("No", "no"),
						).
						Value(&profilingChoice),
					huh.NewSelect[string]().
						Title("Set up a data contract?").
						Options(
							huh.NewOption("AI-generated contract (Copilot)", "copilot"),
							huh.NewOption("Skeleton contract (empty template)", "skeleton"),
							huh.NewOption("No contract", "none"),
						).
						Value(&contractChoice),
					huh.NewSelect[string]().
						Title("Enable failed rows collection?").
						Description("Store rows that fail checks in the diagnostics warehouse.\nRequires unique key columns.").
						Options(
							huh.NewOption("Yes", "yes"),
							huh.NewOption("No", "no"),
						).
						Value(&failedRowsChoice),
				),
				huh.NewGroup(
					huh.NewInput().
						Title("Unique key columns").
						Description("Comma-separated list, e.g. id,customer_email").
						Value(&uniqueKeysInput),
				).WithHideFunc(func() bool {
					return failedRowsChoice != "yes"
				}),
			)
			if err := form.Run(); err != nil {
				return output.Errorf(2, "onboarding cancelled")
			}
			enableMonitoring = monitoringChoice == "yes"
			enableProfiling = profilingChoice == "yes"
			contractsMode = contractChoice
			enableCollectFailedRows = failedRowsChoice == "yes"
			if enableCollectFailedRows {
				for _, k := range strings.Split(uniqueKeysInput, ",") {
					if k = strings.TrimSpace(k); k != "" {
						uniqueKeys = append(uniqueKeys, k)
					}
				}
			}
		} else {
			if noMonitoring {
				enableMonitoring = false
			}
			if noProfiling {
				enableProfiling = false
			}
			if contractsMode == "" {
				contractsMode = "none"
			}
			// Treat --unique-keys alone as implicit --collect-failed-rows.
			if len(uniqueKeys) > 0 {
				enableCollectFailedRows = true
			}
		}

		if enableCollectFailedRows && len(uniqueKeys) == 0 {
			return output.Errorf(2, "--unique-keys is required when --collect-failed-rows is set (failed rows collection won't work without unique key columns)")
		}

		// ── Execute ─────────────────────────────────────────────────────────

		// Step 1: Monitoring + Profiling
		if enableMonitoring || enableProfiling {
			label := ""
			switch {
			case enableMonitoring && enableProfiling:
				label = "Enabling monitoring and profiling..."
			case enableMonitoring:
				label = "Enabling default metric monitoring..."
			default:
				label = "Enabling dataset profiling..."
			}
			fmt.Println(output.Dim.Render("  " + label))
			if err := client.EnableDatasetDefaults(datasetID, enableMonitoring, enableProfiling); err != nil {
				fmt.Fprintf(os.Stderr, "  %s Could not enable settings: %v\n", output.Yellow.Render("⚠"), err)
			} else {
				fmt.Println(output.Green.Render("  ✓") + " " + label[:len(label)-3] + "d.")
			}
		} else {
			fmt.Println(output.Dim.Render("  Skipping monitoring and profiling setup."))
		}

		// Step 2: Failed rows collection
		if enableCollectFailedRows {
			fmt.Println(output.Dim.Render("  Enabling failed rows collection..."))
			enabled := true
			cfg := api.DiagnosticsWarehouseConfig{
				ScanAndResultsConfiguration: &api.DiagnosticsScanConfig{Enabled: &enabled},
				FailedRowsConfiguration: &api.DiagnosticsFailedRowsConfig{
					Enabled:              &enabled,
					UniqueKeyColumnNames: uniqueKeys,
				},
			}
			if _, err := client.UpdateDatasetDiagnostics(datasetID, cfg); err != nil {
				fmt.Fprintf(os.Stderr, "  %s Could not enable failed rows collection: %v\n", output.Yellow.Render("⚠"), err)
				if isNotEnabledOnDatasource(err) {
					fmt.Fprintf(os.Stderr, "  %s\n", output.Dim.Render("Set up the diagnostics warehouse on the datasource first:"))
					fmt.Fprintf(os.Stderr, "  %s\n", output.Dim.Render("  sodacli datasource diagnostics <datasource-id> --enable"))
				}
			} else {
				fmt.Println(output.Green.Render("  ✓") + fmt.Sprintf(" Failed rows collection enabled (keys: %s).", strings.Join(uniqueKeys, ", ")))
			}
		}

		// Step 3: Contracts
		var contractFile string
		switch contractsMode {
		case "copilot":
			if qualifiedName == "" {
				fmt.Fprintf(os.Stderr, "  %s Cannot generate AI contract: dataset qualified name not available.\n", output.Yellow.Render("⚠"))
			} else {
				outFile := datasetFileName(qualifiedName)
				if err := runContractCreateCopilot(client, qualifiedName, outFile, false); err != nil {
					fmt.Fprintf(os.Stderr, "  %s Contract generation failed: %v\n", output.Yellow.Render("⚠"), err)
				} else {
					contractFile = outFile
				}
			}
		case "skeleton":
			if qualifiedName == "" {
				fmt.Fprintf(os.Stderr, "  %s Cannot generate skeleton contract: dataset qualified name not available.\n", output.Yellow.Render("⚠"))
			} else {
				outFile := datasetFileName(qualifiedName)
				if err := runContractCreateSkeleton(client, qualifiedName, outFile); err != nil {
					fmt.Fprintf(os.Stderr, "  %s Contract generation failed: %v\n", output.Yellow.Render("⚠"), err)
				} else {
					contractFile = outFile
				}
			}
		case "none":
			fmt.Println(output.Dim.Render("  Skipping contract setup."))
		default:
			return output.Errorf(2, "unknown contracts mode '%s' — use copilot, skeleton, or none", contractsMode)
		}

		// Step 4: Verify contract
		if contractFile != "" {
			fmt.Println()
			fmt.Println(output.Dim.Render("  Verifying contract..."))
			if err := runContractVerify(client, contractFile, false); err != nil {
				fmt.Fprintf(os.Stderr, "  %s Verification failed: %v\n", output.Yellow.Render("⚠"), err)
			}
		}

		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Dataset '%s' onboarding complete.", datasetName), GCtx)
		return nil
	},
}

// datasetFileName returns "tablename.yml" from a qualified name like "ds.schema.table" or "ds/db/schema/table".
func datasetFileName(qualifiedName string) string {
	sep := "."
	if strings.Contains(qualifiedName, "/") {
		sep = "/"
	}
	parts := strings.Split(qualifiedName, sep)
	return parts[len(parts)-1] + ".yml"
}

func init() {
	datasetOnboardCmd.Flags().Bool("monitoring", false, "Enable default metric monitors")
	datasetOnboardCmd.Flags().Bool("no-monitoring", false, "Skip monitoring setup")
	datasetOnboardCmd.Flags().Bool("profiling", false, "Enable dataset profiling")
	datasetOnboardCmd.Flags().Bool("no-profiling", false, "Skip profiling setup")
	datasetOnboardCmd.Flags().String("contracts", "", "Generate contract: copilot|skeleton|none")
	datasetOnboardCmd.Flags().Bool("collect-failed-rows", false, "Enable failed rows collection (requires --unique-keys)")
	datasetOnboardCmd.Flags().Bool("no-collect-failed-rows", false, "Skip failed rows collection setup")
	datasetOnboardCmd.Flags().StringSlice("unique-keys", nil, "Unique key columns for failed rows collection (comma-separated or repeated)")

	datasetCmd.AddCommand(datasetOnboardCmd)
}
