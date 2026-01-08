// UI Primitives - Svelte/Bits UI Components
// Based on AFFiNE patterns, adapted for BusinessOS

// Core Components
export { default as Button } from './button/Button.svelte';
export { default as Input } from './input/Input.svelte';
export { default as Loading } from './loading/Loading.svelte';
export { default as Skeleton } from './skeleton/Skeleton.svelte';
export { default as Separator } from './separator/Separator.svelte';

// Overlay Components
export { default as Modal } from './modal/Modal.svelte';
export { default as Tooltip } from './tooltip/Tooltip.svelte';
export { default as Popover } from './popover/Popover.svelte';

// Menu Components
export {
	Menu,
	MenuItem,
	MenuSeparator,
	MenuLabel,
	MenuGroup
} from './menu';

// Tab Components
export {
	Tabs,
	TabsList,
	TabsTrigger,
	TabsContent
} from './tabs';

// Layout Components
export { default as ScrollArea } from './scroll-area/ScrollArea.svelte';
