"""
Typer/Rich help formatting overrides for the Soda CLI aesthetic.

Must be imported BEFORE any Typer app is constructed so the monkey-patches
land before click builds the formatter objects.
"""
import typer.rich_utils as _ru
from rich.padding import Padding as _Padding
from rich.panel import Panel as _Panel
from rich import box as _box

# ── Color palette ────────────────────────────────────────────────────────────
# No cyan/green/yellow in the help text — bold-only for names, dim for meta.
_ru.STYLE_OPTION = "bold"
_ru.STYLE_SWITCH = "bold"
_ru.STYLE_NEGATIVE_OPTION = "dim"
_ru.STYLE_NEGATIVE_SWITCH = "dim"
_ru.STYLE_METAVAR = "dim"
_ru.STYLE_METAVAR_SEPARATOR = "dim"
_ru.STYLE_USAGE = "dim"
_ru.STYLE_USAGE_COMMAND = "bold"
_ru.STYLE_COMMANDS_TABLE_FIRST_COLUMN = "bold"
_ru.STYLE_HELPTEXT_FIRST_LINE = ""
_ru.STYLE_HELPTEXT = "dim"

# ── Remove section panel titles ───────────────────────────────────────────────
_ru.OPTIONS_PANEL_TITLE = ""   # empty string → no visible title label
_ru.COMMANDS_PANEL_TITLE = ""
_ru.ARGUMENTS_PANEL_TITLE = ""

# ── Replace panels with borderless padding blocks ─────────────────────────────
# rich.panel.Panel requires a non-None box object — box=None crashes in Rich 14.
# Instead, return a Padding renderable so sections still have breathing room
# but no borders, no rounded corners, and no section titles.
_ERRORS_TITLE = _ru.ERRORS_PANEL_TITLE  # "Error" by default


def _flat_panel(renderable=None, *args, title=None, border_style=None, box=_box.ROUNDED, **kwargs):
    if title == _ERRORS_TITLE:
        # Errors: straight-corner box, dim border, keep the "Error" label
        return _Panel(renderable, box=_box.SQUARE, title=title, border_style="dim")
    # Help sections: drop the panel entirely — padded content with a blank line below
    return _Padding(renderable, pad=(0, 1, 1, 1))


_ru.Panel = _flat_panel
