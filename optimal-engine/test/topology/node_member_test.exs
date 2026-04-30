defmodule OptimalEngine.Topology.NodeMemberTest do
  use ExUnit.Case, async: false

  alias OptimalEngine.Identity.Principal
  alias OptimalEngine.Topology.{Node, NodeMember}

  setup do
    suffix = System.unique_integer([:positive])

    {:ok, node} =
      Node.upsert(%{slug: "nm-test-#{suffix}", name: "NM Test", kind: :team})

    {:ok, internal_p} =
      Principal.upsert(%{
        id: "user:internal-#{suffix}",
        kind: :user,
        display_name: "Internal #{suffix}"
      })

    {:ok, external_p} =
      Principal.upsert(%{
        id: "user:external-#{suffix}",
        kind: :user,
        display_name: "External #{suffix}"
      })

    {:ok, agent_p} =
      Principal.upsert(%{
        id: "agent:observer-#{suffix}",
        kind: :agent,
        display_name: "Observer #{suffix}"
      })

    {:ok, node: node, internal: internal_p, external: external_p, agent: agent_p}
  end

  test "add + members_of works for all membership kinds",
       %{node: node, internal: internal, external: external, agent: agent} do
    :ok = NodeMember.add(node.id, internal.id, membership: :internal, role: "lead")
    :ok = NodeMember.add(node.id, external.id, membership: :external)
    :ok = NodeMember.add(node.id, agent.id, membership: :observer)

    assert {:ok, members} = NodeMember.members_of(node.id)
    assert length(members) == 3

    by_kind = Enum.group_by(members, & &1.membership)
    assert length(by_kind[:internal]) == 1
    assert length(by_kind[:external]) == 1
    assert length(by_kind[:observer]) == 1

    assert hd(by_kind[:internal]).role == "lead"
  end

  test "members_of with membership filter",
       %{node: node, internal: internal, external: external} do
    :ok = NodeMember.add(node.id, internal.id, membership: :internal)
    :ok = NodeMember.add(node.id, external.id, membership: :external)

    assert {:ok, [member]} = NodeMember.members_of(node.id, membership: :external)
    assert member.principal_id == external.id
  end

  test "nodes_of returns every active node for a principal",
       %{internal: p} do
    suffix = System.unique_integer([:positive])

    {:ok, node_a} = Node.upsert(%{slug: "nm-a-#{suffix}", name: "A", kind: :team})
    {:ok, node_b} = Node.upsert(%{slug: "nm-b-#{suffix}", name: "B", kind: :project})

    :ok = NodeMember.add(node_a.id, p.id, membership: :internal)
    :ok = NodeMember.add(node_b.id, p.id, membership: :owner)

    assert {:ok, memberships} = NodeMember.nodes_of(p.id)
    node_ids = Enum.map(memberships, & &1.node_id) |> Enum.sort()
    assert node_a.id in node_ids
    assert node_b.id in node_ids
  end

  test "remove/3 closes a membership by setting ended_at",
       %{node: node, internal: p} do
    :ok = NodeMember.add(node.id, p.id, membership: :internal)
    assert {:ok, [_active]} = NodeMember.members_of(node.id, membership: :internal)

    :ok = NodeMember.remove(node.id, p.id, membership: :internal)
    assert {:ok, []} = NodeMember.members_of(node.id, membership: :internal)
  end

  test "add is idempotent on (node, principal, membership)",
       %{node: node, internal: p} do
    :ok = NodeMember.add(node.id, p.id, membership: :internal, role: "first")
    :ok = NodeMember.add(node.id, p.id, membership: :internal, role: "second")

    assert {:ok, [single]} = NodeMember.members_of(node.id, membership: :internal)
    # INSERT OR IGNORE keeps the first row
    assert single.role == "first"
  end
end
