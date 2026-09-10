# soda-cli

The unified Soda CLI: one Go binary (`sodacli`) with a `sodacli <resource> <action>` interface that
combines Cloud API management (datasources, datasets, contracts, monitors, results, IAM, secrets,
runners, jobs, incidents) with local contract verification via the v4 engine (`soda-core`).
Previously this was split between the `soda-core` Python CLI (local execution) and the Soda Cloud
web UI (cloud management).

**Public OSS repo** (Apache-2.0). Review commit content and PR descriptions carefully before
pushing — no internal URLs, customer names, or credentials.

## Tech

- Go 1.22, module `github.com/soda-data-inc/soda-cli`, all code under `go/`
- spf13/cobra (commands), charmbracelet huh + lipgloss (interactive wizards, styling),
  olekukonko/tablewriter (table output), santhosh-tekuri/jsonschema/v6 (contract lint), yaml.v3
- goreleaser (`go/.goreleaser.yml`) — cross-platform builds, GitHub Releases, Homebrew tap
  (`sodadata/homebrew-tap`), triggered by tag via `.github/workflows/release.yml`
- Version/commit/date injected at build time via ldflags (`cmd.Version` etc.) — never hand-edit

## Build & test

```bash
cd go
go build -o sodacli .          # build the binary
go test ./...                  # unit tests — in-process, no binary, no network

# Offline CLI tests (build tag `cli`): build the real binary and drive it as a
# subprocess, asserting on stdout/stderr and exit codes. No credentials, no
# network. This is what CI runs on every PR (.github/workflows/test.yml).
go test -tags cli ./tests/integration/...

# Live tests (build tag `cloud`): the same harness against a real Cloud org.
# They CREATE AND DELETE real resources (secrets, monitors, groups, runners).
# They skip without credentials: SODA_TEST_HOST, SODA_TEST_API_KEY_ID,
# SODA_TEST_API_KEY_SECRET (+ SODA_TEST_DATASOURCE_ID/NAME,
# SODA_TEST_DATASET_ID/DQN), read from env or a repo-root .env. Not run in CI.
go test -tags cloud ./tests/integration/...
```

Both tiers live in `go/tests/integration/`, split by filename suffix
(`*_cli_test.go` / `*_cloud_test.go`) with the shared harness in `helpers_test.go`.
A new test belongs in the `cli` tier unless it genuinely needs a live org —
flag validation and blocked-command guards run *before* the credential check,
so those are offline-testable and exit 2, not 3.

From the soda-punch workspace: `just setup cli` / `just test cli`.

## Layout

- `go/main.go` — entrypoint: runs the cobra tree, maps `output.ExitError` to exit codes, fires telemetry
- `go/cmd/*.go` — one file per resource (`auth.go`, `datasource.go`, `contract.go`, `monitor.go`, …);
  `root.go` holds global flags
- `go/internal/api/` — Cloud API client (`client.go`) + typed per-resource endpoint wrappers.
  These Go structs are the CLI side of the Cloud API contract served by the `soda` backend
- `go/internal/config/` — credential store (`~/.soda/credentials`, named profiles)
- `go/internal/sodacore/` — local-verify delegation: builds the `soda contract verify --contract …
  --data-source …` argv and (for `--push`) a temp soda-cloud YAML; expects the v4 `soda` CLI on PATH
- `go/internal/lint/` — offline contract linting against the JSON schema
- `go/internal/output/` — table/json/csv rendering, TTY auto-detection, `ExitError`
- `go/internal/telemetry/` — anonymous usage events (command, flags, exit code, duration, version)
- `command_tree.txt` — the full command tree with per-command status
  (✅ live API / 🔌 wired, waiting on API / 🏠 local / ❌ no public endpoint). **Keep it updated
  when adding or changing commands** — it is the roadmap-of-record
- `skills/soda-cli/` — agent skill for *using* the CLI (not for developing it), shipped with the repo

## Key facts

- **Exit codes are a CI contract:** `0` all checks passed, `1` checks failed, `2` execution error,
  `3` authentication error. Don't repurpose them.
- **Local verify delegates to soda-core.** `sodacli contract verify --local` shells out to the v4
  `soda` CLI. If a failure reproduces with plain `soda contract verify`, the bug belongs in
  `soda-core`; only auth/API/flag-plumbing/output issues belong here.
- **Agent/CI ergonomics are load-bearing:** every command supports `--no-interactive` and
  `--output json`; output auto-switches to JSON when piped. Don't add a command that can hang on a
  prompt under `--no-interactive`.
- **Credentials** live in `~/.soda/credentials` with named profiles (`--profile` on any command).
  Never log or upload them.
- **Secrets are encrypted client-side** (AES-256-GCM + RSA-OAEP) before hitting the API — plaintext
  must never leave the machine.
- **Telemetry is opt-out** (`SODACLI_TELEMETRY=false`); it must never include personal data, API
  keys, file contents, or query data.
- Some commands are wired but blocked on API endpoints going public (🔌/❌ in `command_tree.txt`);
  they must fail gracefully with exit code 2 and a helpful message (see
  `go/tests/integration/blocked_test.go`).

## Deeper docs

- `README.md` — install, quickstart, per-resource command examples, CI/CD recipes, roadmap
- `command_tree.txt` — full command tree + implementation status
- `skills/soda-cli/SKILL.md` + `command-reference.md` — agent-facing usage guide
