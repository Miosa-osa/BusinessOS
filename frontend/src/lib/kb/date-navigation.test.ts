import { describe, expect, it } from 'vitest';
import { recentKnowledge, sortKnowledge } from './date-navigation';
import type { KBTreeNode } from './types';
const files: KBTreeNode[] = [
	{name:'old',path:'old',type:'file',modified:'2026-08-10T12:00:00Z'},
	{name:'new',path:'new',type:'file',modified:'2026-09-14T12:00:00Z'},
	{name:'unknown',path:'unknown',type:'file',modified:'bad'}
];
describe('Knowledge date browsing',()=>{
	it('sorts timestamps both directions with unknown last',()=>{
		expect(sortKnowledge(files,'newest').map(n=>n.path)).toEqual(['new','old','unknown']);
		expect(sortKnowledge(files,'oldest').map(n=>n.path)).toEqual(['old','new','unknown']);
	});
	it('keeps sections in place and sorts their children without mutation',()=>{
		const tree: KBTreeNode[]=[{name:'Review',path:'@section/0',type:'dir',children:files},{name:'Archive',path:'@section/1',type:'dir',children:[]}];
		expect(sortKnowledge(tree,'newest')[0].children?.[0].path).toBe('new');
		expect(tree[0].children?.[0].path).toBe('old');
	});
	it('groups all dates by month and preserves undated documents',()=>{
		const groups=recentKnowledge(files,'newest',0);
		expect(groups.map(n=>n.path)).toEqual(['@month/2026-09','@month/2026-08','@month/unknown']);
	});
	it('filters a rolling date window using update timestamps, not filename dates',()=>{
		const groups=recentKnowledge(files,'newest',7,Date.parse('2026-09-15T12:00:00Z'));
		expect(groups.flatMap(n=>n.children??[]).map(n=>n.path)).toEqual(['new']);
	});
});
