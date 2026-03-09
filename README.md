# Soda CLI

A unified CLI for all things Soda — manage contracts, and interact with the Soda Cloud API from one binary.

---

## Status

**Walking skeleton.** All commands are wired up with the (WIP) structure, flags, and mock output. The goal is to validate UX — command naming, flag design, output formatting, wizard flows.

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

The full command tree: all subcommands, flags, and exit codes is documented in [`command_tree.txt`](command_tree.txt). Feel free to review and leave any comments.

---

## Design principles

- **Noun → verb** — every command follows `soda <resource> <action>`
- **Auto-detect output** — tables when TTY, JSON when piped; override with `--output`
- **`--no-interactive` everywhere** — safe to run in CI and from AI agents
- **One auth system** — `~/.soda/credentials` for both local and cloud API calls
- **Config precedence** — `--flags` → env vars → `./soda.yml` → `~/.soda/config.yml`
