# Soda CLI

A single command-line tool for [Soda](https://www.soda.io) data quality. Manage datasources, datasets, contracts, monitors, incidents, and permissions from your terminal or pipeline.

Previously this was split between `soda-core` (local execution) and the Soda Cloud web UI (cloud management). Soda CLI unifies both into one `soda <resource> <action>` interface.

> **AI-agent friendly.** Every command supports `--no-interactive`, `--output json`, and structured exit codes, so it works well with LLMs, orchestrators, and CI/CD. This project includes a [SKILL](skills/soda-cli) so Claude/Codex or any other agent can run Soda commands, interpret results, and manage data quality through natural conversation.

## Current Status

**Version:** `0.1.0-dev` (active development)

The CLI is functional for core workflows. Here's where things stand:

| Area | Status |
|---|---|
| Auth (login, logout, status, profiles) | Working |
| Datasource (list, get, create, delete, onboard) | Working |
| Dataset (list, get, update, delete, profiling, permissions, onboard) | Working |
| Contract (list, push, pull, diff, create, verify via cloud) | Working |
| Monitor (list, config, add column/custom, update, delete) | Working |
| Results (list with filtering, sorting, date ranges) | Working |
| Runner (list, get, create, delete) | Working |
| IAM (user list, group CRUD, role list) | Working |
| Job logs | Working |
| Contract verify (local via soda-core) | Planned |
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

### From source (Go 1.22+)

```bash
git clone https://github.com/soda-data-inc/soda-cli.git
cd soda-cli/go
go build -o soda .
sudo mv soda /usr/local/bin/
```

### Verify

```bash
soda version
soda --help
```

## Quickstart

### 1. Authenticate

```bash
# Interactive: prompts for host, API key ID, and secret
soda auth login

# Check that it worked
soda auth status
```

Generate API keys at [docs.soda.io/reference/generate-api-keys](https://docs.soda.io/reference/generate-api-keys).

### 2. Onboard a datasource

```bash
# Full onboard: create datasource, discover datasets, enable monitoring + profiling + contracts
soda datasource onboard warehouse.yml --monitoring --profiling --contracts ai
```

Or step by step:

```bash
soda datasource create warehouse.yml           # register datasource, returns ID
soda dataset list --datasource my_warehouse    # see discovered datasets
soda datasource onboard <datasource-id> --monitoring --profiling --contracts skeleton
```

### 3. Verify a contract

```bash
# Run checks against your data
soda contract verify orders.yml

# Check results
soda results list --status failing
soda job logs <scan-id>
```

## Essential Commands

### Authentication

```bash
soda auth login                  # interactive setup
soda auth login --host cloud.us.soda.io --api-key-id <id> --api-key-secret <secret>
soda auth status                 # check connection health
soda auth switch <profile>       # switch between profiles (planned)
```

### Datasources

```bash
soda datasource list
soda datasource get <id>
soda datasource create config.yml                          # register from YAML config
soda datasource onboard config.yml --monitoring --profiling --contracts skeleton  # full setup
soda datasource delete <id>
```

### Datasets

```bash
soda dataset list --datasource <name> --status onboarded --limit 50
soda dataset get <id>
soda dataset update <id> --tag production --tag critical
soda dataset profiling <id> --enable --schedule "0 6 * * *"
soda dataset time-partition <id> --column created_at
soda dataset diagnostics <id> --collect-results --collect-failed-rows
soda dataset permissions list <id>
soda dataset permissions assign <id> --role <role-id> --user <user-email>
```

### Contracts

```bash
soda contract list
soda contract create --dataset ds/db/schema/table --mode skeleton     # generate from schema
soda contract create --dataset ds/db/schema/table --mode copilot      # AI-generated checks
soda contract pull ds/db/schema/table                                 # download from cloud
soda contract push my_table.yml                                       # upload to cloud
soda contract diff my_table.yml                                       # local vs cloud diff
soda contract lint my_table.yml                                       # validate syntax
soda contract verify my_table.yml                                     # run checks via cloud Runner
soda contract verify my_table.yml --no-wait                           # fire and forget
soda contract verify my_table.yml --runner soda-core --datasource config.yml  # run locally (planned)
```

### Monitors

```bash
soda monitor list --dataset <id>
soda monitor config <dataset-id> --enable --schedule "0 */6 * * *" --timezone "UTC"
soda monitor add --dataset <id> --type column --column revenue --metric avg
soda monitor add --dataset <id> --type column --column order_id --metric count --group-by region
soda monitor add --dataset <id> --type custom --name "dup check" \
  --sql "SELECT count(*) as c FROM t" --result-metric c
soda monitor update <monitor-id> --dataset <id> --disable
soda monitor delete <monitor-id> --dataset <id>
```

### Results & Jobs

```bash
soda results list
soda results list --dataset-name "orders" --status failing --from 2026-03-01 --limit 20
soda job logs <scan-id>
```

### IAM

```bash
soda iam user list
soda iam group create --name "Data Engineers" --member alice@co.com --member bob@co.com
soda iam group update <id> --add-member carol@co.com
soda iam role list --scope dataset
```

### Runners

```bash
soda runner list
soda runner get <id>
soda runner create --name "prod-runner"    # returns credentials (shown once)
soda runner delete <id>
```

## CI/CD Integration

Every command works non-interactively:

```bash
# Authenticate
soda auth login \
  --host cloud.soda.io \
  --api-key-id "$SODA_API_KEY_ID" \
  --api-key-secret "$SODA_API_KEY_SECRET" \
  --no-interactive

# Run contract checks
soda contract verify contracts/orders.yml --no-interactive --output json

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
    soda auth login --host cloud.soda.io \
      --api-key-id ${{ secrets.SODA_KEY_ID }} \
      --api-key-secret ${{ secrets.SODA_KEY_SECRET }} \
      --no-interactive
    soda contract verify contracts/orders.yml --no-interactive
```

## Output Formats

The CLI picks the right format automatically:

- **TTY** (interactive terminal): human-readable tables with color
- **Piped** (`soda dataset list | jq .`): JSON
- **Override**: `--output json|table|csv` on any command

```bash
soda dataset list                    # colored table
soda dataset list --output json      # JSON
soda dataset list --output csv       # CSV
soda dataset list | jq '.[] | .id'  # auto-JSON when piped
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

- **Local contract execution.** `soda contract verify --runner soda-core` runs checks locally via the soda-core Python engine, no Soda Cloud needed.
- **Real contract linting.** `soda contract lint` using soda-core for full schema validation.
- **Dashboard.** Org-level overview of datasets, results, and incidents.
- **Contract proposals.** PR-style review flow for contract changes.

### Vision

The goal is one CLI that covers the full data quality lifecycle:

1. **Connect.** `soda datasource onboard` sets up a database connection with monitoring, profiling, and contracts in one command.
2. **Define.** `soda contract create --mode copilot` uses AI to generate meaningful checks from your schema and data profile.
3. **Verify.** `soda contract verify` runs checks locally or in the cloud, from CI/CD or your terminal.
4. **Monitor.** `soda monitor` adds ML anomaly detection that fires alerts when metrics drift.
5. **Respond.** `soda incident` and `soda notification` close the loop from detection to resolution.
6. **Govern.** `soda iam` and `soda dataset permissions` control who can do what.

All of this works the same way for humans typing commands and for AI agents calling them programmatically. Same interface, same exit codes, same JSON output.
