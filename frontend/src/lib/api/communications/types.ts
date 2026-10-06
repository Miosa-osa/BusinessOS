export type CommunicationProvider = 'gmail' | 'slack' | 'calendar' | 'whatsapp';

export type CommunicationState =
	| 'triage'
	| 'waiting'
	| 'delegated'
	| 'scheduled'
	| 'resolved'
	| 'archived';

export interface CommunicationItem {
	id?: string;
	provider: CommunicationProvider;
	source_kind: string;
	source_external_id: string;
	title: string;
	preview?: string;
	participant?: string;
	participant_email?: string;
	occurred_at: string;
	state?: CommunicationState;
	assigned_to?: string;
	waiting_on?: string;
	project_id?: string;
	client_id?: string;
	context_id?: string;
	snoozed_until?: string;
}

export interface CommunicationOverview {
	items: CommunicationItem[];
	unrouted_items: CommunicationItem[];
	connection_state: Record<CommunicationProvider, boolean>;
}

export interface SaveCommunicationItemInput {
	provider: CommunicationProvider;
	source_kind: string;
	source_external_id: string;
	state?: CommunicationState;
	assigned_to?: string;
	waiting_on?: string;
	project_id?: string;
	client_id?: string;
	context_id?: string;
	snoozed_until?: string;
}
