package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

var defaultSodaYml = `# soda.yml — project configuration
# Docs: https://github.com/sodadata/soda-cli

# datasource: my_datasource
#   type: postgres
#   host: localhost
#   port: "5432"
#   database: my_db
`

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold soda.yml and project structure",
	Long:  "Create soda.yml, contracts/, and configs/ directories in the current project.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !GCtx.NoInteractive {
			fmt.Println("Scaffolding soda project structure...")
			fmt.Println()
		}

		// Create soda.yml if it doesn't already exist
		if _, err := os.Stat("soda.yml"); err == nil {
			return fmt.Errorf("soda.yml already exists in this directory")
		}
		if err := os.WriteFile("soda.yml", []byte(defaultSodaYml), 0644); err != nil {
			return fmt.Errorf("failed to create soda.yml: %w", err)
		}
		fmt.Println("  Created  soda.yml")

		// Create contracts/ and configs/ directories
		for _, dir := range []string{"contracts", "configs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create %s/: %w", dir, err)
			}
			fmt.Printf("  Created  %s/\n", dir)
		}

		fmt.Println()
		output.PrintSuccess("Project initialized. Edit soda.yml to configure your datasources.", GCtx)
		return nil
	},
}

func init() {}
