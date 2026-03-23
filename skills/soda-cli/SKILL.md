---
name: soda-cli
description: How to use the Soda CLI for data quality management — authentication, datasources, datasets, contracts, monitors, results, permissions, and CI/CD integration. Use when working with Soda, data quality, or the sodacli command.
allowed-tools: Read, Bash(sodacli *), Bash(cat *), Glob, Grep
---

# Soda CLI Guide

The Soda CLI (`sodacli`) manages data quality from the command line. It follows `sodacli <resource> <action>` consistently.

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
sodacli auth login

# Non-interactive login
sodacli auth login --host cloud.soda.io --api-key-id <id> --api-key-secret <secret>

# US region
sodacli auth login --host cloud.us.soda.io --api-key-id <id> --api-key-secret <secret>

# Check connection health
sodacli auth status

# Use a named profile
sodacli auth login --profile production --host cloud.soda.io --api-key-id <id> --api-key-secret <secret>
sodacli auth switch production

# All commands accept --profile to override active profile
sodacli dataset list --profile production
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
sodacli datasource test-connection warehouse.yml

# Full onboard: create datasource + discover datasets + enable monitoring + profiling + generate contracts
sodacli datasource onboard warehouse.yml --monitoring --profiling --contracts skeleton

# Or step by step:
sodacli datasource create warehouse.yml                    # creates datasource, returns ID
sodacli datasource onboard <datasource-id> --monitoring    # onboard discovered datasets
```

### 2. Explore existing data

```bash
# List datasources
sodacli datasource list

# List datasets (default: 10 rows)
sodacli dataset list
sodacli dataset list --limit 50
sodacli dataset list --datasource snowflakeproduct
sodacli dataset list --status onboarded --filter orders
sodacli dataset list --from 2026-03-01 --until 2026-03-13

# Get dataset details
sodacli dataset get <dataset-id>

# View profiling data
sodacli dataset profiling <dataset-id>
```

### 3. Create, edit, and push a contract

```bash
# Generate a contract from live schema
sodacli contract create --dataset datasource/db/schema/table --mode skeleton --output my_table.yml

# Or pull an existing contract from cloud
sodacli contract pull datasource/db/schema/table

# Edit the contract YAML locally, then push back
sodacli contract push my_table.yml

# Check diff before pushing
sodacli contract diff my_table.yml

# Validate syntax locally (no network)
sodacli contract lint my_table.yml
```

### 4. Run contract checks

```bash
# Verify locally
sodacli contract verify my_table.yml

# Verify and push results to Soda Cloud
sodacli contract verify my_table.yml --push

# Override runtime variables
sodacli contract verify my_table.yml --set date=2026-03-13 --push
```

### 5. Add monitors to a dataset

```bash
# List existing monitors
sodacli monitor list --dataset <dataset-id>

# View monitoring config (enabled, schedule)
sodacli monitor config <dataset-id>

# Enable monitoring with a schedule
sodacli monitor config <dataset-id> --enable --schedule "0 6 * * *" --timezone "UTC"

# Add a column monitor
sodacli monitor add --dataset <dataset-id> --type column --column order_amount --metric avg

# Add a column monitor with group-by
sodacli monitor add --dataset <dataset-id> --type column --column revenue --metric sum --group-by region --group-by product_line

# Add a custom SQL monitor
sodacli monitor add --dataset <dataset-id> --type custom --name "duplicate orders" \
  --sql "SELECT order_id, COUNT(*) as cnt FROM orders GROUP BY order_id HAVING COUNT(*) > 1" \
  --result-metric cnt

# Column monitor metrics: count, missing-pct, duplicate-pct, distinct-count,
#   min, max, avg, sum, std-dev, variance, q1, median, q3,
#   min-length, max-length, avg-length, freshness

# Delete a monitor
sodacli monitor delete <monitor-id> --dataset <dataset-id>
```

### 6. View check results

```bash
# Recent results (default: 10, sorted by date desc)
sodacli results list

# Filter by status
sodacli results list --status failing

# Filter by dataset
sodacli results list --dataset <dataset-id>
sodacli results list --dataset-name orders

# Date range
sodacli results list --from 2026-03-01 --until 2026-03-13

# Sort and paginate
sodacli results list --limit 50 --sort name --order asc
```

### 7. Set up dataset permissions

```bash
# List available roles
sodacli iam role list

# List users to find user IDs
sodacli iam user list

# List current permissions
sodacli dataset permissions list <dataset-id>

# Grant a role
sodacli dataset permissions assign <dataset-id> --role <role-id> --user <user-email>

# Grant to a group
sodacli dataset permissions assign <dataset-id> --role <role-id> --group <group-id>

# Revoke
sodacli dataset permissions revoke <dataset-id> --role <role-id> --user <user-email>
```

### 8. Configure profiling and diagnostics

```bash
# Enable profiling
sodacli dataset profiling <dataset-id> --enable --schedule "0 6 * * *" --sampling-rows 1000000

# View profiling data
sodacli dataset profiling <dataset-id>

# Set time-partition column
sodacli dataset time-partition <dataset-id> --column created_at

# Configure diagnostics warehouse
sodacli dataset diagnostics <dataset-id> --collect-results --collect-failed-rows
```

### 9. Manage users and groups

```bash
# List users
sodacli iam user list

# Create a group with initial members
sodacli iam group create --name "Data Engineers" --member alice@example.com --member bob@example.com

# Add/remove members
sodacli iam group update <group-id> --add-member carol@example.com
sodacli iam group update <group-id> --remove-member bob@example.com

# List groups
sodacli iam group list
```

### 10. View job logs

```bash
sodacli job logs <scan-id>
sodacli job logs <scan-id> --follow    # stream live
```

## CI/CD Pattern

For non-interactive pipelines, always use `--no-interactive` and `--output json`:

```bash
sodacli auth login --host cloud.soda.io \
  --api-key-id "$SODA_API_KEY_ID" \
  --api-key-secret "$SODA_API_KEY_SECRET" \
  --no-interactive

sodacli contract verify contracts/ --push --no-interactive --output json

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
