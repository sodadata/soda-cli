from typing import Optional

import typer

from soda.context import GlobalContext
from soda.output import render, render_one, print_success, print_error, get_console, output_option
from soda.mock import INCIDENTS

app = typer.Typer(help="Manage data quality incidents.", no_args_is_help=True)


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("list")
def list_cmd(
    ctx: typer.Context,
    output: str = output_option(),
    status: str = typer.Option("open", "--status", help="Filter: open|closed|all"),
    dataset: Optional[str] = typer.Option(None, "--dataset", help="Filter by dataset ID"),
):
    """List incidents."""
    gctx = _get_ctx(ctx)
    if output != "auto": gctx.output = output
    data = INCIDENTS
    if status != "all":
        data = [i for i in data if i["status"] == status]
    if dataset:
        data = [i for i in data if dataset in i["dataset"]]
    cols = ["id", "title", "dataset", "severity", "status", "opened", "updated"]
    render(data, cols, gctx, title="Incidents")


@app.command("get")
def get(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Incident ID"),
):
    """Get incident details."""
    gctx = _get_ctx(ctx)
    item = next((i for i in INCIDENTS if i["id"] == id), None)
    if not item:
        print_error(f"Incident '{id}' not found.", gctx)
    render_one(item, gctx, title=f"Incident {id}")


@app.command("update")
def update(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Incident ID"),
    status: Optional[str] = typer.Option(None, "--status", help="New status: open|closed"),
    note: Optional[str] = typer.Option(None, "--note", help="Add a note"),
):
    """Update an incident status or add a note."""
    gctx = _get_ctx(ctx)
    updates = []
    if status:
        updates.append(f"status → {status}")
    if note:
        updates.append(f"note added")
    desc = ", ".join(updates) if updates else "no changes"
    print_success(f"Incident [bold]{id}[/bold] updated ({desc}).", gctx)
