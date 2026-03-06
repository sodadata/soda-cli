package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
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
	Use:   "create",
	Short: "Create a local datasource YAML config",
	RunE: func(cmd *cobra.Command, args []string) error {
		dsType, _ := cmd.Flags().GetString("type")
		name, _ := cmd.Flags().GetString("name")

		if dsType == "" || name == "" {
			if GCtx.NoInteractive {
				return output.Errorf(2, "--type and --name are required in non-interactive mode")
			}
			types := []huh.Option[string]{
				huh.NewOption("PostgreSQL", "postgres"),
				huh.NewOption("Snowflake", "snowflake"),
				huh.NewOption("BigQuery", "bigquery"),
				huh.NewOption("DuckDB", "duckdb"),
			}
			form := huh.NewForm(huh.NewGroup(
				huh.NewInput().Title("Datasource name").Placeholder("my_warehouse").Value(&name),
				huh.NewSelect[string]().Title("Type").Options(types...).Value(&dsType),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
		}

		output.PrintSuccess(fmt.Sprintf("Created configs/%s.yml — edit it to add your connection details.", name), GCtx)
		return nil
	},
}

var dsOnboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Register a cloud datasource via Soda Agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		agent, _ := cmd.Flags().GetString("agent")
		dsType, _ := cmd.Flags().GetString("type")

		if agent == "" || dsType == "" {
			if GCtx.NoInteractive {
				return output.Errorf(2, "--agent and --type are required in non-interactive mode")
			}
		}

		fmt.Println(output.Dim.Render("  Registering datasource via agent '" + agent + "'..."))
		output.PrintSuccess("Datasource onboarded. It will appear in Soda Cloud once the agent confirms the connection.", GCtx)
		return nil
	},
}

var dsTestCmd = &cobra.Command{
	Use:   "test [name-or-file]",
	Short: "Test a datasource connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "default"
		if len(args) > 0 {
			target = args[0]
		}
		fmt.Printf(output.Dim.Render("  Testing connection to '%s'...\n"), target)
		output.PrintSuccess("Connection successful.", GCtx)
		return nil
	},
}

var dsDiagnosticsCmd = &cobra.Command{
	Use:   "diagnostics <id>",
	Short: "View or configure diagnostics warehouse for a datasource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("  %-22s %s\n", output.Bold.Render("Datasource"), args[0])
		fmt.Printf("  %-22s %s\n", output.Bold.Render("Diagnostics warehouse"), "pg_prod.soda_diagnostics")
		fmt.Printf("  %-22s %s\n", output.Bold.Render("Schema"), "soda")
		fmt.Printf("  %-22s %s\n", output.Bold.Render("Retention"), "30 days")
		return nil
	},
}

func init() {
	dsCreateCmd.Flags().String("type", "", "Datasource type: postgres|snowflake|bigquery|duckdb")
	dsCreateCmd.Flags().String("name", "", "Datasource name")
	dsOnboardCmd.Flags().String("agent", "", "Soda Agent name (required)")
	dsOnboardCmd.Flags().String("type", "", "Datasource type (required)")

	datasourceCmd.AddCommand(dsListCmd, dsCreateCmd, dsOnboardCmd, dsTestCmd, dsDiagnosticsCmd)
	rootCmd.AddCommand(datasourceCmd)
}
