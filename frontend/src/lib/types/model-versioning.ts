/**
 * Model Versioning Type Definitions
 *
 * Comprehensive TypeScript types for process model version management.
 * Enables type-safe operations across frontend and backend integration.
 */

export interface ProcessModelVersion {
	id: string;
	model_id: string;
	version: string; // "2.1.3+a7c3e9f1"
	major: number;
	minor: number;
	patch: number;
	content_hash: string; // First 8 chars of SHA256

	created_at: string; // ISO8601
	created_by: string;
	discovery_source?: string | null; // 'inductive', 'heuristic', 'alpha', 'manual', 'ml_rl'
	previous_version_id?: string | null;

	model_json: Record<string, unknown>;
	delta_json?: Record<string, unknown> | null;

	nodes_count: number;
	edges_count: number;
	variants: number;
	fitness: number; // 0.0 to 1.0
	average_duration: number; // minutes
	covered_traces: number;

	change_type: 'major' | 'minor' | 'patch';
	nodes_added: number;
	nodes_removed: number;
	edges_added: number;
	edges_removed: number;

	description: string;
	tags: string[];

	is_released: boolean;
	release_notes?: string | null;
	released_at?: string | null; // ISO8601
	archived_at?: string | null; // ISO8601
}

export interface ModelMetrics {
	nodes_count: number;
	edges_count: number;
	variants: number;
	fitness: number;
	average_duration: number;
	covered_traces: number;
}

export interface ChangeSummary {
	nodes_added: number;
	nodes_removed: number;
	edges_added: number;
	edges_removed: number;
}

export interface VersionDiffResult {
	from_version: string;
	to_version: string;
	structural_diff: StructuralDiff;
	metrics_diff: MetricsDiffResult;
	change_summary: ChangeSummary;
	breaking_changes: string[];
}

export interface StructuralDiff {
	nodes_added: NodeChange[];
	nodes_removed: NodeChange[];
	edges_added: EdgeChange[];
	edges_removed: EdgeChange[];
}

export interface NodeChange {
	id: string;
	type: 'task' | 'xor_gateway' | 'and_gateway' | 'event' | string;
	label: string;
}

export interface EdgeChange {
	id: string;
	source: string;
	target: string;
	label?: string | null;
}

export interface MetricsDiffResult {
	nodes_count: DiffValue;
	edges_count: DiffValue;
	variants: DiffValue;
	fitness: DiffValue;
	average_duration: DiffValue;
	covered_traces: DiffValue;
}

export interface DiffValue {
	before: number | string;
	after: number | string;
	delta: number | string;
}

export interface RollbackRequest {
	model_id: string;
	target_version: string;
	reason: string;
	approved_by: string;
	running_instances?: 'pause' | 'continue' | 'replay';
	backup_current?: boolean;
}

export interface RollbackImpact {
	current_version: string;
	target_version: string;
	breaking_changes: string[];
	instances_to_pause: number;
	compatible_instances: number;
	incompatible_instances: number;
}

export interface ReleaseRequest {
	release_notes: string;
}

export interface CreateVersionRequest {
	model: Record<string, unknown>;
	metrics: ModelMetrics;
	change_type: 'major' | 'minor' | 'patch';
	description: string;
	created_by: string;
	discovery_source?: string | null;
	tags?: string[];
}

/**
 * API Response Wrappers
 */

export interface VersionHistoryResponse {
	versions: ProcessModelVersion[];
	total: number;
	limit: number;
	offset: number;
}

export interface ErrorResponse {
	error: string;
	status: number;
	timestamp: string;
}

/**
 * Type Guards
 */

export function isProcessModelVersion(obj: unknown): obj is ProcessModelVersion {
	if (typeof obj !== 'object' || obj === null) return false;
	const v = obj as Record<string, unknown>;
	return (
		typeof v.id === 'string' &&
		typeof v.version === 'string' &&
		typeof v.fitness === 'number' &&
		typeof v.is_released === 'boolean'
	);
}

export function isVersionDiffResult(obj: unknown): obj is VersionDiffResult {
	if (typeof obj !== 'object' || obj === null) return false;
	const d = obj as Record<string, unknown>;
	return (
		typeof d.from_version === 'string' &&
		typeof d.to_version === 'string' &&
		typeof d.breaking_changes === 'object'
	);
}

/**
 * Utility Functions
 */

export function getVersionChangeIcon(changeType: string): string {
	switch (changeType) {
		case 'major':
			return '⚠️'; // Breaking change
		case 'minor':
			return '✨'; // Feature
		case 'patch':
			return '🔧'; // Fix
		default:
			return '📝';
	}
}

export function getVersionBadgeClass(changeType: string): string {
	switch (changeType) {
		case 'major':
			return 'bg-red-100 text-red-800';
		case 'minor':
			return 'bg-yellow-100 text-yellow-800';
		case 'patch':
			return 'bg-green-100 text-green-800';
		default:
			return 'bg-gray-100 text-gray-800';
	}
}

export function formatFitnessScore(fitness: number): string {
	const percent = (fitness * 100).toFixed(1);
	return `${percent}%`;
}

export function formatDuration(minutes: number): string {
	if (minutes < 60) {
		return `${minutes.toFixed(1)} min`;
	}
	const hours = minutes / 60;
	return `${hours.toFixed(1)} hrs`;
}

export function isFitnessGate(fitness: number): boolean {
	return fitness >= 0.85;
}
