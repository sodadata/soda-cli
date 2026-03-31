//go:build integration

package integration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestSecretList(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "secret", "list")
	if r.ExitCode != 0 {
		t.Errorf("secret list failed: exit=%d output=%s", r.ExitCode, r.Output())
	}
}

func TestSecretListJSON(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "secret", "list", "--output", "json")
	if r.ExitCode != 0 {
		t.Errorf("secret list --output json failed: exit=%d output=%s", r.ExitCode, r.Output())
	}
}

func TestSecretCRUD(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	// Use a unique name per test run
	secretName := fmt.Sprintf("CLI_TEST_%d", time.Now().UnixMilli())

	// Create
	r := run(t, "secret", "create", "--name", secretName, "--value", "test123")
	t.Logf("secret create: exit=%d output=%s", r.ExitCode, r.Output())
	assertExitCode(t, r, 0)
	assertOutputContains(t, r, secretName)

	// List and verify it exists
	r = run(t, "secret", "list", "--output", "json")
	assertExitCode(t, r, 0)
	assertOutputContains(t, r, secretName)

	// Get the secret ID from JSON list
	var secrets []map[string]string
	if err := json.Unmarshal([]byte(r.Stdout), &secrets); err != nil {
		t.Fatalf("could not parse JSON: %v\nstdout: %s", err, r.Stdout)
	}
	var secretID string
	for _, s := range secrets {
		if s["name"] == secretName {
			secretID = s["id"]
			break
		}
	}
	if secretID == "" {
		t.Fatalf("could not find %s in secret list", secretName)
	}

	// Update
	r = run(t, "secret", "update", secretID, "--value", "updated456")
	t.Logf("secret update: exit=%d output=%s", r.ExitCode, r.Output())
	assertExitCode(t, r, 0)

	// Delete
	r = run(t, "secret", "delete", secretID)
	t.Logf("secret delete: exit=%d output=%s", r.ExitCode, r.Output())
	assertExitCode(t, r, 0)
}

func TestSecretCreateMissingFlags(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "secret", "create")
	assertExitCode(t, r, 2)
	assertOutputContains(t, r, "--name is required")
}

func TestSecretGetBadID(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	r := run(t, "secret", "get", "nonexistent-id")
	if r.ExitCode == 0 {
		t.Error("expected non-zero exit code for nonexistent secret")
	}
}
