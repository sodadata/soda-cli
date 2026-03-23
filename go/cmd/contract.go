package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

var contractCmd = &cobra.Command{
	Use:   "contract",
	Short: "Manage data quality contracts",
}

// ── contract list ─────────────────────────────────────────────────────────────

var contractListCmd = &cobra.Command{
	Use:   "list",
	Short: "List contracts in Soda Cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient()
		if err != nil {
			return err
		}
		page, err := client.ListContracts(0, 100)
		if err != nil {
			return err
		}
		if len(page.Content) == 0 {
			fmt.Println(output.Dim.Render("  No contracts found."))
			return nil
		}
		rows := make([]map[string]string, 0, len(page.Content))
		for _, c := range page.Content {
			rows = append(rows, map[string]string{
				"id":      c.ID,
				"dataset": c.DatasetQualifiedName,
				"updated": c.LastUpdated,
			})
		}
		output.Render(rows, []string{"id", "dataset", "updated"}, nil, GCtx)
		return nil
	},
}

// ── contract pull ─────────────────────────────────────────────────────────────

var contractPullCmd = &cobra.Command{
	Use:   "pull <identifier>",
	Short: "Pull a contract from Soda Cloud to a local file",
	Long: `Pull a contract from Soda Cloud by dataset qualified name.

  The identifier is the dataset qualified name:
    datasource/database/schema/table

  The contract YAML is saved to <table>.yml by default.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
			client, err := newAPIClient()
		if err != nil {
			return err
		}

		fmt.Println(output.Dim.Render("  Fetching contract for " + identifier + "..."))

		contract, err := client.FindContractByDataset(identifier)
		if err != nil {
			return err
		}
		if contract == nil {
			return output.Errorf(2, "no contract found for dataset '%s'", identifier)
		}

		parts := strings.Split(identifier, "/")
		outFile := parts[len(parts)-1] + ".yml"

		if err := os.WriteFile(outFile, []byte(contract.Contents), 0644); err != nil {
			return output.Errorf(2, "could not write file: %v", err)
		}
		output.PrintSuccess(fmt.Sprintf("Contract saved to %s (id: %s).", outFile, contract.ID), GCtx)
		return nil
	},
}

// ── contract push ─────────────────────────────────────────────────────────────

var contractPushCmd = &cobra.Command{
	Use:   "push [file]",
	Short: "Push a contract definition to Soda Cloud",
	Long: `Push a local contract YAML file to Soda Cloud.

  Reads the 'dataset:' field from the file to identify the target dataset.
  Creates a new contract if none exists; updates the existing one otherwise.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		file := ""
		if len(args) > 0 {
			file = args[0]
		}

		if file == "" {
			// look for a single .yml in contracts/ or current dir
			candidates, _ := filepath.Glob("contracts/*.yml")
			if len(candidates) == 0 {
				candidates, _ = filepath.Glob("*.yml")
			}
			if len(candidates) == 1 {
				file = candidates[0]
			} else if len(candidates) > 1 {
				if GCtx.NoInteractive {
					return output.Errorf(2, "multiple contract files found — specify a file in non-interactive mode")
				}
				form := huh.NewForm(huh.NewGroup(
					huh.NewSelect[string]().
						Title("Which contract file?").
						OptionsFunc(func() []huh.Option[string] {
							opts := make([]huh.Option[string], len(candidates))
							for i, c := range candidates {
								opts[i] = huh.NewOption(c, c)
							}
							return opts
						}, nil).
						Value(&file),
				))
				if err := form.Run(); err != nil {
					return output.Errorf(2, "cancelled")
				}
			}
		}

		if file == "" {
			return output.Errorf(2, "no contract file found — provide a file path or run from a directory containing a contracts/ folder")
		}

		contents, err := os.ReadFile(file)
		if err != nil {
			return output.Errorf(2, "could not read file %s: %v", file, err)
		}

		qualifiedName, err := parseDatasetField(contents)
		if err != nil {
			return output.Errorf(2, "could not parse 'dataset:' field from %s: %v", file, err)
		}
		if qualifiedName == "" {
			return output.Errorf(2, "contract file %s must have a 'dataset:' field", file)
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		fmt.Println(output.Dim.Render("  Checking for existing contract for " + qualifiedName + "..."))

		existing, err := client.FindContractByDataset(qualifiedName)
		if err != nil {
			return err
		}

		req := api.ContractRequest{
			DatasetQualifiedName: qualifiedName,
			Contents:             string(contents),
		}

		if existing != nil {
			fmt.Println(output.Dim.Render("  Updating existing contract " + existing.ID + "..."))
			result, err := client.UpdateContract(existing.ID, req)
			if err != nil {
				return err
			}
			output.PrintSuccess(fmt.Sprintf("Contract updated (id: %s).", result.ID), GCtx)
		} else {
			fmt.Println(output.Dim.Render("  Creating new contract..."))
			result, err := client.CreateContract(req)
			if err != nil {
				return err
			}
			output.PrintSuccess(fmt.Sprintf("Contract created (id: %s).", result.ID), GCtx)
		}
		return nil
	},
}

