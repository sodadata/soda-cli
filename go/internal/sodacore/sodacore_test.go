package sodacore

import (
	"os"
	"strings"
	"testing"
)

func TestBuildVerifyArgs_Basic(t *testing.T) {
	args := BuildVerifyArgs(VerifyOpts{
		ContractFile:   "contract.yml",
		DatasourceFile: "ds.yml",
	})
	want := []string{"contract", "verify", "--contract", "contract.yml", "--data-source", "ds.yml"}
	if len(args) != len(want) {
		t.Fatalf("got %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Errorf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestBuildVerifyArgs_WithPublish(t *testing.T) {
	args := BuildVerifyArgs(VerifyOpts{
		ContractFile:   "contract.yml",
		DatasourceFile: "ds.yml",
		Publish:        true,
		SodaCloudFile:  "/tmp/sc.yml",
	})
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--soda-cloud /tmp/sc.yml") {
		t.Errorf("expected --soda-cloud flag, got: %s", joined)
	}
	if !strings.Contains(joined, "--publish") {
		t.Errorf("expected --publish flag, got: %s", joined)
	}
}

func TestBuildVerifyArgs_WithSetVars(t *testing.T) {
	args := BuildVerifyArgs(VerifyOpts{
		ContractFile:   "contract.yml",
		DatasourceFile: "ds.yml",
		SetVars:        []string{"KEY=val1", "KEY2=val2"},
	})
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--set KEY=val1") {
		t.Errorf("expected --set KEY=val1, got: %s", joined)
	}
	if !strings.Contains(joined, "--set KEY2=val2") {
		t.Errorf("expected --set KEY2=val2, got: %s", joined)
	}
}

func TestBuildVerifyArgs_Verbose(t *testing.T) {
	args := BuildVerifyArgs(VerifyOpts{
		ContractFile:   "contract.yml",
		DatasourceFile: "ds.yml",
		Verbose:        true,
	})
	found := false
	for _, a := range args {
		if a == "--verbose" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected --verbose flag, got: %v", args)
	}
}

func TestMapExitCode(t *testing.T) {
	tests := []struct {
		in     int
		stderr string
		want   int
	}{
		{0, "", 0},
		{1, "", 1},
		{2, "", 0},                                    // warnings non-fatal
		{2, "[Error] unknown flag: --foo", 2},         // usage error on stderr → treat as error
		{2, "unknown shorthand flag: 'x'", 2},         // cobra shorthand error
		{2, "soda: error: unrecognized arguments: foo", 2}, // Python argparse error
		{3, "", 2},
		{127, "", 2},
		{-1, "", 2},
	}
	for _, tt := range tests {
		got := MapExitCode(tt.in, tt.stderr)
		if got != tt.want {
			t.Errorf("MapExitCode(%d, %q) = %d, want %d", tt.in, tt.stderr, got, tt.want)
		}
	}
}

func TestInstallHint(t *testing.T) {
	hint := InstallHint()
	if !strings.Contains(hint, "pip install") {
		t.Error("expected pip install instructions in hint")
	}
	if !strings.Contains(hint, "docs.soda.io") {
		t.Error("expected docs link in hint")
	}
}

func TestWriteTempCloudConfig(t *testing.T) {
	path, cleanup, err := WriteTempCloudConfig("cloud.soda.io", "key_id", "key_secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cleanup()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read temp file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "cloud.soda.io") {
		t.Error("expected host in config")
	}
	if !strings.Contains(content, "key_id") {
		t.Error("expected api_key_id in config")
	}

	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode 0600, got %o", info.Mode().Perm())
	}
}
