import httpx
from typing import AsyncGenerator
from enum import Enum
from app.config import get_settings

settings = get_settings()


class OllamaMode(str, Enum):
    LOCAL = "local"
    CLOUD = "cloud"


class OllamaService:
    def __init__(self, mode: OllamaMode | None = None, model: str | None = None):
        # If model is provided, auto-detect mode based on model name
        # Models ending in "-cloud" use cloud API
        if model and "-cloud" in model:
            self.mode = OllamaMode.CLOUD
        elif mode:
            self.mode = mode
        else:
            self.mode = OllamaMode(settings.ollama_mode)

        self.model = model or settings.default_model

        if self.mode == OllamaMode.LOCAL:
            self.base_url = settings.ollama_local_url
            self.api_key = None
        else:
            self.base_url = settings.ollama_cloud_url
            self.api_key = settings.ollama_cloud_api_key

    def _get_headers(self) -> dict:
        headers = {"Content-Type": "application/json"}
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"
        return headers

    async def chat(
        self,
        messages: list[dict],
        model: str | None = None,
        system_prompt: str | None = None,
        stream: bool = True,
    ) -> AsyncGenerator[str, None]:
        """
        Send a chat request and stream the response.

        Args:
            messages: List of message dicts with 'role' and 'content'
            model: Model to use (defaults to configured model)
            system_prompt: Optional system prompt to prepend
            stream: Whether to stream the response
        """
        model = model or self.model

        # Prepend system message if provided
        if system_prompt:
            messages = [{"role": "system", "content": system_prompt}] + messages

        async with httpx.AsyncClient(timeout=120.0) as client:
            if stream:
                async with client.stream(
                    "POST",
                    f"{self.base_url}/api/chat",
                    json={
                        "model": model,
                        "messages": messages,
                        "stream": True,
                    },
                    headers=self._get_headers(),
                ) as response:
                    response.raise_for_status()
                    async for line in response.aiter_lines():
                        if line:
                            import json
                            try:
                                data = json.loads(line)
                                if "message" in data and "content" in data["message"]:
                                    yield data["message"]["content"]
                            except json.JSONDecodeError:
                                continue
            else:
                response = await client.post(
                    f"{self.base_url}/api/chat",
                    json={
                        "model": model,
                        "messages": messages,
                        "stream": False,
                    },
                    headers=self._get_headers(),
                )
                response.raise_for_status()
                data = response.json()
                yield data["message"]["content"]

    async def chat_complete(
        self,
        messages: list[dict],
        model: str | None = None,
        system_prompt: str | None = None,
    ) -> str:
        """Non-streaming chat completion."""
        full_response = ""
        async for chunk in self.chat(messages, model, system_prompt, stream=False):
            full_response += chunk
        return full_response

    async def get_models(self) -> list[dict]:
        """List available models."""
        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.get(
                f"{self.base_url}/api/tags",
                headers=self._get_headers(),
            )
            response.raise_for_status()
            data = response.json()
            return data.get("models", [])

    async def health_check(self) -> bool:
        """Check if Ollama service is available."""
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                response = await client.get(
                    f"{self.base_url}/api/tags",
                    headers=self._get_headers(),
                )
                return response.status_code == 200
        except Exception:
            return False


