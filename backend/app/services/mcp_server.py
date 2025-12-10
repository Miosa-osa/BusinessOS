"""
MCP (Model Context Protocol) Server for Business OS

This allows AI models to interact with external tools and integrations:
- File system access
- Database queries
- API integrations
- Custom business tools
"""

import json
from typing import Any
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, TextContent, Resource

# Initialize MCP server
server = Server("business-os")


# Built-in tools for Business OS
BUILTIN_TOOLS = {
    "search_conversations": {
        "description": "Search through past conversations for relevant context",
        "input_schema": {
            "type": "object",
            "properties": {
                "query": {"type": "string", "description": "Search query"},
                "limit": {"type": "integer", "description": "Max results", "default": 10},
            },
            "required": ["query"],
        },
    },
    "get_project_context": {
        "description": "Get context and details about a specific project",
        "input_schema": {
            "type": "object",
            "properties": {
                "project_name": {"type": "string", "description": "Name of the project"},
            },
            "required": ["project_name"],
        },
    },
    "create_artifact": {
        "description": "Create a new artifact (code, document, etc.)",
        "input_schema": {
            "type": "object",
            "properties": {
                "title": {"type": "string", "description": "Artifact title"},
                "type": {
                    "type": "string",
                    "enum": ["code", "document", "markdown", "html"],
                    "description": "Type of artifact",
                },
                "content": {"type": "string", "description": "Artifact content"},
                "language": {"type": "string", "description": "Programming language (for code)"},
            },
            "required": ["title", "type", "content"],
        },
    },
    "add_to_daily_log": {
        "description": "Add an entry to today's daily log",
        "input_schema": {
            "type": "object",
            "properties": {
                "content": {"type": "string", "description": "Log entry content"},
            },
            "required": ["content"],
        },
    },
    "get_context_profile": {
        "description": "Get a context profile by name (person, business, project)",
        "input_schema": {
            "type": "object",
            "properties": {
                "name": {"type": "string", "description": "Context profile name"},
            },
            "required": ["name"],
        },
    },
}


@server.list_tools()
async def list_tools() -> list[Tool]:
    """List all available tools."""
    tools = []
    for name, config in BUILTIN_TOOLS.items():
        tools.append(
            Tool(
                name=name,
                description=config["description"],
                inputSchema=config["input_schema"],
            )
        )
    return tools


@server.call_tool()
async def call_tool(name: str, arguments: dict[str, Any]) -> list[TextContent]:
    """Execute a tool and return results."""

    if name == "search_conversations":
        # TODO: Implement actual search
        return [TextContent(type="text", text=f"Searching for: {arguments.get('query')}")]

    elif name == "get_project_context":
        # TODO: Implement project lookup
        return [TextContent(type="text", text=f"Getting project: {arguments.get('project_name')}")]

    elif name == "create_artifact":
        # TODO: Implement artifact creation
        return [TextContent(
            type="text",
            text=f"Created {arguments.get('type')} artifact: {arguments.get('title')}"
        )]

    elif name == "add_to_daily_log":
        # TODO: Implement daily log addition
        return [TextContent(type="text", text="Added to daily log")]

    elif name == "get_context_profile":
        # TODO: Implement context profile lookup
        return [TextContent(type="text", text=f"Getting context: {arguments.get('name')}")]

    else:
        return [TextContent(type="text", text=f"Unknown tool: {name}")]


@server.list_resources()
async def list_resources() -> list[Resource]:
    """List available resources (contexts, projects, etc.)."""
    # TODO: Implement dynamic resource listing
    return []


class MCPManager:
    """Manager for MCP server connections and tool routing."""

    def __init__(self):
        self.connected_servers: dict[str, Any] = {}
        self.custom_tools: dict[str, dict] = {}

    def register_tool(self, name: str, description: str, input_schema: dict, handler: callable):
        """Register a custom tool."""
        self.custom_tools[name] = {
            "description": description,
            "input_schema": input_schema,
            "handler": handler,
        }

    async def execute_tool(self, tool_name: str, arguments: dict, db_session=None, user_id=None) -> str:
        """Execute a tool by name with given arguments."""

        # Check custom tools first
        if tool_name in self.custom_tools:
            handler = self.custom_tools[tool_name]["handler"]
            return await handler(arguments, db_session, user_id)

        # Fall back to built-in tools
        if tool_name in BUILTIN_TOOLS:
            result = await call_tool(tool_name, arguments)
            return result[0].text if result else "No result"

        return f"Unknown tool: {tool_name}"

    def get_all_tools(self) -> list[dict]:
        """Get all available tools (built-in + custom)."""
        tools = []

        # Add built-in tools
        for name, config in BUILTIN_TOOLS.items():
            tools.append({
                "name": name,
                "description": config["description"],
                "input_schema": config["input_schema"],
                "source": "builtin",
            })

        # Add custom tools
        for name, config in self.custom_tools.items():
            tools.append({
                "name": name,
                "description": config["description"],
                "input_schema": config["input_schema"],
                "source": "custom",
            })

        return tools


# Global MCP manager instance
mcp_manager = MCPManager()


def get_mcp_manager() -> MCPManager:
    """Get the MCP manager instance."""
    return mcp_manager


async def run_mcp_server():
    """Run the MCP server (for stdio connections)."""
    async with stdio_server() as (read_stream, write_stream):
        await server.run(read_stream, write_stream, server.create_initialization_options())
