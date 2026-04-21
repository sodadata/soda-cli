//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestMonitorList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("requires_dataset", func(t *testing.T) {
		r := run(t, "monitor", "list")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "required")
	})

	t.Run("with_dataset", func(t *testing.T) {
		r := run(t, "monitor", "list", "--dataset", testDatasetID())
		assertExitCode(t, r, 0)
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "monitor", "list", "--dataset", testDatasetID(), "--output", "json")
		assertExitCode(t, r, 0)
	})

	t.Run("type_column", func(t *testing.T) {
		r := run(t, "monitor", "list", "--dataset", testDatasetID(), "--type", "column")
		assertExitCode(t, r, 0)
	})

	t.Run("type_custom", func(t *testing.T) {
		r := run(t, "monitor", "list", "--dataset", testDatasetID(), "--type", "custom")
		assertExitCode(t, r, 0)
	})

	t.Run("type_dataset", func(t *testing.T) {
		r := run(t, "monitor", "list", "--dataset", testDatasetID(), "--type", "dataset")
		assertExitCode(t, r, 0)
	})
}

func TestMonitorConfig(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("view", func(t *testing.T) {
		r := run(t, "monitor", "config", testDatasetID())
		assertExitCode(t, r, 0)
	})

	t.Run("view_json", func(t *testing.T) {
		r := run(t, "monitor", "config", testDatasetID(), "--output", "json")
		assertExitCode(t, r, 0)
	})
}

func TestMonitorAddColumnAndDelete(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	// Add a column monitor (CLOSING_BALANCE is a known column on ACCOUNT_BALANCES)
	r := run(t, "monitor", "add",
		"--dataset", testDatasetID(),
		"--type", "column",
		"--column", "CLOSING_BALANCE",
		"--metric", "avg",
	)
	if r.ExitCode != 0 {
		t.Fatalf("monitor add column failed: %s", r.Output())
	}
	assertOutputContains(t, r, "created")

	// The API may return an empty ID — find monitor by listing
	monitorID := findMonitorID(t, testDatasetID(), "column", "avg")
	if monitorID == "" {
		t.Skip("could not find created column monitor ID from list — API may not return IDs")
	}
	t.Logf("Found column monitor: %s", monitorID)

	// Update: disable
	t.Run("update_disable", func(t *testing.T) {
		r := run(t, "monitor", "update", monitorID, "--dataset", testDatasetID(), "--disable")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "updated")
	})

	// Update: re-enable
	t.Run("update_enable", func(t *testing.T) {
		r := run(t, "monitor", "update", monitorID, "--dataset", testDatasetID(), "--enable")
		assertExitCode(t, r, 0)
	})

	// Delete
	t.Run("delete", func(t *testing.T) {
		r := run(t, "monitor", "delete", monitorID, "--dataset", testDatasetID())
		assertExitCode(t, r, 0)
	})
}

func TestMonitorAddCustomAndDelete(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	// Add custom monitor
	r := run(t, "monitor", "add",
		"--dataset", testDatasetID(),
		"--type", "custom",
		"--name", "integration-test-custom",
		"--sql", "SELECT count(*) as cnt FROM ACCOUNT_BALANCES",
		"--result-metric", "cnt",
	)
	if r.ExitCode != 0 {
		t.Fatalf("monitor add custom failed: %s", r.Output())
	}
	assertOutputContains(t, r, "created")

	// The API may return empty ID — find monitor by listing
	monitorID := findMonitorID(t, testDatasetID(), "custom", "integration-test-custom")
	if monitorID == "" {
		t.Skip("could not find created custom monitor ID from list — API may not return IDs")
	}
	t.Logf("Found custom monitor: %s", monitorID)

	// Update name and SQL
	t.Run("update_custom", func(t *testing.T) {
		r := run(t, "monitor", "update", monitorID,
			"--dataset", testDatasetID(),
			"--name", "integration-test-renamed",
			"--sql", "SELECT count(*) as cnt FROM ACCOUNT_BALANCES WHERE 1=1",
		)
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "updated")
	})

	// Delete
	t.Run("delete", func(t *testing.T) {
		r := run(t, "monitor", "delete", monitorID, "--dataset", testDatasetID())
		assertExitCode(t, r, 0)
	})
}

