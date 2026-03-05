import json
import sys
import time
from typing import Optional

import typer
from rich.prompt import Prompt

from soda.context import GlobalContext
from soda.output import render, render_one, print_success, print_error, get_console
from soda.mock import PROPOSALS

app = typer.Typer(help="Manage and verify data contracts.", no_args_is_help=True)

# Sub-app for contract proposal
proposal_app = typer.Typer(help="Manage contract change proposals (PR flow).", no_args_is_help=True)
app.add_typer(proposal_app, name="proposal")


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("create")
def create(
    ctx: typer.Context,
    dataset: str = typer.Option(..., "--dataset", help="Dataset FQN (datasource/db/schema/table)"),
    mode: str = typer.Option("skeleton", "--mode", help="Generation mode: skeleton|copilot"),
    output_file: Optional[str] = typer.Option(None, "--output", "-o", help="Output file path"),
):
    """Bootstrap a contract YAML from a live dataset schema.

    [dim]--mode skeleton[/dim]  Schema-only bootstrap. No AI. Always works.
    [dim]--mode copilot[/dim]   AI-generated contract with checks. Requires license.
    """
    gctx = _get_ctx(ctx)
    console = get_console(gctx)

    parts = dataset.split("/")
    table = parts[-1] if parts else dataset
    out = output_file or f"contracts/{table}.yml"

    if mode == "copilot":
        console.print(f"[dim]Generating AI contract for [bold]{dataset}[/bold]...[/dim]")
        time.sleep(0.3)  # simulate
    else:
        console.print(f"[dim]Reading schema for [bold]{dataset}[/bold]...[/dim]")

    console.print(f"  [dim]create[/dim]  {out}")
    print_success(f"Contract written to [bold]{out}[/bold]", gctx)


@app.command("lint")
def lint(
    ctx: typer.Context,
    file: Optional[str] = typer.Argument(None, help="Contract YAML file (default: auto-discover)"),
):
    """Syntax-check a contract YAML. No network required. (alias: validate)"""
    gctx = _get_ctx(ctx)
    target = file or "contracts/*.yml"
    get_console(gctx).print(f"[dim]Linting [bold]{target}[/bold]...[/dim]")
    print_success("Contract syntax is valid.", gctx)


@app.command("validate", hidden=True)
def validate(
    ctx: typer.Context,
    file: Optional[str] = typer.Argument(None, help="Contract YAML file"),
):
    """Alias for lint."""
    lint(ctx, file)


@app.command("push")
def push(
    ctx: typer.Context,
    file: Optional[str] = typer.Argument(None, help="Contract YAML file to push"),
):
    """Push contract definition to Soda Cloud."""
    gctx = _get_ctx(ctx)
    target = file or "contracts/*.yml"
    get_console(gctx).print(f"[dim]Pushing [bold]{target}[/bold] to Soda Cloud...[/dim]")
    print_success("Contract definition pushed to Soda Cloud.", gctx)


@app.command("pull")
def pull(
    ctx: typer.Context,
    dataset: str = typer.Option(..., "--dataset", help="Dataset FQN to pull contract for"),
    output_file: Optional[str] = typer.Option(None, "--output", "-o", help="Output file path"),
):
    """Pull contract definition from Soda Cloud to a local file."""
    gctx = _get_ctx(ctx)
    parts = dataset.split("/")
    table = parts[-1] if parts else dataset
    out = output_file or f"contracts/{table}.yml"
    get_console(gctx).print(f"[dim]Pulling contract for [bold]{dataset}[/bold]...[/dim]")
    get_console(gctx).print(f"  [dim]write[/dim]  [cyan]{out}[/cyan]")
    print_success(f"Contract saved to [bold]{out}[/bold]", gctx)


@app.command("diff")
def diff(
    ctx: typer.Context,
    file: Optional[str] = typer.Argument(None, help="Local contract file"),
    dataset: str = typer.Option(..., "--dataset", help="Dataset FQN to diff against"),
):
    """Show diff between local contract and Soda Cloud version."""
    gctx = _get_ctx(ctx)
    console = get_console(gctx)
    console.print(f"[dim]Fetching cloud contract for [bold]{dataset}[/bold]...[/dim]")
    console.print()
    diff_text = """\
--- cloud/orders.yml
+++ local/contracts/orders.yml
@@ -12,6 +12,9 @@
   - row_count > 0
   - no_nulls:
       columns: [order_id]
+  - freshness:
+      column: created_at
+      max_age: 1d
   - valid_values:
       column: status
"""
    from rich.syntax import Syntax
    console.print(Syntax(diff_text, "diff", theme="monokai"))


