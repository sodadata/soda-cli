---
name: soda-cli
description: How to use the Soda CLI for data quality management — authentication, datasources, datasets, contracts, monitors, results, permissions, and CI/CD integration. Use when working with Soda, data quality, or the sodacli command.
allowed-tools: Read, Bash(sodacli *), Bash(cat *), Glob, Grep
---

# Soda CLI Guide

The Soda CLI (`sodacli`) manages data quality from the command line. All commands follow `sodacli <resource> <action>`.

## Important: always use `--output json` when running commands

When calling `sodacli` from an agent, **always pass `--output json`** to get structured, parseable output. Without it, the CLI outputs human-readable tables that are harder to parse.

## Quick Reference — What Works

**Fully working (tested against live API):**
- `auth` — login, logout, status, switch profiles
- `datasource` — list, get, create, delete, onboard (full wizard)
- `dataset` — list (with filters), update, profiling, diagnostics, permissions, onboard
- `contract` — list, push, pull, diff, lint, create (skeleton/copilot), verify
- `monitor` — list, config, add (column/custom), update, delete
- `results` — list (with all filters, sorting, date ranges)
- `job` — logs
- `runner` — list, create (returns API keys for Kubernetes deployment)
- `iam` — user list, group CRUD, role list

**Not yet available (API blocked):**
- `incident` — list, get, update (API returns HTML)
- `notification` — rules and integrations
- `secret` — CRUD
- `job list`, `job cancel`
- `dataset get` (API returns HTML instead of JSON)
- `datasource update`, `datasource test-connection`
- `contract proposal` — list, pull, push, close
- `monitor add --type dataset` (dataset monitors exist by default, no write endpoint)

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success / all checks passed |
| `1` | One or more checks failed |
| `2` | Execution error |
| `3` | Authentication error — run `sodacli auth login` |

## Authentication

```bash
# Non-interactive (for agents and CI/CD)
sodacli auth login --host cloud.soda.io --api-key-id <id> --api-key-secret <secret>

# Check connection
sodacli auth status

# Named profiles
sodacli auth login --profile prod --host cloud.soda.io --api-key-id <id> --api-key-secret <secret>
sodacli auth switch prod
sodacli dataset list --profile prod    # any command accepts --profile
```

Credentials stored in `~/.soda/credentials`. Generate API keys at https://docs.soda.io/reference/generate-api-keys

## Core Workflows

### Discover what's there

```bash
# List datasources
sodacli datasource list --output json

# List datasets (default limit: 10)
sodacli dataset list --output json
sodacli dataset list --datasource <name> --status onboarded --limit 50 --output json

# Filter datasets by name, date range, tag
sodacli dataset list --filter "orders" --from 2026-01-01 --until 2026-12-31 --output json

# View profiling data (column stats, row count, missing %)
sodacli dataset profiling <dataset-id> --output json

# NOTE: `sodacli dataset get <id>` is NOT available yet (API returns HTML).
# Use `dataset list` with filters to find dataset details instead.
```

### Onboard a datasource (end-to-end)

```bash
# Full onboard: create datasource + discover + monitoring + profiling + contracts + verify
sodacli datasource onboard warehouse.yml --monitoring --profiling --contracts skeleton

# Or step by step:
sodacli datasource create warehouse.yml
# ⚠ Discovery is async — datasets won't appear immediately.
# Use `sodacli job logs <scan-id>` (printed by create) to confirm completion,
# or wait ~10-15 seconds before listing datasets.
sodacli datasource onboard <datasource-id> --monitoring --profiling --contracts none
```

**Onboarding a single dataset from a new datasource:**
`dataset onboard` only works after `datasource onboard` has run at least once. If you only need one table, you must still run `datasource onboard` first (which onboards all discovered datasets), then manage individual datasets afterwards.

```bash
# This will NOT work right after `datasource create`, even if the dataset appears in `dataset list`:
# sodacli dataset onboard <dataset-id>   ← returns "dataset not found"

# Instead, onboard the datasource first, then work with individual datasets:
sodacli datasource onboard <datasource-id> --monitoring --profiling --contracts copilot
# Now `dataset onboard`, `dataset profiling`, `monitor add`, etc. work on individual datasets.
```

**Contract generation modes:**
- `skeleton` — schema structure + not-null checks only (fast, deterministic)
- `copilot` — AI analyzes your data and generates format, range, uniqueness, and business-logic checks (takes longer, may need manual fixes — see Troubleshooting below)

When `--contracts skeleton` or `--contracts copilot` is used, the onboard flow automatically verifies the generated contracts against your data and displays pass/fail results.

Datasource config file format:
```yaml
type: snowflake           # snowflake, postgres, bigquery, mysql, redshift, etc.
name: my_warehouse
connection:
  host: account.snowflakecomputing.com
  database: ANALYTICS
  schema: PUBLIC
  user: soda_user
  password: secret
  role: SODA_ROLE
  warehouse: COMPUTE_WH
```

### Contracts (data quality checks)

```bash
# List contracts
sodacli contract list --output json

# Generate a contract from live schema
sodacli contract create --dataset datasource/db/schema/table --mode skeleton --output my_table.yml

# Pull existing contract from cloud
sodacli contract pull datasource/db/schema/table

# Edit locally, then push back
sodacli contract push my_table.yml

# Compare local vs cloud
sodacli contract diff my_table.yml

# Validate syntax (no network)
sodacli contract lint my_table.yml

# Run checks via cloud Runner
sodacli contract verify my_table.yml --output json

# Fire and forget (returns immediately)
sodacli contract verify my_table.yml --no-wait
```

