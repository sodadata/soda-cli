from typing import Optional

import typer

from soda.context import GlobalContext
from soda.output import render, output_option
from soda.mock import RESULTS

app = typer.Typer(help="Query data quality signals (checks + monitor alerts).", no_args_is_help=True)


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("list")
def list_cmd(
    ctx: typer.Context,
    output: str = output_option(),
    datasource: Optional[str] = typer.Option(None, "--datasource", help="Filter by datasource"),
    dataset: Optional[str] = typer.Option(None, "--dataset", help="Filter by dataset ID"),
    type: Optional[str] = typer.Option(None, "--type", help="Filter by type: check|monitor|all"),
    status: Optional[str] = typer.Option(None, "--status", help="Filter by status: passing|failing|all"),
    from_date: Optional[str] = typer.Option(None, "--from", help="Start date (YYYY-MM-DD)"),
    to_date: Optional[str] = typer.Option(None, "--to", help="End date (YYYY-MM-DD)"),
):
    """List all data quality signals: contract checks and monitor alerts.

    [dim]Note: 'results' is a temporary name. Covers both contract checks and ML monitor alerts.[/dim]
    """
    gctx = _get_ctx(ctx)
    if output != "auto":
        gctx.output = output

    results = RESULTS
    if datasource:
        results = [r for r in results if r["dataset"].startswith(datasource)]
    if dataset:
        results = [r for r in results if dataset in r["dataset"]]
    if type and type != "all":
        results = [r for r in results if r["type"] == type]
    if status and status != "all":
        results = [r for r in results if r["status"] == status]

    cols = ["dataset", "type", "name", "status", "value", "date"]
    render(results, cols, gctx, title="Results")
