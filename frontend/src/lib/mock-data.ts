/**
 * Mock data for development — populates stores when backend is unavailable.
 * Remove this file (and all references) before shipping to production.
 */

import type { TeamMemberListResponse } from '$lib/api/team';
import type { Project } from '$lib/api/projects';
import type { ClientListResponse } from '$lib/api';
import type { Pipeline, PipelineStage, Deal, DealStats } from '$lib/api/crm';

// ─── Team Members ─────────────────────────────────────────────
export const mockTeamMembers: TeamMemberListResponse[] = [
	{
		id: 'tm-001',
		name: 'Maya Chen',
		email: 'maya@miosa.ai',
		role: 'Lead Designer',
		avatar_url: null,
		status: 'available',
		capacity: 85,
		manager_id: null,
		active_projects: 4,
		open_tasks: 8,
		joined_at: '2024-06-15T09:00:00Z'
	},
	{
		id: 'tm-002',
		name: 'Alex Kim',
		email: 'alex@miosa.ai',
		role: 'Backend Engineer',
		avatar_url: null,
		status: 'busy',
		capacity: 95,
		manager_id: 'tm-001',
		active_projects: 3,
		open_tasks: 12,
		joined_at: '2024-07-01T09:00:00Z'
	},
	{
		id: 'tm-003',
		name: 'Tom Rivera',
		email: 'tom@miosa.ai',
		role: 'DevOps Engineer',
		avatar_url: null,
		status: 'overloaded',
		capacity: 100,
		manager_id: 'tm-001',
		active_projects: 5,
		open_tasks: 11,
		joined_at: '2024-08-10T09:00:00Z'
	},
	{
		id: 'tm-004',
		name: 'Sam Park',
		email: 'sam@miosa.ai',
		role: 'Product Manager',
		avatar_url: null,
		status: 'available',
		capacity: 30,
		manager_id: null,
		active_projects: 2,
		open_tasks: 3,
		joined_at: '2024-05-20T09:00:00Z'
	},
	{
		id: 'tm-005',
		name: 'Nia Johnson',
		email: 'nia@miosa.ai',
		role: 'Frontend Engineer',
		avatar_url: null,
		status: 'busy',
		capacity: 70,
		manager_id: 'tm-001',
		active_projects: 3,
		open_tasks: 7,
		joined_at: '2024-09-01T09:00:00Z'
	},
	{
		id: 'tm-006',
		name: 'Riku Tanaka',
		email: 'riku@miosa.ai',
		role: 'Data Scientist',
		avatar_url: null,
		status: 'ooo',
		capacity: 0,
		manager_id: 'tm-004',
		active_projects: 1,
		open_tasks: 2,
		joined_at: '2024-10-15T09:00:00Z'
	},
	{
		id: 'tm-007',
		name: 'Priya Nair',
		email: 'priya@miosa.ai',
		role: 'UX Researcher',
		avatar_url: null,
		status: 'available',
		capacity: 45,
		manager_id: 'tm-001',
		active_projects: 2,
		open_tasks: 5,
		joined_at: '2024-11-01T09:00:00Z'
	},
	{
		id: 'tm-008',
		name: 'Jordan Blake',
		email: 'jordan@miosa.ai',
		role: 'Full-Stack Engineer',
		avatar_url: null,
		status: 'available',
		capacity: 55,
		manager_id: 'tm-001',
		active_projects: 3,
		open_tasks: 6,
		joined_at: '2025-01-10T09:00:00Z'
	},
	{
		id: 'tm-009',
		name: 'Lena Ortiz',
		email: 'lena@miosa.ai',
		role: 'QA Engineer',
		avatar_url: null,
		status: 'busy',
		capacity: 80,
		manager_id: 'tm-001',
		active_projects: 4,
		open_tasks: 9,
		joined_at: '2025-02-01T09:00:00Z'
	},
	{
		id: 'tm-010',
		name: 'Marcus Lee',
		email: 'marcus@miosa.ai',
		role: 'VP Engineering',
		avatar_url: null,
		status: 'available',
		capacity: 40,
		manager_id: null,
		active_projects: 6,
		open_tasks: 4,
		joined_at: '2024-03-01T09:00:00Z'
	}
];

