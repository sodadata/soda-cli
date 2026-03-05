import typer
from rich.console import Console

from soda.context import GlobalContext
from soda.output import print_success, get_console


def run(ctx: typer.Context, force: bool = False):
    """Scaffold soda.yml, configs/, and contracts/ in the current directory."""
    gctx = ctx.obj or GlobalContext()
    console = get_console(gctx)

    files = [
        ("soda.yml", _SODA_YML),
        ("contracts/", None),
        ("configs/", None),
    ]

    for name, content in files:
        if content is None:
            console.print(f"  [dim]create[/dim]  [cyan]{name}[/cyan]")
        else:
            console.print(f"  [dim]create[/dim]  [cyan]{name}[/cyan]")

    console.print()
    print_success(
        "Project initialized. Edit [bold]soda.yml[/bold] to configure your datasources.",
        gctx,
    )


_SODA_YML = """\
version: 1
profile: default
datasources:
  default: warehouse
  warehouse:
    type: postgres
    config: ./configs/warehouse.yml
contracts:
  directory: ./contracts
cloud:
  organization: my-org
"""
