<script lang="ts">
  import { Bot } from 'lucide-svelte';

  interface PresetTemplate {
    id: string;
    name: string;
    description: string;
    category: string;
    role: string;
    system_prompt: string;
    model?: string;
    tags?: string[];
  }

  interface Props {
    preset: PresetTemplate;
    onUse: (preset: PresetTemplate) => void;
  }

  let { preset, onUse }: Props = $props();

  function getCategoryColor(category: string): string {
    const colors: Record<string, string> = {
      business: 'bg-blue-100 text-blue-700',
      creative: 'bg-purple-100 text-purple-700',
      technical: 'bg-green-100 text-green-700',
      research: 'bg-orange-100 text-orange-700',
      support: 'bg-pink-100 text-pink-700',
      general: 'bg-gray-100 text-gray-700'
    };
    return colors[category.toLowerCase()] || 'bg-gray-100 text-gray-700';
  }
</script>

<div class="group border rounded-lg p-5 hover:shadow-lg hover:border-blue-300 transition-all bg-white h-full flex flex-col">
  <!-- Icon/Avatar Section -->
  <div class="flex items-start gap-4 mb-4">
    <div class="flex-shrink-0">
      <div class="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
        <Bot class="w-6 h-6 text-white" />
      </div>
    </div>
    <div class="flex-1 min-w-0">
      <div class="flex items-start justify-between gap-2 mb-1">
        <h3 class="font-semibold text-gray-900 leading-tight">{preset.name}</h3>
        <span class="flex-shrink-0 text-xs px-2 py-1 rounded-full font-medium {getCategoryColor(preset.category)}">
          {preset.category}
        </span>
      </div>
      <p class="text-xs text-gray-500 font-mono">{preset.role}</p>
    </div>
  </div>

  <!-- Description -->
  <p class="text-sm text-gray-600 mb-4 flex-grow line-clamp-3">{preset.description}</p>

  <!-- Tags/Capabilities -->
  {#if preset.tags && preset.tags.length > 0}
    <div class="flex flex-wrap gap-1.5 mb-4">
      {#each preset.tags.slice(0, 3) as tag}
        <span class="text-xs px-2 py-1 bg-gray-100 text-gray-700 rounded">
          {tag}
        </span>
      {/each}
      {#if preset.tags.length > 3}
        <span class="text-xs px-2 py-1 bg-gray-100 text-gray-500 rounded">
          +{preset.tags.length - 3}
        </span>
      {/if}
    </div>
  {/if}

  <!-- Model Info (if available) -->
  {#if preset.model}
    <div class="mb-4 pb-4 border-b border-gray-100">
      <div class="flex items-center gap-1.5 text-xs text-gray-500">
        <span class="font-medium">Model:</span>
        <span class="font-mono">{preset.model}</span>
      </div>
    </div>
  {/if}

  <!-- Use Template Button -->
  <button
    onclick={() => onUse(preset)}
    class="mt-auto w-full px-4 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-all font-medium shadow-sm hover:shadow group-hover:scale-[1.02] focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
  >
    Use Template
  </button>
</div>

<style>
  .line-clamp-3 {
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>