@app.command("copilot")
def copilot(
    ctx: typer.Context,
    file: Optional[str] = typer.Argument(None, help="Existing contract file to improve"),
    prompt: Optional[str] = typer.Argument(None, help="What to generate or improve"),
    dataset: Optional[str] = typer.Option(None, "--dataset", help="Dataset FQN for generation"),
    output_file: Optional[str] = typer.Option(None, "--output", "-o", help="Output file"),
    no_interactive: bool = typer.Option(False, "--no-interactive", help="Never prompt; fail if required args missing"),
):
    """AI-powered contract generation and improvement.

    [dim]No args[/dim]           → wizard: generate or improve?
    [dim]file, no prompt[/dim]   → wizard: what to improve?
    [dim]--dataset only[/dim]    → generate from scratch
    [dim]file + prompt[/dim]     → improve existing contract
    [dim]--no-interactive[/dim]  → requires --dataset or file + prompt
    """
    gctx = _get_ctx(ctx)
    if no_interactive:
        gctx.no_interactive = True
    console = get_console(gctx)

    # Determine mode
    if file is None and prompt is None and dataset is None:
        if gctx.no_interactive:
            print_error(
                "In non-interactive mode, provide either --dataset (to generate) or "
                "a file path + prompt (to improve an existing contract).",
                gctx,
            )
        console.print("\n  [bold]Soda Copilot[/bold]")
        console.print()
        mode = Prompt.ask("  What would you like to do?", choices=["generate", "improve"], default="generate")
        if mode == "generate":
            dataset = Prompt.ask("Dataset FQN (datasource/db/schema/table)")
        else:
            file = Prompt.ask("Path to existing contract file")
            prompt = Prompt.ask("What should Copilot improve?")

    elif file is not None and prompt is None:
        if gctx.no_interactive:
            print_error(
                "In non-interactive mode, provide a prompt as the second argument. "
                "Example: soda contract copilot orders.yml 'add freshness checks'",
                gctx,
            )
        prompt = Prompt.ask(f"What should Copilot improve in [bold]{file}[/bold]?")

    console.print(f"[dim]Running Copilot...[/dim]")
    time.sleep(0.3)  # simulate AI call

    if file:
        out = output_file or file
        console.print(f"  [dim]update[/dim]  {out}")
        print_success(f"Contract updated: [bold]{out}[/bold]", gctx)
    else:
        parts = (dataset or "dataset").split("/")
        table = parts[-1]
        out = output_file or f"contracts/{table}.yml"
        console.print(f"  [dim]create[/dim]  {out}")
        print_success(f"Contract generated: [bold]{out}[/bold]", gctx)


