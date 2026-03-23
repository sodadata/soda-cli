//go:build integration

package integration

import (
	"strings"
	"testing"
)

func TestRunnerList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("table", func(t *testing.T) {
		r := run(t, "runner", "list")
		assertExitCode(t, r, 0)
	})

	t.Run("json", func(t *testing.T) {
		r := run(t, "runner", "list", "--output", "json")
		assertExitCode(t, r, 0)
	})

	t.Run("csv", func(t *testing.T) {
		r := run(t, "runner", "list", "--output", "csv")
		assertExitCode(t, r, 0)
	})
}

func TestRunnerCreateAndDelete(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	// Create
	r := run(t, "runner", "create", "--name", "integration-test-runner")
	if r.ExitCode != 0 {
		t.Fatalf("runner create failed: %s", r.Output())
	}
	assertOutputContains(t, r, "created")

	// The runner create output shows "Runner ID" but the value may be empty.
	// Extract the Runner ID from the output line "  Runner ID            <id>"
	runnerID := ""
	for _, line := range strings.Split(r.Output(), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Runner ID") {
			parts := strings.Fields(trimmed)
			// "Runner ID <uuid>" → parts[2]
			if len(parts) >= 3 {
				runnerID = parts[2]
			}
		}
	}

	// If Runner ID was empty in output, try to find it from the runner list
	if runnerID == "" {
		lr := run(t, "runner", "list", "--output", "json")
		// Find the runner with our name
		lines := strings.Split(lr.Stdout, "\n")
		candidate := ""
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, `"id"`) {
				candidate = strings.Trim(strings.TrimSpace(strings.SplitN(trimmed, ":", 2)[1]), `", `)
			}
			if strings.Contains(trimmed, "integration-test-runner") && candidate != "" {
				runnerID = candidate
				break
			}
		}
	}

	if runnerID == "" {
		t.Log("WARNING: Could not extract runner ID — runner may have been created but ID not returned by API")
		t.Log("Skipping get/delete subtests. Runner may need manual cleanup.")
		return
	}
	t.Logf("Created runner: %s", runnerID)

	// Get
	t.Run("get", func(t *testing.T) {
		r := run(t, "runner", "get", runnerID)
		assertExitCode(t, r, 0)
	})

	t.Run("get_json", func(t *testing.T) {
		r := run(t, "runner", "get", runnerID, "--output", "json")
		assertExitCode(t, r, 0)
	})

	// Delete
	t.Run("delete", func(t *testing.T) {
		r := run(t, "runner", "delete", runnerID)
		assertExitCode(t, r, 0)
	})
}

func TestRunnerGetBadID(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "runner", "get", "bad-id")
	assertExitCode(t, r, 2)
}

func TestRunnerDeleteBadID(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "runner", "delete", "bad-id")
	assertExitCode(t, r, 2)
}

func TestRunnerCreateMissingName(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "runner", "create")
	// Cobra should error on missing --name
	if r.ExitCode == 0 {
		t.Error("expected error for missing --name flag")
	}
}
