# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

A new unified `soda` CLI that replaces `soda-core` and wraps the Soda Cloud public API into a single binary. The goal is to unify:
- Local data quality execution (previously `soda-core`)
- Soda Cloud API management (checks, datasets, incidents, users, etc.)
- AI-powered contract generation (Copilot/Autopilot)

Target users: data engineers running it manually, and AI agents running it programmatically in CI/CD pipelines.

## Design principles — must be respected in all implementation decisions

1. **Noun → Verb structure** — all commands follow `soda <resource> <action>` consistently
2. **Context-aware defaults** — `soda.yml` in project root eliminates repetitive flags; config precedence: `--flags` > env vars > `./soda.yml` > `~/.soda/config.yml`
3. **Human output when TTY, JSON when piped** — auto-detect, overridable with `--output`
4. **One auth system** — `~/.soda/credentials` serves both local CLI and API calls
5. **AI-agent first** — `--no-interactive` must work on every command; structured exit codes everywhere; `--output json` always available

## Auth & config architecture

```
~/.soda/
  credentials     # API keys per profile (never committed)
  config.yml      # defaults, cloud host, profile selection

./soda.yml        # project-level config (committed)
```

`soda.yml` structure:
```yaml
version: 1
profile: production
datasources:
  default: warehouse
  warehouse:
    type: snowflake
    config: ./configs/snowflake.yml
contracts:
  directory: ./contracts
cloud:
  organization: acme-corp
```

## Key domain concepts

- **Datasource** — a database connection. Can be a local YAML config (`datasource create`) or a cloud-registered connection via Runner (`datasource onboard`). One datasource contains many datasets.
- **Dataset** — a specific table registered in Soda Cloud, within a datasource.
- **Contract** — a YAML file defining data quality expectations for a dataset. Lives locally, published to Soda Cloud via `contract push`.
- **Job** — an execution run (contract verify or monitor cycle). Produces results. Alias: `scan`.
- **Results** — individual data quality signals (check pass/fail, monitor alerts) queryable across all jobs.
- **Diagnostics warehouse** — stores failed rows and data quality issues. Configured at datasource level, with per-dataset overrides.
- **Monitor** — ML anomaly detection check on a dataset column. Fires alerts. Distinct from contract checks (which are manual/rule-based).

## Global flags (all commands must support these)

```
--output table|json|csv   # table=default when TTY, json=default when piped
--profile <name>
--version
--no-color
--quiet
--verbose
--no-interactive          # never prompt; fail with clear error if input missing
```

Exit codes: `0`=pass, `1`=checks failed, `2`=error, `3`=auth error

## Command tree

See `command_tree.txt` for the current authoritative tree. Key structural decisions:

- `contract verify` is the primary execution command (no separate `run` command)
- `contract verify --push` sends results to cloud; `contract push` sends the contract definition
- `contract pull/push` are the cloud sync pair (not fetch/publish)
- `job` (alias: `scan`) is read-only — lists execution history and streams logs
- `results` is the unified view of all signals (checks + monitor alerts) across jobs
- `notification rule` for alert rules; `notification integration` for connections (Slack, Teams, webhook)
- `iam` is the single namespace for identity/access: `iam user`, `iam group`, `iam role`, `iam service-account`
- `datasource create <config-file>` registers a datasource from a YAML connection config
- `monitor` is top-level (not nested under `dataset`); `monitor list` requires `--dataset`

## Implementation

**Language**: Go. Source lives in `go/`. Builds to a single static binary (`go build -o soda .`).

**Stack**: `cobra` (commands) + `lipgloss` (styling) + `huh` (interactive wizards) + `tablewriter` (tables)

**Layout**:
- `go/cmd/` — flat package, one file per resource (root.go, auth.go, dataset.go, contract.go, etc.)
- `go/internal/api/` — API client methods per resource (client.go, datasets.go, contracts.go, agents.go, datasources.go)
- `go/internal/ctx/` — GlobalCtx struct
- `go/internal/output/` — rendering helpers (Render, RenderOne, PrintSuccess, ExitError)

**Auth**: Basic Auth (`api_key_id:api_key_secret`) against Soda Cloud API. Credentials stored in `~/.soda/credentials`.

**API hosts**: `cloud.soda.io` (EU), `cloud.us.soda.io` (US), `dev.sodadata.io` (dev)

## Pending decisions / known TODOs

- **Binary name**: new `soda` replaces `soda-core`'s `soda` binary. Migration strategy TBD.
- **`results`**: temp name — needs a better name that covers both checks and monitor alerts without the "manual check" connotation.
- **`datasource diagnostics` / `dataset diagnostics`**: what exactly is configurable at each level needs spec.
- **`contract verify`**: help text needs to make clear this can be cloud-connected (not obvious from the name alone).

## Known API gaps (as of 2026-03-18)

These Soda Cloud public API endpoints do not exist yet, blocking CLI implementation:
- `monitor config` write — no POST /metricMonitoring (read-only)
- `monitor add --type dataset` — dataset monitors exist by default but can't be enabled via API
- `incident *` — endpoint returns SPA HTML (CLI wired, waiting on API team)
- `job list` — no list endpoint (only `job logs <id>`)
- `job cancel` — POST /scans/{id}/actions/cancel returns 404 (CLI wired, waiting on API team)
- `datasource update` — PATCH /datasources/{id} returns 404 (CLI wired, waiting on API team)
- `datasource test-connection` — POST /datasources/actions/testConnection returns HTML (CLI wired, waiting on API team)
- `notification *`, `secret *` — no endpoints
- `contract proposal *` — endpoint returns SPA HTML
- `datasource get <name>` / `runner get <name>` — API only supports lookup by ID, not name (requested from Michael)

Recently unblocked:
- `contract verify` — POST /contracts/{id}/verify + GET /scans/{id} polling works end-to-end (2026-03-18)
- `runner create` — POST /runners works (returns apiKeyId + apiKeySecret) (2026-03-18)
- `runner delete` — DELETE /runners/{id} works (2026-03-18)
- `datasource get` — GET /datasources/{id} works (2026-03-18)
- `datasource list` — GET /api/v1/datasources now works (paginated, returns id/name/label/type/timestamps)
- `discoveredDatasets` — GET /api/v1/discoveredDatasets (with ?datasourceId= filter) now works
- `onboardDatasets` — POST /api/v1/datasources/{id}/onboardDatasets now works (accepts discoveredDatasetIds)
- `datasource onboard` — now wired end-to-end (create → discover → select → onboard → monitoring → contracts)
