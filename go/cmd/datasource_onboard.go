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
		enableMonitoring, _ := cmd.Flags().GetBool("monitoring")
		noMonitoring, _ := cmd.Flags().GetBool("no-monitoring")
		contractsMode, _ := cmd.Flags().GetString("contracts")

		// Non-interactive validation
		if GCtx.NoInteractive {
			if !cmd.Flags().Changed("monitoring") && !cmd.Flags().Changed("no-monitoring") {
				return output.Errorf(2, "--monitoring or --no-monitoring is required in non-interactive mode")
			}
			if !cmd.Flags().Changed("contracts") {
				return output.Errorf(2, "--contracts is required in non-interactive mode (ai|skeleton|none)")
			}
		}

		if noMonitoring {
			enableMonitoring = false
		}
		if contractsMode == "" {
			contractsMode = "none"
		}

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

		// ── Resolve agent ────────────────────────────────────────────────
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

		// ── Create datasource ────────────────────────────────────────────
		fmt.Println(output.Dim.Render("  Creating datasource '" + name + "'..."))
		createResult, err := client.CreateDatasource(api.CreateDatasourceRequest{
			Name:                      name,
			AgentID:                   agentID,
			ConfigurationFileContents: string(configBytes),
		})
		if err != nil {
			return err
		}
		fmt.Printf("  Datasource ID: %s\n", createResult.Datasource.ID)
		datasourceID := createResult.Datasource.ID

		// ── Poll for discovered datasets ─────────────────────────────────
		fmt.Println(output.Dim.Render("  Waiting for dataset discovery..."))
		var discovered []api.DiscoveredDataset
		deadline := time.Now().Add(5 * time.Minute)
		for {
			page, err := client.ListDiscoveredDatasets(datasourceID, 0, 500)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  %s Could not list discovered datasets: %v\n", output.Yellow.Render("⚠"), err)
				break
			}
			if len(page.Content) > 0 {
				discovered = page.Content
				break
			}
			if time.Now().After(deadline) {
				fmt.Fprintf(os.Stderr, "  %s Discovery timed out after 5 minutes. Datasets may still appear later.\n", output.Yellow.Render("⚠"))
				fmt.Println(output.Dim.Render("  Check with: soda dataset list --datasource " + name))
				fmt.Println()
				output.PrintSuccess(fmt.Sprintf("Datasource '%s' created. No datasets discovered yet.", name), GCtx)
				return nil
			}
			fmt.Println(output.Dim.Render("  Still discovering datasets..."))
			time.Sleep(3 * time.Second)
		}

		if len(discovered) == 0 {
			fmt.Println()
			output.PrintSuccess(fmt.Sprintf("Datasource '%s' created. No datasets discovered.", name), GCtx)
			return nil
		}

		fmt.Printf("  Found %d datasets.\n\n", len(discovered))

		// ── Select datasets to onboard ───────────────────────────────────
		// Build a map from qualifiedName → discovered dataset for lookup
		discoveredByQN := map[string]api.DiscoveredDataset{}
		for _, d := range discovered {
			discoveredByQN[d.QualifiedName] = d
		}

		selectedNames := make([]string, len(discovered))
		for i, d := range discovered {
			selectedNames[i] = d.QualifiedName
		}

		if !GCtx.NoInteractive {
			options := make([]huh.Option[string], len(discovered))
			for i, d := range discovered {
				label := d.Name
				if d.QualifiedName != "" && d.QualifiedName != d.Name {
					label = d.QualifiedName
				}
				options[i] = huh.NewOption(label, d.QualifiedName).Selected(true)
			}
			form := huh.NewForm(huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select datasets to onboard").
					Options(options...).
					Value(&selectedNames),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
		}

		if len(selectedNames) == 0 {
			fmt.Println(output.Dim.Render("  No datasets selected."))
			fmt.Println()
			output.PrintSuccess(fmt.Sprintf("Datasource '%s' created.", name), GCtx)
			return nil
		}

		// ── Onboard selected datasets ────────────────────────────────────
		selectedIDs := make([]string, 0, len(selectedNames))
		for _, qn := range selectedNames {
			if d, ok := discoveredByQN[qn]; ok {
				selectedIDs = append(selectedIDs, d.ID)
			}
		}

		fmt.Printf(output.Dim.Render("  Onboarding %d datasets...")+"\n", len(selectedIDs))
		if err := client.OnboardDiscoveredDatasets(datasourceID, api.OnboardDatasetsRequest{
			DiscoveredDatasetIDs: selectedIDs,
		}); err != nil {
			return err
		}
		fmt.Println(output.Green.Render("  ✓") + " Datasets onboarded.")

		// ── Interactive: monitoring & contracts settings ──────────────────
		if !cmd.Flags().Changed("monitoring") && !cmd.Flags().Changed("no-monitoring") && !GCtx.NoInteractive {
			monitoringChoice := "yes"
			form := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().
					Title("Enable default metric monitoring for onboarded datasets?").
					Options(
						huh.NewOption("Yes", "yes"),
						huh.NewOption("No", "no"),
					).
					Value(&monitoringChoice),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
			enableMonitoring = monitoringChoice == "yes"
		}

		if !cmd.Flags().Changed("contracts") && !GCtx.NoInteractive {
			form := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().
					Title("Generate contracts for onboarded datasets?").
					Options(
						huh.NewOption("AI-generated contract (Copilot)", "ai"),
						huh.NewOption("Skeleton contract (empty template)", "skeleton"),
						huh.NewOption("No contract", "none"),
					).
					Value(&contractsMode),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
		}

		// ── Fetch onboarded dataset IDs from Soda Cloud ──────────────────
		fmt.Println(output.Dim.Render("  Fetching onboarded datasets..."))
		cloudDatasets, err := client.ListDatasets(api.ListDatasetsParams{
			DatasourceName: name,
			Size:           500,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s Could not list datasets: %v\n", output.Yellow.Render("⚠"), err)
		}

		// Build map from qualifiedName → cloud dataset
		// The discovered dataset qualified name may be dot-separated (db.schema.table)
		// while the contract API needs slash-separated (datasource/db/schema/table).
		type cloudDS struct {
			ID                   string
			QualifiedName        string // dot-separated from API
			ContractQualifiedName string // slash-separated for contract APIs
		}
		cloudMap := map[string]cloudDS{}
		if cloudDatasets != nil {
			for _, d := range cloudDatasets.Content {
				cqn := d.Datasource.Name + "/" + strings.ReplaceAll(d.QualifiedName, ".", "/")
				entry := cloudDS{ID: d.ID, QualifiedName: d.QualifiedName, ContractQualifiedName: cqn}
				cloudMap[d.QualifiedName] = entry
				// Also index by contract-style name for flexibility
				cloudMap[cqn] = entry
			}
		}

		var hadErrors bool

		// ── Monitoring ───────────────────────────────────────────────────
		if enableMonitoring {
			fmt.Println(output.Dim.Render("  Enabling monitoring..."))
			for _, qn := range selectedNames {
				ds, ok := cloudMap[qn]
				if !ok {
					fmt.Fprintf(os.Stderr, "  %s Dataset '%s' not found in cloud — skipping monitoring.\n", output.Yellow.Render("⚠"), qn)
					hadErrors = true
					continue
				}
				if _, err := client.UpdateMetricMonitoring(ds.ID, api.UpdateMetricMonitoringRequest{Enabled: boolPtr(true)}); err != nil {
					fmt.Fprintf(os.Stderr, "  %s Monitoring for '%s': %v\n", output.Yellow.Render("⚠"), qn, err)
					hadErrors = true
				}
			}
			fmt.Println(output.Green.Render("  ✓") + " Monitoring enabled.")
		}

		// ── Contracts ────────────────────────────────────────────────────
		switch contractsMode {
		case "ai":
			// Batch: GenerateContract accepts multiple qualified names
			// Use contract-style qualified names (datasource/db/schema/table)
			cqns := make([]string, 0, len(selectedNames))
			for _, qn := range selectedNames {
				if ds, ok := cloudMap[qn]; ok {
					cqns = append(cqns, ds.ContractQualifiedName)
				}
			}
			if len(cqns) > 0 {
				fmt.Println(output.Dim.Render("  Generating AI contracts..."))
				opID, err := client.GenerateContract(api.GenerateContractRequest{
					DatasetQualifiedNames: cqns,
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "  %s AI contract generation failed: %v\n", output.Yellow.Render("⚠"), err)
					hadErrors = true
				} else {
					// Poll
					for {
						status, err := client.GetGenerateStatus(opID)
						if err != nil {
							fmt.Fprintf(os.Stderr, "  %s Could not check generation status: %v\n", output.Yellow.Render("⚠"), err)
							hadErrors = true
							break
						}
						if status.State == "completed" {
							break
						}
						if status.State == "failed" || status.State == "canceled" {
							fmt.Fprintf(os.Stderr, "  %s AI generation %s\n", output.Yellow.Render("⚠"), status.State)
							hadErrors = true
							break
						}
						fmt.Println(output.Dim.Render("  Waiting for AI generation..."))
						time.Sleep(3 * time.Second)
					}
					// Pull contracts locally
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
				ds, ok := cloudMap[qn]
				if !ok {
					continue
				}
				outFile := datasetFileName(ds.ContractQualifiedName)
				if err := runContractCreateSkeleton(client, ds.ContractQualifiedName, outFile); err != nil {
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
			return output.Errorf(2, "some steps had errors — check the warnings above")
		}
		return nil
	},
}

func init() {
	dsOnboardCmd.Flags().String("agent", "", "Route connection through a Soda Agent")
	dsOnboardCmd.Flags().Bool("monitoring", false, "Enable default metric monitors for all datasets")
	dsOnboardCmd.Flags().Bool("no-monitoring", false, "Skip monitoring setup")
	dsOnboardCmd.Flags().String("contracts", "", "Generate contracts: ai|skeleton|none")

	datasourceCmd.AddCommand(dsOnboardCmd)
}
