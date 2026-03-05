import csv
import json
import sys
from typing import Optional

import typer
from rich import box as rich_box
from rich.console import Console
from rich.text import Text

from soda.context import GlobalContext

# Soda brand palette — green accent, red for failures only, dim for secondary info
STATUS_STYLES = {
    "passing":  ("[green]✓[/green]", ""),
    "failing":  ("[red]✗[/red]",   ""),
    "error":    ("[red]✗[/red]",   ""),
    "alert":    ("[yellow]⚠[/yellow]", ""),
    "running":  ("[dim]⟳[/dim]",   ""),
    "open":     ("[yellow]●[/yellow]", ""),
    "closed":   ("[dim]●[/dim]",   ""),
    "connected":    ("[green]●[/green]", ""),
    "degraded":     ("[yellow]●[/yellow]", ""),
    "disconnected": ("[dim]●[/dim]", ""),
    "active":   ("[green]●[/green]", ""),
}


def _effective_format(ctx: GlobalContext) -> str:
    if ctx.output != "auto":
        return ctx.output
    return "table" if sys.stdout.isatty() else "json"


def output_option() -> str:
    """Shared --output/-o option for commands that produce output."""
    return typer.Option("auto", "--output", "-o", help="Output format: table|json|csv")


def _strip_markup(text: str) -> str:
    """Remove Rich markup tags from a string."""
    return Text.from_markup(text).plain


def _sanitize(data: list[dict]) -> list[dict]:
    """Strip Rich markup from all string values (for JSON/CSV output)."""
    result = []
    for row in data:
        clean = {}
        for k, v in row.items():
            clean[k] = _strip_markup(str(v)) if isinstance(v, str) and "[" in v else v
        result.append(clean)
    return result


def _fmt_status(raw: str) -> str:
    """Apply color to a known status value for table display."""
    key = raw.lower().strip()
    if key in STATUS_STYLES:
        icon, _ = STATUS_STYLES[key]
        return f"{icon} {raw}"
    return raw


def render(
    data: list[dict],
    columns: list[str],
    ctx: GlobalContext,
    title: Optional[str] = None,
):
    """Render a list of dicts as table, json, or csv."""
    fmt = _effective_format(ctx)

    if fmt == "json":
        print(json.dumps(_sanitize(data), indent=2, default=str))
        return

    if fmt == "csv":
        clean = _sanitize(data)
        writer = csv.DictWriter(sys.stdout, fieldnames=columns, extrasaction="ignore")
        writer.writeheader()
        for row in clean:
            writer.writerow({k: row.get(k, "") for k in columns})
        return

    # table — clean, minimal, Soda-branded
    # Use at least 160 cols so tables aren't mangled in narrow test environments;
    # in a real terminal Console picks up the actual width automatically.
    import shutil
    term_width = max(shutil.get_terminal_size((160, 40)).columns, 160)
    console = Console(no_color=ctx.no_color, width=term_width)
    if title:
        console.print(f"  [dim]{title}[/dim]")

    # Long free-text columns get ellipsis truncation; compact ones stay nowrap
    long_cols = {"dataset", "fqn", "name", "check", "title", "message", "permissions", "url", "channel"}
    dim_cols = {"id", "date", "last_scan", "opened", "updated", "created", "last_active", "last_run", "duration"}

    table = _make_table()
    for col in columns:
        if col in long_cols:
            table.add_column(col.replace("_", " ").upper(), overflow="ellipsis", no_wrap=True, max_width=48)
        else:
            table.add_column(col.replace("_", " ").upper(), no_wrap=True)

    for row in data:
        cells = []
        for col in columns:
            val = str(row.get(col, ""))
            if col == "status":
                val = _fmt_status(val)
            elif col in dim_cols:
                val = f"[dim]{val}[/dim]"
            cells.append(val)
        table.add_row(*cells)

    console.print(table)


def render_one(item: dict, ctx: GlobalContext, title: Optional[str] = None):
    """Render a single record as aligned key-value pairs or JSON."""
    fmt = _effective_format(ctx)

    if fmt == "json":
        clean = {k: (_strip_markup(str(v)) if isinstance(v, str) and "[" in v else v) for k, v in item.items()}
        print(json.dumps(clean, indent=2, default=str))
        return

    if fmt == "csv":
        writer = csv.DictWriter(sys.stdout, fieldnames=list(item.keys()))
        writer.writeheader()
        writer.writerow(item)
        return

    console = Console(no_color=ctx.no_color)
    if title:
        console.print(f"  [dim]{title}[/dim]")

    key_width = max((len(k) for k in item), default=12) + 2
    for k, v in item.items():
        val = str(v)
        if k == "status":
            val = _fmt_status(val)
        # Pad the plain key, then wrap in dim markup after measuring
        console.print(f"  [dim]{k:<{key_width}}[/dim]  {val}")


def _make_table():
    """Shared table style — minimal separator, no outer borders."""
    from rich.table import Table
    return Table(
        show_header=True,
        header_style="bold",
        box=rich_box.SIMPLE_HEAD,
        show_edge=False,
        pad_edge=False,
        padding=(0, 2),
    )


def print_success(msg: str, ctx: GlobalContext):
    if ctx.quiet:
        return
    fmt = _effective_format(ctx)
    if fmt == "json":
        print(json.dumps({"status": "success", "message": _strip_markup(msg)}))
    else:
        Console(no_color=ctx.no_color).print(f"[green]✓[/green]  {msg}")


def print_error(msg: str, ctx: GlobalContext, exit_code: int = 2):
    fmt = _effective_format(ctx)
    if fmt == "json":
        print(json.dumps({"status": "error", "message": _strip_markup(msg)}), file=sys.stderr)
    else:
        Console(no_color=ctx.no_color, stderr=True).print(f"[red]✗[/red]  {msg}")
    raise SystemExit(exit_code)


def get_console(ctx: GlobalContext) -> Console:
    return Console(no_color=ctx.no_color)
