//go:build integration

package integration

import (
	"testing"
)

func TestGlobalFlags(t *testing.T) {
	skipIfNoCredentials(t)
	loginForTest(t)

	t.Run("output_json", func(t *testing.T) {
		r := run(t, "datasource", "list", "-o", "json")
		assertExitCode(t, r, 0)
		assertStdoutContains(t, r, "[")
	})

	t.Run("output_csv", func(t *testing.T) {
		r := run(t, "datasource", "list", "-o", "csv")
		assertExitCode(t, r, 0)
	})

	t.Run("output_table", func(t *testing.T) {
		r := run(t, "datasource", "list", "-o", "table")
		assertExitCode(t, r, 0)
	})

	t.Run("quiet", func(t *testing.T) {
		r := run(t, "datasource", "list", "--quiet")
		assertExitCode(t, r, 0)
	})

	t.Run("quiet_short", func(t *testing.T) {
		r := run(t, "datasource", "list", "-q")
		assertExitCode(t, r, 0)
	})

	t.Run("verbose", func(t *testing.T) {
		r := run(t, "datasource", "list", "--verbose")
		assertExitCode(t, r, 0)
	})

	t.Run("verbose_short", func(t *testing.T) {
		r := run(t, "datasource", "list", "-v")
		assertExitCode(t, r, 0)
	})

	t.Run("no_color", func(t *testing.T) {
		r := run(t, "datasource", "list", "--no-color")
		assertExitCode(t, r, 0)
	})

	t.Run("no_interactive", func(t *testing.T) {
		r := run(t, "datasource", "list", "--no-interactive")
		assertExitCode(t, r, 0)
	})

	t.Run("profile_default", func(t *testing.T) {
		r := run(t, "datasource", "list", "--profile", "default")
		assertExitCode(t, r, 0)
	})

	t.Run("profile_nonexistent", func(t *testing.T) {
		r := run(t, "datasource", "list", "--profile", "nonexistent")
		// Should fail with auth error
		assertExitCode(t, r, 3)
	})

	t.Run("combined_flags", func(t *testing.T) {
		r := run(t, "datasource", "list", "--no-color", "--no-interactive", "--output", "json")
		assertExitCode(t, r, 0)
	})

	t.Run("quiet_plus_json", func(t *testing.T) {
		r := run(t, "datasource", "list", "-o", "json", "--quiet")
		assertExitCode(t, r, 0)
	})
}

func TestUtility(t *testing.T) {
	t.Run("version", func(t *testing.T) {
		r := run(t, "version")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "sodacli version")
	})

	t.Run("version_flag", func(t *testing.T) {
		r := run(t, "--version")
		assertExitCode(t, r, 0)
	})

	t.Run("help", func(t *testing.T) {
		r := run(t, "--help")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "sodacli")
	})

	t.Run("completion_bash", func(t *testing.T) {
		r := run(t, "completion", "bash")
		assertExitCode(t, r, 0)
	})

	t.Run("completion_zsh", func(t *testing.T) {
		r := run(t, "completion", "zsh")
		assertExitCode(t, r, 0)
	})

	t.Run("completion_fish", func(t *testing.T) {
		r := run(t, "completion", "fish")
		assertExitCode(t, r, 0)
	})
}

func TestInit(t *testing.T) {
	// init scaffolds files — run in a temp dir to avoid polluting the repo
	// For now just check it doesn't crash
	t.Run("basic", func(t *testing.T) {
		r := run(t, "init")
		assertExitCode(t, r, 0)
	})
}
