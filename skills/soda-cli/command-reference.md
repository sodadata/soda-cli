# Soda CLI Command Reference

All commands support these global flags:

| Flag | Description |
|---|---|
| `--output table\|json\|csv` | Output format (auto-detects TTY) |
| `--profile <name>` | Override active auth profile |
| `--no-color` | Disable color output |
| `--quiet` | Suppress non-essential output |
| `--verbose` | Show detailed output |
| `--no-interactive` | Never prompt; fail if input missing |

---

## auth

### `soda auth login`
Authenticate with Soda Cloud. Stores credentials in `~/.soda/credentials`.

| Flag | Description |
|---|---|
| `--host <host>` | Soda Cloud host (default: cloud.soda.io; US: cloud.us.soda.io) |
| `--api-key-id <id>` | API key ID |
| `--api-key-secret <secret>` | API key secret |
| `--profile <name>` | Save to named profile |

### `soda auth status`
Show active profile and connection health.

### `soda auth logout`
Remove stored credentials for the active profile.

### `soda auth switch <profile>`
Switch the active auth profile.

---

## datasource (alias: ds)

### `soda datasource list`
List all datasources. Returns: id, name, label, type, created, updated.

### `soda datasource create <config-file>`
Register a datasource from a YAML connection config.

| Flag | Description |
|---|---|
| `--runner <id>` | Soda Runner ID (auto-detects if only one) |

### `soda datasource test-connection <config-file>`
Test a datasource connection via Soda Runner (async poll, 2min timeout).

| Flag | Description |
|---|---|
| `--runner <id>` | Soda Runner ID (auto-detects if only one) |

### `soda datasource onboard <config-file-or-datasource-id>`
Guided setup: create datasource + discover datasets + onboard.

| Flag | Description |
|---|---|
| `--runner <id>` | Soda Runner ID (only for new datasources) |
| `--monitoring` | Enable default metric monitors |
| `--no-monitoring` | Skip monitoring setup |
| `--profiling` | Enable dataset profiling |
| `--no-profiling` | Skip profiling setup |
| `--contracts ai\|skeleton\|none` | Generate contracts for all datasets |

When all action flags provided, runs fully non-interactively.

### `soda datasource delete <id>`
Delete a datasource (schedules deletion).

### `soda datasource diagnostics <id>`
View or configure the diagnostics warehouse.

| Flag | Description |
|---|---|
| `--enable` / `--disable` | Toggle diagnostics warehouse |
| `--warehouse same\|<config-file>` | Warehouse connection |
| `--schema <name>` | Schema for diagnostic tables |
| `--collect-results` / `--no-collect-results` | Store check results |
| `--collect-failed-rows` / `--no-collect-failed-rows` | Store failed rows |

---

## dataset

### `soda dataset list`
List datasets (onboarded + discovered).

| Flag | Default | Description |
|---|---|---|
| `--filter <query>` | | Fuzzy search on dataset name |
| `--datasource <name>` | | Filter by datasource name |
| `--id <substring>` | | Filter by dataset ID |
| `--status onboarded\|not-onboarded` | all | Filter by onboard status |
| `--limit <n>` | 10 | Max rows (0 = unlimited) |
| `--from <date>` | | Updated on or after (YYYY-MM-DD) |
| `--until <date>` | | Updated on or before (YYYY-MM-DD) |
| `--tag <tag>` | | Filter by tag |

### `soda dataset get <id>`
Show details: name, qualified name, datasource, DQ status, checks, incidents, partition column, tags, cloud URL.

### `soda dataset update <id>`

| Flag | Description |
|---|---|
| `--owner <user-id>` | Set dataset owner |
| `--tag <tag>` | Set tags (repeatable; replaces all existing) |

### `soda dataset delete <id>`
Delete a dataset from Soda Cloud.

### `soda dataset onboard <id>`
Guided setup for an existing dataset: monitoring, profiling, contracts.

| Flag | Description |
|---|---|
| `--monitoring` / `--no-monitoring` | Toggle monitoring |
| `--profiling` / `--no-profiling` | Toggle profiling |
| `--contracts ai\|skeleton\|none` | Contract generation |

### `soda dataset time-partition <id>`

| Flag | Description |
|---|---|
| `--column <col>` | Set partition column (omit to view current) |

### `soda dataset profiling <id>`

| Flag | Description |
|---|---|
| `--enable` / `--disable` | Toggle profiling |
| `--schedule <cron>` | Cron expression (e.g. '0 6 * * *') |
| `--timezone <tz>` | Timezone (default: UTC) |
| `--sampling-rows <n>` | Number of rows to sample |

No flags = view current profiling data and column stats.

### `soda dataset diagnostics <id>`

| Flag | Description |
|---|---|
| `--collect-results` / `--no-collect-results` | Store check results |
| `--collect-failed-rows` / `--no-collect-failed-rows` | Store failed rows |

No flags = view current settings. Requires diagnostics enabled on the datasource first.

### `soda dataset permissions list <id>`
List permissions (principal, type, role).

### `soda dataset permissions assign <id>`

| Flag | Description |
|---|---|
| `--role <role-id>` | Role ID (required) |
| `--user <email>` | User email |
| `--group <group-id>` | Group ID |

