import sys
from typing import Optional

import typer
from rich.console import Console

from soda.context import GlobalContext
from soda.commands import (
    auth,
    init_cmd,
    dashboard,
    datasource,
    contract,
    job,
    results,
    dataset,
    incident,
    notification,
    role,
    users,
    completion,
)

__version__ = "0.1.0-alpha"

app = typer.Typer(
    name="soda",
    help="Soda — data quality CLI. Run checks, manage contracts, and monitor your data.",
    no_args_is_help=True,
    rich_markup_mode="rich",
    add_completion=False,
)

# Register subcommand groups
app.add_typer(auth.app, name="auth", help="Manage authentication and profiles.")
app.add_typer(datasource.app, name="datasource", help="Manage datasource connections.", rich_help_panel="Resources")
app.add_typer(datasource.app, name="ds", help="Alias for datasource.", hidden=True)
app.add_typer(contract.app, name="contract", help="Manage and verify data contracts.", rich_help_panel="Core")
app.add_typer(job.app, name="job", help="View execution history.", rich_help_panel="Observability")
app.add_typer(job.app, name="scan", help="Alias for job.", hidden=True)
app.add_typer(results.app, name="results", help="Query data quality signals.", rich_help_panel="Observability")
app.add_typer(dataset.app, name="dataset", help="Manage datasets.", rich_help_panel="Resources")
app.add_typer(incident.app, name="incident", help="Manage incidents.", rich_help_panel="Observability")
app.add_typer(notification.app, name="notification", help="Manage notifications and channels.", rich_help_panel="Settings")
app.add_typer(role.app, name="role", help="Manage roles and permissions.", rich_help_panel="Settings")
app.add_typer(users.app, name="users", help="Manage users and groups.", rich_help_panel="Settings")
app.add_typer(completion.app, name="completion", help="Shell completion scripts.")


@app.command("version")
def version_cmd():
    """Print version information."""
    console = Console()
    console.print(f"soda version [bold]{__version__}[/bold]")


@app.command("init")
def init_command(
    ctx: typer.Context,
    force: bool = typer.Option(False, "--force", help="Overwrite existing soda.yml"),
):
    """Scaffold soda.yml, configs/, and contracts/ in the current directory."""
    init_cmd.run(ctx, force=force)


@app.command("dashboard")
def dashboard_command(ctx: typer.Context):
    """Org-level overview: datasets, results, incidents, recent jobs."""
    dashboard.run(ctx)


@app.callback(invoke_without_command=True, no_args_is_help=True)
def main(
    ctx: typer.Context,
    output: str = typer.Option("auto", "--output", "-o", help="Output format: table|json|csv (default: auto-detect TTY)"),
    profile: Optional[str] = typer.Option(None, "--profile", help="Override active auth profile"),
    no_color: bool = typer.Option(False, "--no-color", help="Disable color output"),
    quiet: bool = typer.Option(False, "--quiet", "-q", help="Suppress non-essential output"),
    verbose: bool = typer.Option(False, "--verbose", "-v", help="Show detailed output"),
    no_interactive: bool = typer.Option(False, "--no-interactive", help="Never prompt; fail with clear error if input missing"),
    version: bool = typer.Option(False, "--version", is_eager=True, help="Show version and exit"),
):
    """[bold]soda[/bold] — unified data quality CLI.

    Run [bold]soda <command> --help[/bold] for details on any command.
    """
    if version:
        console = Console(no_color=no_color)
        console.print(f"soda version [bold]{__version__}[/bold]")
        raise typer.Exit(0)

    ctx.ensure_object(dict)
    ctx.obj = GlobalContext(
        output=output,
        profile=profile,
        no_color=no_color,
        quiet=quiet,
        verbose=verbose,
        no_interactive=no_interactive,
    )


if __name__ == "__main__":
    app()
