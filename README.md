# Soda CLI

A unified CLI for all things Soda — manage contracts, and interact with the Soda Cloud API from one binary.

---

## Status

**Walking skeleton.** All commands are wired up with the (WIP) structure, flags, and mock output. The goal is to validate UX — command naming, flag design, output formatting, wizard flows.

---

## Install

Requires [Go](https://go.dev/dl/) 1.21 or later.

```bash
git clone <repo>
cd soda-cli/go

go build -o soda .
```

Move the binary somewhere on your `PATH`:

```bash
mv soda /usr/local/bin/soda
```

Or run it directly from the build directory:

```bash
./soda --help
```

---

## Usage

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

  version                      print version info

  auth
    login   [--api-key] [--profile] [--host]
    logout  [--profile]
    status
    switch  <profile>

  init                         scaffold soda.yml + configs/ + contracts/

  dashboard                    org overview: datasets, results, incidents, jobs

  datasource  (alias: ds)
    create           [--type] [--name]
    onboard          --agent <name> --type <t>
    test-connection  [<name-or-file>]
    diagnostics      <id>
    list
    delete           <id>

  contract
    create    --dataset <fqn>  [--mode skeleton|copilot]
    lint      [<file>]           (alias: validate)
    push      [<file>]
    pull      <identifier>
    diff      [<file>]
    copilot   [<file>]  [<prompt>]  [--dataset <fqn>]
    verify    [<file-or-dir>|--dataset <identifier>]  [--datasource <file>]  [--push]  [--agent]  [--set key=val]
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
    list             [--filter] [--tag]
    update           <id>  [--owner] [--tag] [--description]
    delete           <id>
    time-partition   <id>  [--column]
      get            <id>
    profiling        <id>  [--enable|--disable] [--execution] [--schedule] [--strategy]
      get            <id>
      refresh        <id>
    diagnostics      <id>  [--schema] [--collect-results] [--collect-failed-rows]
      get            <id>
    permissions
      list    <id>
      assign  <id>  --role <id>  --user <email>|--group <id>
      revoke  <id>  --role <id>  --user <email>|--group <id>

  monitor
    list    [--dataset] [--type] [--status]
    config  <dataset-id>  [--enable|--disable] [--schedule] [--historical] [--historical-days]
      get   <dataset-id>
    add     --dataset <id>  --type dataset|column|group-by|custom  [...]
    update  <id>
    delete  <id>

  incident
    list    [--status reported|investigating|fixing|resolved]  [--dataset]
    get     <id>
    update  <id>  [--title] [--severity] [--description] [--assigned-to] [--status]

  notification
    rule
      list
      add     --name <n>  --source check|monitor|all  --alert warn-fail|fail-only|anomaly  --notify <recipient>
      update  <id>
      delete  <id>
    integration
      list
      add     slack|teams|webhook
      test    <id>
      delete  <id>

  iam
    role
      list    [--scope global|dataset]
      create  --name <n>  --scope global|dataset
      delete  <id>
      show    <id>
    user
      list
      invite  --email <email>
      remove  <user-id>
      assign  <user-id>  --role <role-id>
      revoke  <user-id>  --role <role-id>
    group
      list
      create  --name <n>  [--member <email>]
      update  <id>  [--name] [--add-member] [--remove-member]
      delete  <id>
      assign  <group-id>  --role <role-id>
      revoke  <group-id>  --role <role-id>
    service-account
      list
      create  --name <n>  --email <email>
      delete  <id>

  agent
    list

  secret
    create  --name <n>  --value <v>
    update  <id>  --value <v>
    delete  <id>

  completion  bash|zsh|fish
```

Exit codes: `0` pass · `1` checks failed · `2` error · `3` auth error

---

## Design principles

- **Noun → verb** — every command follows `soda <resource> <action>`
- **Auto-detect output** — tables when TTY, JSON when piped; override with `--output`
- **`--no-interactive` everywhere** — safe to run in CI and from AI agents
- **One auth system** — `~/.soda/credentials` for both local and cloud API calls
- **Config precedence** — `--flags` → env vars → `./soda.yml` → `~/.soda/config.yml`