### `soda dataset permissions revoke <id>`
Same flags as assign.

---

## contract

### `soda contract list`
List all contracts in Soda Cloud. Returns: id, dataset, updated.

### `soda contract create`
Generate a contract YAML from a live dataset schema.

| Flag | Description |
|---|---|
| `--dataset <fqn>` | Dataset FQN: datasource/db/schema/table (required) |
| `--mode skeleton\|copilot` | Generation mode (default: skeleton) |
| `--output <file>` | Output file path |
| `--no-wait` | Return immediately (copilot mode only) |

### `soda contract push [file]`
Push a local contract YAML to Soda Cloud (upsert). Reads `dataset:` field from the file.

### `soda contract pull <identifier>`
Pull a contract by dataset qualified name (`datasource/db/schema/table`). Saves to `<table>.yml`.

### `soda contract diff [file]`
Show diff between local contract and Soda Cloud version.

### `soda contract lint [file]` (alias: validate)
Validate contract syntax locally (no network required).

### `soda contract verify [file-or-dir]`
Run contract checks against your data.

| Flag | Description |
|---|---|
| `--push` | Push results to Soda Cloud |
| `--datasource <file>` | Datasource config file override |
| `--runner` | Delegate execution to Soda Runner |
| `--set key=value` | Runtime variable overrides (repeatable) |

Exit codes: 0=pass, 1=checks failed, 2=error, 3=auth error.

---

## monitor

### `soda monitor list`

| Flag | Description |
|---|---|
| `--dataset <id>` | Dataset ID (required) |
| `--type column\|custom\|dataset` | Filter by monitor type |

### `soda monitor config <dataset-id>`
View or update dataset-level monitoring settings.

| Flag | Description |
|---|---|
| `--enable` / `--disable` | Toggle monitoring |
| `--schedule <cron>` | Cron expression |
| `--timezone <tz>` | Timezone (default: UTC) |

No flags = view current config.

### `soda monitor add`
Add a monitor to a dataset.

| Flag | Description |
|---|---|
| `--dataset <id>` | Dataset ID (required) |
| `--type column\|custom\|dataset` | Monitor type (required) |

**Column monitors** (`--type column`):

| Flag | Description |
|---|---|
| `--column <col>` | Column name (required) |
| `--metric <type>` | Metric (required): count, missing-pct, duplicate-pct, distinct-count, min, max, avg, sum, std-dev, variance, q1, median, q3, min-length, max-length, avg-length, freshness |
| `--group-by <col>` | Partition by column values (repeatable) |

**Custom SQL monitors** (`--type custom`):

| Flag | Description |
|---|---|
| `--name <name>` | Monitor name (required) |
| `--sql <query>` | SQL query (required unless --sql-file) |
| `--sql-file <path>` | Path to SQL file |
| `--result-metric <col>` | Result metric column (required) |
| `--column <col>` | Associated column (optional) |

**Dataset monitors** (`--type dataset`): Not yet available via API.

### `soda monitor update <id>`

| Flag | Description |
|---|---|
| `--dataset <id>` | Dataset ID (required) |
| `--enable` / `--disable` | Toggle monitor |
| `--sql <query>` | Update SQL (custom only) |
| `--name <name>` | Update name (custom only) |
| `--result-metric <col>` | Update result metric (custom only) |

### `soda monitor delete <id>`

| Flag | Description |
|---|---|
| `--dataset <id>` | Dataset ID (required) |

---

## results

### `soda results list`
List check results across datasets.

| Flag | Default | Description |
|---|---|---|
| `--dataset <id>` | | Filter by dataset ID |
| `--dataset-name <pattern>` | | Substring match on qualified name |
| `--status passing\|failing\|error` | all | Filter by status |
| `--type check\|monitor\|all` | check | Filter by type |
| `--limit <n>` | 10 | Max results |
| `--sort dataset\|name\|column\|status\|date` | date | Sort column |
| `--order asc\|desc` | desc | Sort order |
| `--from <date>` | | On or after (YYYY-MM-DD) |
| `--until <date>` | | On or before (YYYY-MM-DD) |

---

## job (alias: scan)

### `soda job logs <id>`

| Flag | Description |
|---|---|
| `--follow` | Stream logs as they arrive |

### `soda job list`
Not yet available in the API.

---

## runner

### `soda runner list`
List registered Soda Runners. Returns: id, name, status.

### `soda runner get <runner-id>`
Show details for a specific runner.

---

## iam

### `soda iam user list`
List users in the organization.

### `soda iam group list`
List groups with members.

### `soda iam group create`

| Flag | Description |
|---|---|
| `--name <name>` | Group name (required) |
| `--member <email>` | Initial member (repeatable) |

### `soda iam group update <id>`

| Flag | Description |
|---|---|
| `--name <name>` | New group name |
| `--add-member <email>` | Add member (repeatable) |
| `--remove-member <email>` | Remove member (repeatable) |

### `soda iam group delete <id>`
Delete a group.

### `soda iam role list`
List dataset-scoped roles.

| Flag | Description |
|---|---|
| `--scope global\|dataset` | Filter by scope |
