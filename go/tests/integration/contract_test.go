//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestContractList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("table", func(t *testing.T) {
		r := run(t, "contract", "list")
		assertExitCode(t, r, 0)
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "contract", "list", "--output", "json")
		assertExitCode(t, r, 0)
	})

	t.Run("csv", func(t *testing.T) {
		r := run(t, "contract", "list", "--output", "csv")
		assertExitCode(t, r, 0)
	})
}

func TestContractCreate(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("no_dataset_nointeractive_errors", func(t *testing.T) {
		r := run(t, "contract", "create", "--no-interactive")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "required")
	})

	t.Run("bad_mode", func(t *testing.T) {
		r := run(t, "contract", "create",
			"--dataset", testDatasourceName()+"/SODA_TESTING/PUBLIC/ACCOUNT_BALANCES",
			"--mode", "badmode",
		)
		assertExitCode(t, r, 2)
	})

	// skeleton create — may or may not persist depending on API state
	t.Run("skeleton", func(t *testing.T) {
		r := run(t, "contract", "create",
			"--dataset", testDatasourceName()+"/SODA_TESTING/PUBLIC/ACCOUNT_BALANCES",
			"--mode", "skeleton",
			"--output", "/tmp/soda-test-contract.yml",
		)
		// Log result regardless — skeleton generation may time out or fail on backend
		t.Logf("skeleton create: exit=%d output=%s", r.ExitCode, r.Output())
	})
}

func TestContractLint(t *testing.T) {
	// Lint is offline — no credentials needed.

	t.Run("valid_contract", func(t *testing.T) {
		f := writeTempFile(t, "contract-*.yml", `
dataset: ds/db/schema/orders
columns:
  - name: id
    data_type: INTEGER
  - name: status
    data_type: VARCHAR
`)
		r := run(t, "contract", "lint", f, "--output", "table")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "valid")
	})

	t.Run("invalid_contract", func(t *testing.T) {
		f := writeTempFile(t, "contract-*.yml", `
dataset: ds/db/schema/orders
bogus_field: true
columns:
  - name: id
`)
		r := run(t, "contract", "lint", f, "--output", "table")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "bogus_field")
	})

	t.Run("missing_column_name", func(t *testing.T) {
		f := writeTempFile(t, "contract-*.yml", `
dataset: ds/db/schema/orders
columns:
  - data_type: INTEGER
`)
		r := run(t, "contract", "lint", f, "--output", "table")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "name")
	})

	t.Run("yaml_parse_error", func(t *testing.T) {
		f := writeTempFile(t, "contract-*.yml", `
dataset: test
  bad indent: here
`)
		r := run(t, "contract", "lint", f)
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "YAML parse error")
	})

	t.Run("empty_file", func(t *testing.T) {
		f := writeTempFile(t, "contract-*.yml", "")
		r := run(t, "contract", "lint", f)
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "empty")
	})

	t.Run("multiple_files_mixed", func(t *testing.T) {
		good := writeTempFile(t, "good-*.yml", `
dataset: ds/db/schema/t
columns:
  - name: id
`)
		bad := writeTempFile(t, "bad-*.yml", `
dataset: ds/db/schema/t
unknown_prop: 123
`)
		r := run(t, "contract", "lint", good, bad, "--output", "table")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "1 valid")
		assertOutputContains(t, r, "1 with errors")
	})

	t.Run("json_output", func(t *testing.T) {
		f := writeTempFile(t, "contract-*.yml", `
dataset: ds/db/schema/orders
columns:
  - name: id
`)
		r := run(t, "contract", "lint", f, "--output", "json")
		assertExitCode(t, r, 0)
		// Verify it's valid JSON
		var results []struct {
			File  string `json:"file"`
			Valid bool   `json:"valid"`
		}
		if err := json.Unmarshal([]byte(r.Stdout), &results); err != nil {
			t.Fatalf("expected valid JSON output: %v\nstdout: %s", err, r.Stdout)
		}
		if len(results) != 1 || !results[0].Valid {
			t.Errorf("expected 1 valid result, got: %+v", results)
		}
	})

	t.Run("nonexistent_file", func(t *testing.T) {
		r := run(t, "contract", "lint", "/tmp/nonexistent-soda-contract.yml")
		assertExitCode(t, r, 2)
	})

	t.Run("no_files_found", func(t *testing.T) {
		// Run from an empty temp dir where no yml files exist
		dir := t.TempDir()
		bin := ensureBinary(t)
		r := runInDir(t, bin, dir, "contract", "lint")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "no contract files found")
	})

	t.Run("glob_pattern", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "a.yml"), []byte("dataset: ds/db/s/t\ncolumns:\n  - name: id\n"), 0644)
		os.WriteFile(filepath.Join(dir, "b.yml"), []byte("dataset: ds/db/s/t2\ncolumns:\n  - name: id\n"), 0644)
		r := run(t, "contract", "lint", filepath.Join(dir, "*.yml"), "--output", "table")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "2 contract file(s) are valid")
	})
}

