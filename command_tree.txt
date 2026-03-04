Full Command Tree (WIP)

  soda

  ├── auth
  │   ├── login  [--profile <name>]          # Wizard: enters API key, tests, saves
  │   ├── logout [--profile <name>]
  │   ├── status                             # Shows active profile, org, expiry
  │   └── switch <profile>                   # Swap active profile

  ├── init                                   # Scaffolds soda.yml + configs/ + contracts/

  ├── datasource  (alias: ds)
  │   ├── create  [--type pg|sf|bq|duck...]  # Interactive wizard or --type flag
  │   ├── test    [<name-or-file>]           # Test connectivity
  │   └── list                               # Show all configured datasources

  ├── contract
  │   ├── create  --dataset <ds/db/schema/table>  # Bootstrap YAML from live schema
  │   ├── validate [<file>]                  # Dry-run syntax check (no network)
  │   ├── publish  [<file>]                  # Push to Soda Cloud
  │   ├── fetch   --dataset <fqn>            # Pull from cloud → local file
  │   ├── diff    [<file>] --dataset <fqn>  # Diff local vs cloud version
  │   └── verify  [<file>]                  # Execute checks
  │       --agent                            # Delegate to Soda Agent
  │       --publish                          # Push results to cloud
  │       --set   key=value                  # Runtime variable overrides
  │       --check <dot.path>                 # Run specific checks only
  │       --watch                            # Re-run on file changes

  ├── scan
  │   ├── get   <id>                         # Status + summary
  │   └── logs  <id>  [--follow]             # Stream execution logs

  ├── checks
  │   ├── list   [--dataset <id>] [--status passing|failing|all]
  │   └── delete <id>

  ├── dataset                                # Cloud management (wraps API)
  │   ├── list   [--filter <query>]  [--tag <tag>]
  │   ├── update <id>  [--name] [--tag] [--attr key=value]
  │   ├── delete <id>
  │   ├── profiling    <id>                  # Show profiling insights
  │   ├── permissions
  │   │   ├── list   <id>
  │   │   └── set    <id>  --role <role> --user <email>
  │   └── monitor
  │       ├── list   <dataset-id>
  │       ├── add    <dataset-id>            # Interactive or flags
  │       └── update <dataset-id> <monitor-id>

  ├── incident
  │   ├── list   [--status open|closed|all]  [--dataset <id>]
  │   ├── get    <id>
  │   └── update <id>  [--status] [--note <text>]

  ├── request                               # Contract collaboration
  │   ├── list   [--status open|done|all]
  │   ├── fetch  <id>  [--proposal <n>]     # Download → local file
  │   ├── push   <id>  [<file>]  [-m <msg>] # Upload proposal
  │   └── close  <id>  [--status done|wontdo]

  └── users
      ├── list
      └── group
          ├── list
          ├── create  --name <n>  --members <emails...>
          ├── update  <id>
          └── delete  <id>


UX Ideas

  1. No flag soup — smart defaults

  Current (soda-core):
  soda contract verify -ds ./configs/postgres.yml -c ./contracts/orders.yml -sc ./soda-cloud.yml --publish

  Proposed (with soda.yml project config):
  soda contract verify orders.yml --publish

  2. soda run — pipeline command

  A power command that orchestrates verify → publish → wait for scan completion → report:

  soda run orders.yml                   # verify + publish, stream results
  soda run orders.yml --agent           # via Soda Agent
  soda run ./contracts/ --parallel      # all contracts in dir
  soda run orders.yml --output json     # for CI/CD pipelines

  Exit codes: 0 = all checks pass, 1 = checks failed, 2 = error, 3 = auth error

  3. soda status — instant dashboard

  $ soda status

    Organization   acme-corp
    Profile        production

    Datasets       142 total    3 failing
    Checks          87 passing   12 failing
    Incidents        2 open

    Last scan      orders_daily   2m ago   ✓ passed
                   users_weekly   1h ago   ✗ 3 failures → incident #47

  4. Consistent output flags (all commands)

  --output table     # default when TTY (human-friendly)
  --output json      # machine-readable (default when piped)
  --output csv       # for spreadsheet export
  --quiet            # suppress all output except errors
  --verbose          # debug-level detail
  --no-color         # for logs/CI
  --no-interactive   # fail instead of prompting (agent mode)

  5. First-class shell completion

  soda completion bash   >> ~/.bashrc
  soda completion zsh    >> ~/.zshrc
  soda completion fish   >> ~/.config/fish/completions/soda.fish

  Tab-completes dataset names, contract files, profile names by calling the API.


Auth & Config Architecture

  ~/.soda/
    credentials          # API keys per profile (never committed)
    config.yml           # Defaults, cloud host, profile selection

  ./soda.yml             # Project-level config (committed)
    └── datasources, contracts dir, default profile override