// ─── Projects ─────────────────────────────────────────────────
export const mockProjects: Project[] = [
	{
		id: 'proj-001',
		name: 'BusinessOS v2 Launch',
		description: 'Complete redesign and relaunch of the core platform with the new Foundation design system.',
		status: 'active',
		priority: 'critical',
		client_name: null,
		project_type: 'product',
		project_metadata: { progress: 72, team_size: 6 },
		created_at: '2025-11-01T09:00:00Z',
		updated_at: '2026-03-06T14:00:00Z',
		notes: [
			{ id: 'n1', content: 'Sprint 4 kicked off — focus on CRM and team modules.', created_at: '2026-03-05T10:00:00Z' }
		]
	},
	{
		id: 'proj-002',
		name: 'Mobile App MVP',
		description: 'Native iOS and Android companion app for BusinessOS field workers.',
		status: 'active',
		priority: 'high',
		client_name: null,
		project_type: 'product',
		project_metadata: { progress: 35, team_size: 3 },
		created_at: '2026-01-15T09:00:00Z',
		updated_at: '2026-03-04T11:00:00Z',
		notes: []
	},
	{
		id: 'proj-003',
		name: 'Acme Corp Integration',
		description: 'Custom API integration and onboarding for Acme Corp enterprise deployment.',
		status: 'active',
		priority: 'high',
		client_name: 'Acme Corp',
		project_type: 'client',
		project_metadata: { progress: 60, team_size: 2 },
		created_at: '2026-02-01T09:00:00Z',
		updated_at: '2026-03-05T16:00:00Z',
		notes: [
			{ id: 'n2', content: 'SSO integration complete. Working on data migration.', created_at: '2026-03-03T15:00:00Z' }
		]
	},
	{
		id: 'proj-004',
		name: 'AI Agent Framework',
		description: 'Build out the custom agent creation and orchestration framework with MCP support.',
		status: 'active',
		priority: 'critical',
		client_name: null,
		project_type: 'product',
		project_metadata: { progress: 48, team_size: 4 },
		created_at: '2025-12-01T09:00:00Z',
		updated_at: '2026-03-06T09:00:00Z',
		notes: [
			{ id: 'n3', content: 'Tool calling and MCP protocol layer working. Need to build UI.', created_at: '2026-03-06T09:00:00Z' }
		]
	},
	{
		id: 'proj-005',
		name: 'Documentation Overhaul',
		description: 'Rewrite all developer docs, API reference, and user guides.',
		status: 'paused',
		priority: 'medium',
		client_name: null,
		project_type: 'internal',
		project_metadata: { progress: 20, team_size: 1 },
		created_at: '2026-01-01T09:00:00Z',
		updated_at: '2026-02-15T11:00:00Z',
		notes: []
	},
	{
		id: 'proj-006',
		name: 'NovaTech CRM Setup',
		description: 'Deploy and customize CRM module for NovaTech including custom pipeline stages.',
		status: 'active',
		priority: 'medium',
		client_name: 'NovaTech Solutions',
		project_type: 'client',
		project_metadata: { progress: 80, team_size: 2 },
		created_at: '2026-01-20T09:00:00Z',
		updated_at: '2026-03-06T12:00:00Z',
		notes: []
	},
	{
		id: 'proj-007',
		name: 'Security Audit Q1',
		description: 'Quarterly penetration testing and compliance review.',
		status: 'completed',
		priority: 'high',
		client_name: null,
		project_type: 'internal',
		project_metadata: { progress: 100, team_size: 2 },
		created_at: '2026-01-05T09:00:00Z',
		updated_at: '2026-02-28T17:00:00Z',
		notes: [
			{ id: 'n4', content: 'All critical issues resolved. Report filed.', created_at: '2026-02-28T17:00:00Z' }
		]
	},
	{
		id: 'proj-008',
		name: 'Sunset Analytics Dashboard',
		description: 'Build real-time analytics dashboard with charts, KPIs, and export features.',
		status: 'active',
		priority: 'medium',
		client_name: 'Sunset Media',
		project_type: 'client',
		project_metadata: { progress: 55, team_size: 3 },
		created_at: '2026-02-10T09:00:00Z',
		updated_at: '2026-03-05T10:00:00Z',
		notes: []
	}
];

