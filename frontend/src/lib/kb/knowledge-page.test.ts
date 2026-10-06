import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import KnowledgePage from '../../routes/(app)/knowledge/+page.svelte';

vi.mock('$lib/stores/workspaces', async () => {
	const { writable } = await import('svelte/store');
	return { currentWorkspace: writable({id:'test',slug:'example',name:'Example company'}) };
});
vi.mock('$lib/kb/client', async importOriginal => {
	const original = await importOriginal<typeof import('./client')>();
	const nav={version:1,home:'review/today.md',sections:[{title:'Current review',items:[
		{path:'review/today.md',label:'Today'}, {path:'review/slides',label:'Visual slides'}
	]}]};
	return {...original,
		fetchWorkspaces:vi.fn(async()=>[{slug:'example',name:'Example company'}]),
		fetchTree:vi.fn(async()=>[
			{name:'knowledge',path:'knowledge',type:'dir',children:[{name:'navigation.md',path:'knowledge/navigation.md',type:'file'}]},
			{name:'review',path:'review',type:'dir',children:[
				{name:'today.md',path:'review/today.md',type:'file'},
				{name:'slides',path:'review/slides',type:'dir',children:[{name:'one.md',title:'First slide',path:'review/slides/one.md',type:'file'}]}
			]}, {name:'older.md',path:'older.md',type:'file'}
		]),
		fetchFile:vi.fn(async(_workspace:string,path:string)=>({content:path==='knowledge/navigation.md'?'```json\n'+JSON.stringify(nav)+'\n```':path==='review/today.md'?'# Today\nReview the main script.':'# First slide\nSlide content.',path,workspace:'example'})),
		fetchSources:vi.fn(async()=>({sources:{},counts:{engine:0,cloud:0,synced:0}})),
		getStorage:vi.fn(async()=>({activated:true,over_limit:false,bytes_used:0,bytes_limit:1000}))
	};
});
afterEach(cleanup);
describe('Knowledge organized sidebar',()=>{
	it('opens the home, expands nested folders, finds collapsed files and preserves all files',async()=>{
		render(KnowledgePage);
		await screen.findByText('Review the main script.');
		expect(screen.getByRole('button',{name:'Current review'})).toBeTruthy();
		await fireEvent.click(screen.getByRole('button',{name:'Visual slides'}));
		await fireEvent.click(screen.getByRole('button',{name:'First slide'}));
		await screen.findByText('Slide content.');
		await fireEvent.input(screen.getByPlaceholderText('Search document titles…'),{target:{value:'older'}});
		await waitFor(()=>expect(screen.getByRole('button',{name:'older.md'})).toBeTruthy());
		await fireEvent.input(screen.getByPlaceholderText('Search document titles…'),{target:{value:''}});
		await fireEvent.click(screen.getByRole('button',{name:'All files'}));
		expect(screen.getByRole('button',{name:'navigation.md'})).toBeTruthy();
		await fireEvent.click(screen.getByRole('button',{name:'Recent'}));
		expect(screen.getByRole('button',{name:'No update date'})).toBeTruthy();
		await fireEvent.change(screen.getByRole('combobox',{name:'Updated date range'}),{target:{value:'7'}});
		await screen.findByText('Nothing in this layer yet.');
	});
});
