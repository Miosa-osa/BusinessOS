export * from './types';
export * from './integrations';

import * as integrationsApi from './integrations';

export const api = {
  // Google
  initiateGoogleAuth: integrationsApi.initiateGoogleAuth,
  getGoogleConnectionStatus: integrationsApi.getGoogleConnectionStatus,
  disconnectGoogle: integrationsApi.disconnectGoogle,
  // Slack
  initiateSlackAuth: integrationsApi.initiateSlackAuth,
  getSlackConnectionStatus: integrationsApi.getSlackConnectionStatus,
  disconnectSlack: integrationsApi.disconnectSlack,
  getSlackChannels: integrationsApi.getSlackChannels,
  getSlackNotifications: integrationsApi.getSlackNotifications,
  // Notion
  initiateNotionAuth: integrationsApi.initiateNotionAuth,
  getNotionConnectionStatus: integrationsApi.getNotionConnectionStatus,
  disconnectNotion: integrationsApi.disconnectNotion,
  getNotionDatabases: integrationsApi.getNotionDatabases,
  getNotionPages: integrationsApi.getNotionPages,
  syncNotionDatabase: integrationsApi.syncNotionDatabase,
};

export default api;
