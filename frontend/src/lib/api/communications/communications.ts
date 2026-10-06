import { request } from '../base';
import type { CommunicationOverview, SaveCommunicationItemInput } from './types';

const workspaceBase = (workspaceId: string) =>
	`/communications/workspaces/${encodeURIComponent(workspaceId)}`;

export function getCommunicationOverview(workspaceId: string) {
	return request<CommunicationOverview>(`${workspaceBase(workspaceId)}/overview`, {
		skipCache: true
	});
}

export function saveCommunicationItem(workspaceId: string, input: SaveCommunicationItemInput) {
	return request<{ id: string; state: string }>(`${workspaceBase(workspaceId)}/items`, {
		method: 'PUT',
		body: input
	});
}

export function removeCommunicationItem(workspaceId: string, itemId: string) {
	return request<void>(`${workspaceBase(workspaceId)}/items/${encodeURIComponent(itemId)}`, {
		method: 'DELETE'
	});
}
