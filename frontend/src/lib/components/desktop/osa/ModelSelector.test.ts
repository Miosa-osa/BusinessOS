import { cleanup, fireEvent, render, waitFor } from '@testing-library/svelte';
import { afterEach, expect, it, vi } from 'vitest';
import { writable } from 'svelte/store';
import ModelSelector from './ModelSelector.svelte';
const mocks=vi.hoisted(()=>({loadModelCatalog:vi.fn(),setModel:vi.fn(),loadHealth:vi.fn()}));
vi.mock('$lib/stores/osa',()=>({osaStore:{...writable({activeModel:'current:cloud',activeProvider:'ollama'}),...mocks}}));
afterEach(()=>{cleanup();vi.clearAllMocks();});
const catalog={current:'current:cloud',provider:'ollama',providers:[{slug:'ollama',name:'Ollama',configured:true,connected:true},{slug:'ollama_cloud',name:'Ollama Cloud',configured:true,connected:true},{slug:'anthropic',name:'Anthropic',configured:false,connected:false}],models:[{name:'current:cloud',provider:'ollama'},{name:'new-runtime-model',provider:'ollama_cloud'}]};
it('loads OSA providers and models and submits the real provider ID',async()=>{
 mocks.loadModelCatalog.mockResolvedValue(catalog);mocks.setModel.mockResolvedValue(undefined);
 const ui=render(ModelSelector);
 await fireEvent.click(ui.getByRole('combobox',{name:/Select model/}));
 await waitFor(()=>expect(mocks.loadModelCatalog).toHaveBeenCalledOnce());
 const provider=await ui.findByLabelText('Provider');
 await fireEvent.change(provider,{target:{value:'ollama_cloud'}});
 await fireEvent.click(await ui.findByRole('button',{name:'new-runtime-model'}));
 await fireEvent.click(ui.getByRole('button',{name:'Apply model'}));
 await waitFor(()=>expect(mocks.setModel).toHaveBeenCalledWith('ollama_cloud','new-runtime-model'));
 expect(await ui.findByText('Model applied')).toBeTruthy();
});
it('shows catalog errors instead of inventing fallback presets',async()=>{
 mocks.loadModelCatalog.mockRejectedValue(new Error('OSA offline'));
 const ui=render(ModelSelector);await fireEvent.click(ui.getByRole('combobox',{name:/Select model/}));
 expect(await ui.findByRole('alert')).toHaveTextContent('OSA offline');
 expect(ui.queryByText('qwen3:32b')).toBeNull();
});
