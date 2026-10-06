import { fireEvent, render, cleanup } from '@testing-library/svelte';
import { afterEach, expect, it, vi } from 'vitest';
import OsaOrb from './OsaOrb.svelte';
const voice = vi.hoisted(() => ({ toggle:vi.fn(),destroy:vi.fn() }));
vi.mock('$app/environment',()=>({browser:true}));
vi.mock('$lib/services/osaOrbVoice',()=>({OsaOrbVoice:class {toggle=voice.toggle;destroy=voice.destroy;}}));
vi.mock('$lib/stores/osa',()=>({osaStore:{setError:vi.fn()}}));
afterEach(()=>{cleanup();localStorage.clear();vi.clearAllMocks();});
function pointer(type: string,x:number,y:number){const e=new Event(type,{bubbles:true});Object.assign(e,{button:0,isPrimary:true,pointerId:1,clientX:x,clientY:y});return e;}
it('moves only during a captured drag and does not start voice when dropped',async()=>{
 const {getByRole,container}=render(OsaOrb); const button=getByRole('button',{name:'Talk to OSA'});
 Object.assign(button,{setPointerCapture:vi.fn(),hasPointerCapture:()=>true,releasePointerCapture:vi.fn()});
 await fireEvent(button,pointer('pointerdown',20,20));
 await fireEvent(button,pointer('pointermove',300,250));
 await fireEvent(button,pointer('pointerup',300,250));
 await fireEvent.click(button,{detail:1});
 expect(voice.toggle).not.toHaveBeenCalled();
 const anchor=container.querySelector('.osa-orb') as HTMLElement;
 expect(anchor.style.left).toBe('280px'); expect(anchor.style.top).toBe('230px');
 await fireEvent(button,pointer('pointermove',500,500));
 expect(anchor.style.left).toBe('280px');
 await fireEvent.click(button,{detail:1});expect(voice.toggle).toHaveBeenCalledOnce();
});
it('clamps an old offscreen saved position into the viewport',()=>{
 localStorage.setItem('osaOrbPosition',JSON.stringify({x:100000,y:-500}));
 const {container}=render(OsaOrb); const anchor=container.querySelector('.osa-orb') as HTMLElement;
 expect(anchor.style.left).toBe(`${window.innerWidth-96}px`);expect(anchor.style.top).toBe('8px');
});