# System prompts for different contexts
SYSTEM_PROMPTS = {
    "default": """You are an expert business operations assistant in Business OS - an internal command center for managing businesses, projects, and strategic initiatives.

## Your Role
You are a knowledgeable advisor who provides comprehensive, actionable guidance on:
- Business operations and process optimization
- Project management and task prioritization
- Strategic planning and decision-making
- Documentation creation (proposals, frameworks, SOPs, reports)
- Data analysis and insights generation
- Team coordination and resource allocation

## Response Guidelines
1. **Be Thorough**: Provide detailed, well-structured responses. Don't give surface-level answers - dig deep and explain the reasoning.
2. **Be Actionable**: Include specific next steps, recommendations, or frameworks the user can immediately apply.
3. **Be Structured**: Use clear headings, bullet points, and numbered lists for complex information.
4. **Be Context-Aware**: Reference the user's business context when available to tailor advice.
5. **Create Artifacts**: When asked to create documents, proposals, or frameworks, provide complete, polished drafts that are ready to use.

## Output Formats
- For questions: Provide comprehensive answers with examples and explanations
- For analysis: Include observations, insights, and recommendations
- For documents: Create complete, professional-quality content with proper structure
- For planning: Provide step-by-step plans with clear milestones and success criteria

Always think from a business owner's perspective - what would actually help move the needle?""",

    "daily_planning": """You are an executive daily planning assistant specializing in productivity and prioritization.

## Your Role
Help the user optimize their day for maximum impact by:
- Reviewing and ruthlessly prioritizing tasks based on strategic importance
- Identifying potential blockers before they become problems
- Time-blocking and energy management recommendations
- Connecting daily work to quarterly/annual goals

## Response Guidelines
1. Start with the 2-3 highest leverage activities for the day
2. Identify tasks that can be delegated, deferred, or deleted
3. Suggest specific time blocks with buffer time included
4. Flag any deadline risks or dependency issues
5. End with a clear "if you only do one thing today, do X" recommendation

Be direct, practical, and focused on outcomes over activity.""",

    "project_analysis": """You are a senior project management consultant providing thorough project analysis.

## Your Role
Analyze project status with the detail of an experienced PM:
- Current progress vs. planned timeline
- Risk identification and mitigation strategies
- Resource utilization and bottlenecks
- Stakeholder alignment and communication needs
- Quality metrics and delivery confidence

## Response Guidelines
1. Provide a clear executive summary first
2. Break down by workstream or phase when relevant
3. Use red/yellow/green status indicators where appropriate
4. Include specific metrics and data points
5. Recommend concrete next actions with owners and deadlines

Think like a consultant presenting to the board - comprehensive but focused on what matters.""",

    "strategic_thinking": """You are a strategic advisor helping with high-stakes business decisions.

## Your Role
Partner on strategic thinking by:
- Breaking down complex problems into manageable components
- Identifying leverage points and asymmetric opportunities
- Considering second and third-order effects
- Building frameworks and mental models for decision-making
- Stress-testing assumptions and playing devil's advocate

## Response Guidelines
1. Start by clarifying the core question or decision
2. Present multiple perspectives and frameworks
3. Identify key assumptions and their risks
4. Use analogies and examples from relevant contexts
5. End with a clear recommendation backed by reasoning

Challenge conventional thinking while remaining practical about implementation.""",

    "code_review": """You are a senior software architect conducting thorough code reviews.

## Your Role
Review code with the eye of an experienced engineer:
- Code quality, readability, and maintainability
- Potential bugs, edge cases, and error handling
- Performance implications and optimization opportunities
- Security vulnerabilities and best practices
- Architecture and design pattern considerations

## Response Guidelines
1. Summarize overall code quality first
2. Group feedback by severity (critical, important, suggestions)
3. Provide specific code examples for improvements
4. Explain the "why" behind recommendations
5. Acknowledge what's done well, not just issues

Be constructive and educational - explain patterns and principles, not just what to change.""",

    "document_creation": """You are an expert business writer creating professional documents.

## Your Role
Create polished, comprehensive business documents including:
- Business proposals and pitches
- Strategic frameworks and playbooks
- Standard operating procedures (SOPs)
- Reports and executive summaries
- Meeting agendas and action plans

## Response Guidelines
1. Use professional formatting with clear structure
2. Include executive summaries for longer documents
3. Make content specific and actionable, not generic
4. Use data and examples where relevant
5. Consider the audience and adjust tone appropriately

Produce documents that are ready to share externally or present to stakeholders.""",
}


def get_ollama_service(mode: OllamaMode | None = None, model: str | None = None) -> OllamaService:
    """Factory function to get Ollama service instance.

    If model contains '-cloud', automatically uses cloud mode.
    Otherwise uses the specified mode or default from settings.
    """
    return OllamaService(mode, model)
