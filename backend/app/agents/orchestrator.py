"""
Orchestrator Agent for Business OS

The Orchestrator is the main entry point for all user requests.
It analyzes requests and delegates to specialized sub-agents as needed.
"""

from typing import AsyncGenerator
from uuid import UUID
from sqlalchemy.ext.asyncio import AsyncSession

from app.agents.base import BaseAgent, AgentResponse


ORCHESTRATOR_PROMPT = """You are the Orchestrator - the main AI assistant for Business OS.

## Your Role
You are the central coordinator that handles all user requests. You:
1. Understand what the user needs
2. Decide the best approach to help them
3. Either respond directly OR delegate to a specialized agent
4. Synthesize results and present them to the user

## Available Sub-Agents
You can delegate tasks to these specialized agents:

### DocumentAgent
- Creates business documents: proposals, SOPs, frameworks, agendas, reports
- Use when user asks to CREATE or WRITE any kind of document
- Trigger words: "write", "create", "draft", "make me a", "document", "SOP", "proposal", "framework", "agenda"

### AnalysisAgent
- Analyzes data, situations, and provides insights
- Use for questions about performance, trends, comparisons
- Trigger words: "analyze", "review", "assess", "evaluate", "compare", "what do you think about"

### PlanningAgent
- Helps with planning, prioritization, and strategy
- Use for planning sessions, goal setting, prioritization
- Trigger words: "plan", "prioritize", "schedule", "roadmap", "strategy", "goals", "OKRs"

## Decision Framework

1. **Direct Response** - Handle these yourself:
   - Simple questions
   - General conversation
   - Clarifying questions
   - Quick advice

2. **Delegate to DocumentAgent** when:
   - User explicitly asks to create a document
   - User wants something they can share/present
   - Task requires structured, formal output

3. **Delegate to AnalysisAgent** when:
   - User wants to understand data or metrics
   - User asks for assessment or evaluation
   - Task requires research or investigation

4. **Delegate to PlanningAgent** when:
   - User needs help organizing or prioritizing
   - User wants to set goals or make plans
   - Task involves scheduling or roadmapping

## Response Format

When you decide to delegate, respond with:
```
[DELEGATE:AgentName]
Task: Brief description of what the sub-agent should do
Context: Any relevant context for the sub-agent
```

When responding directly, just provide your response naturally.

## Guidelines
- Be helpful and proactive
- Ask clarifying questions if the request is unclear
- Provide thorough, actionable responses
- Always consider which approach best serves the user
- If in doubt about delegation, handle it yourself and offer to create a document if it would help
"""


class OrchestratorAgent(BaseAgent):
    """
    The Orchestrator agent that coordinates all other agents.
    """

    name = "orchestrator"
    description = "Main coordinator that handles requests and delegates to sub-agents"

    def __init__(
        self,
        db: AsyncSession,
        user_id: UUID,
        conversation_id: UUID | None = None,
        model: str | None = None,
    ):
        super().__init__(db, user_id, conversation_id, model)
        self._sub_agents: dict[str, BaseAgent] = {}

    @property
    def system_prompt(self) -> str:
        return ORCHESTRATOR_PROMPT

    def _get_sub_agent(self, agent_name: str) -> BaseAgent | None:
        """Get or create a sub-agent by name."""
        if agent_name in self._sub_agents:
            return self._sub_agents[agent_name]

        agent_class = None
        if agent_name.lower() == "documentagent":
            from app.agents.document_agent import DocumentAgent
            agent_class = DocumentAgent
        elif agent_name.lower() == "analysisagent":
            from app.agents.analysis_agent import AnalysisAgent
            agent_class = AnalysisAgent
        elif agent_name.lower() == "planningagent":
            from app.agents.planning_agent import PlanningAgent
            agent_class = PlanningAgent

        if agent_class:
            agent = agent_class(
                db=self.db,
                user_id=self.user_id,
                conversation_id=self.conversation_id,
                model=self.model,
            )
            self._sub_agents[agent_name] = agent
            return agent

        return None

    def _parse_delegation(self, response: str) -> tuple[str | None, str | None]:
        """Parse a delegation instruction from the response."""
        if "[DELEGATE:" not in response:
            return None, None

        try:
            # Extract agent name
            start = response.find("[DELEGATE:") + len("[DELEGATE:")
            end = response.find("]", start)
            agent_name = response[start:end].strip()

            # Extract the rest as context
            remaining = response[end + 1:].strip()

            return agent_name, remaining
        except Exception:
            return None, None

    async def run(
        self,
        messages: list[dict[str, str]],
        stream: bool = True,
    ) -> AsyncGenerator[str, None]:
        """
        Run the orchestrator. Will delegate to sub-agents if needed.
        """
        # First, get orchestrator's decision
        full_response = ""
        async for chunk in super().run(messages, stream=False):
            full_response += chunk

        # Check if we should delegate
        agent_name, context = self._parse_delegation(full_response)

        if agent_name:
            # Delegate to sub-agent
            sub_agent = self._get_sub_agent(agent_name)
            if sub_agent:
                # Build context for sub-agent
                sub_messages = messages.copy()
                if context:
                    sub_messages.append({
                        "role": "system",
                        "content": f"Orchestrator context: {context}"
                    })

                # Stream sub-agent response
                async for chunk in sub_agent.run(sub_messages, stream=stream):
                    yield chunk
            else:
                # Unknown agent, return original response
                if stream:
                    yield full_response
                else:
                    yield full_response
        else:
            # No delegation, return orchestrator's response
            if stream:
                # Re-stream the response for consistency
                async for chunk in super().run(messages, stream=True):
                    yield chunk
            else:
                yield full_response
