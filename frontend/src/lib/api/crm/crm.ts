// CRM API Client
import { request } from '../base';
import type {
  Company,
  CompaniesResponse,
  CreateCompanyData,
  UpdateCompanyData,
  Pipeline,
  PipelinesResponse,
  CreatePipelineData,
  UpdatePipelineData,
  PipelineStage,
  StagesResponse,
  CreateStageData,
  UpdateStageData,
  Deal,
  DealsResponse,
  DealStats,
  CreateDealData,
  UpdateDealData,
  CRMActivity,
  ActivitiesResponse,
  CreateActivityData,
  ContactsResponse,
  LinkContactData,
  ContactCompanyRelation
} from './types';

// ============================================================================
// Companies
// ============================================================================

export async function getCompanies(filters?: {
  industry?: string;
  lifecycle_stage?: string;
  limit?: number;
  offset?: number;
}): Promise<CompaniesResponse> {
  const params = new URLSearchParams();
  if (filters?.industry) params.set('industry', filters.industry);
  if (filters?.lifecycle_stage) params.set('lifecycle_stage', filters.lifecycle_stage);
  if (filters?.limit) params.set('limit', filters.limit.toString());
  if (filters?.offset) params.set('offset', filters.offset.toString());
  const query = params.toString();
  return request<CompaniesResponse>(`/crm/companies${query ? `?${query}` : ''}`);
}

export async function getCompany(id: string): Promise<Company> {
  return request<Company>(`/crm/companies/${id}`);
}

export async function createCompany(data: CreateCompanyData): Promise<Company> {
  return request<Company>('/crm/companies', { method: 'POST', body: data });
}

export async function updateCompany(id: string, data: UpdateCompanyData): Promise<Company> {
  return request<Company>(`/crm/companies/${id}`, { method: 'PUT', body: data });
}

export async function deleteCompany(id: string): Promise<void> {
  return request<void>(`/crm/companies/${id}`, { method: 'DELETE' });
}

export async function searchCompanies(query: string, limit?: number): Promise<CompaniesResponse> {
  const params = new URLSearchParams();
  params.set('q', query);
  if (limit) params.set('limit', limit.toString());
  return request<CompaniesResponse>(`/crm/companies/search?${params.toString()}`);
}

// ============================================================================
// Contact-Company Relations
// ============================================================================

export async function getCompanyContacts(companyId: string): Promise<ContactsResponse> {
  return request<ContactsResponse>(`/crm/companies/${companyId}/contacts`);
}

export async function linkContactToCompany(companyId: string, data: LinkContactData): Promise<ContactCompanyRelation> {
  return request<ContactCompanyRelation>(`/crm/companies/${companyId}/contacts`, { method: 'POST', body: data });
}

export async function unlinkContactFromCompany(companyId: string, relationId: string): Promise<void> {
  return request<void>(`/crm/companies/${companyId}/contacts/${relationId}`, { method: 'DELETE' });
}

// ============================================================================
// Pipelines
// ============================================================================

export async function getPipelines(): Promise<PipelinesResponse> {
  return request<PipelinesResponse>('/crm/pipelines');
}

export async function getPipeline(id: string): Promise<Pipeline> {
  return request<Pipeline>(`/crm/pipelines/${id}`);
}

export async function createPipeline(data: CreatePipelineData): Promise<Pipeline> {
  return request<Pipeline>('/crm/pipelines', { method: 'POST', body: data });
}

export async function updatePipeline(id: string, data: UpdatePipelineData): Promise<Pipeline> {
  return request<Pipeline>(`/crm/pipelines/${id}`, { method: 'PUT', body: data });
}

export async function deletePipeline(id: string): Promise<void> {
  return request<void>(`/crm/pipelines/${id}`, { method: 'DELETE' });
}

// ============================================================================
// Pipeline Stages
// ============================================================================

export async function getPipelineStages(pipelineId: string): Promise<StagesResponse> {
  return request<StagesResponse>(`/crm/pipelines/${pipelineId}/stages`);
}

export async function createPipelineStage(pipelineId: string, data: CreateStageData): Promise<PipelineStage> {
  return request<PipelineStage>(`/crm/pipelines/${pipelineId}/stages`, { method: 'POST', body: data });
}

