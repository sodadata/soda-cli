# Soda CLI

A unified CLI for all things Soda — manage contracts, monitors, datasets, and teams from one binary connected to Soda Cloud.

---

## Status

This is a working Go CLI that connects to live Soda Cloud instances. Authentication, contracts, datasets, monitors, and IAM commands all make real API calls.

Implementation status for every command is tracked in [`command_tree.txt`](command_tree.txt) using this legend:

```
✅  implemented — real API call wired up
🔌  API endpoint exists, not yet wired in CLI
🏠  local operation, no API needed
❌  no public API endpoint yet
```

---

## Install

Requires [Go](https://go.dev/dl/) 1.21 or later.

```bash
git clone https://github.com/sodadata/soda-cli.git
cd soda-cli/go

go build -o soda .
```

Move the binary somewhere on your `PATH`:

```bash
mv soda /usr/local/bin/soda
```

---

## Quickstart

```bash
# Authenticate (wizard) — or pass flags for CI
soda auth login
soda auth status

# Browse your datasets
soda dataset list
soda dataset list --datasource snowflakeproduct

# Pull a contract, edit it locally, push it back
soda contract pull "snowflakeproduct/db/schema/orders"
soda contract diff orders.yml
soda contract push orders.yml

# Explore monitors and IAM
soda monitor list --dataset <id>
soda iam user list
soda iam group list
```

---

## Design principles

- **Noun → verb** — every command follows `soda <resource> <action>`
- **Auto-detect output** — tables when TTY, JSON when piped; override with `--output`
- **`--no-interactive` everywhere** — safe to run in CI and from AI agents
- **One auth system** — `~/.soda/credentials` for both local and cloud API calls
- **Config precedence** — `--flags` → env vars → `./soda.yml` → `~/.soda/config.yml`

---

## Command reference

Full command tree with all subcommands, flags, and implementation status: [`command_tree.txt`](command_tree.txt).
