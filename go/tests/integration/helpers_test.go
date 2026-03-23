//go:build integration

package integration

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// ── Environment ──────────────────────────────────────────────────────────────

var (
	binPath    string
	buildOnce  sync.Once
	buildErr   error
	envLoaded  bool
	envOnce    sync.Once
)

// env returns an environment variable, falling back to .env file at repo root.
func env(key string) string {
	envOnce.Do(func() {
		// Try loading from ../../.env (repo root relative to go/tests/integration/)
		envFile := filepath.Join("..", "..", "..", ".env")
		if f, err := os.Open(envFile); err == nil {
			defer f.Close()
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				k, v, ok := strings.Cut(line, "=")
				if ok {
					// Don't overwrite already-set env vars
					if os.Getenv(k) == "" {
						os.Setenv(k, v)
					}
				}
			}
			envLoaded = true
		}
	})
	return os.Getenv(key)
}

func testHost() string      { return env("SODA_TEST_HOST") }
func testKeyID() string     { return env("SODA_TEST_API_KEY_ID") }
func testKeySecret() string { return env("SODA_TEST_API_KEY_SECRET") }
func testDatasourceID() string   { return env("SODA_TEST_DATASOURCE_ID") }
func testDatasourceName() string { return env("SODA_TEST_DATASOURCE_NAME") }
func testDatasetID() string      { return env("SODA_TEST_DATASET_ID") }
func testDSConfig() string {
	v := env("SODA_TEST_DS_CONFIG")
	if v == "" {
		return ""
	}
	// Resolve relative paths against the repo root (.env location)
	if !filepath.IsAbs(v) {
		repoRoot, err := filepath.Abs(filepath.Join("..", "..", ".."))
		if err == nil {
			return filepath.Join(repoRoot, v)
		}
	}
	return v
}

// ── Binary ───────────────────────────────────────────────────────────────────

// ensureBinary builds the sodacli binary once and returns its path.
func ensureBinary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		// Resolve go/ directory (two levels up from tests/integration/)
		goDir, err := filepath.Abs(filepath.Join("..", ".."))
		if err != nil {
			buildErr = fmt.Errorf("could not resolve go dir: %v", err)
			return
		}
		binPath = filepath.Join(goDir, "sodacli-test")
		cmd := exec.Command("go", "build", "-o", binPath, ".")
		cmd.Dir = goDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = fmt.Errorf("build failed: %v\n%s", err, out)
		}
	})
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	return binPath
}

// ── Command runner ───────────────────────────────────────────────────────────

// Result holds the output of a CLI command.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// run executes sodacli with the given args and returns the result.
// It never fails the test on non-zero exit — callers check ExitCode.
func run(t *testing.T, args ...string) Result {
	t.Helper()
	bin := ensureBinary(t)

	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run command: %v", err)
		}
	}

	return Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

// Output returns combined stdout+stderr for simple checks.
func (r Result) Output() string {
	return r.Stdout + r.Stderr
}

// ── Assertions ───────────────────────────────────────────────────────────────

func assertExitCode(t *testing.T, r Result, want int) {
	t.Helper()
	if r.ExitCode != want {
		t.Errorf("exit code = %d, want %d\nstdout: %s\nstderr: %s",
			r.ExitCode, want, r.Stdout, r.Stderr)
	}
}

func assertOutputContains(t *testing.T, r Result, substr string) {
	t.Helper()
	if !strings.Contains(r.Output(), substr) {
		t.Errorf("output does not contain %q\nstdout: %s\nstderr: %s",
			substr, r.Stdout, r.Stderr)
	}
}

func assertOutputNotContains(t *testing.T, r Result, substr string) {
	t.Helper()
	if strings.Contains(r.Output(), substr) {
		t.Errorf("output should not contain %q\nstdout: %s\nstderr: %s",
			substr, r.Stdout, r.Stderr)
	}
}

func assertStdoutContains(t *testing.T, r Result, substr string) {
	t.Helper()
	if !strings.Contains(r.Stdout, substr) {
		t.Errorf("stdout does not contain %q\nstdout: %s", substr, r.Stdout)
	}
}

func assertStdoutNotEmpty(t *testing.T, r Result) {
	t.Helper()
	if strings.TrimSpace(r.Stdout) == "" {
		t.Error("expected non-empty stdout")
	}
}

// ── Test setup ───────────────────────────────────────────────────────────────

// loginForTest ensures the CLI is authenticated before tests run.
func loginForTest(t *testing.T) {
	t.Helper()
	r := run(t, "auth", "login",
		"--host", testHost(),
		"--api-key-id", testKeyID(),
		"--api-key-secret", testKeySecret(),
		"--no-interactive",
	)
	if r.ExitCode != 0 {
		t.Fatalf("auth login failed: %s", r.Output())
	}
}

// skipIfNoCredentials skips the test if credentials are not set.
func skipIfNoCredentials(t *testing.T) {
	t.Helper()
	if testKeyID() == "" || testKeySecret() == "" {
		t.Skip("SODA_TEST_API_KEY_ID / SODA_TEST_API_KEY_SECRET not set")
	}
}
