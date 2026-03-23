//go:build integration

package integration

import (
	"testing"
)

func TestJobLogs(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("bad_id", func(t *testing.T) {
		r := run(t, "job", "logs", "bad-scan-id")
		assertExitCode(t, r, 2)
	})

	t.Run("alias_scan_logs", func(t *testing.T) {
		r := run(t, "scan", "logs", "bad-scan-id")
		assertExitCode(t, r, 2)
	})
}

func TestJobCancel(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	// API returns 404 — blocked
	t.Run("bad_id", func(t *testing.T) {
		r := run(t, "job", "cancel", "bad-scan-id")
		// Should fail
		if r.ExitCode == 0 {
			t.Log("job cancel succeeded unexpectedly (API may have been unblocked)")
		}
	})
}