### Monitors (ML anomaly detection)

```bash
# List monitors on a dataset (--dataset is required)
sodacli monitor list --dataset <dataset-id> --output json

# View monitoring config
sodacli monitor config <dataset-id> --output json

# Enable monitoring with schedule
sodacli monitor config <dataset-id> --enable --schedule "0 6 * * *" --timezone "UTC"

# Add column monitor
sodacli monitor add --dataset <id> --type column --column revenue --metric avg
sodacli monitor add --dataset <id> --type column --column order_id --metric count --group-by region

# Add custom SQL monitor
sodacli monitor add --dataset <id> --type custom \
  --name "dup check" \
  --sql "SELECT count(*) as c FROM t GROUP BY id HAVING count(*) > 1" \
  --result-metric c

# Update a monitor (enable/disable, change SQL)
sodacli monitor update <monitor-id> --dataset <id> --disable
sodacli monitor update <monitor-id> --dataset <id> --name "renamed" --sql "SELECT 1 as c"

# Delete
sodacli monitor delete <monitor-id> --dataset <id>
```

Column metric types: `count`, `missing-pct`, `duplicate-pct`, `distinct-count`, `min`, `max`, `avg`, `sum`, `std-dev`, `variance`, `q1`, `median`, `q3`, `min-length`, `max-length`, `avg-length`, `freshness`

### Results (check outcomes)

```bash
# Recent results
sodacli results list --output json

# Filter by dataset, status, date
sodacli results list --dataset <id> --status failing --output json
sodacli results list --dataset-name "orders" --from 2026-03-01 --until 2026-03-31 --output json

# Sort and paginate
sodacli results list --limit 50 --sort name --order asc --output json
```

### Permissions

```bash
# List roles and users (to find IDs)
sodacli iam role list --output json
sodacli iam user list --output json

# List dataset permissions
sodacli dataset permissions list <dataset-id> --output json

# Grant / revoke
sodacli dataset permissions assign <dataset-id> --role <role-id> --user <user-id>
sodacli dataset permissions revoke <dataset-id> --role <role-id> --user <user-id>
```

### Groups

```bash
sodacli iam group list --output json
sodacli iam group create --name "Data Engineers" --member alice@co.com --member bob@co.com
sodacli iam group update <id> --add-member carol@co.com
sodacli iam group update <id> --remove-member bob@co.com
sodacli iam group delete <id>
```

### Job logs

```bash
sodacli job logs <scan-id>
sodacli job logs <scan-id> --follow    # stream live
```

### Profiling and diagnostics

```bash
# Enable profiling
sodacli dataset profiling <dataset-id> --enable --schedule "0 6 * * *" --sampling-rows 1000000

# View profiling data
sodacli dataset profiling <dataset-id> --output json

# Set time-partition column
sodacli dataset time-partition <dataset-id> --column created_at

# Configure diagnostics
sodacli dataset diagnostics <dataset-id> --collect-results --collect-failed-rows
```

## CI/CD Pattern

```bash
# Authenticate (non-interactive)
sodacli auth login --host cloud.soda.io \
  --api-key-id "$SODA_API_KEY_ID" \
  --api-key-secret "$SODA_API_KEY_SECRET" \
  --no-interactive

# Run contract checks
sodacli contract verify contracts/orders.yml --no-interactive --output json

# Exit codes drive pipeline:
# 0 = all checks passed
# 1 = checks failed  →  fail the pipeline
# 2 = execution error →  retry or alert
# 3 = auth error      →  check credentials
```

## Troubleshooting

### Contract verification fails with exit code 2
When `contract verify` returns exit code 2, the output may only show "0 checks passed" without details. Always check the scan logs for the actual error:
```bash
sodacli job logs <scan-id>
```
The scan ID is printed by the `contract verify` command.

### Copilot contracts: UUID regex checks fail on Postgres
Copilot may generate `valid_format` regex checks on UUID columns (e.g., `'^[0-9a-fA-F]{8}-...'`). On Postgres, the `~` regex operator doesn't work on `uuid` type columns, causing a SQL error. **Fix:** Remove `invalid` / `valid_format` checks on UUID columns — the database data type already enforces UUID format. Keep `missing` and `duplicate` checks, which work fine.

### Discovery scan not complete
After `datasource create`, the discovery scan runs asynchronously. If `dataset list` returns empty or `dataset onboard` returns "not found", the scan may still be running. Check with:
```bash
sodacli job logs <scan-id>   # scan-id is printed by datasource create
```

## Global Flags (all commands)

| Flag | Short | Description |
|------|-------|-------------|
| `--output table\|json\|csv` | `-o` | Output format (auto-detects TTY vs pipe) |
| `--profile <name>` | | Override active auth profile |
| `--no-color` | | Disable color output |
| `--quiet` | `-q` | Suppress non-essential output |
| `--verbose` | `-v` | Show detailed output |
| `--no-interactive` | | Never prompt; fail with clear error if input missing |

## Full Command Reference

For detailed flags and usage of every command, see [command-reference.md](command-reference.md).
