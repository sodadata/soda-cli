import time
from typing import Optional

import typer

from soda.context import GlobalContext
from soda.output import render, print_error, get_console, output_option
from soda.mock import JOBS, JOB_LOGS

app = typer.Typer(help="View execution history (read-only).", no_args_is_help=True)


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("list")
def list_cmd(
    ctx: typer.Context,
    output: str = output_option(),
    datasource: Optional[str] = typer.Option(None, "--datasource", help="Filter by datasource ID"),
    dataset: Optional[str] = typer.Option(None, "--dataset", help="Filter by dataset ID"),
    type: Optional[str] = typer.Option(None, "--type", help="Filter by type: contract|monitor|all"),
    status: Optional[str] = typer.Option(None, "--status", help="Filter by status: passing|failing|running|error"),
):
    """List recent execution jobs."""
    gctx = _get_ctx(ctx)
    if output != "auto": gctx.output = output

    jobs = JOBS
    if datasource:
        jobs = [j for j in jobs if j["datasource"] == datasource]
    if dataset:
        jobs = [j for j in jobs if j["dataset"] == dataset]
    if type and type != "all":
        jobs = [j for j in jobs if j["type"] == type]
    if status and status != "all":
        jobs = [j for j in jobs if j["status"] == status]

    # Add status icons for table display
    display = []
    for j in jobs:
        row = dict(j)
        icon = {"passing": "✓", "failing": "✗", "error": "⚠", "running": "⟳"}.get(j["status"], "")
        row["status"] = f"{icon} {j['status']}"
        display.append(row)

    cols = ["id", "datasource", "dataset", "type", "status", "duration", "date"]
    render(display, cols, gctx, title="Jobs")


@app.command("logs")
def logs(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Job ID"),
    follow: bool = typer.Option(False, "--follow", "-f", help="Stream logs (simulated)"),
):
    """Stream logs for a job."""
    gctx = _get_ctx(ctx)
    console = get_console(gctx)

    job_logs = JOB_LOGS.get(id)
    if not job_logs:
        # fallback for unknown IDs
        job_logs = [
            f"[{id}] No detailed logs available for this job.",
            f"[{id}] Status: completed",
        ]

    for line in job_logs:
        console.print(line)
        if follow:
            time.sleep(0.05)
