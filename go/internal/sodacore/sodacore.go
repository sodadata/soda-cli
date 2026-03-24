package sodacore

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/soda-data-inc/soda-cli/internal/output"
)

// VerifyOpts holds the options for building soda-core verify arguments.
type VerifyOpts struct {
	ContractFile   string
	DatasourceFile string
	SetVars        []string // key=value pairs
	Verbose        bool
	Publish        bool   // --push was set
	SodaCloudFile  string // temp file path for soda-cloud config
}

// FindBinary locates the soda-core binary on PATH.
// Returns the path or an ExitError with install instructions.
func FindBinary() (string, error) {
	path, err := exec.LookPath("soda")
	if err != nil {
		return "", output.Errorf(2, InstallHint())
	}
	return path, nil
}

// InstallHint returns a message with install instructions for soda-core.
func InstallHint() string {
	return `soda-core is not installed or not on your PATH.

Install it with pip (choose your connector):
  pip install soda-postgres      # PostgreSQL
  pip install soda-snowflake     # Snowflake
  pip install soda-bigquery      # BigQuery
  pip install soda-sparkdf      # Spark / Databricks

Full list: https://docs.soda.io/reference/data-source-reference-for-soda-core`
}

// CheckVersion runs "soda --version" and returns the version string.
func CheckVersion(binPath string) string {
	out, err := exec.Command(binPath, "--version").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(out))
}

// BuildVerifyArgs constructs the soda-core CLI arguments for contract verification.
// Uses Python soda-core flags: --contract/-c and --data-source/-ds.
func BuildVerifyArgs(opts VerifyOpts) []string {
	args := []string{
		"contract", "verify",
		"--contract", opts.ContractFile,
		"--data-source", opts.DatasourceFile,
	}
	if opts.Verbose {
		args = append(args, "--verbose")
	}
	for _, s := range opts.SetVars {
		args = append(args, "--set", s)
	}
	if opts.Publish && opts.SodaCloudFile != "" {
		args = append(args, "--soda-cloud", opts.SodaCloudFile, "--publish")
	}
	return args
}

// RunResult holds the output of a soda-core execution.
type RunResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Run executes soda-core with the given arguments.
// When stream is true, stdout is piped to the terminal and stderr is
// both displayed and captured (so callers can detect errors).
// When false, all output is captured and returned.
func Run(binPath string, args []string, stream bool) (*RunResult, error) {
	cmd := exec.Command(binPath, args...)
	cmd.Env = os.Environ()

	if stream {
		var stderr bytes.Buffer
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
		err := cmd.Run()
		return &RunResult{
			ExitCode: exitCodeFromErr(err),
			Stderr:   stderr.String(),
		}, nil
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return &RunResult{
		ExitCode: exitCodeFromErr(err),
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}

func exitCodeFromErr(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return 2
}

// MapExitCode maps a soda-core exit code to a sodacli exit code.
// stderr is checked to distinguish real warnings (exit 2) from usage errors
// (which also use exit 2 in cobra-based CLIs).
//
//	soda-core 0 (OK)              → 0
//	soda-core 1 (CHECK_FAILURES)  → 1
//	soda-core 2 (CHECK_WARNINGS)  → 0 (warnings are non-fatal)
//	soda-core 3 (LOG_ERRORS)      → 2 (error)
//	anything else                  → 2 (error)
func MapExitCode(sodaCoreExit int, stderr string) int {
	switch sodaCoreExit {
	case 0:
		return 0
	case 1:
		return 1
	case 2:
		// Exit code 2 can mean CHECK_WARNINGS (non-fatal) or a CLI usage error.
		// If stderr contains error indicators, treat as error.
		if strings.Contains(stderr, "Error") || strings.Contains(stderr, "error:") || strings.Contains(stderr, "unknown flag") || strings.Contains(stderr, "unknown shorthand") || strings.Contains(stderr, "unrecognized arguments") {
			return 2
		}
		return 0
	case 3:
		return 2
	default:
		return 2
	}
}
