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
	Use:   "onboard <config-file>",
	Short: "Guided setup: create datasource + configure all datasets",
	Long: `Create a datasource from a YAML config file, wait for dataset discovery,
then onboard discovered datasets with optional monitoring and contracts.

Interactive mode walks through each step. Use flags for CI/CD or AI agents:

  soda datasource onboard config.yml --no-interactive --no-monitoring --contracts none`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := args[0]
		agentID, _ := cmd.Flags().GetString("agent")

		// ── Read config file ─────────────────────────────────────────────
		configBytes, err := os.ReadFile(configFile)
		if err != nil {
			return output.Errorf(2, "could not read config file: %v", err)
		}
		var configMap map[string]interface{}
		if err := yaml.Unmarshal(configBytes, &configMap); err != nil {
			return output.Errorf(2, "invalid YAML in %s: %v", configFile, err)
		}
		name, _ := configMap["name"].(string)
		if name == "" {
			return output.Errorf(2, "'name' field is required in the config file")
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		// ── Step 1: Resolve agent ────────────────────────────────────────
		if agentID == "" {
			agents, err := client.ListAgents(100)
			if err != nil {
				return output.Errorf(2, "--agent is required (could not list agents: %v)", err)
			}
			if len(agents.Content) == 0 {
				return output.Errorf(2, "--agent is required. No agents found — set up an agent in Soda Cloud first.")
			}
			if len(agents.Content) == 1 {
				agentID = agents.Content[0].ID
				fmt.Printf("  Using agent: %s (%s)\n", output.Bold.Render(agents.Content[0].Name), agentID)
			} else if GCtx.NoInteractive {
				fmt.Println("  Available agents:")
				for _, a := range agents.Content {
					fmt.Printf("    %s  %s\n", a.ID, a.Name)
				}
				return output.Errorf(2, "--agent is required (multiple agents found)")
			} else {
				options := make([]huh.Option[string], len(agents.Content))
				for i, a := range agents.Content {
					label := a.Name
					if a.Label != "" {
						label = a.Label + " (" + a.Name + ")"
					}
					options[i] = huh.NewOption(label, a.ID)
				}
				form := huh.NewForm(huh.NewGroup(
					huh.NewSelect[string]().
						Title("Which agent should route this connection?").
						Options(options...).
						Value(&agentID),
				))
				if err := form.Run(); err != nil {
					return output.Errorf(2, "cancelled")
				}
			}
		}

		// ── Step 2: Create datasource ────────────────────────────────────
		fmt.Println(output.Dim.Render("  Creating datasource '" + name + "'..."))
		createResult, err := client.CreateDatasource(api.CreateDatasourceRequest{
			Name:                      name,
			AgentID:                   agentID,
			ConfigurationFileContents: string(configBytes),
		})
		if err != nil {
			return err
		}
		datasourceID := createResult.Datasource.ID
		fmt.Printf("  Datasource ID: %s\n", datasourceID)

		// ── Step 3: Discover datasets ────────────────────────────────────
		spinner := output.NewSpinner("Waiting for dataset discovery...")
		spinner.Start()
		discovered, err := pollDiscoveredDatasets(client, datasourceID, spinner)
		spinner.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s %v\n", output.Yellow.Render("⚠"), err)
			fmt.Println(output.Dim.Render("  Check with: soda dataset list --datasource " + name))
			fmt.Println()
			output.PrintSuccess(fmt.Sprintf("Datasource '%s' created.", name), GCtx)
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
			output.PrintSuccess(fmt.Sprintf("Datasource '%s' created. No new datasets to onboard.", name), GCtx)
			return nil
		}

		// ── Step 4: Select datasets to onboard ───────────────────────────
		byQN := map[string]api.DiscoveredDataset{}
		for _, d := range candidates {
			byQN[d.QualifiedName] = d
		}

		var selectedNames []string
		if GCtx.NoInteractive {
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
			output.PrintSuccess(fmt.Sprintf("Datasource '%s' created.", name), GCtx)
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
		} else if GCtx.NoInteractive {
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

		// ── Step 7: Contracts ────────────────────────────────────────────
		contractsMode := ""
		if cmd.Flags().Changed("contracts") {
			contractsMode, _ = cmd.Flags().GetString("contracts")
		} else if GCtx.NoInteractive {
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

		// ── Step 8: Execute monitoring + contracts ───────────────────────
		// Fetch cloud dataset IDs so we can call monitoring/contract APIs
		cloudDatasets, err := client.ListDatasets(api.ListDatasetsParams{
			DatasourceName: name,
			Size:           500,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s Could not list datasets: %v\n", output.Yellow.Render("⚠"), err)
		}

		// Map: discovered qualifiedName → cloud dataset ID + contract-style QN
		type cloudInfo struct {
			ID        string
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

		if enableMonitoring {
			fmt.Println(output.Dim.Render("  Enabling monitoring..."))
			for _, qn := range selectedNames {
				ci, ok := cloud[qn]
				if !ok {
					fmt.Fprintf(os.Stderr, "  %s '%s' not found in cloud — skipping.\n", output.Yellow.Render("⚠"), qn)
					hadErrors = true
					continue
				}
				if _, err := client.UpdateMetricMonitoring(ci.ID, api.UpdateMetricMonitoringRequest{Enabled: boolPtr(true)}); err != nil {
					fmt.Fprintf(os.Stderr, "  %s Monitoring for '%s': %v\n", output.Yellow.Render("⚠"), qn, err)
					hadErrors = true
				}
			}
			if !hadErrors {
				fmt.Println(output.Green.Render("  ✓") + " Monitoring enabled.")
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
				aiSpinner := output.NewSpinner(fmt.Sprintf("Generating AI contracts for %d datasets...", len(cqns)))
				aiSpinner.Start()
				opID, err := client.GenerateContract(api.GenerateContractRequest{
					DatasetQualifiedNames: cqns,
				})
				if err != nil {
					aiSpinner.Stop()
					fmt.Fprintf(os.Stderr, "  %s AI contract generation failed: %v\n", output.Yellow.Render("⚠"), err)
					hadErrors = true
				} else {
					elapsed := 0
					for {
						time.Sleep(3 * time.Second)
						elapsed += 3
						status, err := client.GetGenerateStatus(opID)
						if err != nil {
							aiSpinner.Stop()
							fmt.Fprintf(os.Stderr, "  %s Could not check generation status: %v\n", output.Yellow.Render("⚠"), err)
							hadErrors = true
							break
						}
						if status.State == "completed" {
							aiSpinner.Stop()
							break
						}
						if status.State == "failed" || status.State == "canceled" {
							aiSpinner.Stop()
							fmt.Fprintf(os.Stderr, "  %s AI generation %s\n", output.Yellow.Render("⚠"), status.State)
							hadErrors = true
							break
						}
						// Show which datasets are still pending
						pending := make([]string, 0)
						for _, ds := range status.Datasets {
							if ds.ScanState != "completed" {
								pending = append(pending, datasetFileName(ds.DatasetQualifiedName))
							}
						}
						if len(pending) > 0 && len(pending) < len(cqns) {
							aiSpinner.SetMessage(fmt.Sprintf("Generating AI contracts... %d/%d done (%ds)", len(cqns)-len(pending), len(cqns), elapsed))
						} else {
							aiSpinner.SetMessage(fmt.Sprintf("Generating AI contracts for %d datasets... (%ds)", len(cqns), elapsed))
						}
					}
					for _, cqn := range cqns {
						contract, err := client.FindContractByDataset(cqn)
						if err != nil || contract == nil {
							fmt.Fprintf(os.Stderr, "  %s Could not fetch contract for '%s'\n", output.Yellow.Render("⚠"), cqn)
							hadErrors = true
							continue
						}
						outFile := datasetFileName(cqn)
						if err := os.WriteFile(outFile, []byte(contract.Contents), 0644); err != nil {
							fmt.Fprintf(os.Stderr, "  %s Could not write %s: %v\n", output.Yellow.Render("⚠"), outFile, err)
							hadErrors = true
							continue
						}
						fmt.Printf("  %s Contract saved to %s\n", output.Green.Render("✓"), outFile)
					}
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
		// Quick check: just fetch the first page to see if anything exists yet.
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

		// Results exist — fetch remaining pages.
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
	dsOnboardCmd.Flags().String("agent", "", "Route connection through a Soda Agent")
	dsOnboardCmd.Flags().Bool("monitoring", false, "Enable default metric monitors for all datasets")
	dsOnboardCmd.Flags().Bool("no-monitoring", false, "Skip monitoring setup")
	dsOnboardCmd.Flags().String("contracts", "", "Generate contracts: ai|skeleton|none")

	datasourceCmd.AddCommand(dsOnboardCmd)
}
