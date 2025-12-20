// Entrypoint for the `projects` domain folder
export * from './types';
export * from './projects';

import * as projectsApi from './projects';

export const api = {
  getProjects: projectsApi.getProjects,
  getProject: projectsApi.getProject,
  createProject: projectsApi.createProject,
  updateProject: projectsApi.updateProject,
  deleteProject: projectsApi.deleteProject,
  addProjectNote: projectsApi.addProjectNote,
};

export default api;
 
