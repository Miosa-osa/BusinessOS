import { request } from '../base';
import type {
  Node,
  NodeTree,
  NodeDetail,
  NodeActivateResponse,
  CreateNodeData,
  UpdateNodeData
} from './types';

export async function getNodes(includeArchived = false) {
  const params = includeArchived ? '?include_archived=true' : '';
  return request<Node[]>(`/nodes${params}`);
}

export async function getNodeTree(includeArchived = false) {
  const params = includeArchived ? '?include_archived=true' : '';
  return request<NodeTree[]>(`/nodes/tree${params}`);
}

export async function getActiveNode() {
  return request<Node | null>('/nodes/active');
}

export async function getNode(id: string) {
  return request<NodeDetail>(`/nodes/${id}`);
}

export async function createNode(data: CreateNodeData) {
  return request<Node>('/nodes', { method: 'POST', body: data });
}

export async function updateNode(id: string, data: UpdateNodeData) {
  return request<Node>(`/nodes/${id}`, { method: 'PATCH', body: data });
}

export async function activateNode(id: string) {
  return request<NodeActivateResponse>(`/nodes/${id}/activate`, { method: 'POST' });
}

export async function deactivateNode(id: string) {
  return request<Node>(`/nodes/${id}/deactivate`, { method: 'POST' });
}

export async function deleteNode(id: string) {
  return request(`/nodes/${id}`, { method: 'DELETE' });
}

export async function getNodeChildren(id: string, includeArchived = false) {
  const params = includeArchived ? '?include_archived=true' : '';
  return request<Node[]>(`/nodes/${id}/children${params}`);
}

export async function reorderNode(id: string, newOrder: number) {
  return request(`/nodes/${id}/reorder?new_order=${newOrder}`, { method: 'POST' });
}
