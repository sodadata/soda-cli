package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/soda-data-inc/soda-cli/internal/mock"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var contractCmd = &cobra.Command{
	Use:   "contract",
	Short: "Manage data quality contracts",
}

var contractCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a contract YAML from a live schema",
	Long: `Bootstrap a contract from a live dataset schema.

  --mode skeleton  (default) generates basic structure checks with no AI.
  --mode copilot   uses AI to generate meaningful checks (requires license).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataset, _ := cmd.Flags().GetString("dataset")
		mode, _ := cmd.Flags().GetString("mode")
		outFile, _ := cmd.Flags().GetString("output")

		if dataset == "" {
			if GCtx.NoInteractive {
				return output.Errorf(2, "--dataset is required in non-interactive mode")
			}
			form := huh.NewForm(huh.NewGroup(
				huh.NewInput().
					Title("Dataset (datasource/db/schema/table)").
					Placeholder("pg_prod/mydb/public/orders").
					Value(&dataset),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
		}

		if outFile == "" {
			parts := strings.Split(dataset, "/")
			outFile = parts[len(parts)-1] + ".yml"
		}

		if mode == "copilot" {
			fmt.Println(output.Dim.Render("  Connecting to Soda Copilot..."))
			fmt.Println(output.Dim.Render("  Analyzing schema for " + dataset + "..."))
			fmt.Println(output.Dim.Render("  Generating AI-powered checks..."))
		} else {
			fmt.Println(output.Dim.Render("  Reading schema for " + dataset + "..."))
		}

		output.PrintSuccess(fmt.Sprintf("Contract written to %s", outFile), GCtx)
		return nil
	},
}

var contractLintCmd = &cobra.Command{
	Use:     "lint [file]",
	Aliases: []string{"validate"},
	Short:   "Validate contract syntax (no network required)",
	RunE: func(cmd *cobra.Command, args []string) error {
		file := "contracts/*.yml"
		if len(args) > 0 {
			file = args[0]
		}
		fmt.Println(output.Dim.Render("  Linting " + file + "..."))
		output.PrintSuccess("Contract syntax is valid.", GCtx)
		return nil
	},
}

var contractPushCmd = &cobra.Command{
	Use:   "push [file]",
	Short: "Push contract definition to Soda Cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		file := "contracts/*.yml"
		if len(args) > 0 {
			file = args[0]
		}
		fmt.Println(output.Dim.Render("  Pushing " + file + " to Soda Cloud..."))
		output.PrintSuccess("Contract published to Soda Cloud.", GCtx)
		return nil
	},
}

var contractPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull contract from Soda Cloud to a local file",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataset, _ := cmd.Flags().GetString("dataset")
		if dataset == "" {
			return output.Errorf(2, "--dataset is required")
		}
		parts := strings.Split(dataset, "/")
		outFile := parts[len(parts)-1] + ".yml"
		fmt.Println(output.Dim.Render("  Pulling contract for " + dataset + "..."))
		output.PrintSuccess(fmt.Sprintf("Contract saved to %s", outFile), GCtx)
		return nil
	},
}

var contractDiffCmd = &cobra.Command{
	Use:   "diff [file]",
	Short: "Show diff between local contract and Soda Cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		file := "contracts/*.yml"
		if len(args) > 0 {
			file = args[0]
		}
		fmt.Printf("  Comparing %s with cloud version...\n\n", file)
		fmt.Println(output.Green.Render("  + freshness(created_at) < 24h"))
		fmt.Println(output.Red.Render("  - no_nulls(shipping_address)"))
		fmt.Println()
		fmt.Println(output.Dim.Render("  2 changes. Run `soda contract push` to publish."))
		return nil
	},
}

var contractCopilotCmd = &cobra.Command{
	Use:   "copilot [file] [prompt]",
	Short: "AI-powered contract generation and improvement",
	Long: `Use Soda Copilot to generate or improve contracts with AI.

  No args           → wizard: generate or improve?
  file, no prompt   → wizard: what to improve?
  --dataset only    → generate from scratch
  file + prompt     → improve existing contract
  --no-interactive  → fails with clear error if required args missing`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataset, _ := cmd.Flags().GetString("dataset")

		if len(args) == 0 && dataset == "" {
			if GCtx.NoInteractive {
				return output.Errorf(2, "provide a file, a prompt, or --dataset in non-interactive mode")
			}
			mode := "generate"
			form := huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to do?").
					Options(
						huh.NewOption("Generate a new contract from a dataset", "generate"),
						huh.NewOption("Improve an existing contract", "improve"),
					).
					Value(&mode),
			))
			if err := form.Run(); err != nil {
				return output.Errorf(2, "cancelled")
			}
			if mode == "generate" {
				return runCopilotGenerate(dataset)
			}
			return runCopilotImprove("", "")
		}

		file := ""
		prompt := ""
		if len(args) >= 1 {
			file = args[0]
		}
		if len(args) >= 2 {
			prompt = strings.Join(args[1:], " ")
		}

		if file != "" && prompt == "" {
			if GCtx.NoInteractive {
				return output.Errorf(2, "provide a prompt when passing a file in non-interactive mode")
			}
			return runCopilotImprove(file, "")
		}

		if dataset != "" {
			return runCopilotGenerate(dataset)
		}

		return runCopilotImprove(file, prompt)
	},
}

func runCopilotGenerate(dataset string) error {
	if dataset == "" {
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Dataset (datasource/db/schema/table)").
				Placeholder("pg_prod/mydb/public/orders").
				Value(&dataset),
		))
		if err := form.Run(); err != nil {
			return output.Errorf(2, "cancelled")
		}
	}
	fmt.Println(output.Dim.Render("  Connecting to Soda Copilot..."))
	fmt.Println(output.Dim.Render("  Analyzing schema and data profile for " + dataset + "..."))
	fmt.Println(output.Dim.Render("  Generating contract..."))
	fmt.Println()
	output.PrintSuccess("Contract generated and saved to orders.yml", GCtx)
	return nil
}

func runCopilotImprove(file, prompt string) error {
	if file == "" {
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Contract file to improve").
				Placeholder("contracts/orders.yml").
				Value(&file),
		))
		if err := form.Run(); err != nil {
			return output.Errorf(2, "cancelled")
		}
	}
	if prompt == "" {
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("What would you like to improve?").
				Placeholder("Add freshness checks and tighten null constraints").
				Value(&prompt),
		))
		if err := form.Run(); err != nil {
			return output.Errorf(2, "cancelled")
		}
	}
	fmt.Println(output.Dim.Render("  Sending to Soda Copilot..."))
	fmt.Println(output.Dim.Render("  Applying: " + prompt))
	fmt.Println()
	output.PrintSuccess("Contract updated in "+file, GCtx)
	return nil
}

var contractVerifyCmd = &cobra.Command{
	Use:   "verify [file-or-dir]",
	Short: "Run contract checks against your data",
	Long: `Execute data quality checks defined in a contract file.

  Connects to the configured datasource and runs all checks.
  Use --push to send results to Soda Cloud after verification.

  Exit codes: 0=all passing, 1=checks failed, 2=error, 3=auth error`,
	RunE: func(cmd *cobra.Command, args []string) error {
		file := "contracts/*.yml"
		if len(args) > 0 {
			file = args[0]
		}
		push, _ := cmd.Flags().GetBool("push")

		fmt.Printf("  Verifying %s\n\n", output.Bold.Render(file))

		failCount := 0
		for _, check := range mock.ContractChecks {
			if check.Status == "pass" {
				fmt.Printf("  %s  %-45s  %s\n",
					output.Green.Render("✓"),
					check.Name,
					output.Dim.Render(check.Value),
				)
			} else {
				fmt.Printf("  %s  %-45s  %s\n",
					output.Red.Render("✗"),
					check.Name,
					output.Dim.Render(check.Value),
				)
				failCount++
			}
		}

		passing := len(mock.ContractChecks) - failCount
		fmt.Println()

		if failCount == 0 {
			fmt.Printf("  %s  All %d checks passed.\n", output.Green.Render("✓"), passing)
		} else {
			fmt.Printf("  %s  %d/%d checks failed.\n",
				output.Red.Render("✗"),
				failCount,
				len(mock.ContractChecks),
			)
		}

		if push {
			fmt.Println()
			fmt.Println(output.Dim.Render("  Pushing results to Soda Cloud..."))
			output.PrintSuccess("Results pushed. Job ID: sc_abc123", GCtx)
		}

		if failCount > 0 {
			return &output.ExitError{Code: 1, Msg: ""}
		}
		return nil
	},
}

// contract proposal sub-group
var contractProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage contract change proposals",
}

var proposalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List open proposals",
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")
		rows := mock.Proposals
		if status != "" && status != "all" {
			filtered := []map[string]string{}
			for _, p := range rows {
				if p["status"] == status {
					filtered = append(filtered, p)
				}
			}
			rows = filtered
		}
		cols := []string{"id", "dataset", "status", "message", "created"}
		output.Render(rows, cols, map[string]bool{"status": true}, GCtx)
		return nil
	},
}

var proposalPullCmd = &cobra.Command{
	Use:   "pull <id>",
	Short: "Download a proposal locally",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		revision, _ := cmd.Flags().GetInt("revision")
		suffix := ""
		if revision > 0 {
			suffix = fmt.Sprintf(" (revision %d)", revision)
		}
		output.PrintSuccess(fmt.Sprintf("Proposal %s%s saved to proposal_%s.yml", args[0], suffix, args[0]), GCtx)
		return nil
	},
}

var proposalPushCmd = &cobra.Command{
	Use:   "push <id> [file]",
	Short: "Submit changes for a proposal",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, _ := cmd.Flags().GetString("message")
		if msg == "" {
			msg = "Updated contract"
		}
		output.PrintSuccess(fmt.Sprintf("Proposal %s updated: %s", args[0], msg), GCtx)
		return nil
	},
}

var proposalCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a proposal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		status, _ := cmd.Flags().GetString("status")
		if status == "" {
			status = "done"
		}
		output.PrintSuccess(fmt.Sprintf("Proposal %s closed as '%s'.", args[0], status), GCtx)
		return nil
	},
}

func init() {
	contractCreateCmd.Flags().String("dataset", "", "Dataset FQN: datasource/db/schema/table")
	contractCreateCmd.Flags().String("mode", "skeleton", "Generation mode: skeleton|copilot")
	contractCreateCmd.Flags().String("output", "", "Output file path")

	contractPullCmd.Flags().String("dataset", "", "Dataset FQN (required)")

	contractDiffCmd.Flags().String("dataset", "", "Dataset FQN for cloud comparison")

	contractCopilotCmd.Flags().String("dataset", "", "Dataset FQN to generate from")
	contractCopilotCmd.Flags().String("output", "", "Output file path")

	contractVerifyCmd.Flags().String("datasource", "", "Datasource config file override")
	contractVerifyCmd.Flags().Bool("agent", false, "Delegate execution to Soda Agent")
	contractVerifyCmd.Flags().Bool("push", false, "Push results to Soda Cloud after verification")
	contractVerifyCmd.Flags().StringArray("set", nil, "Runtime variable overrides (key=value)")

	proposalListCmd.Flags().String("status", "open", "Filter by status: open|done|all")
	proposalPullCmd.Flags().Int("revision", 0, "Specific revision number")
	proposalPushCmd.Flags().String("message", "", "Change message")
	proposalCloseCmd.Flags().String("status", "done", "Close status: done|wontdo")

	contractProposalCmd.AddCommand(proposalListCmd, proposalPullCmd, proposalPushCmd, proposalCloseCmd)

	contractCmd.AddCommand(
		contractCreateCmd,
		contractLintCmd,
		contractPushCmd,
		contractPullCmd,
		contractDiffCmd,
		contractCopilotCmd,
		contractVerifyCmd,
		contractProposalCmd,
	)
	rootCmd.AddCommand(contractCmd)
}
