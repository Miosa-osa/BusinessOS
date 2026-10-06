import { describe, expect, it } from 'vitest';
import { organizeKnowledge, parseNavigation, type KnowledgeNavigation } from './navigation';
import type { KBTreeNode } from './types';

const tree: KBTreeNode[] = [{name:'packages',path:'packages',type:'dir',children:[
	{name:'current',path:'packages/current',type:'dir',children:[
		{name:'main.md',path:'packages/current/main.md',type:'file'},
		{name:'notes.md',path:'packages/current/notes.md',type:'file'}
	]}, {name:'old.md',path:'packages/old.md',type:'file'}
]}];
const nav: KnowledgeNavigation = {version:1,home:'packages/current/main.md',sections:[
	{title:'Review',items:[{path:'packages/current/main.md',label:'Main script'}]},
	{title:'Production',items:[{path:'packages/current',label:'Production files'}]}
]};
function files(nodes: KBTreeNode[]): string[] {return nodes.flatMap(n=>n.type==='file'?[n.path]:files(n.children??[]));}
describe('workspace Knowledge navigation',()=>{
	it('parses a workspace-owned manifest',()=>expect(parseNavigation('```json\n'+JSON.stringify(nav)+'\n```')).toEqual(nav));
	it.each(['../secret','/absolute','https://host','a\\b','a/../b','a//b'])('rejects unsafe path %s',path=>{
		expect(parseNavigation('```json\n'+JSON.stringify({...nav,home:path})+'\n```')).toBeNull();
	});
	it('rejects invalid or unsupported manifests without crashing',()=>{
		for(const content of ['no config','```json\n{bad}\n```','```json\n{"version":2}\n```']) expect(parseNavigation(content)).toBeNull();
	});
	it('preserves every unselected document once and does not mutate source',()=>{
		const original=JSON.stringify(tree);
		const result=organizeKnowledge(tree,nav);
		expect(result.map(n=>n.name)).toEqual(['Review','Production','Other documents']);
		expect(files(result).sort()).toEqual(files(tree).sort());
		expect(JSON.stringify(tree)).toBe(original);
	});
	it('ignores missing or repeated targets',()=>{
		const result=organizeKnowledge(tree,{...nav,sections:[{title:'Review',items:[
			{path:'missing',label:'Missing'},{path:'packages/current',label:'Current'},
			{path:'packages/current/main.md',label:'Duplicate'}
		]}]});
		expect(result[0].children).toHaveLength(1);
		expect(files(result).sort()).toEqual(files(tree).sort());
	});
});
