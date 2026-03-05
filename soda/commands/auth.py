import sys
from typing import Optional

import typer
from rich.prompt import Prompt

from soda.context import GlobalContext
from soda.output import render_one, print_success, print_error, get_console, output_option
from soda.mock import AUTH_STATUS

app = typer.Typer(help="Manage authentication and profiles.", no_args_is_help=True)


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("login")
def login(
    ctx: typer.Context,
    profile: Optional[str] = typer.Option(None, "--profile", "-p", help="Profile name to create or update"),
    api_key: Optional[str] = typer.Option(None, "--api-key", help="Soda Cloud API key (skips prompt)"),
    host: str = typer.Option("cloud.soda.io", "--host", help="Soda Cloud host"),
    no_interactive: bool = typer.Option(False, "--no-interactive", help="Never prompt; fail if --api-key not provided"),
):
    """Authenticate with Soda Cloud. Saves credentials to ~/.soda/credentials."""
    gctx = _get_ctx(ctx)
    if no_interactive:
        gctx.no_interactive = True

    console = get_console(gctx)
    profile_name = profile or gctx.profile or "default"

    if api_key is None:
        if gctx.no_interactive:
            print_error(
                "--api-key is required in non-interactive mode. "
                "Get your API key at https://cloud.soda.io/settings/api-keys",
                gctx,
                exit_code=2,
            )
        console.print(f"\n  [bold]Soda Cloud Login[/bold]")
        console.print(f"  [dim]profile[/dim]  {profile_name}")
        console.print(f"  [dim]host[/dim]     {host}")
        console.print()
        api_key = Prompt.ask("  API key")

    console.print(f"\n  [dim]Connecting to {host}…[/dim]")
    # mock: always succeeds
    console.print(f"  [dim]Authenticated as alice@acme.com (acme-corp)[/dim]")
    console.print()
    print_success(f"Profile [bold]{profile_name}[/bold] saved to [dim]~/.soda/credentials[/dim]", gctx)


@app.command("logout")
def logout(
    ctx: typer.Context,
    profile: Optional[str] = typer.Option(None, "--profile", "-p", help="Profile to remove"),
):
    """Remove credentials for a profile."""
    gctx = _get_ctx(ctx)
    profile_name = profile or gctx.profile or "default"
    print_success(f"Profile [bold]{profile_name}[/bold] removed from [dim]~/.soda/credentials[/dim]", gctx)


@app.command("status")
def status(
    ctx: typer.Context,
    output: str = output_option(),
):
    """Show active profile, org, and connection health."""
    gctx = _get_ctx(ctx)
    if output != "auto":
        gctx.output = output
    render_one(AUTH_STATUS, gctx)


@app.command("switch")
def switch(
    ctx: typer.Context,
    profile: str = typer.Argument(..., help="Profile name to switch to"),
):
    """Switch the active auth profile."""
    gctx = _get_ctx(ctx)
    print_success(f"Active profile switched to [bold]{profile}[/bold]", gctx)
