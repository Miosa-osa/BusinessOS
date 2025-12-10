"""
Planning Agent for Business OS

Specialized agent for planning and strategy:
- Project plans, roadmaps, prioritization, goals, OKRs
"""

from app.agents.base import BaseAgent
from app.prompts.planner import PLANNER_PROMPT


class PlanningAgent(BaseAgent):
    """
    Agent specialized in planning and strategy.
    """

    name = "planning"
    description = "Helps with planning, prioritization, and strategy"

    @property
    def system_prompt(self) -> str:
        return PLANNER_PROMPT
