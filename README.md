# Soda CLI

A unified CLI for all things Soda — manage contracts, and interact with the Soda Cloud API from one binary.

---

## Status

**Walking skeleton.** All commands are wired up with the (WIP) structure, flags, and mock output. The goal is to validate UX — command naming, flag design, output formatting, wizard flows.

---

## Install

Uses [uv](https://github.com/astral-sh/uv) for fast installs. (TBD is actual CLI will be build in Python or any other language).

```bash
git clone <repo>
cd soda-cli

uv venv --python 3.11
source .venv/bin/activate

uv pip install -e .
```

The `soda` binary is now available:

```bash
soda --help
soda contract --help
soda contract verify contracts/orders.yml
```

---

## Command tree

```
soda
  Global flags (all commands):
    --output table|json|csv    auto-detects TTY; override per-command
    --profile <name>
    --no-color
    --quiet / --verbose
    --no-interactive           for CI/CD and AI agents
    --version

  auth
    login   [--api-key] [--profile] [--host]
    logout  [--profile]
    status
    switch  <profile>

  init                         scaffold soda.yml + configs/ + contracts/

  dashboard                    org overview: datasets, results, incidents, jobs

  datasource  (alias: ds)
    create      [--type] [--name]
    onboard     --agent <name> --type <t>
    diagnostics <id>
    test        [<name-or-file>]
    list

  contract
    create    --dataset <fqn>  [--mode skeleton|copilot]
    lint      [<file>]
    push      [<file>]
    pull      --dataset <fqn>
    diff      [<file>]  --dataset <fqn>
    copilot   [<file>]  [<prompt>]  [--dataset <fqn>]
    verify    [<file-or-dir>]  [--push]  [--agent]  [--set key=val]
    proposal
      list    [--status open|done|all]
      pull    <id>  [--revision <n>]
      push    <id>  [<file>]  [--message]
      close   <id>  [--status done|wontdo]

  job  (alias: scan)
    list    [--datasource] [--dataset] [--type] [--status]
    logs    <id>  [--follow]

  results
    list    [--datasource] [--dataset] [--type] [--status] [--from] [--to]

  dataset
    list         [--filter] [--tag]
    update       <id>  [--name] [--tag] [--attr key=val]
    delete       <id>
    profiling    show   <id>
    profiling    refresh <id>
    diagnostics  <id>
    permissions
      list  <id>
      set   <id>  --role <id>  --user <email>|--group <id>
    monitor
      list    <dataset-id>
      add     <dataset-id>  [--column] [--metric]
      update  <dataset-id> <monitor-id>
      delete  <dataset-id> <monitor-id>

  incident
    list    [--status open|closed|all]  [--dataset]
    get     <id>
    update  <id>  [--status]  [--note]

  notification
    list    [--channel] [--dataset]
    add     --channel <id>  --trigger <event>  [--dataset]
    update  <id>
    delete  <id>
    channel
      list
      add     slack|teams|webhook
      delete  <id>
      test    <id>

  role
    list    [--scope global|dataset]
    create  --name <n>  --scope global|dataset
    delete  <id>
    show    <id>

  users
    list
    assign  <user-id>  --role <role-id>
    revoke  <user-id>  --role <role-id>
    group
      list
      create  --name <n>  [--members <emails>]
      update  <id>
      delete  <id>
      assign  <group-id>  --role <role-id>
      revoke  <group-id>  --role <role-id>

  completion  bash|zsh|fish
```

Exit codes: `0` pass · `1` checks failed · `2` error · `3` auth error

---

## Design principles

- **Noun → verb** — every command follows `soda <resource> <action>`
- **Auto-detect output** — Rich tables when TTY, JSON when piped; override with `--output`
- **`--no-interactive` everywhere** — safe to run in CI and from AI agents
- **One auth system** — `~/.soda/credentials` for both local and cloud API calls
- **Config precedence** — `--flags` → env vars → `./soda.yml` → `~/.soda/config.yml`
