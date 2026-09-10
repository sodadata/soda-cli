//go:build cloud

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

	// Needs credentials: `contract verify` resolves the profile before it reads
	// the file, so without one this exits 3 rather than 2.
	t.Run("nonexistent_file", func(t *testing.T) {
		r := run(t, "contract", "verify", "nonexistent.yml")
		assertExitCode(t, r, 2)
	})
}

func TestContractVerifyDQN(t *testing.T) {
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
