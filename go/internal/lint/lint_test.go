package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintFile_ValidContract(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "contract.yml")
	os.WriteFile(f, []byte(`
dataset: my_ds/db/schema/orders
columns:
  - name: id
    data_type: INTEGER
  - name: status
    data_type: VARCHAR
`), 0644)

	result, err := LintFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		for _, e := range result.Errors {
			t.Logf("  %s: %s", e.Path, e.Message)
		}
		t.Fatalf("expected valid, got %d errors", len(result.Errors))
	}
}

func TestLintFile_InvalidProperty(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "contract.yml")
	os.WriteFile(f, []byte(`
dataset: my_ds/db/schema/orders
bogus_field: true
columns:
  - name: id
`), 0644)

	result, err := LintFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid, got valid")
	}
	found := false
	for _, e := range result.Errors {
		t.Logf("  %s: %s", e.Path, e.Message)
		if e.Path == "$" || e.Path == "$.bogus_field" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected an error about bogus_field")
	}
}

func TestLintFile_YAMLParseError(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "bad.yml")
	os.WriteFile(f, []byte(`
dataset: test
  bad indent: here
`), 0644)

	result, err := LintFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid for malformed YAML")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected at least one error")
	}
	if result.Errors[0].Path != "$" {
		t.Fatalf("expected path '$', got %q", result.Errors[0].Path)
	}
}

func TestLintFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "empty.yml")
	os.WriteFile(f, []byte(""), 0644)

	result, err := LintFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid for empty file")
	}
}

func TestLintFile_MissingColumnName(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "contract.yml")
	os.WriteFile(f, []byte(`
dataset: my_ds/db/schema/orders
columns:
  - data_type: INTEGER
`), 0644)

	result, err := LintFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid — column missing required 'name'")
	}
	for _, e := range result.Errors {
		t.Logf("  %s: %s", e.Path, e.Message)
	}
}

func TestLintFiles_Mixed(t *testing.T) {
	dir := t.TempDir()

	good := filepath.Join(dir, "good.yml")
	os.WriteFile(good, []byte(`
dataset: ds/db/schema/t
columns:
  - name: id
`), 0644)

	bad := filepath.Join(dir, "bad.yml")
	os.WriteFile(bad, []byte(`
dataset: ds/db/schema/t
unknown_prop: true
`), 0644)

	results, err := LintFiles([]string{good, bad})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if !results[0].Valid {
		t.Error("expected first file to be valid")
	}
	if results[1].Valid {
		t.Error("expected second file to be invalid")
	}
}

func TestLintFile_AcceptsSodaRunner(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "contract.yml")
	os.WriteFile(f, []byte(`
dataset: my_ds/db/schema/orders
columns:
  - name: id
soda_runner:
  checks_schedule:
    cron: "0 0 * * *"
    timezone: UTC
`), 0644)

	result, err := LintFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		for _, e := range result.Errors {
			t.Logf("  %s: %s", e.Path, e.Message)
		}
		t.Fatalf("expected valid contract with soda_runner, got %d errors", len(result.Errors))
	}
}

func TestLintFile_AcceptsSodaAgentLegacyAlias(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "contract.yml")
	os.WriteFile(f, []byte(`
dataset: my_ds/db/schema/orders
columns:
  - name: id
soda_agent:
  checks_schedule:
    cron: "0 0 * * *"
    timezone: UTC
`), 0644)

	result, err := LintFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		for _, e := range result.Errors {
			t.Logf("  %s: %s", e.Path, e.Message)
		}
		t.Fatalf("expected valid contract with deprecated soda_agent, got %d errors", len(result.Errors))
	}
}

func TestLintFile_RejectsBothSodaRunnerAndSodaAgent(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "contract.yml")
	os.WriteFile(f, []byte(`
dataset: my_ds/db/schema/orders
columns:
  - name: id
soda_runner:
  checks_schedule:
    cron: "0 0 * * *"
soda_agent:
  checks_schedule:
    cron: "0 0 * * *"
`), 0644)

	result, err := LintFile(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Fatal("expected invalid when both soda_runner and soda_agent are set")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected at least one validation error")
	}
	rootError := false
	for _, e := range result.Errors {
		t.Logf("  %s: %s", e.Path, e.Message)
		if e.Path == "$" {
			rootError = true
		}
	}
	if !rootError {
		t.Fatal("expected a validation error at the root path '$' for the not-both constraint")
	}
}

func TestSegmentsToPath(t *testing.T) {
	tests := []struct {
		in   []string
		want string
	}{
		{nil, "$"},
		{[]string{"columns", "0", "name"}, "$.columns[0].name"},
		{[]string{"checks", "2"}, "$.checks[2]"},
		{[]string{"dataset"}, "$.dataset"},
	}
	for _, tt := range tests {
		got := segmentsToPath(tt.in)
		if got != tt.want {
			t.Errorf("segmentsToPath(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
