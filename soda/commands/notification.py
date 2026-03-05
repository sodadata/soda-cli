from typing import Optional

import typer
from rich.prompt import Prompt

from soda.context import GlobalContext
from soda.output import render, print_success, print_error, get_console, output_option
from soda.mock import NOTIFICATIONS, CHANNELS

app = typer.Typer(help="Manage notifications and channels.", no_args_is_help=True)

channel_app = typer.Typer(help="Manage notification channels.", no_args_is_help=True)
app.add_typer(channel_app, name="channel")

TRIGGER_CHOICES = ["check-failure", "incident-opened", "incident-closed", "monitor-alert", "check-recovery"]


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


@app.command("list")
def list_cmd(
    ctx: typer.Context,
    output: str = output_option(),
    channel: Optional[str] = typer.Option(None, "--channel", help="Filter by channel ID"),
    dataset: Optional[str] = typer.Option(None, "--dataset", help="Filter by dataset ID"),
):
    """List notification rules."""
    gctx = _get_ctx(ctx)
    if output != "auto": gctx.output = output
    data = NOTIFICATIONS
    if channel:
        data = [n for n in data if n["channel"] == channel]
    if dataset:
        data = [n for n in data if dataset in n["dataset"]]
    render(data, ["id", "channel", "trigger", "dataset", "status"], gctx, title="Notifications")


@app.command("add")
def add(
    ctx: typer.Context,
    channel: Optional[str] = typer.Option(None, "--channel", help="Channel ID"),
    trigger: Optional[str] = typer.Option(None, "--trigger", help=f"Trigger: {', '.join(TRIGGER_CHOICES)}"),
    dataset: Optional[str] = typer.Option(None, "--dataset", help="Scope to specific dataset (optional)"),
):
    """Add a notification rule."""
    gctx = _get_ctx(ctx)
    console = get_console(gctx)

    if channel is None:
        if gctx.no_interactive:
            print_error("--channel is required in non-interactive mode", gctx)
        channel_names = [c["name"] for c in CHANNELS]
        channel = Prompt.ask("Channel", choices=channel_names)

    if trigger is None:
        if gctx.no_interactive:
            print_error("--trigger is required in non-interactive mode", gctx)
        trigger = Prompt.ask("Trigger event", choices=TRIGGER_CHOICES, default="check-failure")

    scope = f" on [bold]{dataset}[/bold]" if dataset else " (all datasets)"
    print_success(
        f"Notification added: [bold]{trigger}[/bold] → [bold]{channel}[/bold]{scope}",
        gctx,
    )


@app.command("update")
def update(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Notification ID"),
    channel: Optional[str] = typer.Option(None, "--channel"),
    trigger: Optional[str] = typer.Option(None, "--trigger"),
):
    """Update a notification rule."""
    gctx = _get_ctx(ctx)
    print_success(f"Notification [bold]{id}[/bold] updated.", gctx)


@app.command("delete")
def delete(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Notification ID"),
):
    """Delete a notification rule."""
    gctx = _get_ctx(ctx)
    print_success(f"Notification [bold]{id}[/bold] deleted.", gctx)


# ── channel subcommands ───────────────────────────────────────────────────────

@channel_app.command("list")
def channel_list(
    ctx: typer.Context,
    output: str = output_option(),
):
    """List notification channels."""
    gctx = _get_ctx(ctx)
    if output != "auto": gctx.output = output
    render(CHANNELS, ["id", "name", "type", "status"], gctx, title="Channels")


@channel_app.command("add")
def channel_add(
    ctx: typer.Context,
    type: str = typer.Argument(..., help="Channel type: slack|teams|webhook"),
    name: Optional[str] = typer.Option(None, "--name", help="Channel name"),
):
    """Add a notification channel."""
    gctx = _get_ctx(ctx)
    console = get_console(gctx)

    ch_name = name or f"my-{type}"

    if type == "slack":
        if gctx.no_interactive:
            print_error("Slack setup requires interactive OAuth flow. Use --type webhook for non-interactive setup.", gctx)
        console.print("[dim]Opening Slack OAuth flow...[/dim]")
        console.print("[dim](Mock) Slack connected.[/dim]")
    elif type == "teams":
        if gctx.no_interactive:
            print_error("Teams setup requires interactive OAuth flow.", gctx)
        console.print("[dim]Opening Microsoft Teams OAuth flow...[/dim]")
    elif type == "webhook":
        if gctx.no_interactive:
            print_error("--url is required for webhook in non-interactive mode", gctx)
        url = Prompt.ask("Webhook URL")
        console.print(f"[dim]Testing webhook {url}...[/dim]")
    else:
        print_error(f"Unknown channel type: {type}. Valid: slack, teams, webhook", gctx)

    print_success(f"Channel [bold]{ch_name}[/bold] ({type}) added.", gctx)


@channel_app.command("delete")
def channel_delete(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Channel ID"),
):
    """Delete a notification channel."""
    gctx = _get_ctx(ctx)
    print_success(f"Channel [bold]{id}[/bold] deleted.", gctx)


@channel_app.command("test")
def channel_test(
    ctx: typer.Context,
    id: str = typer.Argument(..., help="Channel ID"),
):
    """Send a test message to a channel."""
    gctx = _get_ctx(ctx)
    ch = next((c for c in CHANNELS if c["id"] == id), None)
    name = ch["name"] if ch else id
    get_console(gctx).print(f"[dim]Sending test message to [bold]{name}[/bold]...[/dim]")
    print_success(f"Test message sent to [bold]{name}[/bold].", gctx)
