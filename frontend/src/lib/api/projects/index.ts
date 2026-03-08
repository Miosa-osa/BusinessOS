// Entrypoint for the `projects` domain folder
export * from './types';
export * from './projects';
export * from './members';

import * as projectsApi from './projects';
import * as membersApi from './members';

export const api = {
  getProjects: projectsApi.getProjects,
  getProject: projectsApi.getProject,
  createProject: projectsApi.createProject,
  updateProject: projectsApi.updateProject,
  deleteProject: projectsApi.deleteProject,
  addProjectNote: projectsApi.addProjectNote,
  // Project members
  listProjectMembers: membersApi.listProjectMembers,
  addProjectMember: membersApi.addProjectMember,
  updateProjectMemberRole: membersApi.updateProjectMemberRole,
  removeProjectMember: membersApi.removeProjectMember,
  checkProjectAccess: membersApi.checkProjectAccess,
};

export default api;
 
