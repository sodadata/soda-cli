---
name: soda-cli
description: How to use the Soda CLI for data quality management — authentication, datasources, datasets, contracts, monitors, results, permissions, and CI/CD integration. Use when working with Soda, data quality, or the soda command.
allowed-tools: Read, Bash(soda *), Bash(cat *), Glob, Grep
---

# Soda CLI Guide

The Soda CLI (`soda`) manages data quality from the command line. It follows `soda <resource> <action>` consistently.

## Key Concepts

- **Datasource** — a database connection (Snowflake, Postgres, BigQuery, etc.)
- **Dataset** — a table registered in Soda Cloud, within a datasource
- **Contract** — a YAML file defining data quality checks for a dataset
- **Monitor** — ML anomaly detection on a dataset column (distinct from contract checks)
- **Runner** — an agent that executes checks against your database
- **Results** — check pass/fail signals across all jobs

## Output Behavior

- TTY: human-readable tables. Piped: JSON.
- Override with `--output json|table|csv`
- All commands support `--no-interactive` for CI/CD and AI agents
- Exit codes: `0`=pass, `1`=checks failed, `2`=error, `3`=auth error

## Authentication

```bash
# Interactive login (prompts for host, API key ID, API key secret)
soda auth login

# Non-interactive login
soda auth login --host cloud.soda.io --api-key-id <id> --api-key-secret <secret>

# US region
soda auth login --host cloud.us.soda.io --api-key-id <id> --api-key-secret <secret>

# Check connection health
soda auth status

# Use a named profile
soda auth login --profile production --host cloud.soda.io --api-key-id <id> --api-key-secret <secret>
soda auth switch production

# All commands accept --profile to override active profile
soda dataset list --profile production
```

Credentials are stored in `~/.soda/credentials`. Generate API keys at https://docs.soda.io/reference/generate-api-keys

## Common Workflows

### 1. Onboard a new datasource end-to-end

```bash
# Create a datasource config file
cat > warehouse.yml <<EOF
type: snowflake
name: my_warehouse
connection:
  host: account.snowflakecomputing.com
  database: ANALYTICS
  schema: PUBLIC
  user: soda_user
  password: secret
  role: SODA_ROLE
  warehouse: COMPUTE_WH
EOF

# Test the connection first
soda datasource test-connection warehouse.yml

# Full onboard: create datasource + discover datasets + enable monitoring + profiling + generate contracts
soda datasource onboard warehouse.yml --monitoring --profiling --contracts skeleton

# Or step by step:
soda datasource create warehouse.yml                    # creates datasource, returns ID
soda datasource onboard <datasource-id> --monitoring    # onboard discovered datasets
```

### 2. Explore existing data

```bash
# List datasources
soda datasource list

# List datasets (default: 10 rows)
soda dataset list
soda dataset list --limit 50
soda dataset list --datasource snowflakeproduct
soda dataset list --status onboarded --filter orders
soda dataset list --from 2026-03-01 --until 2026-03-13

# Get dataset details
soda dataset get <dataset-id>

# View profiling data
soda dataset profiling <dataset-id>
```

### 3. Create, edit, and push a contract

```bash
# Generate a contract from live schema
soda contract create --dataset datasource/db/schema/table --mode skeleton --output my_table.yml

# Or pull an existing contract from cloud
soda contract pull datasource/db/schema/table

# Edit the contract YAML locally, then push back
soda contract push my_table.yml

# Check diff before pushing
soda contract diff my_table.yml

# Validate syntax locally (no network)
soda contract lint my_table.yml
```

### 4. Run contract checks

```bash
# Verify locally
soda contract verify my_table.yml

# Verify and push results to Soda Cloud
soda contract verify my_table.yml --push

# Override runtime variables
soda contract verify my_table.yml --set date=2026-03-13 --push
```

### 5. Add monitors to a dataset

