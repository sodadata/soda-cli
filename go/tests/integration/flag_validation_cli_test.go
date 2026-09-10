//go:build cli

package integration

import (
	"testing"
)

// TestFlagValidation asserts that flag and argument validation happens before
// any Cloud call, so a misconfigured invocation fails with exit code 2
// (execution error) rather than 3 (authentication error).
//
// That ordering is what makes these cases testable without credentials, and it
// matters for CI ergonomics: a pipeline with a bad flag should be told about
// the bad flag, not about its credentials. The results are the same whether or
// not ~/.soda/credentials exists.
func TestFlagValidation(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExit     int
		wantContains string
	}{
		{
			name:         "auth_login_missing_keys",
			args:         []string{"auth", "login", "--no-interactive"},
			wantExit:     2,
			wantContains: "required",
		},
		{
			name:         "contract_create_no_dataset",
			args:         []string{"contract", "create", "--no-interactive"},
			wantExit:     2,
			wantContains: "required",
		},
		{
			name:     "contract_verify_no_args",
			args:     []string{"contract", "verify"},
			wantExit: 2,
		},
		{
			name:         "secret_create_missing_name",
			args:         []string{"secret", "create"},
			wantExit:     2,
			wantContains: "--name is required",
		},
		{
			name:         "runner_create_missing_name",
			args:         []string{"runner", "create"},
			wantExit:     2,
			wantContains: "--name is required",
		},
		{
			name:     "datasource_create_bad_file",
			args:     []string{"datasource", "create", "nonexistent.yml"},
			wantExit: 2,
		},
		{
			name:         "monitor_add_no_dataset",
			args:         []string{"monitor", "add", "--type", "column", "--column", "x", "--metric", "count"},
			wantExit:     2,
			wantContains: "--dataset is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := run(t, tc.args...)
			assertExitCode(t, r, tc.wantExit)
			if tc.wantContains != "" {
				assertOutputContains(t, r, tc.wantContains)
			}
		})
	}
}
