export type ClientType = 'company' | 'individual';
export type ClientStatus = 'lead' | 'prospect' | 'active' | 'inactive' | 'churned';
export type InteractionType = 'call' | 'email' | 'meeting' | 'note';
export type DealStage = 'qualification' | 'proposal' | 'negotiation' | 'closed_won' | 'closed_lost';

export interface ContactResponse {
  id: string;
  client_id: string;
  name: string;
  email: string | null;
  phone: string | null;
  role: string | null;
  is_primary: boolean;
  notes: string | null;
  created_at: string;
  updated_at: string;
}

export interface InteractionResponse {
  id: string;
  client_id: string;
  contact_id: string | null;
  type: InteractionType;
  subject: string;
  description: string | null;
  outcome: string | null;
  occurred_at: string;
  created_at: string;
}

export interface DealResponse {
  id: string;
  client_id: string;
  name: string;
  value: number;
  stage: DealStage;
  probability: number;
  expected_close_date: string | null;
  notes: string | null;
  created_at: string;
  updated_at: string;
  closed_at: string | null;
}

export interface ClientResponse {
  id: string;
  user_id: string;
  name: string;
  type: ClientType;
  email: string | null;
  phone: string | null;
  website: string | null;
  industry: string | null;
  company_size: string | null;
  address: string | null;
  city: string | null;
  state: string | null;
  zip_code: string | null;
  country: string | null;
  status: ClientStatus;
  source: string | null;
  assigned_to: string | null;
  lifetime_value: number | null;
  tags: string[] | null;
  custom_fields: Record<string, unknown> | null;
  notes: string | null;
  created_at: string;
  updated_at: string;
  last_contacted_at: string | null;
}

export interface ClientListResponse {
  id: string;
  name: string;
  type: ClientType;
  email: string | null;
  phone: string | null;
  status: ClientStatus;
  source: string | null;
  assigned_to: string | null;
  lifetime_value: number | null;
  tags: string[] | null;
  created_at: string;
  last_contacted_at: string | null;
  contacts_count: number;
  interactions_count: number;
  deals_count: number;
  active_deals_value: number;
}

export interface ClientDetailResponse extends ClientResponse {
  contacts: ContactResponse[];
  interactions: InteractionResponse[];
  deals: DealResponse[];
}

export interface CreateClientData {
  name: string;
  type?: ClientType;
  email?: string;
  phone?: string;
  website?: string;
  industry?: string;
  company_size?: string;
  address?: string;
  city?: string;
  state?: string;
  zip_code?: string;
  country?: string;
  status?: ClientStatus;
  source?: string;
  assigned_to?: string;
  tags?: string[];
  custom_fields?: Record<string, unknown>;
  notes?: string;
}

export interface UpdateClientData {
  name?: string;
  type?: ClientType;
  email?: string;
  phone?: string;
  website?: string;
  industry?: string;
  company_size?: string;
  address?: string;
  city?: string;
  state?: string;
  zip_code?: string;
  country?: string;
  status?: ClientStatus;
  source?: string;
  assigned_to?: string;
  lifetime_value?: number;
  tags?: string[];
  custom_fields?: Record<string, unknown>;
  notes?: string;
}

export interface CreateContactData {
  name: string;
  email?: string;
  phone?: string;
  role?: string;
  is_primary?: boolean;
  notes?: string;
}

export interface UpdateContactData {
  name?: string;
  email?: string;
  phone?: string;
  role?: string;
  is_primary?: boolean;
  notes?: string;
}

export interface CreateInteractionData {
  type: InteractionType;
  subject: string;
  description?: string;
  outcome?: string;
  contact_id?: string;
  occurred_at?: string;
}

export interface CreateDealData {
  name: string;
  value?: number;
  stage?: DealStage;
  probability?: number;
  expected_close_date?: string;
  notes?: string;
}

export interface UpdateDealData {
  name?: string;
  value?: number;
  stage?: DealStage;
  probability?: number;
  expected_close_date?: string;
  notes?: string;
}
