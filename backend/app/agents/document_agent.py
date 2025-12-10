"""
Document Agent for Business OS

Specialized agent for creating business documents:
- Proposals, SOPs, Frameworks, Agendas, Reports, Plans
"""

from app.agents.base import BaseAgent
from app.prompts.document_creator import DOCUMENT_CREATOR_PROMPT


class DocumentAgent(BaseAgent):
    """
    Agent specialized in creating business documents.
    """

    name = "document"
    description = "Creates professional business documents"

    @property
    def system_prompt(self) -> str:
        return DOCUMENT_CREATOR_PROMPT
