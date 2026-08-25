//go:build integration

package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	t.Run("creates_files_and_dirs", func(t *testing.T) {
		dir := t.TempDir()
		r := runInDir(t, dir, "init", "--no-interactive")
		assertExitCode(t, r, 0)
		assertOutputContains(t, r, "Created  soda.yml")
		assertOutputContains(t, r, "Created  contracts/")
		assertOutputContains(t, r, "Created  configs/")

		// Verify soda.yml was actually created
		info, err := os.Stat(filepath.Join(dir, "soda.yml"))
		if err != nil {
			t.Fatalf("soda.yml not created: %v", err)
		}
		if info.Size() == 0 {
			t.Error("soda.yml is empty")
		}

		// Verify directories were created
		for _, d := range []string{"contracts", "configs"} {
			info, err := os.Stat(filepath.Join(dir, d))
			if err != nil {
				t.Fatalf("%s/ not created: %v", d, err)
			}
			if !info.IsDir() {
				t.Errorf("%s is not a directory", d)
			}
		}
	})

	t.Run("fails_if_soda_yml_exists", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "soda.yml"), []byte("existing"), 0644)

		r := runInDir(t, dir, "init", "--no-interactive")
		assertExitCode(t, r, 1)
		assertOutputContains(t, r, "already exists")
	})
}
