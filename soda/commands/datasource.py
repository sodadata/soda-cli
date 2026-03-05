from typing import Optional

import typer
from rich.prompt import Prompt

from soda.context import GlobalContext
from soda.output import render, render_one, print_success, print_error, get_console, output_option
from soda.mock import DATASOURCES

app = typer.Typer(help="Manage datasource connections.", no_args_is_help=True)

SUPPORTED_TYPES = ["postgres", "snowflake", "bigquery", "duckdb", "redshift", "mysql", "spark"]


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("list")
def list_cmd(
    ctx: typer.Context,
    output: str = output_option(),
):
    """List all configured datasources."""
    gctx = _get_ctx(ctx)
    if output != "auto": gctx.output = output
    cols = ["id", "name", "type", "host", "status", "datasets"]
    render(DATASOURCES, cols, gctx, title="Datasources")


@app.command("create")
def create(
    ctx: typer.Context,
    type: Optional[str] = typer.Option(None, "--type", "-t", help=f"Datasource type: {', '.join(SUPPORTED_TYPES)}"),
    name: Optional[str] = typer.Option(None, "--name", "-n", help="Datasource name"),
    output_file: Optional[str] = typer.Option(None, "--output", "-o", help="Output YAML file path"),
    no_interactive: bool = typer.Option(False, "--no-interactive", help="Never prompt; require --type and --name"),
):
    """Create a local YAML datasource config (no cloud registration)."""
    gctx = _get_ctx(ctx)
    if no_interactive:
        gctx.no_interactive = True
    console = get_console(gctx)

    if type is None:
        if gctx.no_interactive:
            print_error("--type is required in non-interactive mode", gctx)
        type = Prompt.ask("Datasource type", choices=SUPPORTED_TYPES, default="postgres")

    if name is None:
        if gctx.no_interactive:
            print_error("--name is required in non-interactive mode", gctx)
        name = Prompt.ask("Datasource name", default=f"my_{type}")

    out = output_file or f"configs/{name}.yml"
    console.print(f"  [dim]create[/dim]  [cyan]{out}[/cyan]")
    print_success(f"Datasource config written to [bold]{out}[/bold]. Edit it to add credentials.", gctx)


@app.command("onboard")
def onboard(
    ctx: typer.Context,
    agent: str = typer.Option(..., "--agent", help="Soda Agent name"),
    type: str = typer.Option(..., "--type", "-t", help=f"Datasource type: {', '.join(SUPPORTED_TYPES)}"),
    name: Optional[str] = typer.Option(None, "--name", "-n", help="Datasource name"),
):
    """Register a new cloud datasource connection via a Soda Agent."""
    gctx = _get_ctx(ctx)
    console = get_console(gctx)

    ds_name = name or f"{type}_via_{agent}"
    console.print(f"[dim]Registering datasource via agent [bold]{agent}[/bold]...[/dim]")
    print_success(
        f"Datasource [bold]{ds_name}[/bold] registered. "
        "Connection test will run when the agent picks up the task.",
        gctx,
    )


@app.command("test")
def test(
    ctx: typer.Context,
    name_or_file: Optional[str] = typer.Argument(None, help="Datasource name or config file path"),
):
    """Test a datasource connection."""
    gctx = _get_ctx(ctx)
    console = get_console(gctx)
    target = name_or_file or "default"
    console.print(f"[dim]Connecting to [bold]{target}[/bold]...[/dim]")
    print_success(f"Connection to [bold]{target}[/bold] successful.", gctx)


@app.command("diagnostics")
def diagnostics(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Datasource ID"),
):
    """Show primary diagnostics warehouse config for a datasource."""
    gctx = _get_ctx(ctx)
    ds = next((d for d in DATASOURCES if d["id"] == id), None)
    if not ds:
        print_error(f"Datasource '{id}' not found.", gctx)

    diag = {
        "datasource": id,
        "diagnostics_warehouse": "pg_prod (same datasource)",
        "schema": "soda_diagnostics",
        "retention_days": 90,
        "store_failed_rows": True,
    }
    render_one(diag, gctx, title=f"Diagnostics Config — {id}")
