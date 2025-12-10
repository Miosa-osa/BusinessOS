# Agent Prompts for Business OS
# Each prompt is designed for a specific agent role

from .business_ops import BUSINESS_OPS_PROMPT
from .document_creator import DOCUMENT_CREATOR_PROMPT
from .analyst import ANALYST_PROMPT
from .planner import PLANNER_PROMPT
from .default import DEFAULT_PROMPT

__all__ = [
    "BUSINESS_OPS_PROMPT",
    "DOCUMENT_CREATOR_PROMPT",
    "ANALYST_PROMPT",
    "PLANNER_PROMPT",
    "DEFAULT_PROMPT",
]
