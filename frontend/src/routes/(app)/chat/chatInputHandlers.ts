import type { chatConversationStore } from "$lib/stores/chat/chatConversationStore.svelte";
import type { chatContextStore } from "$lib/stores/chat/chatContextStore.svelte";
import type { chatAgentStore } from "$lib/stores/chat/chatAgentStore.svelte";
import type { chatUIStore } from "$lib/stores/chat/chatUIStore.svelte";
import type { SlashCommand, AgentPreset } from "$lib/stores/chat/types";

type CS = typeof chatConversationStore;
type CX = typeof chatContextStore;
type AG = typeof chatAgentStore;
type UI = typeof chatUIStore;

export interface InputHandlerContext {
  cs: CS;
  cx: CX;
  ag: AG;
  ui: UI;
  getInputRef: () => HTMLTextAreaElement | undefined;
  getShowModelDropdown: () => boolean;
  handleSendMessage: () => void;
}

function scrollCommandIntoView(index: number) {
  setTimeout(() => {
    const item = document.querySelector(`[data-command-index="${index}"]`);
    item?.scrollIntoView({ block: "nearest", behavior: "smooth" });
  }, 0);
}

export function selectCommand(ctx: InputHandlerContext, cmd: SlashCommand) {
  ctx.cs.inputValue = ctx.ag.selectCommand(cmd);
  ctx.getInputRef()?.focus();
}

export function clearActiveCommand(ctx: InputHandlerContext) {
  ctx.ag.clearActiveCommand();
  ctx.cs.inputValue = "";
  ctx.getInputRef()?.focus();
}

export function selectAgent(ctx: InputHandlerContext, agent: AgentPreset) {
  ctx.cs.inputValue = ctx.ag.selectAgent(agent);
  ctx.getInputRef()?.focus();
}

export function clearDetectedAgent(ctx: InputHandlerContext) {
  ctx.ag.clearDetectedAgent();
  ctx.cs.inputValue = "";
  ctx.getInputRef()?.focus();
}

export function handleInput(
  ctx: Pick<InputHandlerContext, "cs" | "ag" | "getInputRef">,
) {
  const inputRef = ctx.getInputRef();
  if (inputRef) {
    inputRef.style.height = "auto";
    inputRef.style.height = Math.min(inputRef.scrollHeight, 200) + "px";
  }
  ctx.ag.parseInputForSuggestions(ctx.cs.inputValue);
}

export function handleKeydown(ctx: InputHandlerContext, e: KeyboardEvent) {
  const { ag, cx, ui, cs } = ctx;

  if (ag.showCommandSuggestions && ag.filteredCommands.length > 0) {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      ag.commandDropdownIndex =
        (ag.commandDropdownIndex + 1) % ag.filteredCommands.length;
      scrollCommandIntoView(ag.commandDropdownIndex);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      ag.commandDropdownIndex =
        ag.commandDropdownIndex <= 0
          ? ag.filteredCommands.length - 1
          : ag.commandDropdownIndex - 1;
      scrollCommandIntoView(ag.commandDropdownIndex);
    } else if (e.key === "Enter" || e.key === "Tab") {
      e.preventDefault();
      const cmd = ag.filteredCommands[ag.commandDropdownIndex];
      if (cmd) selectCommand(ctx, cmd);
    } else if (e.key === "Escape") {
      e.preventDefault();
      ag.showCommandSuggestions = false;
    }
    return;
  }

  if (ag.showAgentSuggestions && ag.filteredAgents.length > 0) {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      ag.agentDropdownIndex =
        (ag.agentDropdownIndex + 1) % ag.filteredAgents.length;
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      ag.agentDropdownIndex =
        ag.agentDropdownIndex <= 0
          ? ag.filteredAgents.length - 1
          : ag.agentDropdownIndex - 1;
    } else if (e.key === "Enter" || e.key === "Tab") {
      e.preventDefault();
      const agent = ag.filteredAgents[ag.agentDropdownIndex];
      if (agent) selectAgent(ctx, agent);
    } else if (e.key === "Escape") {
      e.preventDefault();
      ag.showAgentSuggestions = false;
    }
    return;
  }

  if (ui.showInlineProjectPicker) {
    const totalItems = cx.projectsList.length + 1;
    if (e.key === "ArrowDown") {
      e.preventDefault();
      cx.projectDropdownIndex = (cx.projectDropdownIndex + 1) % totalItems;
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      cx.projectDropdownIndex =
        cx.projectDropdownIndex <= 0
          ? totalItems - 1
          : cx.projectDropdownIndex - 1;
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (cx.projectDropdownIndex === cx.projectsList.length) {
        ui.showInlineProjectPicker = false;
        cx.showNewProjectModal = true;
      } else {
        const project = cx.projectsList[cx.projectDropdownIndex];
        if (project) cx.selectedProjectId = project.id;
      }
      ui.showInlineProjectPicker = false;
    } else if (e.key === "Escape") {
      e.preventDefault();
      ui.showInlineProjectPicker = false;
    }
    return;
  }

  if (
    cx.showProjectDropdown ||
    cx.showContextDropdown ||
    ctx.getShowModelDropdown()
  )
    return;

  if (e.key === "Enter" && !e.shiftKey) {
    e.preventDefault();
    if (!cx.selectedProjectId && cs.inputValue.trim()) {
      cx.projectDropdownIndex = 0;
      ui.showInlineProjectPicker = true;
      return;
    }
    ctx.handleSendMessage();
  }
}

export function handleProjectDropdownKeydown(
  ctx: Pick<InputHandlerContext, "cx" | "cs" | "handleSendMessage">,
  e: KeyboardEvent,
) {
  const { cx, cs } = ctx;
  if (!cx.showProjectDropdown) return;

  const totalItems = cx.projectsList.length + 1;
  const createNewIndex = cx.projectsList.length;

  if (e.key === "ArrowDown") {
    e.preventDefault();
    cx.projectDropdownIndex = (cx.projectDropdownIndex + 1) % totalItems;
  } else if (e.key === "ArrowUp") {
    e.preventDefault();
    cx.projectDropdownIndex =
      cx.projectDropdownIndex <= 0
        ? totalItems - 1
        : cx.projectDropdownIndex - 1;
  } else if (e.key === "Enter") {
    e.preventDefault();
    if (cx.projectDropdownIndex === createNewIndex) {
      cx.showProjectDropdown = false;
      cx.showNewProjectModal = true;
    } else {
      const project = cx.projectsList[cx.projectDropdownIndex];
      if (project) {
        cx.selectedProjectId = project.id;
        cx.showProjectDropdown = false;
        if (cs.inputValue.trim()) {
          setTimeout(() => ctx.handleSendMessage(), 50);
        }
      }
    }
  } else if (e.key === "Escape") {
    e.preventDefault();
    cx.showProjectDropdown = false;
  }
}
