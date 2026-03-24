# Soda CLI

A single command-line tool for [Soda](https://www.soda.io) data quality. Manage datasources, datasets, contracts, monitors, incidents, and permissions from your terminal or pipeline.

<p align="center">
  <img src="assets/hero.png" alt="Soda CLI" width="580">
</p>

Previously this was split between `soda-core` (local execution) and the Soda Cloud web UI (cloud management). Soda CLI unifies both into one `sodacli <resource> <action>` interface.

> **AI-agent friendly.** Every command supports `--no-interactive`, `--output json`, and structured exit codes, so it works well with LLMs, orchestrators, and CI/CD. This project includes a [SKILL](skills/soda-cli) so Claude/Codex or any other agent can run Soda commands, interpret results, and manage data quality through natural conversation.

## Current Status

**Version:** `v0.2.0` (active development)

The CLI is functional for core workflows. Here's where things stand:

| Area | Status |
|---|---|
| Auth (login, logout, status, profiles) | Working |
| Datasource (list, get, create, delete, onboard) | Working |
| Dataset (list, get, update, delete, profiling, permissions, onboard) | Working |
| Contract (list, push, pull, diff, create, lint, verify via cloud or local) | Working |
| Monitor (list, config, add column/custom, update, delete) | Working |
| Results (list with filtering, sorting, date ranges) | Working |
| Runner (list, get, create, delete) | Working |
| IAM (user list, group CRUD, role list) | Working |
| Job logs | Working |
| Contract verify (local via soda-core) | Working |
| Incidents | Planned |
| Notifications, Secrets | Planned |
| Dashboard | Planned |

Per-command status is tracked in [`command_tree.txt`](command_tree.txt):

```
✅  implemented with real API call
🔌  CLI wired, waiting on API endpoint
🏠  local operation, no API needed
❌  no public API endpoint yet
```

## Install

### Homebrew (macOS/Linux)

```bash
brew tap sodadata/tap
brew install sodacli
```

### Install script (macOS/Linux)

```bash
curl -sSL https://raw.githubusercontent.com/sodadata/soda-cli/main/install.sh | sh
```

### Windows

Download the latest `.zip` for your architecture from [GitHub Releases](https://github.com/sodadata/soda-cli/releases), extract `sodacli.exe`, and add it to your PATH.

### Download binary (any platform)

Grab the archive for your OS/arch from [GitHub Releases](https://github.com/sodadata/soda-cli/releases), extract, and add to your PATH. Available for Linux, macOS, and Windows (amd64 + arm64).

### From source (Go 1.22+)

```bash
git clone https://github.com/sodadata/soda-cli.git
cd soda-cli/go
go build -o sodacli .
sudo mv sodacli /usr/local/bin/   # macOS/Linux
```

### Verify

```bash
sodacli version
sodacli --help
```

## Quickstart

### 1. Authenticate

```bash
# Interactive: prompts for host, API key ID, and secret
sodacli auth login

# Check that it worked
sodacli auth status
```

Generate API keys at [docs.soda.io/reference/generate-api-keys](https://docs.soda.io/reference/generate-api-keys).

### 2. Onboard a datasource

```bash
# Full onboard: create datasource, discover datasets, enable monitoring + profiling + contracts + verify
sodacli datasource onboard warehouse.yml --monitoring --profiling --contracts copilot
```

Or step by step:

```bash
sodacli datasource create warehouse.yml           # register datasource, returns ID
sodacli dataset list --datasource my_warehouse    # see discovered datasets
sodacli datasource onboard <datasource-id> --monitoring --profiling --contracts skeleton
```

### 3. Verify a contract

```bash
# Run checks via Soda Cloud Runner
sodacli contract verify orders.yml

# Or run locally via soda-core (no cloud needed)
sodacli contract verify orders.yml --local --datasource datasource.yml

# Check results
sodacli results list --status failing
sodacli job logs <scan-id>
```

## Essential Commands

### Authentication

```bash
sodacli auth login                  # interactive setup
sodacli auth login --host cloud.us.soda.io --api-key-id <id> --api-key-secret <secret>
sodacli auth status                 # check connection health
sodacli auth switch <profile>       # switch between profiles (planned)
```

### Datasources

```bash
sodacli datasource list
sodacli datasource get <id>
sodacli datasource create config.yml                          # register from YAML config
sodacli datasource onboard config.yml --monitoring --profiling --contracts skeleton  # full setup
sodacli datasource delete <id>
```

### Datasets

```bash
sodacli dataset list --datasource <name> --status onboarded --limit 50
sodacli dataset get <id>
sodacli dataset update <id> --tag production --tag critical
sodacli dataset profiling <id> --enable --schedule "0 6 * * *"
sodacli dataset time-partition <id> --column created_at
sodacli dataset diagnostics <id> --collect-results --collect-failed-rows
sodacli dataset permissions list <id>
sodacli dataset permissions assign <id> --role <role-id> --user <user-email>
```

### Contracts

```bash
sodacli contract list
sodacli contract create --dataset ds/db/schema/table --mode skeleton     # generate from schema
sodacli contract create --dataset ds/db/schema/table --mode copilot      # AI-generated checks
sodacli contract pull ds/db/schema/table                                 # download from cloud
sodacli contract push my_table.yml                                       # upload to cloud
sodacli contract diff my_table.yml                                       # local vs cloud diff
sodacli contract lint my_table.yml                                       # validate syntax (offline)
sodacli contract lint contracts/*.yml                                    # lint multiple files
sodacli contract verify my_table.yml                                     # run checks via cloud Runner
sodacli contract verify my_table.yml --no-wait                           # fire and forget
sodacli contract verify my_table.yml --local --datasource config.yml     # run locally via soda-core
sodacli contract verify my_table.yml --local --datasource config.yml --push  # run locally + push results to cloud
```

### Monitors

```bash
sodacli monitor list --dataset <id>
sodacli monitor config <dataset-id> --enable --schedule "0 */6 * * *" --timezone "UTC"
sodacli monitor add --dataset <id> --type column --column revenue --metric avg
sodacli monitor add --dataset <id> --type column --column order_id --metric count --group-by region
sodacli monitor add --dataset <id> --type custom --name "dup check" \
  --sql "SELECT count(*) as c FROM t" --result-metric c
sodacli monitor update <monitor-id> --dataset <id> --disable
sodacli monitor delete <monitor-id> --dataset <id>
```

### Results & Jobs

```bash
sodacli results list
sodacli results list --dataset-name "orders" --status failing --from 2026-03-01 --limit 20
sodacli job logs <scan-id>
```

### IAM

```bash
sodacli iam user list
sodacli iam group create --name "Data Engineers" --member alice@co.com --member bob@co.com
sodacli iam group update <id> --add-member carol@co.com
sodacli iam role list --scope dataset
```

### Runners

```bash
sodacli runner list
sodacli runner get <id>
sodacli runner create --name "prod-runner"    # returns credentials (shown once)
sodacli runner delete <id>
```

## CI/CD Integration

Every command works non-interactively:

```bash
# Authenticate
sodacli auth login \
  --host cloud.soda.io \
  --api-key-id "$SODA_API_KEY_ID" \
  --api-key-secret "$SODA_API_KEY_SECRET" \
  --no-interactive

# Run contract checks (via cloud Runner)
sodacli contract verify contracts/orders.yml --no-interactive --output json

# Or run locally (no cloud auth needed, just soda-core on PATH)
sodacli contract verify contracts/orders.yml --local --datasource datasource.yml

# Exit codes
# 0 = all checks passed
# 1 = one or more checks failed  →  fail the pipeline
# 2 = execution error             →  retry or alert
# 3 = authentication error        →  check credentials
```

### GitHub Actions example

```yaml
- name: Verify data contracts
  run: |
    sodacli auth login --host cloud.soda.io \
      --api-key-id ${{ secrets.SODA_KEY_ID }} \
      --api-key-secret ${{ secrets.SODA_KEY_SECRET }} \
      --no-interactive
    sodacli contract verify contracts/orders.yml --no-interactive
```

## Output Formats

The CLI picks the right format automatically:

- **TTY** (interactive terminal): human-readable tables with color
- **Piped** (`sodacli dataset list | jq .`): JSON
- **Override**: `--output json|table|csv` on any command

```bash
sodacli dataset list                    # colored table
sodacli dataset list --output json      # JSON
sodacli dataset list --output csv       # CSV
sodacli dataset list | jq '.[] | .id'  # auto-JSON when piped
```

## Global Flags

These work on every command:

| Flag | Description |
|---|---|
| `--output table\|json\|csv` | Output format (auto-detects TTY) |
| `--profile <name>` | Override active auth profile |
| `--no-color` | Disable color output |
| `--quiet` | Suppress non-essential output |
| `--verbose` | Show detailed output |
| `--no-interactive` | Never prompt, fail with clear error if input is missing |

## Telemetry

Soda CLI collects anonymous usage data (command names, exit codes, duration, OS/arch) to help us understand which features are used and improve the tool. No personal information, API keys, file contents, or query data is ever collected.

To opt out:

```bash
export SODACLI_TELEMETRY=false
```

## What's Missing & Roadmap

### Waiting on Soda Cloud API

The CLI code is written for these. They'll work as soon as the API endpoints ship:

- **Incidents** (list, get, update)
- **Notifications** (rules and integrations CRUD)
- **Secrets** (CRUD for cloud-stored secrets)
- **Job list** (scan history)
- **Job cancel** (cancel running scans)
- **Datasource update** (change labels, runners, connection configs)
- **Datasource test-connection** (async connection test)

### Planned Features

- **Dashboard.** Org-level overview of datasets, results, and incidents.
- **Contract proposals.** PR-style review flow for contract changes.

### Vision

The goal is one CLI that covers the full data quality lifecycle:

1. **Connect.** `sodacli datasource onboard` sets up a database connection with monitoring, profiling, contracts, and verification in one command.
2. **Define.** `sodacli contract create --mode copilot` uses AI to generate meaningful checks from your schema and data profile.
3. **Import.** `sodacli contract translate` translates existing data quality definitions from other formats (ODCS, dbt tests, Great Expectations, SodaCL v3) into Soda contracts.
4. **Verify.** `sodacli contract verify` runs checks locally or in the cloud, from CI/CD or your terminal.
5. **Monitor.** `sodacli monitor` adds ML anomaly detection that fires alerts when metrics drift.
6. **Respond.** `sodacli incident` and `sodacli notification` close the loop from detection to resolution.
7. **Govern.** `sodacli iam` and `sodacli dataset permissions` control who can do what.

All of this works the same way for humans typing commands and for AI agents calling them programmatically. Same interface, same exit codes, same JSON output.

## Soda CLI vs soda-core

| | Soda CLI (`sodacli`) | soda-core (`soda`) |
|---|---|---|
| **Language** | Go (single binary, no dependencies) | Python (requires pip + DB connectors) |
| **Execution** | Cloud via Soda Runner, or local via `--local` | Local only |
| **Scope** | Full platform: datasources, datasets, contracts, monitors, results, IAM, incidents | Contract verification and data source testing |
| **Contract generation** | `contract create --mode copilot` (AI) or `skeleton` | Manual authoring only |
| **CI/CD** | `--no-interactive`, `--output json`, structured exit codes | Basic exit codes |

**Why use Soda CLI?** If you only need to run checks locally, soda-core is enough. If you want to manage your entire data quality lifecycle from one tool — generate contracts with AI, monitor anomalies, track results, control permissions, and integrate with CI/CD — use sodacli. It shells out to soda-core for local execution when needed (`--local`), so you get both.
