import { request } from '../base';
import type { Project, CreateProjectData, ProjectNote } from './types';

export async function getProjects(status?: string): Promise<Project[]> {
  const params = status ? `?status_filter=${status}` : '';
  const raw = await request<Record<string, unknown>>(`/projects${params}`);

  // Backend returns paginated format { data: [...], pagination: {...} }
  if (raw && typeof raw === 'object' && 'data' in raw && Array.isArray(raw.data)) {
    return raw.data as Project[];
  }
  if (Array.isArray(raw)) {
    return raw as unknown as Project[];
  }
  return [];
}

export async function getProject(id: string): Promise<Project> {
  return request<Project>(`/projects/${id}`);
}

export async function createProject(data: CreateProjectData): Promise<Project> {
  return request<Project>('/projects', { method: 'POST', body: data });
}

export async function updateProject(id: string, data: Partial<CreateProjectData>): Promise<Project> {
  return request<Project>(`/projects/${id}`, { method: 'PUT', body: data });
}

export async function deleteProject(id: string): Promise<void> {
  return request(`/projects/${id}`, { method: 'DELETE' }) as unknown as void;
}

export async function addProjectNote(projectId: string, content: string): Promise<ProjectNote> {
  return request<ProjectNote>(`/projects/${projectId}/notes`, { method: 'POST', body: { content } });
}
