from typing import Optional

import typer

from soda.context import GlobalContext
from soda.output import render, render_one, print_success, print_error, get_console, output_option
from soda.mock import DATASETS, MONITORS, PROFILE_DATA

app = typer.Typer(help="Manage datasets.", no_args_is_help=True)

# Sub-apps for nested commands
profiling_app = typer.Typer(help="Dataset profiling commands.", no_args_is_help=True)
permissions_app = typer.Typer(help="Dataset permission commands.", no_args_is_help=True)
monitor_app = typer.Typer(help="Dataset monitor commands.", no_args_is_help=True)

app.add_typer(profiling_app, name="profiling")
app.add_typer(permissions_app, name="permissions")
app.add_typer(monitor_app, name="monitor")


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("list")
def list_cmd(
    ctx: typer.Context,
    output: str = output_option(),
    filter: Optional[str] = typer.Option(None, "--filter", help="Filter query"),
    tag: Optional[str] = typer.Option(None, "--tag", help="Filter by tag"),
):
    """List all datasets."""
    gctx = _get_ctx(ctx)
    if output != "auto": gctx.output = output
    data = DATASETS
    if filter:
        data = [d for d in data if filter.lower() in d["name"].lower() or filter.lower() in d["fqn"].lower()]

    cols = ["id", "fqn", "rows", "status", "last_scan"]
    render(data, cols, gctx, title="Datasets")


@app.command("update")
def update(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Dataset ID"),
    name: Optional[str] = typer.Option(None, "--name", help="New name"),
    tag: Optional[str] = typer.Option(None, "--tag", help="Add tag"),
    attr: Optional[list[str]] = typer.Option(None, "--attr", help="Set attribute (key=value)"),
):
    """Update dataset metadata."""
    gctx = _get_ctx(ctx)
    print_success(f"Dataset [bold]{id}[/bold] updated.", gctx)


@app.command("delete")
def delete(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Dataset ID"),
):
    """Delete a dataset from Soda Cloud."""
    gctx = _get_ctx(ctx)
    print_success(f"Dataset [bold]{id}[/bold] deleted.", gctx)


@app.command("diagnostics")
def diagnostics(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Dataset ID"),
):
    """Show dataset-level diagnostics warehouse overrides."""
    gctx = _get_ctx(ctx)
    diag = {
        "dataset": id,
        "diagnostics_override": "none (inherits from datasource)",
        "store_failed_rows": True,
        "retention_days": 30,
    }
    render_one(diag, gctx, title=f"Diagnostics Config — {id}")


# ── profiling ─────────────────────────────────────────────────────────────────

@profiling_app.command("show")
def profiling_show(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Dataset ID"),
):
    """View cached profiling data for a dataset."""
    gctx = _get_ctx(ctx)
    profile = PROFILE_DATA.get(id)
    if not profile:
        profile = {"dataset": id, "message": "No profiling data available. Run: soda dataset profiling refresh"}
        render_one(profile, gctx)
        return

    console = get_console(gctx)
    summary = {k: v for k, v in profile.items() if k != "columns_detail"}
    render_one(summary, gctx, title=f"Profiling — {profile['dataset']}")

    if profile.get("columns_detail"):
        console.print()
        render(profile["columns_detail"], ["column", "type", "nulls", "unique", "min", "max"], gctx, title="Columns")


@profiling_app.command("refresh")
def profiling_refresh(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Dataset ID"),
):
    """Trigger a new profiling run for a dataset."""
    gctx = _get_ctx(ctx)
    get_console(gctx).print(f"[dim]Triggering profiling run for [bold]{id}[/bold]...[/dim]")
    print_success(f"Profiling job queued for [bold]{id}[/bold]. Check status with: soda job list", gctx)


# ── permissions ───────────────────────────────────────────────────────────────

@permissions_app.command("list")
def permissions_list(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Dataset ID"),
):
    """List permissions for a dataset."""
    gctx = _get_ctx(ctx)
    data = [
        {"principal": "alice@acme.com", "type": "user", "role": "Dataset Owner"},
        {"principal": "data-platform", "type": "group", "role": "Data Engineer"},
        {"principal": "analytics", "type": "group", "role": "Viewer"},
    ]
    render(data, ["principal", "type", "role"], gctx, title=f"Permissions — {id}")


@permissions_app.command("set")
def permissions_set(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Dataset ID"),
    role: str = typer.Option(..., "--role", help="Role ID"),
    user: Optional[str] = typer.Option(None, "--user", help="User email"),
    group: Optional[str] = typer.Option(None, "--group", help="Group ID"),
):
    """Set a permission on a dataset."""
    gctx = _get_ctx(ctx)
    if not user and not group:
        print_error("Provide either --user <email> or --group <id>", gctx)
    principal = user or group
    print_success(f"Role [bold]{role}[/bold] assigned to [bold]{principal}[/bold] on dataset [bold]{id}[/bold].", gctx)


# ── monitor ───────────────────────────────────────────────────────────────────

@monitor_app.command("list")
def monitor_list(
    ctx: typer.Context,
    dataset_id: str = typer.Argument(..., help="Dataset ID"),
):
    """List monitors for a dataset."""
    gctx = _get_ctx(ctx)
    data = MONITORS.get(dataset_id, [])
    if not data:
        get_console(gctx).print(f"[dim]No monitors configured for [bold]{dataset_id}[/bold][/dim]")
        return
    render(data, ["id", "column", "metric", "alert_threshold", "status", "last_run"], gctx, title=f"Monitors — {dataset_id}")


@monitor_app.command("add")
def monitor_add(
    ctx: typer.Context,
    dataset_id: str = typer.Argument(..., help="Dataset ID"),
    column: Optional[str] = typer.Option(None, "--column", help="Column to monitor"),
    metric: Optional[str] = typer.Option(None, "--metric", help="Metric: row_count|avg|min|max|count_distinct|..."),
):
    """Add an ML anomaly monitor to a dataset."""
    gctx = _get_ctx(ctx)
    col = column or "(SQL)"
    met = metric or "row_count"
    print_success(f"Monitor added: [bold]{met}[/bold] on [bold]{col}[/bold] for dataset [bold]{dataset_id}[/bold].", gctx)


@monitor_app.command("update")
def monitor_update(
    ctx: typer.Context,
    dataset_id: str = typer.Argument(..., help="Dataset ID"),
    monitor_id: str = typer.Argument(..., help="Monitor ID"),
):
    """Update a monitor configuration."""
    gctx = _get_ctx(ctx)
    print_success(f"Monitor [bold]{monitor_id}[/bold] updated.", gctx)


@monitor_app.command("delete")
def monitor_delete(
    ctx: typer.Context,
    dataset_id: str = typer.Argument(..., help="Dataset ID"),
    monitor_id: str = typer.Argument(..., help="Monitor ID"),
):
    """Delete a monitor."""
    gctx = _get_ctx(ctx)
    print_success(f"Monitor [bold]{monitor_id}[/bold] deleted from dataset [bold]{dataset_id}[/bold].", gctx)