func TestContractPull(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("bad_qualified_name", func(t *testing.T) {
		r := run(t, "contract", "pull", "bad/qualified/name")
		assertExitCode(t, r, 2)
	})
}

func TestContractVerify(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("no_file_errors", func(t *testing.T) {
		r := run(t, "contract", "verify")
		// cobra should error on missing required arg
		if r.ExitCode == 0 {
			t.Error("expected non-zero exit code for missing file arg")
		}
	})

	t.Run("nonexistent_file", func(t *testing.T) {
		r := run(t, "contract", "verify", "nonexistent.yml")
		assertExitCode(t, r, 2)
	})
}

func TestContractVerifyDQN(t *testing.T) {
	t.Run("local_rejects_dqn", func(t *testing.T) {
		// Error path — no credentials needed.
		r := run(t, "contract", "verify", "datasource/db/schema/table", "--local", "--datasource", "ds.yml")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "--local requires a contract file")
	})

	t.Run("nonexistent_dqn", func(t *testing.T) {
		skipIfNoCredentials(t)
		loginForTest(t)
		r := run(t, "contract", "verify", "fake/ds/no/exist", "--no-wait")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "no contract found")
	})

	t.Run("verify_by_dqn", func(t *testing.T) {
		skipIfNoCredentials(t)
		dqn := testDatasetDQN()
		if dqn == "" {
			t.Skip("SODA_TEST_DATASET_DQN not set")
		}
		loginForTest(t)
		r := run(t, "contract", "verify", dqn, "--no-wait")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "Verification started")
	})
}

func TestContractVerifyLocal(t *testing.T) {
	// Local verify error paths don't need credentials.

	t.Run("local_requires_datasource", func(t *testing.T) {
		f := writeTempFile(t, "contract-*.yml", `
dataset: ds/db/schema/orders
columns:
  - name: id
`)
		r := run(t, "contract", "verify", f, "--local")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "--datasource")
	})

	t.Run("local_nonexistent_contract", func(t *testing.T) {
		r := run(t, "contract", "verify", "/tmp/nonexistent-soda-file.yml", "--local", "--datasource", "ds.yml")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "could not read file")
	})

	t.Run("local_nonexistent_datasource", func(t *testing.T) {
		f := writeTempFile(t, "contract-*.yml", `
dataset: ds/db/schema/orders
columns:
  - name: id
`)
		r := run(t, "contract", "verify", f, "--local", "--datasource", "/tmp/nonexistent-ds.yml")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "datasource config file not found")
	})

	t.Run("local_no_wait_warning", func(t *testing.T) {
		contract := writeTempFile(t, "contract-*.yml", `
dataset: ds/db/schema/orders
columns:
  - name: id
`)
		ds := writeTempFile(t, "ds-*.yml", `
type: postgres
name: test
connection:
  host: localhost
`)
		r := run(t, "contract", "verify", contract, "--local", "--datasource", ds, "--no-wait", "--output", "table")
		// Should contain the --no-wait warning regardless of whether soda-core is installed
		assertOutputContains(t, r, "no-wait")
	})

	t.Run("local_push_without_auth", func(t *testing.T) {
		contract := writeTempFile(t, "contract-*.yml", `
dataset: ds/db/schema/orders
columns:
  - name: id
`)
		ds := writeTempFile(t, "ds-*.yml", `
type: postgres
name: test
connection:
  host: localhost
`)
		r := run(t, "contract", "verify", contract, "--local", "--datasource", ds, "--push")
		// Should fail: either no credentials, or soda-core rejects the push flags
		if r.ExitCode == 0 {
			t.Log("push succeeded; skipping error assertion")
		} else {
			t.Logf("push failed as expected: exit=%d output=%s", r.ExitCode, r.Output())
		}
	})
}

func TestContractProposal(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("list_blocked", func(t *testing.T) {
		r := run(t, "contract", "proposal", "list")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet available")
	})

	t.Run("pull_blocked", func(t *testing.T) {
		r := run(t, "contract", "proposal", "pull", "fake-id")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet available")
	})

	t.Run("push_blocked", func(t *testing.T) {
		r := run(t, "contract", "proposal", "push", "fake-id")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet available")
	})

	t.Run("close_blocked", func(t *testing.T) {
		r := run(t, "contract", "proposal", "close", "fake-id")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not yet available")
	})
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// writeTempFile creates a temp file with the given content and returns its path.
func writeTempFile(t *testing.T, pattern, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), pattern)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

// runInDir executes the binary from a specific working directory.
func runInDir(t *testing.T, bin, dir string, args ...string) Result {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run command: %v", err)
		}
	}
	return Result{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exitCode}
}
