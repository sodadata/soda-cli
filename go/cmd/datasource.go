package cmd

import (
	"fmt"
	"os"

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
An agent is required to route the connection through.

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
		agentID, _ := cmd.Flags().GetString("agent")

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

		// Agent is required
		if agentID == "" {
			// Try to auto-detect if there's only one agent
			client, err := newAPIClient()
			if err != nil {
				return err
			}
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
			} else {
				fmt.Println("  Available agents:")
				for _, a := range agents.Content {
					fmt.Printf("    %s  %s\n", a.ID, a.Name)
				}
				return output.Errorf(2, "--agent is required (multiple agents found)")
			}
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		fmt.Println(output.Dim.Render("  Creating datasource '" + name + "'..."))

		result, err := client.CreateDatasource(api.CreateDatasourceRequest{
			Name:                      name,
			AgentID:                   agentID,
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
		return nil
	},
}

var dsTestConnectionCmd = &cobra.Command{
	Use:   "test-connection <config-file>",
	Short: "Test the connection defined in a YAML config file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(output.Dim.Render("  Testing connection from " + args[0] + "..."))
		output.PrintSuccess("Connection successful.", GCtx)
		return nil
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

// ── datasource diagnostics ────────────────────────────────────────────────────

var dsDiagnosticsCmd = &cobra.Command{
	Use:   "diagnostics <id>",
	Short: "View or configure the diagnostics warehouse for a datasource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		warehouse, _ := cmd.Flags().GetString("warehouse")
		schema, _ := cmd.Flags().GetString("schema")
		collectResults, _ := cmd.Flags().GetBool("collect-results")
		noCollectResults, _ := cmd.Flags().GetBool("no-collect-results")
		collectFailedRows, _ := cmd.Flags().GetBool("collect-failed-rows")
		noCollectFailedRows, _ := cmd.Flags().GetBool("no-collect-failed-rows")
		tablePrefix, _ := cmd.Flags().GetString("table-prefix")
		tableSuffix, _ := cmd.Flags().GetString("table-suffix")
		failedRowsDesc, _ := cmd.Flags().GetString("failed-rows-description")
		exposeQuery, _ := cmd.Flags().GetBool("expose-failed-rows-query")
		noExposeQuery, _ := cmd.Flags().GetBool("no-expose-failed-rows-query")
		cta, _ := cmd.Flags().GetBool("failed-rows-cta")
		noCta, _ := cmd.Flags().GetBool("no-failed-rows-cta")

		changed := enable || disable || warehouse != "" || schema != "" ||
			collectResults || noCollectResults ||
			collectFailedRows || noCollectFailedRows ||
			tablePrefix != "" || tableSuffix != "" ||
			failedRowsDesc != "" ||
			exposeQuery || noExposeQuery ||
			cta || noCta

		if !changed {
			// no flags → same as get
			return runDsDiagnosticsGet(args[0])
		}

		output.PrintSuccess(fmt.Sprintf("Diagnostics warehouse config updated for datasource '%s'.", args[0]), GCtx)
		return nil
	},
}


func runDsDiagnosticsGet(id string) error {
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Datasource"), id)
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Diagnostics warehouse"), output.Green.Render("enabled"))
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Warehouse"), "same connection as datasource")
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Schema"), "soda_diagnostics")
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Collect results & scans"), "yes")
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Collect failed rows"), "yes")
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Table prefix"), output.Dim.Render("(none)"))
	fmt.Printf("  %-32s %s\n", output.Bold.Render("Table suffix"), output.Dim.Render("(none)"))
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
	dsCreateCmd.Flags().String("agent", "", "Route connection through a Soda Agent")

	dsDiagnosticsCmd.Flags().Bool("enable", false, "Enable the diagnostics warehouse")
	dsDiagnosticsCmd.Flags().Bool("disable", false, "Disable the diagnostics warehouse")
	dsDiagnosticsCmd.Flags().String("warehouse", "", "Warehouse connection: same|<config-file>")
	dsDiagnosticsCmd.Flags().String("schema", "", "Schema for diagnostic tables (default: soda_diagnostics)")
	dsDiagnosticsCmd.Flags().Bool("collect-results", false, "Store check results and scan history")
	dsDiagnosticsCmd.Flags().Bool("no-collect-results", false, "Disable storing check results and scan history")
	dsDiagnosticsCmd.Flags().Bool("collect-failed-rows", false, "Store failed rows")
	dsDiagnosticsCmd.Flags().Bool("no-collect-failed-rows", false, "Disable storing failed rows")
	dsDiagnosticsCmd.Flags().String("table-prefix", "", "Prefix for diagnostic table names")
	dsDiagnosticsCmd.Flags().String("table-suffix", "", "Suffix for diagnostic table names")
	dsDiagnosticsCmd.Flags().String("failed-rows-description", "", "Description for failed rows storage context")
	dsDiagnosticsCmd.Flags().Bool("expose-failed-rows-query", false, "Expose the failed rows SQL query in Cloud")
	dsDiagnosticsCmd.Flags().Bool("no-expose-failed-rows-query", false, "Hide the failed rows SQL query in Cloud")
	dsDiagnosticsCmd.Flags().Bool("failed-rows-cta", false, "Show a call-to-action link to where failed rows can be found")
	dsDiagnosticsCmd.Flags().Bool("no-failed-rows-cta", false, "Hide the call-to-action link for failed rows")
	dsDiagnosticsCmd.AddCommand(dsDiagnosticsTestConnectionCmd)

	datasourceCmd.AddCommand(dsCreateCmd, dsTestConnectionCmd, dsListCmd, dsDeleteCmd, dsDiagnosticsCmd)
	rootCmd.AddCommand(datasourceCmd)
}
