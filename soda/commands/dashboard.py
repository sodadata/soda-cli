import typer
from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from rich.columns import Columns
import json

from soda.context import GlobalContext
from soda.output import get_console
from soda.mock import DASHBOARD_SUMMARY


def run(ctx: typer.Context):
    """Org-level overview: datasets, results, incidents, recent jobs."""
    gctx = ctx.obj or GlobalContext()
    d = DASHBOARD_SUMMARY

    import sys
    fmt = gctx.output if gctx.output != "auto" else ("table" if sys.stdout.isatty() else "json")

    if fmt == "json":
        print(json.dumps(d, indent=2))
        return

    console = get_console(gctx)
    console.print()
    console.print(Panel.fit(
        f"[bold]Soda Cloud[/bold]  ·  [cyan]{d['org']}[/cyan]  ·  as of {d['as_of']}",
        style="bold",
    ))
    console.print()

    # Summary row
    summary_table = Table.grid(padding=(0, 4))
    summary_table.add_column(justify="center")
    summary_table.add_column(justify="center")
    summary_table.add_column(justify="center")
    summary_table.add_column(justify="center")
    summary_table.add_column(justify="center")
    summary_table.add_row(
        f"[bold]{d['datasets']}[/bold]\n[dim]datasets[/dim]",
        f"[green]{d['passing']}[/green]\n[dim]passing[/dim]",
        f"[red]{d['failing']}[/red]\n[dim]failing[/dim]",
        f"[yellow]{d['erroring']}[/yellow]\n[dim]erroring[/dim]",
        f"[bold]{d['open_incidents']}[/bold]\n[dim]open incidents[/dim]",
    )
    console.print(summary_table)
    console.print()

    # Recent failures
    if d["recent_failures"]:
        console.print("[bold red]Recent Failures[/bold red]")
        t = Table(show_header=True, header_style="bold")
        t.add_column("Dataset")
        t.add_column("Check")
        t.add_column("Time")
        for f in d["recent_failures"]:
            t.add_row(f["dataset"], f["check"], f["time"])
        console.print(t)
