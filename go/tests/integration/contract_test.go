//go:build integration

package integration

import (
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
	skipIfNoCredentials(t)

	t.Run("lint_no_file", func(t *testing.T) {
		r := run(t, "contract", "lint")
		// Depends on whether there are yml files in cwd
		t.Logf("lint no-file: exit=%d", r.ExitCode)
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
