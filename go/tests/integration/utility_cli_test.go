//go:build cli

package integration

import (
	"testing"
)

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
		dir := t.TempDir()
		bin := ensureBinary(t)
		r := runInDir(t, bin, dir, "init")
		assertExitCode(t, r, 0)
	})
}
