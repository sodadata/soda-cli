import typer
from rich.syntax import Syntax

from soda.context import GlobalContext
from soda.output import get_console

app = typer.Typer(help="Generate shell completion scripts.", no_args_is_help=True)


def _get_ctx(ctx: typer.Context) -> GlobalContext:
    return ctx.obj or GlobalContext()


BASH_COMPLETION = """\
# Add to ~/.bashrc or ~/.bash_profile:
# eval "$(soda completion bash)"
_soda_completion() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    COMPREPLY=( $(compgen -W "auth init dashboard datasource ds contract job scan results dataset incident notification role users completion version" -- "$cur") )
}
complete -F _soda_completion soda
"""

ZSH_COMPLETION = """\
# Add to ~/.zshrc:
# eval "$(soda completion zsh)"
autoload -U compinit && compinit
_soda() {
  local -a commands
  commands=(auth init dashboard datasource ds contract job scan results dataset incident notification role users completion version)
  _describe 'command' commands
}
compdef _soda soda
"""

FISH_COMPLETION = """\
# Add to ~/.config/fish/completions/soda.fish
complete -c soda -f -a "auth init dashboard datasource ds contract job scan results dataset incident notification role users completion version"
"""

SCRIPTS = {"bash": BASH_COMPLETION, "zsh": ZSH_COMPLETION, "fish": FISH_COMPLETION}


@app.command("bash")
def bash(ctx: typer.Context):
    """Print bash completion script."""
    print(BASH_COMPLETION, end="")


@app.command("zsh")
def zsh(ctx: typer.Context):
    """Print zsh completion script."""
    print(ZSH_COMPLETION, end="")


@app.command("fish")
def fish(ctx: typer.Context):
    """Print fish completion script."""
    print(FISH_COMPLETION, end="")
