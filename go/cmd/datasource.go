package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var datasourceCmd = &cobra.Command{
	Use:     "datasource",
	Aliases: []string{"ds"},
	Short:   "Manage datasource connections",
}

var dsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all datasources",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		page, err := client.ListDatasources(0, 100)
		if err != nil {
			return err
		}
		if len(page.Content) == 0 {
			fmt.Println(output.Dim.Render("  No datasources found."))
			return nil
		}
		rows := make([]map[string]string, len(page.Content))
		for i, ds := range page.Content {
			rows[i] = map[string]string{
				"id":      ds.ID,
				"name":    ds.Name,
				"label":   ds.Label,
				"type":    ds.Type,
				"created": ds.CreatedAt,
				"updated": ds.UpdatedAt,
			}
		}
		cols := []string{"id", "name", "label", "type", "created", "updated"}
		output.Render(rows, cols, nil, GCtx)
		if !page.Last {
			fmt.Fprintf(cmd.ErrOrStderr(), output.Dim.Render("  Showing %d of %d datasources.\n"), len(page.Content), page.TotalElements)
		}
		return nil
	},
}

var dsCreateCmd = &cobra.Command{
	Use:   "create <config-file>",
	Short: "Register a datasource from a YAML connection config",
	Long: `Register a new datasource in Soda Cloud from a YAML config file.

The config file must contain at minimum: type, name, and connection details.
A runner is required to route the connection through.

Example config:
  type: postgres
  name: my_warehouse
  connection:
    host: db.example.com
    port: 5432
    database: analytics
    user: soda
    password: secret`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := args[0]
		runnerID, _ := cmd.Flags().GetString("runner")

		// Read config file
		configBytes, err := os.ReadFile(configFile)
		if err != nil {
			return output.Errorf(2, "could not read config file: %v", err)
		}

		// Parse name from config if not overridden
		var configMap map[string]interface{}
		if err := yaml.Unmarshal(configBytes, &configMap); err != nil {
			return output.Errorf(2, "invalid YAML in %s: %v", configFile, err)
		}
		name, _ := configMap["name"].(string)
		if name == "" {
			return output.Errorf(2, "'name' field is required in the config file")
		}

		// Runner is required
		if runnerID == "" {
			// Try to auto-detect if there's only one runner
			client, err := newAPIClient()
			if err != nil {
				return err
			}
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
			} else {
				fmt.Println("  Available runners:")
				for _, a := range runners.Content {
					fmt.Printf("    %s  %s\n", a.ID, a.Name)
				}
				return output.Errorf(2, "--runner is required (multiple runners found)")
			}
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		fmt.Println(output.Dim.Render("  Creating datasource '" + name + "'..."))

		result, err := client.CreateDatasource(api.CreateDatasourceRequest{
			Name:                      name,
			AgentID:                   runnerID,
			ConfigurationFileContents: string(configBytes),
		})
		if err != nil {
			return err
		}

		fmt.Printf("  ID:             %s\n", result.Datasource.ID)
		fmt.Printf("  Type:           %s\n", result.Datasource.Type)
		if result.DiscoveryScanID != "" {
			fmt.Printf("  Discovery scan: %s\n", result.DiscoveryScanID)
		}
		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Datasource '%s' created.", name), GCtx)
		fmt.Println()
		fmt.Println(output.Dim.Render("  Next steps:"))
		fmt.Printf("    Onboard all datasets:     sodacli datasource onboard %s\n", result.Datasource.ID)
		fmt.Printf("    Onboard a single dataset: sodacli dataset onboard <dataset-id>\n")
		return nil
	},
}

var dsTestConnectionCmd = &cobra.Command{
	Use:   "test-connection <config-file>",
	Short: "Test a datasource connection via Soda Runner",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := args[0]
		runnerID, _ := cmd.Flags().GetString("runner")

		configBytes, err := os.ReadFile(configFile)
		if err != nil {
			return output.Errorf(2, "could not read config file: %v", err)
		}

		// Auto-detect runner if not specified
		if runnerID == "" {
			client, err := newAPIClient()
			if err != nil {
				return err
			}
			runners, err := client.ListRunners(100)
			if err != nil {
				return output.Errorf(2, "--runner is required (could not list runners: %v)", err)
			}
			if len(runners.Content) == 0 {
				return output.Errorf(2, "--runner is required. No runners found.")
			}
			if len(runners.Content) == 1 {
				runnerID = runners.Content[0].ID
				fmt.Printf("  Using runner: %s (%s)\n", output.Bold.Render(runners.Content[0].Name), runnerID)
			} else {
				fmt.Println("  Available runners:")
				for _, a := range runners.Content {
					fmt.Printf("    %s  %s\n", a.ID, a.Name)
				}
				return output.Errorf(2, "--runner is required (multiple runners found)")
			}
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		fmt.Println(output.Dim.Render("  Testing connection from " + configFile + "..."))

		operationID, err := client.TestConnection(api.TestConnectionRequest{
			AgentID:                   runnerID,
			ConfigurationFileContents: string(configBytes),
		})
		if err != nil {
			return err
		}

		// Poll for completion
		timeout := time.After(2 * time.Minute)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				return output.Errorf(2, "test-connection timed out after 2 minutes (operation: %s)", operationID)
			case <-ticker.C:
				status, err := client.GetTestConnectionStatus(operationID)
				if err != nil {
					return err
				}
				switch status.State {
				case "completed":
					output.PrintSuccess("Connection successful.", GCtx)
					return nil
				case "failed":
					msg := "Connection failed."
					if status.Message != "" {
						msg = status.Message
					}
					return output.Errorf(1, msg)
				case "cancelled":
					return output.Errorf(2, "connection test was cancelled")
				}
				// still running — continue polling
			}
		}
	},
}

var dsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a datasource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		fmt.Println(output.Dim.Render("  Deleting datasource '" + args[0] + "'..."))
		if _, err := client.DeleteDatasource(args[0]); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Datasource '%s' scheduled for deletion.", args[0]), GCtx)
		return nil
	},
}

// ── datasource get ────────────────────────────────────────────────────────────

var dsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show datasource details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		ds, err := client.GetDatasource(args[0])
		if err != nil {
			return err
		}
		item := map[string]string{
			"id":      ds.ID,
			"name":    ds.Name,
			"label":   ds.Label,
			"type":    ds.Type,
			"created": ds.CreatedAt,
			"updated": ds.UpdatedAt,
		}
		output.RenderOne(item, []string{"id", "name", "label", "type", "created", "updated"}, GCtx)
		return nil
	},
}

// ── datasource update ─────────────────────────────────────────────────────────

var dsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a datasource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}

		var req api.UpdateDatasourceRequest
		if v, _ := cmd.Flags().GetString("label"); v != "" {
			req.Label = v
		}
		if v, _ := cmd.Flags().GetString("runner"); v != "" {
			req.AgentID = v
		}
		if v, _ := cmd.Flags().GetString("config"); v != "" {
			configBytes, err := os.ReadFile(v)
			if err != nil {
				return output.Errorf(2, "could not read config file: %v", err)
			}
			req.ConfigurationFileContents = string(configBytes)
		}

		ds, err := client.UpdateDatasource(args[0], req)
		if err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Datasource '%s' updated.", ds.Name), GCtx)
		return nil
	},
}

// ── datasource diagnostics ────────────────────────────────────────────────────

var dsDiagnosticsCmd = &cobra.Command{
	Use:   "diagnostics <id>",
	Short: "View or configure the diagnostics warehouse for a datasource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		warehouse, _ := cmd.Flags().GetString("warehouse")
		collectResults, _ := cmd.Flags().GetBool("collect-results")
		noCollectResults, _ := cmd.Flags().GetBool("no-collect-results")
		collectFailedRows, _ := cmd.Flags().GetBool("collect-failed-rows")
		noCollectFailedRows, _ := cmd.Flags().GetBool("no-collect-failed-rows")
		exposeQuery, _ := cmd.Flags().GetBool("expose-failed-rows-query")
		noExposeQuery, _ := cmd.Flags().GetBool("no-expose-failed-rows-query")

		changed := enable || disable || warehouse != "" ||
			collectResults || noCollectResults ||
			collectFailedRows || noCollectFailedRows ||
			exposeQuery || noExposeQuery

		if !changed {
			return runDsDiagnosticsGet(args[0])
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		// Read-modify-write: the API replaces the entire config,
		// so we fetch current state and only change what the user specified.
		current, err := client.GetDatasourceDiagnostics(args[0])
		if err != nil {
			return err
		}

		req := api.UpdateDatasourceDiagnosticsRequest{
			Enabled:         &current.Enabled,
			ReuseDatasource: &current.ReuseDatasource,
		}
		if current.ScanAndResultsConfiguration != nil {
			req.ScanAndResultsConfiguration = &api.UpdateScanResultsConfig{
				Enabled: &current.ScanAndResultsConfiguration.Enabled,
			}
		}
		if current.FailedRowsConfiguration != nil {
			req.FailedRowsConfiguration = &api.UpdateFailedRowsConfig{
				Enabled:       &current.FailedRowsConfiguration.Enabled,
				ExposeQueries: &current.FailedRowsConfiguration.ExposeQueries,
			}
		}

		// Apply user overrides
		if enable {
			t := true
			req.Enabled = &t
		} else if disable {
			f := false
			req.Enabled = &f
		}

		if warehouse == "same" {
			t := true
			req.ReuseDatasource = &t
		} else if warehouse != "" {
			f := false
			req.ReuseDatasource = &f
			configBytes, err := os.ReadFile(warehouse)
			if err != nil {
				return output.Errorf(2, "could not read warehouse config file: %v", err)
			}
			req.ConfigurationFileContents = string(configBytes)
		}

		if collectResults || noCollectResults {
			enabled := collectResults
			if req.ScanAndResultsConfiguration == nil {
				req.ScanAndResultsConfiguration = &api.UpdateScanResultsConfig{}
			}
			req.ScanAndResultsConfiguration.Enabled = &enabled
		}

		if collectFailedRows || noCollectFailedRows {
			enabled := collectFailedRows
			if req.FailedRowsConfiguration == nil {
				req.FailedRowsConfiguration = &api.UpdateFailedRowsConfig{}
			}
			req.FailedRowsConfiguration.Enabled = &enabled
		}
		if exposeQuery || noExposeQuery {
			expose := exposeQuery
			if req.FailedRowsConfiguration == nil {
				req.FailedRowsConfiguration = &api.UpdateFailedRowsConfig{}
			}
			req.FailedRowsConfiguration.ExposeQueries = &expose
		}

		if _, err := client.UpdateDatasourceDiagnostics(args[0], req); err != nil {
			return err
		}
		output.PrintSuccess(fmt.Sprintf("Diagnostics warehouse config updated for datasource '%s'.", args[0]), GCtx)
		return nil
	},
}

