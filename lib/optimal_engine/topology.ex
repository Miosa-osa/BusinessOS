defmodule OptimalEngine.Topology do
  @moduledoc """
  Top-level facade for the organizational topology layer.

  A tenant's workspace is the first-class model of the company it represents:
  nodes (organizational units), people (principal.kind=:user), agents
  (principal.kind=:agent), skills (capabilities), memberships (who's in
  what node), and tool integrations (the translation layer between the
  human and agent layers).

  This module is a thin facade over:

    * `OptimalEngine.Topology.Node`             — organizational units
    * `OptimalEngine.Topology.NodeMember`       — principal ↔ node memberships
    * `OptimalEngine.Topology.Skill`            — capability registry
    * `OptimalEngine.Topology.PrincipalSkill`   — principal ↔ skill grants

  See `docs/architecture/WORKSPACE.md` for the model.
  """

  alias OptimalEngine.Topology.{Node, NodeMember, PrincipalSkill, Skill}

  # ── Nodes ────────────────────────────────────────────────────────────────

  defdelegate create_node(attrs), to: Node, as: :upsert
  defdelegate get_node(id, tenant_id), to: Node, as: :get
  defdelegate get_node_by_slug(slug, tenant_id), to: Node, as: :get_by_slug
  defdelegate list_nodes(opts), to: Node, as: :list
  defdelegate children(node_id, opts), to: Node
  defdelegate ancestors(node_id, opts), to: Node

  # ── Memberships ──────────────────────────────────────────────────────────

  defdelegate add_member(node_id, principal_id, opts), to: NodeMember, as: :add
  defdelegate remove_member(node_id, principal_id, opts), to: NodeMember, as: :remove
  defdelegate members_of(node_id, opts), to: NodeMember
  defdelegate nodes_of(principal_id, opts), to: NodeMember

  # ── Skills ───────────────────────────────────────────────────────────────

  defdelegate create_skill(attrs), to: Skill, as: :upsert
  defdelegate get_skill(id, tenant_id), to: Skill, as: :get
  defdelegate list_skills(opts), to: Skill, as: :list

  defdelegate grant_skill(principal_id, skill_id, opts), to: PrincipalSkill, as: :grant
  defdelegate revoke_skill(principal_id, skill_id, opts), to: PrincipalSkill, as: :revoke
  defdelegate skills_of(principal_id, opts), to: PrincipalSkill
  defdelegate principals_with_skill(skill_id, opts), to: PrincipalSkill
end
