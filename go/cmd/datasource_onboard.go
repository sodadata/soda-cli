package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

func boolPtr(b bool) *bool { return &b }

var dsOnboardCmd = &cobra.Command{
	Use:   "onboard <config-file-or-datasource-id>",
	Short: "Guided setup: create datasource + configure all datasets",
	Long: `Create or connect to a datasource, wait for dataset discovery,
then onboard discovered datasets with optional monitoring, profiling and contracts.

Pass a YAML config file to create a new datasource, or pass an existing
datasource ID to run the onboarding flow on an already-registered datasource.

When all action flags are provided the command runs fully non-interactively,
selecting all discovered datasets and applying the requested settings.

  sodacli datasource onboard config.yml --monitoring --profiling --contracts ai
  sodacli datasource onboard <datasource-id> --no-monitoring --no-profiling --contracts none`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := args[0]
		runnerID, _ := cmd.Flags().GetString("runner")

		// Infer non-interactive when all action flags are explicitly provided.
		hasMonitoring := cmd.Flags().Changed("monitoring") || cmd.Flags().Changed("no-monitoring")
		hasProfiling := cmd.Flags().Changed("profiling") || cmd.Flags().Changed("no-profiling")
		hasContracts := cmd.Flags().Changed("contracts")
		noInteractive := GCtx.NoInteractive || (hasMonitoring && hasProfiling && hasContracts)

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var datasourceID, name string

		if _, statErr := os.Stat(arg); statErr == nil {
			// ── arg is a config file: create a new datasource ────────────────

			configBytes, err := os.ReadFile(arg)
			if err != nil {
				return output.Errorf(2, "could not read config file: %v", err)
			}
			var configMap map[string]interface{}
			if err := yaml.Unmarshal(configBytes, &configMap); err != nil {
				return output.Errorf(2, "invalid YAML in %s: %v", arg, err)
			}
			name, _ = configMap["name"].(string)
			if name == "" {
				return output.Errorf(2, "'name' field is required in the config file")
			}

			// ── Step 1: Resolve runner ────────────────────────────────────────
			if runnerID == "" {
				runners, err := client.ListRunners(100)
				if err != nil {
					return output.Errorf(2, "--runner is required (could not list runners: %v)", err)
				}
				if len(runners.Content) == 0 {
					return output.Errorf(2, "--runner is required. No runners found — set up a runner in Soda Cloud first.")
				}
				if len(runners.Content) == 1 {
					runnerID = runners.Content[0].ID
					fmt.Printf("  Using runner: %s (%s)\n", output.Bold.Render(runners.Content[0].Name), runnerID)
				} else if noInteractive {
					fmt.Println("  Available runners:")
					for _, a := range runners.Content {
						fmt.Printf("    %s  %s\n", a.ID, a.Name)
					}
					return output.Errorf(2, "--runner is required (multiple runners found)")
				} else {
					options := make([]huh.Option[string], len(runners.Content))
					for i, a := range runners.Content {
						label := a.Name
						if a.Label != "" {
							label = a.Label + " (" + a.Name + ")"
						}
						options[i] = huh.NewOption(label, a.ID)
					}
					form := huh.NewForm(huh.NewGroup(
						huh.NewSelect[string]().
							Title("Which runner should route this connection?").
							Options(options...).
							Value(&runnerID),
					))
					if err := form.Run(); err != nil {
						return output.Errorf(2, "cancelled")
					}
				}
			}

			// ── Step 2: Create datasource ─────────────────────────────────────
			fmt.Println(output.Dim.Render("  Creating datasource '" + name + "'..."))
			createResult, err := client.CreateDatasource(api.CreateDatasourceRequest{
				Name:                      name,
				AgentID:                   runnerID,
				ConfigurationFileContents: string(configBytes),
			})
			if err != nil {
				return err
			}
			datasourceID = createResult.Datasource.ID
			fmt.Printf("  Datasource ID: %s\n", datasourceID)

		} else {
			// ── arg is a datasource ID: use existing datasource ───────────────
			datasourceID = arg
			page, err := client.ListDatasources(0, 500)
			if err != nil {
				return output.Errorf(2, "could not look up datasource: %v", err)
			}
			for _, ds := range page.Content {
				if ds.ID == datasourceID {
					name = ds.Name
					break
				}
			}
			if name == "" {
				return output.Errorf(2, "datasource '%s' not found", datasourceID)
			}
			fmt.Printf("  Datasource: %s (%s)\n", output.Bold.Render(name), datasourceID)
		}

		// ── Step 3: Discover datasets ────────────────────────────────────
		spinner := output.NewSpinner("Waiting for dataset discovery...")
		spinner.Start()
		discovered, err := pollDiscoveredDatasets(client, datasourceID, spinner)
		spinner.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s %v\n", output.Yellow.Render("⚠"), err)
			fmt.Println(output.Dim.Render("  Check with: sodacli dataset list --datasource " + name))
			fmt.Println()
			output.PrintSuccess(fmt.Sprintf("Datasource '%s' ready.", name), GCtx)
			return nil
		}

		// Filter out internal tables
		var candidates []api.DiscoveredDataset
		for _, d := range discovered {
			if d.Onboarded || isInternalDataset(d.Name, d.QualifiedName) {
				continue
			}
			candidates = append(candidates, d)
		}

		hidden := len(discovered) - len(candidates)
		if hidden > 0 {
			fmt.Printf("  Found %d datasets (%d internal tables hidden).\n\n", len(candidates), hidden)
		} else {
			fmt.Printf("  Found %d datasets.\n\n", len(candidates))
		}

		if len(candidates) == 0 {
			output.PrintSuccess(fmt.Sprintf("Datasource '%s' ready. No new datasets to onboard.", name), GCtx)
			return nil
		}

		// ── Step 4: Select datasets to onboard ───────────────────────────
		byQN := map[string]api.DiscoveredDataset{}
		for _, d := range candidates {
			byQN[d.QualifiedName] = d
		}

		var selectedNames []string
		if noInteractive {
			for _, d := range candidates {
				selectedNames = append(selectedNames, d.QualifiedName)
			}
		} else {
			options := make([]huh.Option[string], len(candidates))
			for i, d := range candidates {
				label := d.QualifiedName
				if label == "" {
					label = d.Name
				}
				options[i] = huh.NewOption(label, d.QualifiedName)
			}
			form := huh.NewForm(huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select datasets to onboard").
					Description("Space to toggle, Enter to confirm").
					Options(options...).
					Value(&selectedNames),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
		}

		if len(selectedNames) == 0 {
			fmt.Println(output.Dim.Render("  No datasets selected."))
			output.PrintSuccess(fmt.Sprintf("Datasource '%s' ready.", name), GCtx)
			return nil
		}

		// ── Step 5: Onboard ──────────────────────────────────────────────
		ids := make([]string, 0, len(selectedNames))
		for _, qn := range selectedNames {
			if d, ok := byQN[qn]; ok {
				ids = append(ids, d.ID)
			}
		}

		fmt.Println(output.Dim.Render(fmt.Sprintf("  Onboarding %d datasets...", len(ids))))
		if err := client.OnboardDiscoveredDatasets(datasourceID, api.OnboardDatasetsRequest{
			DiscoveredDatasetIDs: ids,
		}); err != nil {
			return err
		}
		fmt.Println(output.Green.Render("  ✓") + " Datasets onboarded.")

		// ── Step 6: Monitoring ───────────────────────────────────────────
		enableMonitoring := false
		if cmd.Flags().Changed("monitoring") {
			enableMonitoring, _ = cmd.Flags().GetBool("monitoring")
		} else if cmd.Flags().Changed("no-monitoring") {
			enableMonitoring = false
		} else if noInteractive {
			return output.Errorf(2, "--monitoring or --no-monitoring is required in non-interactive mode")
		} else {
			choice := "yes"
			form := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().
					Title("Enable default metric monitoring?").
					Description("Row count, freshness, schema changes, and more.").
					Options(
						huh.NewOption("Yes", "yes"),
						huh.NewOption("No", "no"),
					).
					Value(&choice),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
			enableMonitoring = choice == "yes"
		}

		// ── Step 7: Profiling ─────────────────────────────────────────────
		enableProfiling := false
		if cmd.Flags().Changed("profiling") {
			enableProfiling, _ = cmd.Flags().GetBool("profiling")
		} else if cmd.Flags().Changed("no-profiling") {
			enableProfiling = false
		} else if noInteractive {
			return output.Errorf(2, "--profiling or --no-profiling is required in non-interactive mode")
		} else {
			choice := "yes"
			form := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().
					Title("Enable dataset profiling?").
					Description("Column stats, row counts, and data type distribution.").
					Options(
						huh.NewOption("Yes", "yes"),
						huh.NewOption("No", "no"),
					).
					Value(&choice),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
			enableProfiling = choice == "yes"
		}

		// ── Step 8: Contracts ────────────────────────────────────────────
		contractsMode := ""
		if cmd.Flags().Changed("contracts") {
			contractsMode, _ = cmd.Flags().GetString("contracts")
		} else if noInteractive {
			return output.Errorf(2, "--contracts is required in non-interactive mode (ai|skeleton|none)")
		} else {
			form := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().
					Title("Set up data contracts?").
					Options(
						huh.NewOption("AI-generated (Autopilot)", "ai"),
						huh.NewOption("Skeleton (empty template)", "skeleton"),
						huh.NewOption("None", "none"),
					).
					Value(&contractsMode),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
		}

		// ── Step 9: Execute monitoring + profiling + contracts ────────────
		cloudDatasets, err := client.ListDatasets(api.ListDatasetsParams{
			DatasourceName: name,
			Size:           500,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s Could not list datasets: %v\n", output.Yellow.Render("⚠"), err)
		}

		type cloudInfo struct {
			ID         string
			ContractQN string
		}
		cloud := map[string]cloudInfo{}
		if cloudDatasets != nil {
			for _, d := range cloudDatasets.Content {
				cqn := d.Datasource.Name + "/" + strings.ReplaceAll(d.QualifiedName, ".", "/")
				ci := cloudInfo{ID: d.ID, ContractQN: cqn}
				cloud[d.QualifiedName] = ci
				cloud[cqn] = ci
			}
		}

		var hadErrors bool

		if enableMonitoring || enableProfiling {
			label := ""
			switch {
			case enableMonitoring && enableProfiling:
				label = "Enabling monitoring and profiling..."
			case enableMonitoring:
				label = "Enabling monitoring..."
			default:
				label = "Enabling profiling..."
			}
			fmt.Println(output.Dim.Render("  " + label))
			for _, qn := range selectedNames {
				ci, ok := cloud[qn]
				if !ok {
					fmt.Fprintf(os.Stderr, "  %s '%s' not found in cloud — skipping.\n", output.Yellow.Render("⚠"), qn)
					hadErrors = true
					continue
				}
				if err := client.EnableDatasetDefaults(ci.ID, enableMonitoring, enableProfiling); err != nil {
					fmt.Fprintf(os.Stderr, "  %s Setup for '%s': %v\n", output.Yellow.Render("⚠"), qn, err)
					hadErrors = true
				}
			}
			if !hadErrors {
				fmt.Println(output.Green.Render("  ✓") + " " + label[:len(label)-3] + "d.")
			}
		}

		switch contractsMode {
		case "ai":
			cqns := make([]string, 0, len(selectedNames))
			for _, qn := range selectedNames {
				if ci, ok := cloud[qn]; ok {
					cqns = append(cqns, ci.ContractQN)
				}
			}
			if len(cqns) > 0 {
				_, err := client.GenerateContract(api.GenerateContractRequest{
					DatasetQualifiedNames: cqns,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "  %s AI contract generation failed: %v\n", output.Yellow.Render("⚠"), err)
					hadErrors = true
				} else {
					fmt.Printf("  %s AI contract generation started for %d datasets.\n", output.Green.Render("✓"), len(cqns))
					fmt.Println(output.Dim.Render("  Contracts are being generated in the background and will appear in Soda Cloud when ready."))
					fmt.Println(output.Dim.Render("  Check results:  sodacli results list"))
				}
			}
		case "skeleton":
			for _, qn := range selectedNames {
				ci, ok := cloud[qn]
				if !ok {
					continue
				}
				outFile := datasetFileName(ci.ContractQN)
				if err := runContractCreateSkeleton(client, ci.ContractQN, outFile); err != nil {
					fmt.Fprintf(os.Stderr, "  %s Skeleton for '%s': %v\n", output.Yellow.Render("⚠"), qn, err)
					hadErrors = true
				}
			}
		case "none":
			// skip
		default:
			return output.Errorf(2, "unknown contracts mode '%s' — use ai, skeleton, or none", contractsMode)
		}

		// ── Summary ──────────────────────────────────────────────────────
		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Datasource '%s' onboarded with %d datasets.", name, len(selectedNames)), GCtx)
		if hadErrors {
			fmt.Println(output.Yellow.Render("  Some steps had warnings — check the output above."))
		}
		return nil
	},
}

