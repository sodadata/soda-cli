# Soda CLI

A unified CLI for all things Soda — manage contracts, and interact with the Soda Cloud API from one binary.

---

## Status & how to help

This is a **walking skeleton** — all commands are wired up with mock output so you can feel the UX, but no real API calls happen yet. There is significant API work ahead before this is fully functional, and we are prioritizing in this order:

1. Auth & onboarding
2. Monitors
3. Contracts

**We'd love your input.** If you want to help shape this CLI:

- **Review the command tree** — read [`command_tree.txt`](command_tree.txt) and leave comments on naming, structure, or anything that feels off
- **Try the skeleton** — install it (see below), run some commands, and tell us if the UX feels right

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