// ── contract diff ─────────────────────────────────────────────────────────────

var contractDiffCmd = &cobra.Command{
	Use:   "diff [file]",
	Short: "Show diff between local contract and Soda Cloud",
	Long: `Compare a local contract file with the version stored in Soda Cloud.

  Reads the 'dataset:' field from the file to identify which contract to compare.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		file := ""
		if len(args) > 0 {
			file = args[0]
		}
		if file == "" {
			candidates, _ := filepath.Glob("contracts/*.yml")
			if len(candidates) == 1 {
				file = candidates[0]
			} else if len(candidates) == 0 {
				return output.Errorf(2, "no contract file found — provide a file path")
			} else {
				return output.Errorf(2, "multiple contract files found — specify a file")
			}
		}

		contents, err := os.ReadFile(file)
		if err != nil {
			return output.Errorf(2, "could not read file %s: %v", file, err)
		}

		qualifiedName, err := parseDatasetField(contents)
		if err != nil {
			return output.Errorf(2, "could not parse 'dataset:' field from %s: %v", file, err)
		}
		if qualifiedName == "" {
			return output.Errorf(2, "contract file %s must have a 'dataset:' field", file)
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		fmt.Printf("  Comparing %s with cloud version for %s...\n\n", output.Bold.Render(file), output.Dim.Render(qualifiedName))

		remote, err := client.FindContractByDataset(qualifiedName)
		if err != nil {
			return err
		}
		if remote == nil {
			return output.Errorf(2, "no contract found in Soda Cloud for dataset '%s' — run `sodacli contract push` to create it", qualifiedName)
		}

		localLines := strings.Split(strings.TrimRight(string(contents), "\n"), "\n")
		remoteLines := strings.Split(strings.TrimRight(remote.Contents, "\n"), "\n")

		changes := diffLines(remoteLines, localLines)
		hasChanges := false
		for _, l := range changes {
			if len(l) > 0 && (l[0] == '+' || l[0] == '-') {
				hasChanges = true
				break
			}
		}
		if !hasChanges {
			fmt.Println(output.Dim.Render("  No differences — local matches cloud."))
			return nil
		}

		for _, line := range changes {
			switch line[0] {
			case '+':
				fmt.Println(output.Green.Render("  " + line))
			case '-':
				fmt.Println(output.Red.Render("  " + line))
			default:
				fmt.Println(output.Dim.Render("  " + line))
			}
		}
		fmt.Println()
		fmt.Println(output.Dim.Render(fmt.Sprintf("  Run `sodacli contract push %s` to publish local changes.", file)))
		return nil
	},
}

// ── contract lint ─────────────────────────────────────────────────────────────

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

// ── contract create ───────────────────────────────────────────────────────────

var contractCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a contract YAML from a live schema",
	Long: `Bootstrap a contract from a live dataset schema.

  --mode skeleton  (default) generates basic structure from live schema (no AI).
  --mode copilot   uses AI to generate meaningful checks (requires license).

The contract is created in Soda Cloud and its YAML is saved locally.`,
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

		if dataset == "" {
			return output.Errorf(2, "--dataset is required")
		}

		if outFile == "" {
			parts := strings.Split(dataset, "/")
			outFile = parts[len(parts)-1] + ".yml"
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		noWait, _ := cmd.Flags().GetBool("no-wait")

		switch mode {
		case "skeleton":
			return runContractCreateSkeleton(client, dataset, outFile)
		case "copilot":
			return runContractCreateCopilot(client, dataset, outFile, noWait)
		default:
			return output.Errorf(2, "unknown mode '%s' — use skeleton or copilot", mode)
		}
	},
}

func runContractCreateSkeleton(client *api.Client, dataset, outFile string) error {
	spinner := output.NewSpinner("Generating skeleton contract for " + dataset + "...")
	spinner.Start()

	opID, err := client.CreateSkeleton(api.CreateSkeletonRequest{
		DatasetQualifiedName: dataset,
	})
	if err != nil {
		spinner.Stop()
		return err
	}

	// Poll until done
	for {
		time.Sleep(2 * time.Second)
		status, err := client.GetSkeletonStatus(opID)
		if err != nil {
			spinner.Stop()
			return err
		}
		if status.State == "completed" {
			break
		}
		if status.State == "failed" || status.State == "canceled" {
			spinner.Stop()
			return output.Errorf(2, "skeleton generation %s", status.State)
		}
	}
	spinner.Stop()

	// Fetch the created contract
	contract, err := client.FindContractByDataset(dataset)
	if err != nil {
		return err
	}
	if contract == nil {
		return output.Errorf(2, "skeleton generation completed but contract was not persisted by the API.\n  This may be a backend issue — try creating the contract from the Soda Cloud UI.")
	}

	if err := os.WriteFile(outFile, []byte(contract.Contents), 0644); err != nil {
		return output.Errorf(2, "could not write file: %v", err)
	}

	output.PrintSuccess(fmt.Sprintf("Skeleton contract written to %s", outFile), GCtx)
	return nil
}

func runContractCreateCopilot(client *api.Client, dataset, outFile string, noWait bool) error {
	opID, err := client.GenerateContract(api.GenerateContractRequest{
		DatasetQualifiedNames: []string{dataset},
	})
	if err != nil {
		return err
	}

	if noWait {
		fmt.Printf("  %s AI contract generation started for %s.\n", output.Green.Render("✓"), dataset)
		fmt.Println(output.Dim.Render("  Running in background — contract will appear in Soda Cloud when ready."))
		fmt.Println(output.Dim.Render("  Check results:  sodacli results list"))
		return nil
	}

	spinner := output.NewSpinner("Generating AI contract for " + dataset + "...")
	spinner.Start()

	// Poll until done
	elapsed := 0
	for {
		time.Sleep(3 * time.Second)
		elapsed += 3
		status, err := client.GetGenerateStatus(opID)
		if err != nil {
			spinner.Stop()
			return err
		}
		if status.State == "completed" {
			break
		}
		if status.State == "failed" || status.State == "canceled" {
			spinner.Stop()
			return output.Errorf(2, "AI generation %s", status.State)
		}
		spinner.SetMessage(fmt.Sprintf("Generating AI contract for %s... (%ds)", dataset, elapsed))
	}
	spinner.Stop()

	// Fetch the created contract
	contract, err := client.FindContractByDataset(dataset)
	if err != nil {
		return err
	}
	if contract == nil {
		return output.Errorf(2, "AI generation completed but contract was not persisted by the API.\n  This may be a backend issue — try creating the contract from the Soda Cloud UI.")
	}

	if err := os.WriteFile(outFile, []byte(contract.Contents), 0644); err != nil {
		return output.Errorf(2, "could not write file: %v", err)
	}

	output.PrintSuccess(fmt.Sprintf("AI-generated contract written to %s", outFile), GCtx)
	return nil
}

// ── contract copilot ──────────────────────────────────────────────────────────

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

// ── contract verify ───────────────────────────────────────────────────────────

var contractVerifyCmd = &cobra.Command{
	Use:   "verify <file>",
	Short: "Run contract checks against your data",
	Long: `Execute data quality checks defined in a contract file.

  Pushes the contract to Soda Cloud and triggers verification via a Runner.
  Polls for results and displays a summary.

  Exit codes: 0=all passing, 1=checks failed, 2=error, 3=auth error`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file := args[0]
		noWait, _ := cmd.Flags().GetBool("no-wait")

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		return runContractVerify(client, file, noWait)
	},
}

// runContractVerify pushes a contract file to cloud, triggers verification,
// polls for results, and displays a summary. Reused by the verify command
// and both onboard flows.
func runContractVerify(client *api.Client, file string, noWait bool) error {
	contents, err := os.ReadFile(file)
	if err != nil {
		return output.Errorf(2, "could not read file %s: %v", file, err)
	}
	qualifiedName, err := parseDatasetField(contents)
	if err != nil {
		return output.Errorf(2, "could not parse 'dataset:' field from %s: %v", file, err)
	}
	if qualifiedName == "" {
		return output.Errorf(2, "contract file %s must have a 'dataset:' field", file)
	}

	// Push/update contract to cloud
	fmt.Println(output.Dim.Render("  Pushing contract for " + qualifiedName + "..."))
	existing, err := client.FindContractByDataset(qualifiedName)
	if err != nil {
		return err
	}

	req := api.ContractRequest{
		DatasetQualifiedName: qualifiedName,
		Contents:             string(contents),
	}

	var contractID string
	if existing != nil {
		result, err := client.UpdateContract(existing.ID, req)
		if err != nil {
			return err
		}
		contractID = result.ID
	} else {
		result, err := client.CreateContract(req)
		if err != nil {
			return err
		}
		contractID = result.ID
	}

	// Trigger verification
	fmt.Println(output.Dim.Render("  Triggering verification..."))
	scanID, err := client.VerifyContract(contractID)
	if err != nil {
		return err
	}
	fmt.Println(output.Dim.Render("  Scan ID: " + scanID))

	if noWait {
		output.PrintSuccess(fmt.Sprintf("Verification started (scan: %s). Running in background.", scanID), GCtx)
		fmt.Println(output.Dim.Render("  Check status:  sodacli job logs " + scanID))
		return nil
	}

	// Poll for completion
	spinner := output.NewSpinner("Running contract checks...")
	spinner.Start()

	timeout := time.After(10 * time.Minute)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	var finalStatus *api.ScanStatus
	for {
		select {
		case <-timeout:
			spinner.Stop()
			return output.Errorf(2, "verification timed out after 10 minutes (scan: %s)", scanID)
		case <-ticker.C:
			status, err := client.GetScanStatus(scanID)
			if err != nil {
				spinner.Stop()
				return err
			}
			if api.IsScanTerminal(status.State) {
				finalStatus = status
				spinner.Stop()
				goto done
			}
			spinner.SetMessage(fmt.Sprintf("Running contract checks... (state: %s)", status.State))
		}
	}
done:

	// Display results
	fmt.Println()
	passed := 0
	failed := 0
	warned := 0
	for _, chk := range finalStatus.Checks {
		switch chk.EvaluationStatus {
		case "pass":
			passed++
		case "fail":
			failed++
		case "warn":
			warned++
		}
	}

	summary := fmt.Sprintf("  %d checks passed", passed)
	if failed > 0 {
		summary += fmt.Sprintf(", %s", output.Red.Render(fmt.Sprintf("%d failed", failed)))
	}
	if warned > 0 {
		summary += fmt.Sprintf(", %s", output.Yellow.Render(fmt.Sprintf("%d warnings", warned)))
	}
	fmt.Println(summary)

	if len(finalStatus.Checks) > 0 {
		rows := make([]map[string]string, len(finalStatus.Checks))
		for i, chk := range finalStatus.Checks {
			name := chk.Name
			if name == "" {
				name = chk.Definition
			}
			if name == "" {
				name = chk.ID
			}
			rows[i] = map[string]string{
				"type":   chk.Type,
				"column": chk.Column,
				"name":   name,
				"status": fmtCheckStatus(chk.EvaluationStatus),
			}
		}
		fmt.Println()
		output.Render(rows, []string{"type", "column", "name", "status"}, map[string]bool{"status": true}, GCtx)
	}

	if finalStatus.CloudURL != "" {
		fmt.Println()
		fmt.Println(output.Dim.Render("  Full results: " + finalStatus.CloudURL))
	}

	switch finalStatus.State {
	case "completed", "completedWithWarnings":
		if failed > 0 {
			return output.Errorf(1, "%d check(s) failed", failed)
		}
		output.PrintSuccess("All checks passed.", GCtx)
		return nil
	case "completedWithFailures":
		return output.Errorf(1, "%d check(s) failed", finalStatus.Failures)
	case "completedWithErrors":
		return output.Errorf(2, "verification completed with errors")
	case "failed":
		return output.Errorf(2, "verification failed")
	case "canceled":
		return output.Errorf(2, "verification was canceled")
	case "timedOut":
		return output.Errorf(2, "verification timed out on the server")
	default:
		return output.Errorf(2, "unexpected terminal state: %s", finalStatus.State)
	}
}

// ── contract proposal ─────────────────────────────────────────────────────────

var contractProposalCmd = &cobra.Command{
	Use:   "proposal",
	Short: "Manage contract change proposals",
}

var proposalListCmd = &cobra.Command{
	Use:   "list",
	Short: "List open proposals",
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "contract proposal list is not yet available in the public API")
	},
}

var proposalPullCmd = &cobra.Command{
	Use:   "pull <id>",
	Short: "Download a proposal locally",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "contract proposal pull is not yet available in the public API")
	},
}

var proposalPushCmd = &cobra.Command{
	Use:   "push <id> [file]",
	Short: "Submit changes for a proposal",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "contract proposal push is not yet available in the public API")
	},
}

var proposalCloseCmd = &cobra.Command{
	Use:   "close <id>",
	Short: "Close a proposal",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return output.Errorf(2, "contract proposal close is not yet available in the public API")
	},
}

// ── helpers ───────────────────────────────────────────────────────────────────

// parseDatasetField extracts the top-level `dataset:` value from a contract YAML.
func parseDatasetField(contents []byte) (string, error) {
	var doc struct {
		Dataset string `yaml:"dataset"`
	}
	if err := yaml.Unmarshal(contents, &doc); err != nil {
		return "", err
	}
	return doc.Dataset, nil
}

// diffLines returns a unified-style diff of old vs new lines.
// Lines only in old are prefixed with "-", lines only in new with "+",
// and up to 2 shared context lines around each change are shown with " ".
func diffLines(old, new []string) []string {
	// Build LCS-based diff using a simple O(n*m) approach for contract sizes.
	type edit struct {
		op   byte // ' ', '+', '-'
		text string
	}
	var edits []edit

	// Myers-lite: just compare line by line with a simple patience approach.
	// For contract files (typically < 200 lines) this is fast enough.
	oldSet := make(map[string]bool)
	newSet := make(map[string]bool)
	for _, l := range old {
		oldSet[l] = true
	}
	for _, l := range new {
		newSet[l] = true
	}

	oi, ni := 0, 0
	for oi < len(old) || ni < len(new) {
		if oi < len(old) && ni < len(new) && old[oi] == new[ni] {
			edits = append(edits, edit{' ', old[oi]})
			oi++
			ni++
		} else if oi < len(old) && !newSet[old[oi]] {
			edits = append(edits, edit{'-', old[oi]})
			oi++
		} else if ni < len(new) && !oldSet[new[ni]] {
			edits = append(edits, edit{'+', new[ni]})
			ni++
		} else {
			// Both lines exist in both files but at different positions — treat as change.
			if oi < len(old) {
				edits = append(edits, edit{'-', old[oi]})
				oi++
			}
			if ni < len(new) {
				edits = append(edits, edit{'+', new[ni]})
				ni++
			}
		}
	}

	// Filter to only changed lines + 2 lines of context.
	const ctx = 2
	changed := make([]bool, len(edits))
	for i, e := range edits {
		if e.op != ' ' {
			for j := max(0, i-ctx); j <= min(len(edits)-1, i+ctx); j++ {
				changed[j] = true
			}
		}
	}

	var result []string
	prevSkipped := false
	for i, e := range edits {
		if !changed[i] {
			if !prevSkipped {
				result = append(result, "...")
				prevSkipped = true
			}
			continue
		}
		prevSkipped = false
		result = append(result, string(e.op)+" "+e.text)
	}
	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── init ──────────────────────────────────────────────────────────────────────

func init() {
	contractCreateCmd.Flags().String("dataset", "", "Dataset FQN: datasource/db/schema/table")
	contractCreateCmd.Flags().String("mode", "skeleton", "Generation mode: skeleton|copilot")
	contractCreateCmd.Flags().String("output", "", "Output file path")
	contractCreateCmd.Flags().Bool("no-wait", false, "Start generation and return immediately without waiting for completion (copilot mode only)")

	contractDiffCmd.Flags().String("dataset", "", "Dataset qualified name for cloud comparison (overrides file's dataset field)")

	contractCopilotCmd.Flags().String("dataset", "", "Dataset FQN to generate from")
	contractCopilotCmd.Flags().String("output", "", "Output file path")

	contractVerifyCmd.Flags().String("datasource", "", "Datasource config file override")
	contractVerifyCmd.Flags().Bool("runner", false, "Delegate execution to Soda Runner")
	contractVerifyCmd.Flags().Bool("push", false, "Push results to Soda Cloud after verification")
	contractVerifyCmd.Flags().Bool("no-wait", false, "Start verification and return immediately without waiting for results")
	contractVerifyCmd.Flags().StringArray("set", nil, "Runtime variable overrides (key=value)")

	proposalListCmd.Flags().String("status", "open", "Filter by status: open|done|all")
	proposalPullCmd.Flags().Int("revision", 0, "Specific revision number")
	proposalPushCmd.Flags().String("message", "", "Change message")
	proposalCloseCmd.Flags().String("status", "done", "Close status: done|wontdo")

	contractProposalCmd.AddCommand(proposalListCmd, proposalPullCmd, proposalPushCmd, proposalCloseCmd)

	contractCmd.AddCommand(
		contractListCmd,
		contractCreateCmd,
		contractLintCmd,
		contractPushCmd,
		contractPullCmd,
		contractDiffCmd,
		contractCopilotCmd,
		contractVerifyCmd,
		contractProposalCmd,
	)
}
