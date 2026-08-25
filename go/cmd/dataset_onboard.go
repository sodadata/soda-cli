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

// datasetInfo carries everything we need about each dataset between resolve
// and execute phases.
type datasetInfo struct {
	ID            string
	Name          string
	QualifiedName string // canonical "datasource/db/schema/table"
	Onboarded     bool
	DatasourceID  string // populated only when promotion is needed
}

var datasetOnboardCmd = &cobra.Command{
	Use:   "onboard [dataset-id]",
	Short: "Guided setup: enable monitors, profiling and contracts for one or more datasets",
	Long: `Set up one or more datasets with default monitors, profiling and optionally generate contracts.

Single-dataset mode walks through interactive prompts:

  sodacli dataset onboard <id>

Bulk mode (multiple datasets via --dataset, repeatable) requires non-interactive flags:

  sodacli dataset onboard <id1> --dataset <id2> --dataset <id3> \
      --monitoring --no-profiling --contracts copilot

Failed-rows collection (--collect-failed-rows / --unique-keys) is only supported
in single-dataset mode, since unique keys are dataset-specific.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// ── Collect dataset IDs (positional + repeatable --dataset) ─────────
		ids := []string{}
		if len(args) == 1 && strings.TrimSpace(args[0]) != "" {
			ids = append(ids, strings.TrimSpace(args[0]))
		}
		extra, _ := cmd.Flags().GetStringArray("dataset")
		for _, e := range extra {
			if e = strings.TrimSpace(e); e != "" {
				ids = append(ids, e)
			}
		}
		seen := map[string]bool{}
		dedup := ids[:0]
		for _, id := range ids {
			if !seen[id] {
				seen[id] = true
				dedup = append(dedup, id)
			}
		}
		ids = dedup
		if len(ids) == 0 {
			return output.Errorf(2, "at least one dataset ID is required (positional or --dataset)")
		}
		bulk := len(ids) > 1

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

		// Bulk-mode constraints
		if bulk {
			if !hasMonitoring || !hasProfiling || !hasContracts {
				return output.Errorf(2, "bulk mode (multiple datasets) requires --monitoring/--no-monitoring, --profiling/--no-profiling, and --contracts copilot|skeleton|none")
			}
			if hasFailedRows {
				return output.Errorf(2, "--collect-failed-rows / --unique-keys are not supported in bulk mode (run dataset onboard one at a time for failed-rows setup, since unique keys are dataset-specific)")
			}
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		// ── Resolve all dataset IDs ─────────────────────────────────────────
		fmt.Println(output.Dim.Render(fmt.Sprintf("  Checking %d dataset(s)...", len(ids))))
		infoByID := make(map[string]*datasetInfo, len(ids))
		for _, id := range ids {
			infoByID[id] = &datasetInfo{ID: id}
		}

		// Sweep already-onboarded datasets via paginated ListDatasets.
		unresolved := len(ids)
		page := 0
		for unresolved > 0 {
			datasets, err := client.ListDatasets(api.ListDatasetsParams{Size: 500, Page: page})
			if err != nil {
				return err
			}
			for _, d := range datasets.Content {
				if i, ok := infoByID[d.ID]; ok && !i.Onboarded {
					i.Name = d.Name
					i.QualifiedName = d.Datasource.Name + "/" + strings.ReplaceAll(d.QualifiedName, ".", "/")
					i.Onboarded = true
					unresolved--
				}
			}
			if datasets.Last || len(datasets.Content) == 0 {
				break
			}
			page++
		}

		// Anything still unresolved → look across discovered datasets per datasource.
		if unresolved > 0 {
			dsPage, dsErr := client.ListDatasources(0, 500)
			if dsErr != nil {
				return dsErr
			}
			for _, ds := range dsPage.Content {
				if unresolved == 0 {
					break
				}
				discPage, discErr := client.ListDiscoveredDatasets(ds.ID, 0, 500)
				if discErr != nil {
					continue
				}
				for i := range discPage.Content {
					d := &discPage.Content[i]
					if info, ok := infoByID[d.ID]; ok && !info.Onboarded && info.DatasourceID == "" {
						info.DatasourceID = ds.ID
						info.Name = d.Name
						unresolved--
					}
				}
			}
		}

		var notFound []string
		for _, id := range ids {
			if !infoByID[id].Onboarded && infoByID[id].DatasourceID == "" {
				notFound = append(notFound, id)
			}
		}
		if len(notFound) > 0 {
			return output.Errorf(2, "dataset(s) not found: %s", strings.Join(notFound, ", "))
		}

		// ── Promote any not-yet-onboarded datasets, batched per datasource ──
		toPromoteByDS := map[string][]string{}
		for _, id := range ids {
			if !infoByID[id].Onboarded {
				toPromoteByDS[infoByID[id].DatasourceID] = append(toPromoteByDS[infoByID[id].DatasourceID], id)
			}
		}
		if len(toPromoteByDS) > 0 {
			n := 0
			for _, v := range toPromoteByDS {
				n += len(v)
			}
			fmt.Println(output.Dim.Render(fmt.Sprintf("  Onboarding %d discovered dataset(s)...", n)))
			for dsID, idList := range toPromoteByDS {
				if err := client.OnboardDiscoveredDatasets(dsID, api.OnboardDatasetsRequest{
					DiscoveredDatasetIDs: idList,
				}); err != nil {
					return err
				}
			}
			// Re-fetch each via the standard endpoint so qualifiedName matches
			// the format used by the already-onboarded path (DiscoveredDataset
			// includes the datasource prefix; Dataset does not).
			for _, id := range ids {
				if infoByID[id].Onboarded {
					continue
				}
				detail, err := client.GetDataset(id)
				if err != nil {
					return err
				}
				infoByID[id].Name = detail.Name
				infoByID[id].QualifiedName = detail.Datasource.Name + "/" + strings.ReplaceAll(detail.QualifiedName, ".", "/")
				infoByID[id].Onboarded = true
			}
		}

		// Print resolved datasets
		if bulk {
			fmt.Printf("  Datasets (%d):\n", len(ids))
			for _, id := range ids {
				fmt.Printf("    • %s\n", infoByID[id].Name)
			}
			fmt.Println()
		} else {
			fmt.Printf("  Dataset: %s\n\n", output.Bold.Render(infoByID[ids[0]].Name))
		}

		// ── Determine settings (interactive form only valid for single-dataset) ──
		if !bulk && !hasMonitoring && !hasProfiling && !hasContracts && !hasFailedRows {
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
			// Treat --unique-keys alone as implicit --collect-failed-rows (single-mode only).
			if !bulk && len(uniqueKeys) > 0 {
				enableCollectFailedRows = true
			}
		}

		if enableCollectFailedRows && len(uniqueKeys) == 0 {
			return output.Errorf(2, "--unique-keys is required when --collect-failed-rows is set (failed rows collection won't work without unique key columns)")
		}

		// ── Execute ─────────────────────────────────────────────────────────

		// Step 1: Monitoring + Profiling (per-dataset API call)
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
			var hadErr bool
			for _, id := range ids {
				if err := client.EnableDatasetDefaults(id, enableMonitoring, enableProfiling); err != nil {
					fmt.Fprintf(os.Stderr, "  %s [%s] %v\n", output.Yellow.Render("⚠"), infoByID[id].Name, err)
					hadErr = true
				}
			}
			if !hadErr {
				fmt.Println(output.Green.Render("  ✓") + " " + label[:len(label)-3] + "d.")
			}
		} else {
			fmt.Println(output.Dim.Render("  Skipping monitoring and profiling setup."))
		}

		// Step 2: Failed rows (single-mode only — bulk-mode constraint above blocks this)
		if enableCollectFailedRows {
			id := ids[0]
			fmt.Println(output.Dim.Render("  Enabling failed rows collection..."))
			enabled := true
			cfg := api.DiagnosticsWarehouseConfig{
				ScanAndResultsConfiguration: &api.DiagnosticsScanConfig{Enabled: &enabled},
				FailedRowsConfiguration: &api.DiagnosticsFailedRowsConfig{
					Enabled:              &enabled,
					UniqueKeyColumnNames: uniqueKeys,
				},
			}
			if _, err := client.UpdateDatasetDiagnostics(id, cfg); err != nil {
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
		var contractFiles []string
		switch contractsMode {
		case "copilot":
			qns := make([]string, 0, len(ids))
			outFiles := make(map[string]string, len(ids))
			for _, id := range ids {
				qn := infoByID[id].QualifiedName
				qns = append(qns, qn)
				outFiles[qn] = datasetFileName(qn)
			}
			files, err := runContractCreateCopilotBulk(client, qns, outFiles, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  %s Contract generation failed: %v\n", output.Yellow.Render("⚠"), err)
			} else {
				contractFiles = append(contractFiles, files...)
			}
		case "skeleton":
			for _, id := range ids {
				qn := infoByID[id].QualifiedName
				outFile := datasetFileName(qn)
				if err := runContractCreateSkeleton(client, qn, outFile); err != nil {
					fmt.Fprintf(os.Stderr, "  %s [%s] Skeleton generation failed: %v\n", output.Yellow.Render("⚠"), infoByID[id].Name, err)
				} else {
					contractFiles = append(contractFiles, outFile)
				}
			}
		case "none":
			fmt.Println(output.Dim.Render("  Skipping contract setup."))
		default:
			return output.Errorf(2, "unknown contracts mode '%s' — use copilot, skeleton, or none", contractsMode)
		}

		// Step 4: Verify contracts
		if len(contractFiles) > 0 {
			fmt.Println()
			fmt.Println(output.Dim.Render(fmt.Sprintf("  Verifying %d contract(s)...", len(contractFiles))))
			for _, f := range contractFiles {
				if err := runContractVerify(client, f, false); err != nil {
					fmt.Fprintf(os.Stderr, "  %s [%s] Verification failed: %v\n", output.Yellow.Render("⚠"), f, err)
				}
			}
		}

		fmt.Println()
		if bulk {
			output.PrintSuccess(fmt.Sprintf("Onboarded %d datasets.", len(ids)), GCtx)
		} else {
			output.PrintSuccess(fmt.Sprintf("Dataset '%s' onboarding complete.", infoByID[ids[0]].Name), GCtx)
		}
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
	datasetOnboardCmd.Flags().StringArray("dataset", nil, "Additional dataset ID to onboard (repeatable, enables bulk mode)")
	datasetOnboardCmd.Flags().Bool("monitoring", false, "Enable default metric monitors")
	datasetOnboardCmd.Flags().Bool("no-monitoring", false, "Skip monitoring setup")
	datasetOnboardCmd.Flags().Bool("profiling", false, "Enable dataset profiling")
	datasetOnboardCmd.Flags().Bool("no-profiling", false, "Skip profiling setup")
	datasetOnboardCmd.Flags().String("contracts", "", "Generate contract: copilot|skeleton|none")
	datasetOnboardCmd.Flags().Bool("collect-failed-rows", false, "Enable failed rows collection (single-dataset only; requires --unique-keys)")
	datasetOnboardCmd.Flags().Bool("no-collect-failed-rows", false, "Skip failed rows collection setup")
	datasetOnboardCmd.Flags().StringSlice("unique-keys", nil, "Unique key columns for failed rows collection (single-dataset only; comma-separated or repeated)")

	datasetCmd.AddCommand(datasetOnboardCmd)
}
