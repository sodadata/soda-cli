import csv
import io
import json
import sys
from typing import Any, Optional

import typer
from rich.console import Console
from rich.table import Table

from soda.context import GlobalContext


def _effective_format(ctx: GlobalContext) -> str:
    if ctx.output != "auto":
        return ctx.output
    return "table" if sys.stdout.isatty() else "json"


def output_option() -> str:
    """Shared --output/-o option for commands that produce output."""
    return typer.Option("auto", "--output", "-o", help="Output format: table|json|csv")


def render(
    data: list[dict],
    columns: list[str],
    ctx: GlobalContext,
    title: Optional[str] = None,
):
    """Render a list of dicts as table, json, or csv."""
    fmt = _effective_format(ctx)

    if fmt == "json":
        print(json.dumps(data, indent=2, default=str))
        return

    if fmt == "csv":
        writer = csv.DictWriter(sys.stdout, fieldnames=columns, extrasaction="ignore")
        writer.writeheader()
        for row in data:
            writer.writerow({k: row.get(k, "") for k in columns})
        return

    # table
    console = Console(no_color=ctx.no_color)
    table = Table(title=title, show_header=True, header_style="bold cyan")
    for col in columns:
        table.add_column(col)
    for row in data:
        table.add_row(*[str(row.get(col, "")) for col in columns])
    console.print(table)


def render_one(item: dict, ctx: GlobalContext, title: Optional[str] = None):
    """Render a single record as a key-value table or JSON."""
    fmt = _effective_format(ctx)

    if fmt == "json":
        print(json.dumps(item, indent=2, default=str))
        return

    if fmt == "csv":
        writer = csv.DictWriter(sys.stdout, fieldnames=list(item.keys()))
        writer.writeheader()
        writer.writerow(item)
        return

    console = Console(no_color=ctx.no_color)
    table = Table(title=title, show_header=False)
    table.add_column("Field", style="bold cyan")
    table.add_column("Value")
    for k, v in item.items():
        table.add_row(k, str(v))
    console.print(table)


def _strip_markup(text: str) -> str:
    """Remove Rich markup tags from a string for plain-text/JSON output."""
    from rich.text import Text
    return Text.from_markup(text).plain


def print_success(msg: str, ctx: GlobalContext):
    if ctx.quiet:
        return
    fmt = _effective_format(ctx)
    if fmt == "json":
        print(json.dumps({"status": "success", "message": _strip_markup(msg)}))
    else:
        console = Console(no_color=ctx.no_color)
        console.print(f"[green]✓[/green] {msg}")


def print_error(msg: str, ctx: GlobalContext, exit_code: int = 2):
    fmt = _effective_format(ctx)
    if fmt == "json":
        print(json.dumps({"status": "error", "message": _strip_markup(msg)}), file=sys.stderr)
    else:
        console = Console(no_color=ctx.no_color, stderr=True)
        console.print(f"[red]✗[/red] {msg}")
    raise SystemExit(exit_code)


def get_console(ctx: GlobalContext) -> Console:
    return Console(no_color=ctx.no_color)