// ─── Clients ──────────────────────────────────────────────────
export const mockClients: ClientListResponse[] = [
	{
		id: 'cl-001',
		name: 'Acme Corp',
		type: 'company',
		email: 'contact@acmecorp.com',
		phone: '+1 (555) 100-2000',
		status: 'active',
		source: 'referral',
		assigned_to: 'Sam Park',
		lifetime_value: 128000,
		tags: ['enterprise', 'priority'],
		created_at: '2025-06-10T09:00:00Z',
		last_contacted_at: '2026-03-05T14:00:00Z',
		contacts_count: 4,
		interactions_count: 32,
		deals_count: 2,
		active_deals_value: 45000
	},
	{
		id: 'cl-002',
		name: 'NovaTech Solutions',
		type: 'company',
		email: 'hello@novatech.io',
		phone: '+1 (555) 200-3000',
		status: 'active',
		source: 'inbound',
		assigned_to: 'Maya Chen',
		lifetime_value: 85000,
		tags: ['saas', 'growth'],
		created_at: '2025-09-15T09:00:00Z',
		last_contacted_at: '2026-03-04T11:00:00Z',
		contacts_count: 3,
		interactions_count: 18,
		deals_count: 1,
		active_deals_value: 32000
	},
	{
		id: 'cl-003',
		name: 'Sunset Media',
		type: 'company',
		email: 'team@sunsetmedia.co',
		phone: '+1 (555) 300-4000',
		status: 'active',
		source: 'conference',
		assigned_to: 'Jordan Blake',
		lifetime_value: 62000,
		tags: ['media', 'creative'],
		created_at: '2025-11-01T09:00:00Z',
		last_contacted_at: '2026-03-03T16:00:00Z',
		contacts_count: 2,
		interactions_count: 14,
		deals_count: 1,
		active_deals_value: 28000
	},
	{
		id: 'cl-004',
		name: 'Greenfield Labs',
		type: 'company',
		email: 'info@greenfieldlabs.com',
		phone: '+1 (555) 400-5000',
		status: 'prospect',
		source: 'cold-outreach',
		assigned_to: 'Alex Kim',
		lifetime_value: 0,
		tags: ['biotech', 'startup'],
		created_at: '2026-02-01T09:00:00Z',
		last_contacted_at: '2026-02-28T10:00:00Z',
		contacts_count: 1,
		interactions_count: 5,
		deals_count: 1,
		active_deals_value: 15000
	},
	{
		id: 'cl-005',
		name: 'Derek Hollis',
		type: 'individual',
		email: 'derek@freelance.co',
		phone: '+1 (555) 500-6000',
		status: 'lead',
		source: 'website',
		assigned_to: null,
		lifetime_value: 0,
		tags: ['freelance'],
		created_at: '2026-03-01T09:00:00Z',
		last_contacted_at: null,
		contacts_count: 0,
		interactions_count: 1,
		deals_count: 0,
		active_deals_value: 0
	},
	{
		id: 'cl-006',
		name: 'Pinnacle Finance',
		type: 'company',
		email: 'ops@pinnaclefin.com',
		phone: '+1 (555) 600-7000',
		status: 'active',
		source: 'referral',
		assigned_to: 'Sam Park',
		lifetime_value: 210000,
		tags: ['finance', 'enterprise', 'priority'],
		created_at: '2025-03-20T09:00:00Z',
		last_contacted_at: '2026-03-06T09:00:00Z',
		contacts_count: 6,
		interactions_count: 45,
		deals_count: 3,
		active_deals_value: 75000
	},
	{
		id: 'cl-007',
		name: 'Bright Horizon Education',
		type: 'company',
		email: 'admin@brighthorizon.edu',
		phone: '+1 (555) 700-8000',
		status: 'inactive',
		source: 'conference',
		assigned_to: 'Nia Johnson',
		lifetime_value: 38000,
		tags: ['education', 'nonprofit'],
		created_at: '2025-07-10T09:00:00Z',
		last_contacted_at: '2025-12-15T09:00:00Z',
		contacts_count: 2,
		interactions_count: 9,
		deals_count: 1,
		active_deals_value: 0
	},
	{
		id: 'cl-008',
		name: 'CloudScale Inc',
		type: 'company',
		email: 'partnerships@cloudscale.io',
		phone: '+1 (555) 800-9000',
		status: 'prospect',
		source: 'inbound',
		assigned_to: 'Marcus Lee',
		lifetime_value: 0,
		tags: ['cloud', 'infrastructure'],
		created_at: '2026-02-20T09:00:00Z',
		last_contacted_at: '2026-03-02T14:00:00Z',
		contacts_count: 2,
		interactions_count: 7,
		deals_count: 1,
		active_deals_value: 50000
	}
];

// ─── CRM — Pipelines, Stages & Deals ─────────────────────────
export const mockPipeline: Pipeline = {
	id: 'pipe-001',
	user_id: 'user-001',
	name: 'Sales Pipeline',
	description: 'Main B2B sales pipeline',
	pipeline_type: 'sales',
	currency: 'USD',
	is_default: true,
	is_active: true,
	color: '#3b82f6',
	icon: 'briefcase',
	created_at: '2025-06-01T09:00:00Z',
	updated_at: '2026-03-01T09:00:00Z'
};