@app.command("verify")
def verify(
    ctx: typer.Context,
    file_or_dir: Optional[str] = typer.Argument(None, help="Contract file or directory (default: contracts/)"),
    output: str = typer.Option("auto", "--output", "-o", help="Output format: table|json|csv"),
    datasource: Optional[str] = typer.Option(None, "--datasource", help="Datasource config file (overrides soda.yml)"),
    agent: bool = typer.Option(False, "--agent", help="Delegate execution to Soda Agent"),
    push: bool = typer.Option(False, "--push", help="Push check results to Soda Cloud"),
    set: Optional[list[str]] = typer.Option(None, "--set", help="Runtime variable overrides (key=value)"),
    check: Optional[str] = typer.Option(None, "--check", help="Run only checks matching this pattern"),
):
    """Run contract checks against your data.

    Exit codes: [bold]0[/bold]=pass  [bold]1[/bold]=checks failed  [bold]2[/bold]=error  [bold]3[/bold]=auth error

    Use [bold]--push[/bold] to send results to Soda Cloud.
    Use [bold]--agent[/bold] to delegate execution to a Soda Agent.
    """
    gctx = _get_ctx(ctx)
    if output != "auto":
        gctx.output = output
    console = get_console(gctx)

    target = file_or_dir or "contracts/"

    checks = [
        {"check": "row_count > 0",                  "status": "passing", "value": "2,412,847 rows"},
        {"check": "freshness < 1d",                  "status": "passing", "value": "7m ago"},
        {"check": "no_nulls(order_id)",              "status": "passing", "value": "0 nulls"},
        {"check": "no_nulls(customer_id)",           "status": "failing", "value": "143 nulls"},
        {"check": "valid_values(status)",            "status": "passing", "value": "5 valid values"},
        {"check": "reference(customer_id) → users", "status": "failing", "value": "23 orphaned rows"},
        {"check": "min(amount) >= 0",               "status": "passing", "value": "min=0.99"},
        {"check": "unique(order_id)",               "status": "passing", "value": "100% unique"},
    ]

    fmt = gctx.output if gctx.output != "auto" else ("table" if sys.stdout.isatty() else "json")

    if fmt == "json":
        passed = sum(1 for c in checks if c["status"] == "passing")
        failed = sum(1 for c in checks if c["status"] == "failing")
        print(json.dumps({"file": target, "passed": passed, "failed": failed, "checks": checks}, indent=2))
        raise typer.Exit(1)

    console.print()
    console.print(f"  [dim]verifying[/dim]  {target}")
    if agent:
        console.print(f"  [dim]via agent[/dim]")
    console.print()

    name_width = max(len(c["check"]) for c in checks)
    for c in checks:
        icon = "[green]✓[/green]" if c["status"] == "passing" else "[red]✗[/red]"
        name = c["check"].ljust(name_width)
        value = f"[dim]{c['value']}[/dim]"
        console.print(f"  {icon}  {name}  {value}")

    passed = sum(1 for c in checks if c["status"] == "passing")
    failed = sum(1 for c in checks if c["status"] == "failing")
    console.print()
    console.print(f"  [green]{passed} passed[/green]  [dim]·[/dim]  [red]{failed} failed[/red]")

    if push:
        console.print()
        console.print(f"  [dim]pushing results to Soda Cloud…[/dim]")
        print_success("Results pushed.", gctx)

    raise typer.Exit(1)  # checks failed


# ── proposal subcommands ──────────────────────────────────────────────────────

@proposal_app.command("list")
def proposal_list(
    ctx: typer.Context,
    status: str = typer.Option("open", "--status", help="Filter: open|done|all"),
):
    """List contract change proposals."""
    gctx = _get_ctx(ctx)
    data = PROPOSALS if status == "all" else [p for p in PROPOSALS if status == "all" or p["status"] == status]
    cols = ["id", "dataset", "status", "revision", "author", "message", "created"]
    render(data, cols, gctx, title="Proposals")


@proposal_app.command("pull")
def proposal_pull(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Proposal ID"),
    revision: Optional[int] = typer.Option(None, "--revision", "-r", help="Specific revision number"),
):
    """Download a proposal to a local file."""
    gctx = _get_ctx(ctx)
    rev = f" (revision {revision})" if revision else ""
    get_console(gctx).print(f"[dim]Fetching proposal [bold]{id}[/bold]{rev}...[/dim]")
    print_success(f"Proposal saved to [bold]contracts/proposal_{id}.yml[/bold]", gctx)


@proposal_app.command("push")
def proposal_push(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Proposal ID"),
    file: Optional[str] = typer.Argument(None, help="Local file to submit"),
    message: Optional[str] = typer.Option(None, "--message", "-m", help="Change description"),
):
    """Submit changes for a proposal."""
    gctx = _get_ctx(ctx)
    msg = message or "Updated via CLI"
    get_console(gctx).print(f"[dim]Submitting proposal [bold]{id}[/bold]...[/dim]")
    print_success(f"Proposal [bold]{id}[/bold] updated: \"{msg}\"", gctx)


@proposal_app.command("close")
def proposal_close(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Proposal ID"),
    status: str = typer.Option("done", "--status", help="Closing status: done|wontdo"),
):
    """Close a proposal."""
    gctx = _get_ctx(ctx)
    print_success(f"Proposal [bold]{id}[/bold] closed with status [bold]{status}[/bold].", gctx)
