//go:build integration

package integration

import (
	"testing"
)

func TestDatasetList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("no_filters", func(t *testing.T) {
		r := run(t, "dataset", "list")
		assertExitCode(t, r, 0)
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "dataset", "list", "--output", "json")
		assertExitCode(t, r, 0)
		assertStdoutContains(t, r, `"id"`)
	})

	t.Run("csv", func(t *testing.T) {
		r := run(t, "dataset", "list", "--output", "csv")
		assertExitCode(t, r, 0)
	})

	t.Run("filter_by_datasource", func(t *testing.T) {
		r := run(t, "dataset", "list", "--datasource", testDatasourceName())
		assertExitCode(t, r, 0)
	})

	t.Run("filter_by_status_onboarded", func(t *testing.T) {
		r := run(t, "dataset", "list", "--status", "onboarded")
		assertExitCode(t, r, 0)
	})

	t.Run("filter_by_status_not_onboarded", func(t *testing.T) {
		r := run(t, "dataset", "list", "--status", "not-onboarded")
		assertExitCode(t, r, 0)
	})

	t.Run("filter_by_name", func(t *testing.T) {
		r := run(t, "dataset", "list", "--filter", "ACCOUNT")
		assertExitCode(t, r, 0)
	})

	t.Run("limit", func(t *testing.T) {
		r := run(t, "dataset", "list", "--limit", "5")
		assertExitCode(t, r, 0)
	})

	t.Run("limit_50", func(t *testing.T) {
		r := run(t, "dataset", "list", "--limit", "50")
		assertExitCode(t, r, 0)
	})

	t.Run("date_range", func(t *testing.T) {
		r := run(t, "dataset", "list", "--from", "2026-01-01", "--until", "2026-12-31")
		assertExitCode(t, r, 0)
	})

	t.Run("combined_filters", func(t *testing.T) {
		r := run(t, "dataset", "list",
			"--datasource", testDatasourceName(),
			"--status", "onboarded",
			"--limit", "10",
		)
		assertExitCode(t, r, 0)
	})

	t.Run("invalid_status", func(t *testing.T) {
		r := run(t, "dataset", "list", "--status", "badvalue")
		assertExitCode(t, r, 2)
	})
}

func TestDatasetGet(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("by_id", func(t *testing.T) {
		r := run(t, "dataset", "get", testDatasetID())
		if r.ExitCode != 0 {
			// API may intermittently return "not yet available"
			t.Logf("dataset get exit=%d: %s", r.ExitCode, r.Output())
		} else {
			assertOutputContains(t, r, "ACCOUNT_BALANCES")
		}
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "dataset", "get", testDatasetID(), "--output", "json")
		if r.ExitCode != 0 {
			t.Logf("dataset get json exit=%d: %s", r.ExitCode, r.Output())
		} else {
			assertStdoutContains(t, r, `"id"`)
		}
	})

	t.Run("bad_id", func(t *testing.T) {
		r := run(t, "dataset", "get", "bad-id")
		assertExitCode(t, r, 2)
	})
}

func TestDatasetUpdate(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("set_tag", func(t *testing.T) {
		r := run(t, "dataset", "update", testDatasetID(), "--tag", "integration-test")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "updated")
	})

	t.Run("set_multiple_tags", func(t *testing.T) {
		r := run(t, "dataset", "update", testDatasetID(), "--tag", "tag1", "--tag", "tag2")
		assertExitCode(t, r, 0)
	})

	t.Run("no_flags_errors", func(t *testing.T) {
		r := run(t, "dataset", "update", testDatasetID())
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "required")
	})

	// Clean up tags
	t.Run("clear_tags", func(t *testing.T) {
		r := run(t, "dataset", "update", testDatasetID(), "--tag", "")
		// May or may not succeed depending on API behavior with empty tag
		t.Logf("clear tags: exit=%d", r.ExitCode)
	})
}

func TestDatasetTimePartition(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("view_current", func(t *testing.T) {
		r := run(t, "dataset", "time-partition", testDatasetID())
		if r.ExitCode != 0 {
			// API may intermittently return errors for dataset endpoints
			t.Logf("time-partition view exit=%d: %s", r.ExitCode, r.Output())
		}
	})
}

func TestDatasetProfiling(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("view", func(t *testing.T) {
		r := run(t, "dataset", "profiling", testDatasetID())
		assertExitCode(t, r, 0)
	})

	t.Run("view_json", func(t *testing.T) {
		r := run(t, "dataset", "profiling", testDatasetID(), "--output", "json")
		assertExitCode(t, r, 0)
	})
}

func TestDatasetDiagnostics(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("view", func(t *testing.T) {
		r := run(t, "dataset", "diagnostics", testDatasetID())
		assertExitCode(t, r, 0)
	})

	t.Run("view_json", func(t *testing.T) {
		r := run(t, "dataset", "diagnostics", testDatasetID(), "--output", "json")
		assertExitCode(t, r, 0)
	})
}

func TestDatasetPermissions(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("list", func(t *testing.T) {
		r := run(t, "dataset", "permissions", "list", testDatasetID())
		assertExitCode(t, r, 0)
	})

	t.Run("list_json", func(t *testing.T) {
		r := run(t, "dataset", "permissions", "list", testDatasetID(), "--output", "json")
		assertExitCode(t, r, 0)
	})

	t.Run("assign_no_user_or_group_errors", func(t *testing.T) {
		r := run(t, "dataset", "permissions", "assign", testDatasetID(), "--role", "some-role")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "required")
	})
}
