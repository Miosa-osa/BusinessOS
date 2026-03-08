import { request } from './base';
import type {
  CustomModule,
  ModuleVersion,
  ModuleInstallation,
  ModuleShare,
  CreateModuleData,
  UpdateModuleData,
  ShareModuleData,
  ModuleCategory,
  ModuleVisibility
} from '$lib/types/modules';

// Module CRUD operations

export async function getModules(params?: {
  category?: ModuleCategory;
  search?: string;
  sort?: 'popular' | 'newest' | 'name' | 'installs';
  visibility?: ModuleVisibility;
  limit?: number;
  offset?: number;
}): Promise<{ modules: CustomModule[]; total: number }> {
  const queryParams = new URLSearchParams();
  if (params?.category) queryParams.set('category', params.category);
  if (params?.search) queryParams.set('search', params.search);
  if (params?.sort) queryParams.set('sort', params.sort);
  if (params?.visibility) queryParams.set('visibility', params.visibility);
  if (params?.limit) queryParams.set('limit', params.limit.toString());
  if (params?.offset) queryParams.set('offset', params.offset.toString());

  const query = queryParams.toString();
  const raw = await request<Record<string, unknown>>(`/modules${query ? `?${query}` : ''}`, { skipAuthRedirect: true });

  // Backend returns paginated format { data: [...], pagination: { total_items } }
  if (raw && typeof raw === 'object' && 'data' in raw && 'pagination' in raw) {
    const pagination = raw.pagination as { total_items?: number } | undefined;
    return {
      modules: (raw.data as CustomModule[]) ?? [],
      total: pagination?.total_items ?? 0
    };
  }

  // Fallback: legacy format { modules: [...], total }
  if (raw && typeof raw === 'object' && 'modules' in raw) {
    return { modules: (raw.modules as CustomModule[]) ?? [], total: (raw.total as number) ?? 0 };
  }

  return { modules: [], total: 0 };
}

export async function getModule(id: string): Promise<CustomModule> {
  return request(`/modules/${id}`);
}

export async function createModule(data: CreateModuleData): Promise<CustomModule> {
  return request('/modules', {
    method: 'POST',
    body: data
  });
}

export async function updateModule(id: string, data: UpdateModuleData): Promise<CustomModule> {
  return request(`/modules/${id}`, {
    method: 'PUT',
    body: data
  });
}

export async function deleteModule(id: string): Promise<{ message: string }> {
  return request(`/modules/${id}`, {
    method: 'DELETE'
  });
}

// Module versions

export async function getModuleVersions(moduleId: string): Promise<{ versions: ModuleVersion[] }> {
  return request(`/modules/${moduleId}/versions`);
}

export async function createModuleVersion(
  moduleId: string,
  data: { version: string; manifest: unknown; changelog?: string }
): Promise<ModuleVersion> {
  return request(`/modules/${moduleId}/versions`, {
    method: 'POST',
    body: data
  });
}

// Module installations

export async function getInstallations(): Promise<{ installations: ModuleInstallation[] }> {
  const raw = await request<Record<string, unknown>>('/modules/installed', { skipAuthRedirect: true });

  // Backend returns paginated { data: [...], pagination: {...} }
  if (raw && typeof raw === 'object' && 'data' in raw && Array.isArray(raw.data)) {
    return { installations: raw.data as ModuleInstallation[] };
  }
  if (raw && typeof raw === 'object' && 'installations' in raw) {
    return raw as { installations: ModuleInstallation[] };
  }
  return { installations: [] };
}

export async function installModule(
  moduleId: string,
  config?: Record<string, unknown>
): Promise<ModuleInstallation> {
  return request('/modules/install', {
    method: 'POST',
    body: { module_id: moduleId, config }
  });
}

export async function uninstallModule(moduleId: string): Promise<{ message: string }> {
  return request('/modules/uninstall', {
    method: 'POST',
    body: { module_id: moduleId }
  });
}

export async function updateInstallation(
  installationId: string,
  data: { config?: Record<string, unknown>; is_active?: boolean }
): Promise<ModuleInstallation> {
  return request(`/modules/installations/${installationId}`, {
    method: 'PUT',
    body: data
  });
}

// Module sharing

export async function shareModule(moduleId: string, data: ShareModuleData): Promise<ModuleShare> {
  return request(`/modules/${moduleId}/share`, {
    method: 'POST',
    body: data
  });
}

export async function getModuleShares(moduleId: string): Promise<{ shares: ModuleShare[] }> {
  return request(`/modules/${moduleId}/shares`);
}

export async function revokeShare(shareId: string): Promise<{ message: string }> {
  return request(`/modules/shares/${shareId}`, {
    method: 'DELETE'
  });
}

// Module export/import

export async function exportModule(moduleId: string): Promise<Blob> {
  const response = await fetch(`/api/v1/modules/${moduleId}/export`, {
    method: 'GET',
    credentials: 'include'
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ detail: 'Export failed' }));
    throw new Error(error.detail || 'Export failed');
  }

  return response.blob();
}

export async function importModule(file: File): Promise<CustomModule> {
  const formData = new FormData();
  formData.append('file', file);

  const response = await fetch('/api/v1/modules/import', {
    method: 'POST',
    credentials: 'include',
    body: formData
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ detail: 'Import failed' }));
    throw new Error(error.detail || 'Import failed');
  }

  return response.json();
}

// Module actions (execute)

export async function executeModuleAction(
  moduleId: string,
  actionName: string,
  parameters: Record<string, unknown>
): Promise<{ success: boolean; result: unknown; error?: string }> {
  return request(`/modules/${moduleId}/execute`, {
    method: 'POST',
    body: { action: actionName, parameters }
  });
}

// Module discovery (marketplace-like)

export async function getPopularModules(limit: number = 10): Promise<{ modules: CustomModule[] }> {
  return request(`/modules/popular?limit=${limit}`);
}

export async function getRecentModules(limit: number = 10): Promise<{ modules: CustomModule[] }> {
  return request(`/modules/recent?limit=${limit}`);
}

export async function searchModules(query: string): Promise<{ modules: CustomModule[] }> {
  return request(`/modules/search?q=${encodeURIComponent(query)}`);
}
