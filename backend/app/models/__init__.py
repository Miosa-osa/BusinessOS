# User model is managed by Better Auth (table: "user")
from app.models.conversation import Conversation, Message, ConversationTag
from app.models.project import Project, ProjectNote, ProjectConversation
from app.models.context import Context
from app.models.daily_log import DailyLog
from app.models.artifact import Artifact, ArtifactVersion
from app.models.node import Node, NodeMetric
from app.models.team_member import TeamMember, TeamMemberActivity
from app.models.task import Task, FocusItem
from app.models.user_settings import UserSettings
from app.models.client import Client, ClientContact, ClientInteraction, ClientDeal

__all__ = [
    "Conversation",
    "Message",
    "ConversationTag",
    "Project",
    "ProjectNote",
    "ProjectConversation",
    "Context",
    "DailyLog",
    "Artifact",
    "ArtifactVersion",
    "Node",
    "NodeMetric",
    "TeamMember",
    "TeamMemberActivity",
    "Task",
    "FocusItem",
    "UserSettings",
    "Client",
    "ClientContact",
    "ClientInteraction",
    "ClientDeal",
]
