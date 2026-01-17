"""
User Context Management (Level 1)

Fetches user context from Go backend for session initialization.
"""

import logging
import httpx
from typing import Dict, Optional

from config import config

logger = logging.getLogger(__name__)


async def fetch_user_context(user_id: str, session_token: Optional[str] = None) -> Dict:
    """
    Fetch user context from Go backend.

    This provides Level 1 context: user name, workspace, current node, recent nodes.

    Args:
        user_id: The LiveKit participant identity (user ID from token)
        session_token: Optional auth token (not required for voice agent endpoint)

    Returns:
        dict with user context:
        {
            "name": "User Name",
            "email": "user@example.com",
            "workspace": "Workspace Name",
            "current_node": "Node Name",
            "recent_nodes": ["Node1", "Node2", "Node3"],
            "recent_activity": "Description of recent activity"
        }
    """
    try:
        async with httpx.AsyncClient() as client:
            headers = {}
            if session_token:
                headers["Authorization"] = f"Bearer {session_token}"

            # Parse user_id if it has "user-" prefix
            # LiveKit identity format: "user-{first_8_chars_of_uuid}"
            user_id_clean = user_id.replace("user-", "") if user_id.startswith("user-") else user_id

            url = f"{config.go_backend_url}/api/voice/user-context/{user_id_clean}"

            logger.info(f"[Context] Fetching user context from: {url}")

            response = await client.get(
                url,
                headers=headers,
                timeout=2.0,  # Fast timeout to not delay voice session
            )

            if response.status_code == 200:
                context = response.json()
                logger.info(
                    f"[Context] Fetched context for user: {context.get('name', 'Unknown')}"
                )
                return context
            else:
                logger.warning(
                    f"[Context] Failed to fetch user context: {response.status_code}"
                )
                return {"name": "User"}

    except httpx.TimeoutException:
        logger.error("[Context] Timeout fetching user context")
        return {"name": "User"}
    except Exception as e:
        logger.error(f"[Context] Error fetching user context: {e}")
        return {"name": "User"}


def extract_user_id_from_identity(identity: str) -> str:
    """
    Extract clean user ID from LiveKit participant identity.

    LiveKit identity format: "user-{first_8_chars_of_uuid}"
    We need to extract the full UUID.

    Args:
        identity: LiveKit participant identity string

    Returns:
        Clean user ID (without "user-" prefix if present)
    """
    if identity.startswith("user-"):
        return identity[5:]  # Remove "user-" prefix
    return identity
