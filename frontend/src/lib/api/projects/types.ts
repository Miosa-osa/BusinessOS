// Extracted API types for domain modules

export interface Project {
  id: string;
  name: string;
  description: string | null;
  status: 'active' | 'paused' | 'completed' | 'archived';
  priority: 'critical' | 'high' | 'medium' | 'low';
  client_name: string | null;
  project_type: string;
  project_metadata: Record<string, unknown> | null;
  created_at: string;
  updated_at: string;
  notes: ProjectNote[];
}

export interface ProjectNote {
  id: string;
  content: string;
  created_at: string;
}

export interface CreateProjectData {
  name: string;
  description?: string;
  status?: 'active' | 'paused' | 'completed' | 'archived';
  priority?: 'critical' | 'high' | 'medium' | 'low';
  client_name?: string;
  project_type?: string;
  project_metadata?: Record<string, unknown>;
}

// Re-export other common types as needed in future
