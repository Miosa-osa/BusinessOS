import type { KBTreeNode } from './types';
export type KnowledgeSort = 'manual' | 'newest' | 'oldest' | 'title';
export function modifiedTime(node: KBTreeNode): number {
	const own = Date.parse(node.modified ?? '');
	return Math.max(Number.isFinite(own) ? own : 0, ...(node.children ?? []).map(modifiedTime));
}
export function sortKnowledge(nodes: KBTreeNode[], order: KnowledgeSort): KBTreeNode[] {
	const copy = nodes.map(n => ({ ...n, ...(n.children ? { children: sortKnowledge(n.children, order) } : {}) }));
	if (order === 'manual') return copy;
	return copy.sort((a, b) => {
		// Keep the workspace's top-level subject hierarchy in its editorial order.
		if (a.path.startsWith('@section/') && b.path.startsWith('@section/')) return 0;
		const title = (a.title ?? a.name).localeCompare(b.title ?? b.name, undefined, { numeric: true });
		if (order === 'title') return title;
		const x = modifiedTime(a), y = modifiedTime(b);
		if (!x || !y) return x ? -1 : y ? 1 : title;
		return (order === 'newest' ? y - x : x - y) || title;
	});
}
export function recentKnowledge(nodes: KBTreeNode[], order: KnowledgeSort, days: number, now = Date.now()): KBTreeNode[] {
	const files: KBTreeNode[] = [];
	function walk(items: KBTreeNode[]) {
		for (const n of items) {
			if (n.type === 'dir') walk(n.children ?? []);
			else if (n.path !== 'knowledge/navigation.md' && (!days || (modifiedTime(n) > 0 && modifiedTime(n) >= now - days * 86400000))) files.push(n);
		}
	}
	walk(nodes);
	const groups = new Map<string, KBTreeNode>();
	for (const n of sortKnowledge(files, order === 'oldest' ? 'oldest' : 'newest')) {
		const time = modifiedTime(n), date = new Date(time);
		const key = time ? `${date.getFullYear()}-${String(date.getMonth()+1).padStart(2,'0')}` : 'unknown';
		if (!groups.has(key)) groups.set(key, { name: time ? date.toLocaleDateString(undefined, {month:'long',year:'numeric'}) : 'No update date', path:`@month/${key}`, type:'dir', children:[] });
		groups.get(key)!.children!.push(n);
	}
	return [...groups.values()];
}
