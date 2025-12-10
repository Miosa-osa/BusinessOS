"""
Base Agent class for Business OS

All agents inherit from this base class which provides:
- Tool execution
- LLM interaction
- Response formatting
"""

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Any, AsyncGenerator
from uuid import UUID
from sqlalchemy.ext.asyncio import AsyncSession

from app.services.ollama import OllamaService, get_ollama_service
from app.tools.artifact_tools import (
    create_artifact_tool,
    read_artifact_tool,
    update_artifact_tool,
    list_artifacts_tool,
    TOOL_DEFINITIONS,
)


@dataclass
class AgentResponse:
    """Response from an agent."""
    content: str
    artifacts: list[dict[str, Any]] = field(default_factory=list)
    tool_calls: list[dict[str, Any]] = field(default_factory=list)
    delegated_to: str | None = None
    metadata: dict[str, Any] = field(default_factory=dict)


class BaseAgent(ABC):
    """Base class for all agents."""

    name: str = "base"
    description: str = "Base agent"

    def __init__(
        self,
        db: AsyncSession,
        user_id: UUID,
        conversation_id: UUID | None = None,
        model: str | None = None,
    ):
        self.db = db
        self.user_id = user_id
        self.conversation_id = conversation_id
        self.ollama = get_ollama_service(model=model)
        self.model = model

    @property
    @abstractmethod
    def system_prompt(self) -> str:
        """Return the system prompt for this agent."""
        pass

    @property
    def tools(self) -> list[dict]:
        """Return the tools available to this agent."""
        return TOOL_DEFINITIONS

    async def execute_tool(self, tool_name: str, arguments: dict[str, Any]) -> dict[str, Any]:
        """Execute a tool and return the result."""
        if tool_name == "create_artifact":
            return await create_artifact_tool(
                db=self.db,
                user_id=self.user_id,
                conversation_id=self.conversation_id,
                title=arguments["title"],
                content=arguments["content"],
                artifact_type=arguments["artifact_type"],
                summary=arguments.get("summary"),
            )
        elif tool_name == "read_artifact":
            result = await read_artifact_tool(
                db=self.db,
                user_id=self.user_id,
                artifact_id=arguments["artifact_id"],
            )
            return result or {"error": "Artifact not found"}
        elif tool_name == "update_artifact":
            result = await update_artifact_tool(
                db=self.db,
                user_id=self.user_id,
                artifact_id=arguments["artifact_id"],
                title=arguments.get("title"),
                content=arguments.get("content"),
            )
            return result or {"error": "Artifact not found"}
        elif tool_name == "list_artifacts":
            return {
                "artifacts": await list_artifacts_tool(
                    db=self.db,
                    user_id=self.user_id,
                    artifact_type=arguments.get("artifact_type"),
                    conversation_id=self.conversation_id,
                    limit=arguments.get("limit", 10),
                )
            }
        else:
            return {"error": f"Unknown tool: {tool_name}"}

    async def run(
        self,
        messages: list[dict[str, str]],
        stream: bool = True,
    ) -> AsyncGenerator[str, None]:
        """
        Run the agent with the given messages.
        Yields text chunks for streaming responses.
        """
        # Build messages with system prompt
        full_messages = [{"role": "system", "content": self.system_prompt}] + messages

        # For now, use simple chat without tool calling
        # Tool calling will be added when Ollama supports it properly
        async for chunk in self.ollama.chat(
            messages=full_messages,
            model=self.model,
            stream=stream,
        ):
            yield chunk

    async def run_with_tools(
        self,
        messages: list[dict[str, str]],
    ) -> AgentResponse:
        """
        Run the agent with tool support (non-streaming).
        Returns a complete response with any artifacts created.
        """
        # Build messages with system prompt
        full_messages = [{"role": "system", "content": self.system_prompt}] + messages

        # Get response from LLM
        response = await self.ollama.chat_complete(
            messages=full_messages,
            model=self.model,
        )

        # For now, return simple response
        # Tool calling parsing will be added when needed
        return AgentResponse(
            content=response,
            artifacts=[],
            tool_calls=[],
        )
