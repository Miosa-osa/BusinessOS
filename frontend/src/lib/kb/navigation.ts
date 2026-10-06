import { z } from 'zod';
import type { KBTreeNode } from './types';

export const NAVIGATION_PATH = 'knowledge/navigation.md';
const pathSchema = z.string().min(1).max(500).refine(path =>
	!/[\\?#:\u0000]/.test(path) && path.split('/').every(part => part && part !== '.' && part !== '..')
);
const schema = z.object({
	version: z.literal(1),
	home: pathSchema,
	sections: z.array(z.object({
		title: z.string().trim().min(1).max(80),
		items: z.array(z.object({ path: pathSchema, label: z.string().trim().min(1).max(100) })).max(100)
	})).min(1).max(30)
});
export type KnowledgeNavigation = z.infer<typeof schema>;

export function parseNavigation(content: string): KnowledgeNavigation | null {
	try {
		const json = content.match(/^```json\s*\n([\s\S]*?)^```\s*$/m)?.[1];
		return json ? schema.parse(JSON.parse(json)) : null;
	} catch { return null; }
}

export function organizeKnowledge(tree: KBTreeNode[], navigation: KnowledgeNavigation): KBTreeNode[] {
	const index = new Map<string, KBTreeNode>();
	function collect(nodes: KBTreeNode[]) {
		for (const node of nodes) { index.set(node.path, node); collect(node.children ?? []); }
	}
	collect(tree);
	const used = new Set<string>([NAVIGATION_PATH]);
	function remaining(nodes: KBTreeNode[]): KBTreeNode[] {
		return nodes.flatMap(node => {
			if (used.has(node.path)) return [];
			if (node.type === 'file') return [{ ...node }];
			const children = remaining(node.children ?? []);
			return children.length ? [{ ...node, children }] : [];
		});
	}
	function claim(node: KBTreeNode) {
		used.add(node.path);
		for (const child of node.children ?? []) claim(child);
	}
	const sections: KBTreeNode[] = [];
	for (const [i, section] of navigation.sections.entries()) {
		const children: KBTreeNode[] = [];
		for (const item of section.items) {
			const node = index.get(item.path);
			if (!node || used.has(node.path)) continue;
			const available = remaining([node])[0];
			if (!available) continue;
			claim(available);
			children.push({ ...available, name: item.label, title: item.label });
		}
		if (children.length) sections.push({ name: section.title, title: section.title, path: `@section/${i}`, type: 'dir', children });
	}
	const other = remaining(tree);
	if (other.length) sections.push({ name: 'Other documents', path: '@section/other', type: 'dir', children: other });
	return sections;
}
