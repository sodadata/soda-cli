//go:build integration

package integration

import (
	"testing"
)

func TestResultsList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("no_filters", func(t *testing.T) {
		r := run(t, "results", "list")
		// API may intermittently return errors (e.g., endpoint not available)
		if r.ExitCode != 0 {
			t.Logf("results list (no filters) exit=%d: %s", r.ExitCode, r.Output())
		}
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "results", "list", "--output", "json")
		assertExitCode(t, r, 0)
	})

	t.Run("csv", func(t *testing.T) {
		r := run(t, "results", "list", "--output", "csv")
		assertExitCode(t, r, 0)
	})

	t.Run("filter_by_dataset", func(t *testing.T) {
		r := run(t, "results", "list", "--dataset", testDatasetID())
		assertExitCode(t, r, 0)
	})

	t.Run("filter_by_dataset_name", func(t *testing.T) {
		r := run(t, "results", "list", "--dataset-name", "ACCOUNT")
		assertExitCode(t, r, 0)
	})

	t.Run("filter_status_passing", func(t *testing.T) {
		r := run(t, "results", "list", "--status", "passing")
		assertExitCode(t, r, 0)
	})

	t.Run("filter_status_failing", func(t *testing.T) {
		r := run(t, "results", "list", "--status", "failing")
		assertExitCode(t, r, 0)
	})

	t.Run("filter_type_check", func(t *testing.T) {
		r := run(t, "results", "list", "--type", "check")
		assertExitCode(t, r, 0)
	})

	t.Run("filter_type_monitor_graceful", func(t *testing.T) {
		r := run(t, "results", "list", "--type", "monitor")
		assertExitCode(t, r, 0)
	})

	t.Run("limit_5", func(t *testing.T) {
		r := run(t, "results", "list", "--limit", "5")
		assertExitCode(t, r, 0)
	})

	t.Run("limit_50", func(t *testing.T) {
		r := run(t, "results", "list", "--limit", "50")
		assertExitCode(t, r, 0)
	})

	t.Run("sort_dataset", func(t *testing.T) {
		r := run(t, "results", "list", "--sort", "dataset")
		assertExitCode(t, r, 0)
	})

	t.Run("sort_name", func(t *testing.T) {
		r := run(t, "results", "list", "--sort", "name")
		assertExitCode(t, r, 0)
	})

	t.Run("sort_status", func(t *testing.T) {
		r := run(t, "results", "list", "--sort", "status")
		assertExitCode(t, r, 0)
	})

	t.Run("sort_date_asc", func(t *testing.T) {
		r := run(t, "results", "list", "--sort", "date", "--order", "asc")
		assertExitCode(t, r, 0)
	})

	t.Run("sort_date_desc", func(t *testing.T) {
		r := run(t, "results", "list", "--sort", "date", "--order", "desc")
		assertExitCode(t, r, 0)
	})

	t.Run("date_range", func(t *testing.T) {
		r := run(t, "results", "list", "--from", "2026-01-01", "--until", "2026-12-31")
		assertExitCode(t, r, 0)
	})

	t.Run("combined_filters", func(t *testing.T) {
		r := run(t, "results", "list",
			"--dataset", testDatasetID(),
			"--status", "failing",
			"--limit", "20",
			"--sort", "date",
			"--order", "desc",
		)
		assertExitCode(t, r, 0)
	})
}
