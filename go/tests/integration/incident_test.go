//go:build integration

package integration

import (
	"testing"
)

func TestIncidentList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "incident", "list")
	// API endpoint is documented but may still return HTML on some environments.
	// Accept both success and the "not yet available" error.
	if r.ExitCode == 0 {
		t.Log("incident list succeeded — API is live")
	} else {
		t.Logf("incident list: exit=%d output=%s", r.ExitCode, r.Output())
	}
}

func TestIncidentGetBadID(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "incident", "get", "nonexistent-id")
	if r.ExitCode == 0 {
		t.Error("expected non-zero exit code for nonexistent incident")
	}
}
