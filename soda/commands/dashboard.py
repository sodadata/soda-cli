import json
import sys

import typer

from soda.context import GlobalContext
from soda.output import get_console
from soda.mock import DASHBOARD_SUMMARY


def run(ctx: typer.Context):
    gctx = ctx.obj or GlobalContext()
    d = DASHBOARD_SUMMARY

    fmt = gctx.output if gctx.output != "auto" else ("table" if sys.stdout.isatty() else "json")
    if fmt == "json":
        print(json.dumps(d, indent=2))
        return

    console = get_console(gctx)
    console.print()
    console.print(f"  [bold]{d['org']}[/bold]  [dim]{d['as_of']}[/dim]")
    console.print()

    # Stats row — green for passing, red for failing, dim for neutral
    stats = [
        (str(d["datasets"]),       "datasets"),
        (f"[green]{d['passing']}[/green]", "passing"),
        (f"[red]{d['failing']}[/red]",     "failing"),
        (f"[yellow]{d['erroring']}[/yellow]", "erroring"),
        (str(d["open_incidents"]), "open incidents"),
        (str(d["jobs_today"]),     "jobs today"),
    ]
    parts = "   ".join(f"{v}  [dim]{label}[/dim]" for v, label in stats)
    console.print(f"  {parts}")

    if d["recent_failures"]:
        console.print()
        console.print("  [dim]recent failures[/dim]")
        console.print()
        for f in d["recent_failures"]:
            console.print(f"  [dim]{f['time']:>5}[/dim]  {f['dataset']}  [dim]{f['check']}[/dim]")

    console.print()
