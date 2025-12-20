export * from './types';
export * from './ai';

import * as aiApi from './ai';

export const api = {
  getAIProviders: aiApi.getAIProviders,
  updateAIProvider: aiApi.updateAIProvider,
  getAllModels: aiApi.getAllModels,
  getLocalModels: aiApi.getLocalModels,
  pullModel: aiApi.pullModel,
  warmupModel: aiApi.warmupModel,
  getAISystemInfo: aiApi.getAISystemInfo,
  saveAPIKey: aiApi.saveAPIKey,
  getAgentPrompts: aiApi.getAgentPrompts,
  getAgentPrompt: aiApi.getAgentPrompt,
  getTools: aiApi.getTools,
  executeTool: aiApi.executeTool,
};

export default api;