```bash
# List existing monitors
soda monitor list --dataset <dataset-id>

# View monitoring config (enabled, schedule)
soda monitor config <dataset-id>

# Enable monitoring with a schedule
soda monitor config <dataset-id> --enable --schedule "0 6 * * *" --timezone "UTC"

# Add a column monitor
soda monitor add --dataset <dataset-id> --type column --column order_amount --metric avg

# Add a column monitor with group-by
soda monitor add --dataset <dataset-id> --type column --column revenue --metric sum --group-by region --group-by product_line

# Add a custom SQL monitor
soda monitor add --dataset <dataset-id> --type custom --name "duplicate orders" \
  --sql "SELECT order_id, COUNT(*) as cnt FROM orders GROUP BY order_id HAVING COUNT(*) > 1" \
  --result-metric cnt

# Column monitor metrics: count, missing-pct, duplicate-pct, distinct-count,
#   min, max, avg, sum, std-dev, variance, q1, median, q3,
#   min-length, max-length, avg-length, freshness

# Delete a monitor
soda monitor delete <monitor-id> --dataset <dataset-id>
```

### 6. View check results

```bash
# Recent results (default: 10, sorted by date desc)
soda results list

# Filter by status
soda results list --status failing

# Filter by dataset
soda results list --dataset <dataset-id>
soda results list --dataset-name orders

# Date range
soda results list --from 2026-03-01 --until 2026-03-13

# Sort and paginate
soda results list --limit 50 --sort name --order asc
```

### 7. Set up dataset permissions

```bash
# List available roles
soda iam role list

# List users to find user IDs
soda iam user list

# List current permissions
soda dataset permissions list <dataset-id>

# Grant a role
soda dataset permissions assign <dataset-id> --role <role-id> --user <user-email>

# Grant to a group
soda dataset permissions assign <dataset-id> --role <role-id> --group <group-id>

# Revoke
soda dataset permissions revoke <dataset-id> --role <role-id> --user <user-email>
```

### 8. Configure profiling and diagnostics

```bash
# Enable profiling
soda dataset profiling <dataset-id> --enable --schedule "0 6 * * *" --sampling-rows 1000000

# View profiling data
soda dataset profiling <dataset-id>

# Set time-partition column
soda dataset time-partition <dataset-id> --column created_at

# Configure diagnostics warehouse
soda dataset diagnostics <dataset-id> --collect-results --collect-failed-rows
```

### 9. Manage users and groups

```bash
# List users
soda iam user list

# Create a group with initial members
soda iam group create --name "Data Engineers" --member alice@example.com --member bob@example.com

# Add/remove members
soda iam group update <group-id> --add-member carol@example.com
soda iam group update <group-id> --remove-member bob@example.com

# List groups
soda iam group list
```

### 10. View job logs

```bash
soda job logs <scan-id>
soda job logs <scan-id> --follow    # stream live
```

## CI/CD Pattern

For non-interactive pipelines, always use `--no-interactive` and `--output json`:

```bash
soda auth login --host cloud.soda.io \
  --api-key-id "$SODA_API_KEY_ID" \
  --api-key-secret "$SODA_API_KEY_SECRET" \
  --no-interactive

soda contract verify contracts/ --push --no-interactive --output json

# Check exit code
# 0 = all checks passed
# 1 = one or more checks failed
# 2 = execution error
# 3 = authentication error
```

## Datasource Config File Format

```yaml
type: snowflake           # snowflake, postgres, bigquery, mysql, redshift, etc.
name: my_warehouse        # must match ^[a-zA-Z_][0-9a-zA-Z_]+ (no hyphens)
connection:
  host: account.snowflakecomputing.com
  database: ANALYTICS
  schema: PUBLIC
  user: soda_user
  password: secret
  role: SODA_ROLE
  warehouse: COMPUTE_WH
```

## Full Command Reference

For detailed flags and usage of every command, see [command-reference.md](command-reference.md).