export const mockStages: PipelineStage[] = [
	{ id: 'stg-001', pipeline_id: 'pipe-001', name: 'Lead',          description: 'New inbound lead',       position: 0, probability: 10, stage_type: 'open', rotting_days: 14, color: '#6b7280', created_at: '2025-06-01T09:00:00Z', updated_at: '2025-06-01T09:00:00Z' },
	{ id: 'stg-002', pipeline_id: 'pipe-001', name: 'Qualified',     description: 'Qualified prospect',     position: 1, probability: 25, stage_type: 'open', rotting_days: 10, color: '#3b82f6', created_at: '2025-06-01T09:00:00Z', updated_at: '2025-06-01T09:00:00Z' },
	{ id: 'stg-003', pipeline_id: 'pipe-001', name: 'Proposal',      description: 'Proposal sent',          position: 2, probability: 50, stage_type: 'open', rotting_days: 7,  color: '#8b5cf6', created_at: '2025-06-01T09:00:00Z', updated_at: '2025-06-01T09:00:00Z' },
	{ id: 'stg-004', pipeline_id: 'pipe-001', name: 'Negotiation',   description: 'In negotiation',         position: 3, probability: 75, stage_type: 'open', rotting_days: 5,  color: '#f59e0b', created_at: '2025-06-01T09:00:00Z', updated_at: '2025-06-01T09:00:00Z' },
	{ id: 'stg-005', pipeline_id: 'pipe-001', name: 'Won',           description: 'Deal closed — won',      position: 4, probability: 100, stage_type: 'won',  rotting_days: 0,  color: '#22c55e', created_at: '2025-06-01T09:00:00Z', updated_at: '2025-06-01T09:00:00Z' },
	{ id: 'stg-006', pipeline_id: 'pipe-001', name: 'Lost',          description: 'Deal closed — lost',     position: 5, probability: 0,   stage_type: 'lost', rotting_days: 0,  color: '#ef4444', created_at: '2025-06-01T09:00:00Z', updated_at: '2025-06-01T09:00:00Z' }
];

