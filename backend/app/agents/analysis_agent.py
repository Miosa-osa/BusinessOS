"""
Analysis Agent for Business OS

Specialized agent for data analysis and insights:
- Performance analysis, trends, comparisons, evaluations
"""

from app.agents.base import BaseAgent
from app.prompts.analyst import ANALYST_PROMPT


class AnalysisAgent(BaseAgent):
    """
    Agent specialized in analysis and insights.
    """

    name = "analysis"
    description = "Analyzes data and provides insights"

    @property
    def system_prompt(self) -> str:
        return ANALYST_PROMPT
