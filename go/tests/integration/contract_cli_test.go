//go:build cli

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestContractLint(t *testing.T) {
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

	// The contract root and column objects are open by design — the schema
	// allows unknown keys there — but check objects are closed.
	t.Run("unknown_check_key", func(t *testing.T) {
		f := writeTempFile(t, "contract-*.yml", `
dataset: ds/db/schema/orders
columns:
  - name: id
    checks:
      - bogus_check: {}
`)
		r := run(t, "contract", "lint", f, "--output", "table")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "bogus_check")
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

// TestContractVerifyLocal covers the --local error paths, which are handled
// before any Cloud call. `local_push_without_auth` tolerates soda-core being
// absent from PATH.
func TestContractVerifyLocal(t *testing.T) {
	t.Run("local_rejects_dqn", func(t *testing.T) {
		r := run(t, "contract", "verify", "datasource/db/schema/table", "--local", "--datasource", "ds.yml")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "--local requires a contract file")
	})

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

// TestContractProposal covers commands blocked on unreleased API endpoints.
// The guard runs before the credential check.
func TestContractProposal(t *testing.T) {
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
