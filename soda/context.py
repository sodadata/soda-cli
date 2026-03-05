from dataclasses import dataclass, field
from typing import Optional


@dataclass
class GlobalContext:
    output: str = "auto"
    profile: Optional[str] = None
    no_color: bool = False
    quiet: bool = False
    verbose: bool = False
    no_interactive: bool = False
