"""
Tool Definitions and Handlers (Level 2-3 Context)

Implements hierarchical context fetching via tools that call Go backend.
Tools allow the LLM to fetch specific node/project data on-demand.
"""

import logging
import httpx
from typing import Dict, List, Any, Optional

from config import config

logger = logging.getLogger(__name__)


# =============================================================================
# TOOL DEFINITIONS (for LLM tool calling)
# =============================================================================

TOOL_DEFINITIONS = [
    {
        "type": "function",
        "function": {
            "name": "get_node_context",
            "description": "Get the full context of a specific node including identity, relationships, state, and focus. Use this when user asks about a specific project, team, person, or other node.",
            "parameters": {
                "type": "object",
                "properties": {
                    "node_id": {
                        "type": "string",
                        "description": "The node ID or name to fetch context for"
                    }
                },
                "required": ["node_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_node_children",
            "description": "Get all child nodes of a parent node. Use this when user asks what's inside a node, or what sub-nodes/teams/projects exist under a parent.",
            "parameters": {
                "type": "object",
                "properties": {
                    "node_id": {
                        "type": "string",
                        "description": "The parent node ID"
                    }
                },
                "required": ["node_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "search_nodes",
            "description": "Search for nodes by name, type, or content. Use this when user asks to find something or asks about nodes they don't specify exactly.",
            "parameters": {
                "type": "object",
                "properties": {
                    "query": {
                        "type": "string",
                        "description": "Search query text"
                    },
                    "type": {
                        "type": "string",
                        "description": "Optional node type filter: PROJECT, TEAM, PERSON, ENTITY, etc.",
                        "enum": [
                            "ENTITY", "DEPARTMENT", "TEAM", "PROJECT",
                            "OPERATIONAL", "LEARNING", "PERSON", "PRODUCT",
                            "PARTNERSHIP", "CONTEXT"
                        ]
                    }
                },
                "required": ["query"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_project_tasks",
            "description": "Get tasks for a specific project node. Use this when user asks about tasks, what needs to be done, or project progress.",
            "parameters": {
                "type": "object",
                "properties": {
                    "project_id": {
                        "type": "string",
                        "description": "The project node ID"
                    }
                },
                "required": ["project_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_recent_activity",
            "description": "Get recent activity and updates for a node. Use this when user asks what happened recently, latest updates, or recent changes.",
            "parameters": {
                "type": "object",
                "properties": {
                    "node_id": {
                        "type": "string",
                        "description": "The node ID"
                    },
                    "limit": {
                        "type": "number",
                        "description": "Number of items to return, default 5",
                        "default": 5
                    }
                },
                "required": ["node_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_node_decisions",
            "description": "Get pending decisions and recent decision history for a node. Use this when user asks about decisions, what needs to be decided, or recent choices made.",
            "parameters": {
                "type": "object",
                "properties": {
                    "node_id": {
                        "type": "string",
                        "description": "The node ID"
                    }
                },
                "required": ["node_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "activate_node",
            "description": "Activate (open/switch to) a specific node in the UI. Use this when user asks to open, switch to, navigate to, or activate a node/project/team.",
            "parameters": {
                "type": "object",
                "properties": {
                    "node_id": {
                        "type": "string",
                        "description": "The node ID to activate"
                    }
                },
                "required": ["node_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "list_all_nodes",
            "description": "Get a list of all available nodes/apps/projects. Use this when user asks what's available, what can they open, or for a list of nodes.",
            "parameters": {
                "type": "object",
                "properties": {}
            }
        }
    }
]


# =============================================================================
# TOOL HANDLERS (execute tool calls by calling Go backend)
# =============================================================================

async def execute_tool(tool_name: str, arguments: Dict[str, Any]) -> Dict[str, Any]:
    """
    Execute a tool call by routing to the appropriate handler.

    Args:
        tool_name: Name of the tool to execute
        arguments: Tool arguments dict

    Returns:
        Tool execution result dict
    """
    logger.info(f"[Tools] Executing tool: {tool_name} with args: {arguments}")

    handlers = {
        "get_node_context": handle_get_node_context,
        "get_node_children": handle_get_node_children,
        "search_nodes": handle_search_nodes,
        "get_project_tasks": handle_get_project_tasks,
        "get_recent_activity": handle_get_recent_activity,
        "get_node_decisions": handle_get_node_decisions,
        "activate_node": handle_activate_node,
        "list_all_nodes": handle_list_all_nodes,
    }

    handler = handlers.get(tool_name)
    if not handler:
        logger.error(f"[Tools] Unknown tool: {tool_name}")
        return {"error": f"Unknown tool: {tool_name}"}

    try:
        result = await handler(arguments)
        logger.info(f"[Tools] Tool {tool_name} completed successfully")
        return result
    except Exception as e:
        logger.error(f"[Tools] Error executing {tool_name}: {e}")
        return {"error": str(e)}


# Individual tool handlers

async def handle_get_node_context(args: Dict[str, Any]) -> Dict[str, Any]:
    """Fetch full context for a specific node."""
    node_id = args.get("node_id")

    async with httpx.AsyncClient() as client:
        response = await client.get(
            f"{config.go_backend_url}/api/nodes/{node_id}/context",
            timeout=3.0,
        )

        if response.status_code == 200:
            return response.json()
        else:
            return {
                "error": f"Failed to fetch node context: {response.status_code}",
                "node_id": node_id
            }


async def handle_get_node_children(args: Dict[str, Any]) -> Dict[str, Any]:
    """Fetch child nodes of a parent node."""
    node_id = args.get("node_id")

    async with httpx.AsyncClient() as client:
        response = await client.get(
            f"{config.go_backend_url}/api/nodes/{node_id}/children",
            timeout=3.0,
        )

        if response.status_code == 200:
            return response.json()
        else:
            return {
                "error": f"Failed to fetch node children: {response.status_code}",
                "node_id": node_id
            }


async def handle_search_nodes(args: Dict[str, Any]) -> Dict[str, Any]:
    """Search for nodes by query and optional type filter."""
    query = args.get("query")
    node_type = args.get("type")

    params = {"q": query}
    if node_type:
        params["type"] = node_type

    async with httpx.AsyncClient() as client:
        response = await client.get(
            f"{config.go_backend_url}/api/nodes/search",
            params=params,
            timeout=3.0,
        )

        if response.status_code == 200:
            return response.json()
        else:
            return {
                "error": f"Failed to search nodes: {response.status_code}",
                "query": query
            }


async def handle_get_project_tasks(args: Dict[str, Any]) -> Dict[str, Any]:
    """Fetch tasks for a project node."""
    project_id = args.get("project_id")

    async with httpx.AsyncClient() as client:
        response = await client.get(
            f"{config.go_backend_url}/api/projects/{project_id}/tasks",
            timeout=3.0,
        )

        if response.status_code == 200:
            return response.json()
        else:
            return {
                "error": f"Failed to fetch project tasks: {response.status_code}",
                "project_id": project_id
            }


async def handle_get_recent_activity(args: Dict[str, Any]) -> Dict[str, Any]:
    """Fetch recent activity for a node."""
    node_id = args.get("node_id")
    limit = args.get("limit", 5)

    async with httpx.AsyncClient() as client:
        response = await client.get(
            f"{config.go_backend_url}/api/nodes/{node_id}/activity",
            params={"limit": limit},
            timeout=3.0,
        )

        if response.status_code == 200:
            return response.json()
        else:
            return {
                "error": f"Failed to fetch node activity: {response.status_code}",
                "node_id": node_id
            }


async def handle_get_node_decisions(args: Dict[str, Any]) -> Dict[str, Any]:
    """Fetch pending decisions and recent decision history for a node."""
    node_id = args.get("node_id")

    async with httpx.AsyncClient() as client:
        response = await client.get(
            f"{config.go_backend_url}/api/nodes/{node_id}/decisions",
            timeout=3.0,
        )

        if response.status_code == 200:
            return response.json()
        else:
            return {
                "error": f"Failed to fetch node decisions: {response.status_code}",
                "node_id": node_id
            }


async def handle_activate_node(args: Dict[str, Any]) -> Dict[str, Any]:
    """Activate (switch to) a specific node in the UI."""
    node_id = args.get("node_id")

    async with httpx.AsyncClient() as client:
        response = await client.post(
            f"{config.go_backend_url}/api/nodes/{node_id}/activate",
            timeout=3.0,
        )

        if response.status_code == 200:
            return {
                "success": True,
                "message": f"Node {node_id} activated successfully",
                "node_id": node_id
            }
        else:
            return {
                "error": f"Failed to activate node: {response.status_code}",
                "node_id": node_id
            }


async def handle_list_all_nodes(args: Dict[str, Any]) -> Dict[str, Any]:
    """Get list of all available nodes."""
    async with httpx.AsyncClient() as client:
        response = await client.get(
            f"{config.go_backend_url}/api/nodes",
            timeout=3.0,
        )

        if response.status_code == 200:
            return response.json()
        else:
            return {
                "error": f"Failed to fetch nodes list: {response.status_code}"
            }
