package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
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
		cols := []string{"id", "name", "type", "status", "datasets", "created"}
		output.Render(mock.Datasources, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var dsCreateCmd = &cobra.Command{
	Use:   "create <config-file>",
	Short: "Register a datasource from a YAML connection config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configFile := args[0]
		agent, _ := cmd.Flags().GetString("agent")

		msg := fmt.Sprintf("Datasource registered from %s.", configFile)
		if agent != "" {
			fmt.Println(output.Dim.Render("  Registering via agent '" + agent + "'..."))
			msg = fmt.Sprintf("Datasource registered from %s via agent '%s'.", configFile, agent)
		}
		output.PrintSuccess(msg, GCtx)
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
		output.PrintSuccess(fmt.Sprintf("Datasource '%s' deleted.", args[0]), GCtx)
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