export const mockDeals: Deal[] = [
	{
		id: 'deal-001', user_id: 'user-001', pipeline_id: 'pipe-001', pipeline_name: 'Sales Pipeline',
		stage_id: 'stg-002', stage_name: 'Qualified', name: 'Acme Corp — Platform License',
		description: 'Annual platform license for 200 seats', amount: 45000, currency: 'USD',
		probability: 25, expected_close_date: '2026-04-30T00:00:00Z', actual_close_date: undefined,
		owner_id: 'tm-004', company_id: 'cl-001', company_name: 'Acme Corp',
		primary_contact_id: undefined, status: 'open', lost_reason: undefined,
		priority: 'high', lead_source: 'referral', deal_score: 72,
		custom_fields: {}, created_at: '2026-02-10T09:00:00Z', updated_at: '2026-03-05T14:00:00Z'
	},
	{
		id: 'deal-002', user_id: 'user-001', pipeline_id: 'pipe-001', pipeline_name: 'Sales Pipeline',
		stage_id: 'stg-003', stage_name: 'Proposal', name: 'NovaTech — CRM Module',
		description: 'Custom CRM setup + 1 year support', amount: 32000, currency: 'USD',
		probability: 50, expected_close_date: '2026-04-15T00:00:00Z', actual_close_date: undefined,
		owner_id: 'tm-001', company_id: 'cl-002', company_name: 'NovaTech Solutions',
		primary_contact_id: undefined, status: 'open', lost_reason: undefined,
		priority: 'medium', lead_source: 'inbound', deal_score: 65,
		custom_fields: {}, created_at: '2026-01-20T09:00:00Z', updated_at: '2026-03-04T11:00:00Z'
	},
	{
		id: 'deal-003', user_id: 'user-001', pipeline_id: 'pipe-001', pipeline_name: 'Sales Pipeline',
		stage_id: 'stg-004', stage_name: 'Negotiation', name: 'Pinnacle Finance — Enterprise',
		description: 'Full platform + custom integrations + dedicated support', amount: 75000, currency: 'USD',
		probability: 75, expected_close_date: '2026-03-31T00:00:00Z', actual_close_date: undefined,
		owner_id: 'tm-004', company_id: 'cl-006', company_name: 'Pinnacle Finance',
		primary_contact_id: undefined, status: 'open', lost_reason: undefined,
		priority: 'urgent', lead_source: 'referral', deal_score: 88,
		custom_fields: {}, created_at: '2025-12-01T09:00:00Z', updated_at: '2026-03-06T09:00:00Z'
	},
	{
		id: 'deal-004', user_id: 'user-001', pipeline_id: 'pipe-001', pipeline_name: 'Sales Pipeline',
		stage_id: 'stg-001', stage_name: 'Lead', name: 'Greenfield Labs — Starter',
		description: 'Startup package exploration', amount: 15000, currency: 'USD',
		probability: 10, expected_close_date: '2026-06-30T00:00:00Z', actual_close_date: undefined,
		owner_id: 'tm-002', company_id: 'cl-004', company_name: 'Greenfield Labs',
		primary_contact_id: undefined, status: 'open', lost_reason: undefined,
		priority: 'low', lead_source: 'cold-outreach', deal_score: 30,
		custom_fields: {}, created_at: '2026-02-01T09:00:00Z', updated_at: '2026-02-28T10:00:00Z'
	},
	{
		id: 'deal-005', user_id: 'user-001', pipeline_id: 'pipe-001', pipeline_name: 'Sales Pipeline',
		stage_id: 'stg-005', stage_name: 'Won', name: 'Sunset Media — Dashboard',
		description: 'Analytics dashboard build + 6 month support', amount: 28000, currency: 'USD',
		probability: 100, expected_close_date: '2026-02-15T00:00:00Z', actual_close_date: '2026-02-12T00:00:00Z',
		owner_id: 'tm-008', company_id: 'cl-003', company_name: 'Sunset Media',
		primary_contact_id: undefined, status: 'won', lost_reason: undefined,
		priority: 'medium', lead_source: 'conference', deal_score: 95,
		custom_fields: {}, created_at: '2025-11-01T09:00:00Z', updated_at: '2026-02-12T14:00:00Z'
	},
	{
		id: 'deal-006', user_id: 'user-001', pipeline_id: 'pipe-001', pipeline_name: 'Sales Pipeline',
		stage_id: 'stg-001', stage_name: 'Lead', name: 'CloudScale — Infrastructure',
		description: 'Cloud infrastructure management platform', amount: 50000, currency: 'USD',
		probability: 10, expected_close_date: '2026-07-31T00:00:00Z', actual_close_date: undefined,
		owner_id: 'tm-010', company_id: 'cl-008', company_name: 'CloudScale Inc',
		primary_contact_id: undefined, status: 'open', lost_reason: undefined,
		priority: 'medium', lead_source: 'inbound', deal_score: 40,
		custom_fields: {}, created_at: '2026-02-20T09:00:00Z', updated_at: '2026-03-02T14:00:00Z'
	},
	{
		id: 'deal-007', user_id: 'user-001', pipeline_id: 'pipe-001', pipeline_name: 'Sales Pipeline',
		stage_id: 'stg-006', stage_name: 'Lost', name: 'Bright Horizon — LMS Integration',
		description: 'Learning management system integration', amount: 18000, currency: 'USD',
		probability: 0, expected_close_date: '2025-12-01T00:00:00Z', actual_close_date: '2025-12-15T00:00:00Z',
		owner_id: 'tm-005', company_id: 'cl-007', company_name: 'Bright Horizon Education',
		primary_contact_id: undefined, status: 'lost', lost_reason: 'Budget constraints',
		priority: 'low', lead_source: 'conference', deal_score: 10,
		custom_fields: {}, created_at: '2025-07-10T09:00:00Z', updated_at: '2025-12-15T09:00:00Z'
	},
	{
		id: 'deal-008', user_id: 'user-001', pipeline_id: 'pipe-001', pipeline_name: 'Sales Pipeline',
		stage_id: 'stg-002', stage_name: 'Qualified', name: 'Pinnacle Finance — Analytics Add-on',
		description: 'Real-time analytics dashboard extension', amount: 22000, currency: 'USD',
		probability: 25, expected_close_date: '2026-05-15T00:00:00Z', actual_close_date: undefined,
		owner_id: 'tm-004', company_id: 'cl-006', company_name: 'Pinnacle Finance',
		primary_contact_id: undefined, status: 'open', lost_reason: undefined,
		priority: 'medium', lead_source: 'upsell', deal_score: 60,
		custom_fields: {}, created_at: '2026-03-01T09:00:00Z', updated_at: '2026-03-06T11:00:00Z'
	}
];

export const mockDealStats: DealStats = {
	total_deals: 8,
	open_deals: 5,
	won_deals: 1,
	lost_deals: 1,
	open_value: 239000,
	won_value: 28000,
	lost_value: 18000
};
