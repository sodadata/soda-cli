//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestDatasourceList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("table", func(t *testing.T) {
		r := run(t, "datasource", "list")
		assertExitCode(t, r, 0)
		assertStdoutNotEmpty(t, r)
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "datasource", "list", "--output", "json")
		assertExitCode(t, r, 0)
		assertStdoutContains(t, r, `"id"`)
	})

	t.Run("csv", func(t *testing.T) {
		r := run(t, "datasource", "list", "--output", "csv")
		assertExitCode(t, r, 0)
	})

	t.Run("quiet", func(t *testing.T) {
		r := run(t, "datasource", "list", "--quiet")
		assertExitCode(t, r, 0)
	})

	t.Run("alias_ds", func(t *testing.T) {
		r := run(t, "ds", "list")
		assertExitCode(t, r, 0)
	})
}

func TestDatasourceGet(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("by_id", func(t *testing.T) {
		r := run(t, "datasource", "get", testDatasourceID())
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, testDatasourceName())
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "datasource", "get", testDatasourceID(), "--output", "json")
		assertExitCode(t, r, 0)
		assertStdoutContains(t, r, `"id"`)
	})

	t.Run("bad_id", func(t *testing.T) {
		r := run(t, "datasource", "get", "bad-id-does-not-exist")
		assertExitCode(t, r, 2)
	})

	t.Run("alias_ds_get", func(t *testing.T) {
		r := run(t, "ds", "get", testDatasourceID())
		assertExitCode(t, r, 0)
	})
}

func TestDatasourceCreateAndDelete(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	dsConfig := testDSConfig()
	if dsConfig == "" {
		t.Skip("SODA_TEST_DS_CONFIG not set")
	}

	// Create — may fail if no suitable runner is available (hosted runners can't create datasources)
	r := run(t, "datasource", "create", dsConfig)
	if r.ExitCode != 0 {
		if strings.Contains(r.Output(), "runnerId") || strings.Contains(r.Output(), "runner") {
			t.Skip("datasource create requires a self-hosted runner — skipping")
		}
		t.Fatalf("datasource create failed: %s", r.Output())
	}
	assertOutputContains(t, r, "created")

	// Extract ID from output
	dsID := ""
	for _, line := range strings.Split(r.Output(), "\n") {
		if strings.Contains(line, "ID:") && !strings.Contains(line, "Discovery") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dsID = parts[len(parts)-1]
			}
		}
	}
	if dsID == "" {
		t.Fatal("could not extract datasource ID from create output")
	}
	t.Logf("Created datasource: %s", dsID)

	// Get the newly created datasource
	t.Run("get_created", func(t *testing.T) {
		r := run(t, "datasource", "get", dsID)
		assertExitCode(t, r, 0)
	})

	// Delete
	t.Run("delete", func(t *testing.T) {
		r := run(t, "datasource", "delete", dsID)
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "deletion")
	})
}

func TestDatasourceDelete_BadID(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "datasource", "delete", "bad-id-does-not-exist")
	assertExitCode(t, r, 2)
}

func TestDatasourceUpdate(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("update_label", func(t *testing.T) {
		r := run(t, "datasource", "update", testDatasourceID(), "--label", "test-label")
		t.Logf("datasource update exit=%d output=%s", r.ExitCode, r.Output())
	})
}

func TestDatasourceTestConnection(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	dsConfig := testDSConfig()
	if dsConfig == "" {
		t.Skip("SODA_TEST_DS_CONFIG not set")
	}

	t.Run("test_connection", func(t *testing.T) {
		r := run(t, "datasource", "test-connection", dsConfig)
		t.Logf("test-connection exit=%d output=%s", r.ExitCode, r.Output())
	})
}

func TestDatasourceDiagnostics(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("get", func(t *testing.T) {
		r := run(t, "datasource", "diagnostics", testDatasourceID())
		t.Logf("datasource diagnostics exit=%d output=%s", r.ExitCode, r.Output())
	})
}

func TestDatasourceCreateBadFile(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "datasource", "create", "nonexistent.yml")
	assertExitCode(t, r, 2)
}
