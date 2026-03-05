from typing import Optional

import typer

from soda.context import GlobalContext
from soda.output import render, print_success, print_error, get_console, output_option
from soda.mock import USERS, GROUPS

app = typer.Typer(help="Manage users and groups.", no_args_is_help=True)

group_app = typer.Typer(help="Manage user groups.", no_args_is_help=True)
app.add_typer(group_app, name="group")


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("list")
def list_cmd(
    ctx: typer.Context,
    output: str = output_option(),
):
    """List all users in the organization."""
    gctx = _get_ctx(ctx)
    if output != "auto": gctx.output = output
    render(USERS, ["id", "email", "name", "role", "group", "last_active"], gctx, title="Users")


@app.command("assign")
def assign(
    ctx: typer.Context,
    user_id: str = typer.Argument(..., help="User ID or email"),
    role: str = typer.Option(..., "--role", help="Role ID to assign"),
):
    """Assign a role to a user."""
    gctx = _get_ctx(ctx)
    print_success(f"Role [bold]{role}[/bold] assigned to [bold]{user_id}[/bold].", gctx)


@app.command("revoke")
def revoke(
    ctx: typer.Context,
    user_id: str = typer.Argument(..., help="User ID or email"),
    role: str = typer.Option(..., "--role", help="Role ID to revoke"),
):
    """Revoke a role from a user."""
    gctx = _get_ctx(ctx)
    print_success(f"Role [bold]{role}[/bold] revoked from [bold]{user_id}[/bold].", gctx)


# ── group subcommands ─────────────────────────────────────────────────────────

@group_app.command("list")
def group_list(
    ctx: typer.Context,
    output: str = output_option(),
):
    """List all user groups."""
    gctx = _get_ctx(ctx)
    if output != "auto": gctx.output = output
    render(GROUPS, ["id", "name", "role", "members", "created"], gctx, title="Groups")


@group_app.command("create")
def group_create(
    ctx: typer.Context,
    name: str = typer.Option(..., "--name", help="Group name"),
    members: Optional[list[str]] = typer.Option(None, "--members", help="Member emails"),
):
    """Create a new user group."""
    gctx = _get_ctx(ctx)
    count = len(members) if members else 0
    print_success(f"Group [bold]{name}[/bold] created with {count} members.", gctx)


@group_app.command("update")
def group_update(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Group ID"),
    name: Optional[str] = typer.Option(None, "--name"),
    add_member: Optional[str] = typer.Option(None, "--add-member"),
    remove_member: Optional[str] = typer.Option(None, "--remove-member"),
):
    """Update a user group."""
    gctx = _get_ctx(ctx)
    print_success(f"Group [bold]{id}[/bold] updated.", gctx)


@group_app.command("delete")
def group_delete(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Group ID"),
):
    """Delete a user group."""
    gctx = _get_ctx(ctx)
    print_success(f"Group [bold]{id}[/bold] deleted.", gctx)


@group_app.command("assign")
def group_assign(
    ctx: typer.Context,
    group_id: str = typer.Argument(..., help="Group ID"),
    role: str = typer.Option(..., "--role", help="Role ID to assign"),
):
    """Assign a role to a group."""
    gctx = _get_ctx(ctx)
    print_success(f"Role [bold]{role}[/bold] assigned to group [bold]{group_id}[/bold].", gctx)


@group_app.command("revoke")
def group_revoke(
    ctx: typer.Context,
    group_id: str = typer.Argument(..., help="Group ID"),
    role: str = typer.Option(..., "--role", help="Role ID to revoke"),
):
    """Revoke a role from a group."""
    gctx = _get_ctx(ctx)
    print_success(f"Role [bold]{role}[/bold] revoked from group [bold]{group_id}[/bold].", gctx)
