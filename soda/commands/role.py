from typing import Optional

import typer

from soda.context import GlobalContext
from soda.output import render, render_one, print_success, print_error, get_console, output_option
from soda.mock import ROLES

app = typer.Typer(help="Manage roles and permissions.", no_args_is_help=True)


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("list")
def list_cmd(
    ctx: typer.Context,
    output: str = output_option(),
    scope: Optional[str] = typer.Option(None, "--scope", help="Filter by scope: global|dataset"),
):
    """List roles."""
    gctx = _get_ctx(ctx)
    if output != "auto": gctx.output = output
    data = ROLES
    if scope:
        data = [r for r in data if r["scope"] == scope]
    render(data, ["id", "name", "scope", "permissions", "members"], gctx, title="Roles")


@app.command("create")
def create(
    ctx: typer.Context,
    name: str = typer.Option(..., "--name", help="Role name"),
    scope: str = typer.Option(..., "--scope", help="Scope: global|dataset"),
):
    """Create a new role."""
    gctx = _get_ctx(ctx)
    print_success(f"Role [bold]{name}[/bold] created with scope [bold]{scope}[/bold].", gctx)


@app.command("delete")
def delete(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Role ID"),
):
    """Delete a role."""
    gctx = _get_ctx(ctx)
    print_success(f"Role [bold]{id}[/bold] deleted.", gctx)


@app.command("show")
def show(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Role ID"),
):
    """Show permissions defined in a role."""
    gctx = _get_ctx(ctx)
    role = next((r for r in ROLES if r["id"] == id), None)
    if not role:
        print_error(f"Role '{id}' not found.", gctx)

    render_one(role, gctx, title=f"Role — {role['name']}")

    # Mock permissions list
    console = get_console(gctx)
    console.print()
    perms_data = [
        {"permission": "datasets:read", "granted": True},
        {"permission": "contracts:write", "granted": True},
        {"permission": "contracts:verify", "granted": True},
        {"permission": "incidents:manage", "granted": False},
        {"permission": "users:manage", "granted": False},
    ]
    render(perms_data, ["permission", "granted"], gctx, title="Permissions")