func TestMonitorAddDataset_Blocked(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "monitor", "add",
		"--dataset", testDatasetID(),
		"--type", "dataset",
		"--metric", "row-count",
	)
	assertExitCode(t, r, 2)
	assertOutputContains(t, r, "not yet available")
}

func TestMonitorAddErrors(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("column_no_column_flag", func(t *testing.T) {
		r := run(t, "monitor", "add", "--dataset", testDatasetID(), "--type", "column")
		assertExitCode(t, r, 2)
	})

	t.Run("custom_no_name", func(t *testing.T) {
		r := run(t, "monitor", "add", "--dataset", testDatasetID(), "--type", "custom")
		assertExitCode(t, r, 2)
	})

	t.Run("no_dataset", func(t *testing.T) {
		r := run(t, "monitor", "add", "--type", "column", "--column", "x", "--metric", "count")
		assertExitCode(t, r, 2)
	})

	t.Run("exclude_values_bad_format", func(t *testing.T) {
		r := run(t, "monitor", "add",
			"--dataset", testDatasetID(),
			"--type", "column",
			"--column", "CLOSING_BALANCE",
			"--metric", "avg",
			"--group-by", "CLOSING_BALANCE",
			"--exclude-values", "no-equals-sign",
		)
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "column=val1,val2")
	})

	t.Run("exclude_values_unknown_column", func(t *testing.T) {
		r := run(t, "monitor", "add",
			"--dataset", testDatasetID(),
			"--type", "column",
			"--column", "CLOSING_BALANCE",
			"--metric", "avg",
			"--group-by", "CLOSING_BALANCE",
			"--exclude-values", "NOT_IN_GROUP_BY=a,b",
		)
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not in --group-by")
	})
}

func TestMonitorAddColumnWithExcludeValues(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	// Create a group-by monitor with excluded values. Use CLOSING_BALANCE for
	// both the metric column and the group-by column — the API accepts this
	// and it avoids depending on a second known column name.
	r := run(t, "monitor", "add",
		"--dataset", testDatasetID(),
		"--type", "column",
		"--column", "CLOSING_BALANCE",
		"--metric", "avg",
		"--group-by", "CLOSING_BALANCE",
		"--exclude-values", "CLOSING_BALANCE=0,100",
	)
	if r.ExitCode != 0 {
		t.Fatalf("monitor add with exclude-values failed: %s", r.Output())
	}
	assertOutputContains(t, r, "created")

	// Verify the excluded values round-trip by listing with JSON output.
	list := run(t, "monitor", "list", "--dataset", testDatasetID(), "--type", "column", "--output", "json")
	assertExitCode(t, list, 0)
	if !strings.Contains(list.Stdout, "excl: 0,100") {
		t.Errorf("expected excluded values to round-trip in list output, got:\n%s", list.Stdout)
	}

	// Clean up: find and delete the monitor we just created.
	monitorID := findMonitorID(t, testDatasetID(), "column", "avg")
	if monitorID == "" {
		t.Fatal("could not find created group-by monitor ID from list")
	}
	r = run(t, "monitor", "delete", monitorID, "--dataset", testDatasetID())
	assertExitCode(t, r, 0)
}

func TestMonitorDeleteBadID(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "monitor", "delete", "bad-id", "--dataset", testDatasetID())
	assertExitCode(t, r, 2)
}

// findMonitorID lists monitors on a dataset and finds one matching the given type and metric/name.
func findMonitorID(t *testing.T, datasetID, monType, metricOrName string) string {
	t.Helper()
	r := run(t, "monitor", "list", "--dataset", datasetID, "--type", monType, "--output", "json")
	if r.ExitCode != 0 {
		return ""
	}
	// Parse JSON output to find a monitor with matching metric/name
	// Look for lines with "id" field near lines containing the metric or name
	lines := strings.Split(r.Stdout, "\n")
	lastID := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, `"id"`) {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				lastID = strings.Trim(strings.TrimSpace(parts[1]), `",`)
			}
		}
		if strings.Contains(trimmed, metricOrName) && lastID != "" && lastID != "-" {
			return lastID
		}
	}
	return ""
}