// pollDiscoveredDatasets fetches all discovered datasets for a datasource,
// retrying until results appear or the timeout is reached.
func pollDiscoveredDatasets(client *api.Client, datasourceID string, spinner *output.Spinner) ([]api.DiscoveredDataset, error) {
	deadline := time.Now().Add(5 * time.Minute)
	elapsed := 0
	for {
		first, err := client.ListDiscoveredDatasets(datasourceID, 0, 100)
		if err != nil {
			return nil, fmt.Errorf("could not list discovered datasets: %w", err)
		}
		if len(first.Content) == 0 {
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("discovery timed out after 5 minutes — datasets may still appear later")
			}
			elapsed += 5
			spinner.SetMessage(fmt.Sprintf("Discovering datasets... (%ds)", elapsed))
			time.Sleep(5 * time.Second)
			continue
		}

		spinner.SetMessage("Loading discovered datasets...")
		all := first.Content
		if !first.Last {
			for pg := 1; ; pg++ {
				time.Sleep(200 * time.Millisecond)
				page, err := client.ListDiscoveredDatasets(datasourceID, pg, 100)
				if err != nil {
					break
				}
				all = append(all, page.Content...)
				if page.Last || len(page.Content) == 0 {
					break
				}
			}
		}
		return all, nil
	}
}

// isInternalDataset returns true for Soda temp tables and diagnostics tables.
func isInternalDataset(name, qualifiedName string) bool {
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, "__soda_temp") || strings.HasPrefix(lower, "soda_temp") {
		return true
	}
	lowerQN := strings.ToLower(qualifiedName)
	if strings.Contains(lowerQN, "/soda_diagnostics/") || strings.Contains(lowerQN, ".soda_diagnostics.") {
		return true
	}
	return false
}

func init() {
	dsOnboardCmd.Flags().String("runner", "", "Route connection through a Soda Runner (only used when creating a new datasource)")
	dsOnboardCmd.Flags().Bool("monitoring", false, "Enable default metric monitors for all datasets")
	dsOnboardCmd.Flags().Bool("no-monitoring", false, "Skip monitoring setup")
	dsOnboardCmd.Flags().Bool("profiling", false, "Enable dataset profiling for all datasets")
	dsOnboardCmd.Flags().Bool("no-profiling", false, "Skip profiling setup")
	dsOnboardCmd.Flags().String("contracts", "", "Generate contracts: ai|skeleton|none")

}