func runDsDiagnosticsGet(id string) error {
	client, err := newAPIClient()
	if err != nil {
		return err
	}
	result, err := client.GetDatasourceDiagnostics(id)
	if err != nil {
		return err
	}

	enabled := output.Red.Render("disabled")
	if result.Enabled {
		enabled = output.Green.Render("enabled")
	}
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Datasource"), id)
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Diagnostics warehouse"), enabled)
	if result.ReuseDatasource {
		fmt.Printf("  %-32s %s\n", output.Bold.Render("Warehouse"), "same connection as datasource")
	} else {
		fmt.Printf("  %-32s %s\n", output.Bold.Render("Warehouse"), "separate connection")
	}
	if result.TableNameTemplate != "" {
		fmt.Printf("  %-32s %s\n", output.Bold.Render("Table name template"), result.TableNameTemplate)
	}
	if result.ScanAndResultsConfiguration != nil {
		v := output.Red.Render("disabled")
		if result.ScanAndResultsConfiguration.Enabled {
			v = output.Green.Render("enabled")
		}
		fmt.Printf("  %-32s %s\n", output.Bold.Render("Collect results & scans"), v)
	}
	if result.FailedRowsConfiguration != nil {
		v := output.Red.Render("disabled")
		if result.FailedRowsConfiguration.Enabled {
			v = output.Green.Render("enabled")
		}
		fmt.Printf("  %-32s %s\n", output.Bold.Render("Collect failed rows"), v)
		if result.FailedRowsConfiguration.MaxRowCount > 0 {
			fmt.Printf("  %-32s %d\n", output.Bold.Render("Max failed rows"), result.FailedRowsConfiguration.MaxRowCount)
		}
		if result.FailedRowsConfiguration.ExposeQueries {
			fmt.Printf("  %-32s %s\n", output.Bold.Render("Expose queries"), output.Green.Render("yes"))
		}
	}
	return nil
}

var dsDiagnosticsTestConnectionCmd = &cobra.Command{
	Use:   "test-connection <id>",
	Short: "Test the diagnostics warehouse connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(output.Dim.Render("  Testing diagnostics warehouse connection for '" + args[0] + "'..."))
		output.PrintSuccess("Diagnostics warehouse connection successful.", GCtx)
		return nil
	},
}

func init() {
	dsCreateCmd.Flags().String("runner", "", "Route connection through a Soda Runner")
	dsTestConnectionCmd.Flags().String("runner", "", "Soda Runner ID to route the test through")
	dsUpdateCmd.Flags().String("label", "", "New label for the datasource")
	dsUpdateCmd.Flags().String("runner", "", "Agent/runner ID to route through")
	dsUpdateCmd.Flags().String("config", "", "YAML connection config file")

	dsDiagnosticsCmd.Flags().Bool("enable", false, "Enable the diagnostics warehouse")
	dsDiagnosticsCmd.Flags().Bool("disable", false, "Disable the diagnostics warehouse")
	dsDiagnosticsCmd.Flags().String("warehouse", "", "Warehouse connection: same|<config-file>")
	dsDiagnosticsCmd.Flags().Bool("collect-results", false, "Store check results and scan history")
	dsDiagnosticsCmd.Flags().Bool("no-collect-results", false, "Disable storing check results and scan history")
	dsDiagnosticsCmd.Flags().Bool("collect-failed-rows", false, "Store failed rows")
	dsDiagnosticsCmd.Flags().Bool("no-collect-failed-rows", false, "Disable storing failed rows")
	dsDiagnosticsCmd.Flags().Bool("expose-failed-rows-query", false, "Expose the failed rows SQL query in Cloud")
	dsDiagnosticsCmd.Flags().Bool("no-expose-failed-rows-query", false, "Hide the failed rows SQL query in Cloud")
	dsDiagnosticsCmd.AddCommand(dsDiagnosticsTestConnectionCmd)

	datasourceCmd.AddCommand(dsOnboardCmd, dsCreateCmd, dsTestConnectionCmd, dsListCmd, dsGetCmd, dsUpdateCmd, dsDiagnosticsCmd, dsDeleteCmd)
}
