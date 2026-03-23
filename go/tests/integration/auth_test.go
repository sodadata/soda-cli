//go:build integration

package integration

import (
	"testing"
)

func TestAuth(t *testing.T) {
	skipIfNoCredentials(t)

	t.Run("login_interactive_no_flags_nointeractive_errors", func(t *testing.T) {
		r := run(t, "auth", "login", "--no-interactive")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "required")
	})

	t.Run("login_noninteractive_full_flags", func(t *testing.T) {
		r := run(t, "auth", "login",
			"--host", testHost(),
			"--api-key-id", testKeyID(),
			"--api-key-secret", testKeySecret(),
		)
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "Authenticated")
	})

	t.Run("login_with_profile", func(t *testing.T) {
		r := run(t, "auth", "login",
			"--host", testHost(),
			"--api-key-id", testKeyID(),
			"--api-key-secret", testKeySecret(),
			"--profile", "test-profile",
		)
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "test-profile")
	})

	t.Run("status_default_profile", func(t *testing.T) {
		loginForTest(t)
		r := run(t, "auth", "status")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "connected")
	})

	t.Run("status_json", func(t *testing.T) {
		loginForTest(t)
		r := run(t, "auth", "status", "--output", "json")
		assertExitCode(t, r, 0)
	})

	t.Run("status_with_profile", func(t *testing.T) {
		r := run(t, "auth", "status", "--profile", "test-profile")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "connected")
	})

	t.Run("switch_profile", func(t *testing.T) {
		r := run(t, "auth", "switch", "default")
		assertExitCode(t, r, 0)
	})

	t.Run("switch_nonexistent_profile", func(t *testing.T) {
		r := run(t, "auth", "switch", "nonexistent-profile-xyz")
		// switch currently always succeeds — just sets the name
		assertExitCode(t, r, 0)
	})

	t.Run("logout_named_profile", func(t *testing.T) {
		// logout the test-profile we created
		r := run(t, "auth", "logout", "--profile", "test-profile")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "Logged out")
	})

	t.Run("logout_nonexistent_profile", func(t *testing.T) {
		r := run(t, "auth", "logout", "--profile", "does-not-exist")
		assertExitCode(t, r, 2)
		assertOutputContains(t, r, "not found")
	})
}