export async function updatePipelineStage(pipelineId: string, stageId: string, data: UpdateStageData): Promise<PipelineStage> {
  return request<PipelineStage>(`/crm/pipelines/${pipelineId}/stages/${stageId}`, { method: 'PUT', body: data });
}

export async function deletePipelineStage(pipelineId: string, stageId: string): Promise<void> {
  return request<void>(`/crm/pipelines/${pipelineId}/stages/${stageId}`, { method: 'DELETE' });
}

export async function reorderPipelineStages(pipelineId: string, stageOrders: { id: string; position: number }[]): Promise<void> {
  return request<void>(`/crm/pipelines/${pipelineId}/stages/reorder`, { method: 'POST', body: { stage_orders: stageOrders } });
}

// ============================================================================
// Deals
// ============================================================================

export async function getDeals(filters?: {
  pipeline_id?: string;
  stage_id?: string;
  status?: string;
  owner_id?: string;
  limit?: number;
  offset?: number;
}): Promise<DealsResponse> {
  const params = new URLSearchParams();
  if (filters?.pipeline_id) params.set('pipeline_id', filters.pipeline_id);
  if (filters?.stage_id) params.set('stage_id', filters.stage_id);
  if (filters?.status) params.set('status', filters.status);
  if (filters?.owner_id) params.set('owner_id', filters.owner_id);
  if (filters?.limit) params.set('limit', filters.limit.toString());
  if (filters?.offset) params.set('offset', filters.offset.toString());
  const query = params.toString();
  return request<DealsResponse>(`/crm/deals${query ? `?${query}` : ''}`);
}

export async function getDeal(id: string): Promise<Deal> {
  return request<Deal>(`/crm/deals/${id}`);
}

export async function createDeal(data: CreateDealData): Promise<Deal> {
  return request<Deal>('/crm/deals', { method: 'POST', body: data });
}

export async function updateDeal(id: string, data: UpdateDealData): Promise<Deal> {
  return request<Deal>(`/crm/deals/${id}`, { method: 'PUT', body: data });
}

export async function deleteDeal(id: string): Promise<void> {
  return request<void>(`/crm/deals/${id}`, { method: 'DELETE' });
}

export async function moveDealToStage(dealId: string, stageId: string): Promise<Deal> {
  return request<Deal>(`/crm/deals/${dealId}/stage`, { method: 'PATCH', body: { stage_id: stageId } });
}

export async function updateDealStatus(dealId: string, status: string, lostReason?: string): Promise<Deal> {
  return request<Deal>(`/crm/deals/${dealId}/status`, {
    method: 'PATCH',
    body: { status, lost_reason: lostReason }
  });
}

export async function getDealStats(pipelineId?: string): Promise<DealStats> {
  const params = pipelineId ? `?pipeline_id=${pipelineId}` : '';
  return request<DealStats>(`/crm/deals/stats${params}`);
}

// ============================================================================
// Activities
// ============================================================================

export async function getActivities(filters?: {
  activity_type?: string;
  is_completed?: boolean;
  limit?: number;
  offset?: number;
}): Promise<ActivitiesResponse> {
  const params = new URLSearchParams();
  if (filters?.activity_type) params.set('activity_type', filters.activity_type);
  if (filters?.is_completed !== undefined) params.set('is_completed', filters.is_completed.toString());
  if (filters?.limit) params.set('limit', filters.limit.toString());
  if (filters?.offset) params.set('offset', filters.offset.toString());
  const query = params.toString();
  return request<ActivitiesResponse>(`/crm/activities${query ? `?${query}` : ''}`);
}

export async function getDealActivities(dealId: string): Promise<ActivitiesResponse> {
  return request<ActivitiesResponse>(`/crm/deals/${dealId}/activities`);
}

export async function createActivity(data: CreateActivityData): Promise<CRMActivity> {
  return request<CRMActivity>('/crm/activities', { method: 'POST', body: data });
}

export async function completeActivity(activityId: string, outcome?: string): Promise<CRMActivity> {
  return request<CRMActivity>(`/crm/activities/${activityId}/complete`, {
    method: 'POST',
    body: { outcome }
  });
}

export async function deleteActivity(activityId: string): Promise<void> {
  return request<void>(`/crm/activities/${activityId}`, { method: 'DELETE' });
}
