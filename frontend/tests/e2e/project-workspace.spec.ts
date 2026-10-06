import { expect, test } from '@playwright/test';

test('projects open from both views and expose editable tasks, people and notes', async ({ page }) => {
  let task = { id: 'task-1', title: 'Review production handoff', description: '## Acceptance criteria\n\nConfirm the receiving owner.\n\n[Operating guide](packages/guide.md)', status: 'todo', priority: 'medium', project_id: 'project-1', assignee_id: 'person-1', due_date: null };
  const project = { id: 'project-1', name: 'Production operations', description: 'Weekly production and publishing.', status: 'active', priority: 'medium', project_metadata: { knowledge_path: 'packages/guide.md' }, notes: [] };
  const notes: object[] = [];
  await page.route('**/api/**', async route => {
    const req = route.request(), url = new URL(req.url());
    if (!url.pathname.startsWith('/api/')) return route.continue();
    let body: unknown = {};
    if (url.pathname.endsWith('/projects')) body = [project];
    if (url.pathname.endsWith('/projects/project-1')) body = { project, notes };
    if (url.pathname.endsWith('/team')) body = [{ id: 'person-1', name: 'Casey', role: 'Coordinator' }];
    if (url.pathname.endsWith('/dashboard/tasks')) body = [task];
    if (url.pathname.endsWith('/dashboard/tasks/task-1') && req.method() === 'PUT') { task = { ...task, ...req.postDataJSON() }; body = task; }
    if (url.pathname.endsWith('/projects/project-1/notes') && req.method() === 'POST') { const n = { id: 'note-1', content: req.postDataJSON().content, created_at: '2026-01-01T12:00:00Z' }; notes.push(n); body = n; }
    if (url.pathname.endsWith('/knowledge/file')) body = { content: '# Operating guide\n\nCheck every handoff.\n\n<img src=x onerror="window.unsafe=true">' };
    await route.fulfill({ json: body });
  });
  await page.goto('/login');
  await page.evaluate(async () => {
    const runtimePath: string = '/node_modules/.vite/deps/svelte.js';
    const componentPath: string = '/src/routes/(app)/projects/+page.svelte';
    const storePath: string = '/src/lib/stores/workspaces.ts';
    const { mount } = await import(runtimePath);
    const { default: Projects } = await import(componentPath);
    const { currentWorkspace } = await import(storePath);
    currentWorkspace.set({ id: 'workspace-1', name: 'Operations', slug: 'operations' });
    const host = document.createElement('div');
    host.id = 'project-test-host';
    host.style.cssText = 'position:fixed;inset:0;z-index:99999;background:white;--dbg:#fff;--dt:#171717;--dt2:#555;--dt3:#707070;--dbd:#ddd';
    document.body.append(host);
    mount(Projects, { target: host });
  });
  const host = page.locator('#project-test-host');
  for (const view of ['Board view', 'List view']) {
    await host.getByRole('button', { name: view, exact: true }).click();
    await host.getByRole('button', { name: 'Production operations', exact: true }).click();
    await expect(host.getByRole('heading', { name: 'Production operations' })).toBeVisible();
    await host.getByRole('button', { name: 'Open operating guide' }).click();
    await expect(host.getByRole('heading', { name: 'Operating guide', exact: true })).toBeVisible();
    expect(await page.evaluate(() => Reflect.get(window, 'unsafe'))).toBeUndefined();
    await host.getByRole('button', { name: 'Tasks 1', exact: true }).click();
    await host.getByRole('button', { name: /^Review production handoff/ }).click();
    await expect(host.getByRole('heading', { name: 'Acceptance criteria' })).toBeVisible();
    await host.getByRole('combobox', { name: 'Task status', exact: true }).selectOption('in_progress');
    await expect(host.getByRole('combobox', { name: 'Task status', exact: true })).toHaveValue('in_progress');
    await host.getByRole('button', { name: 'Edit task', exact: true }).click();
    await host.getByLabel('Task brief', { exact: true }).fill('## Acceptance criteria\n\nVerified revised handoff.');
    await host.getByRole('button', { name: 'Save task' }).click();
    await expect(host.getByText('Verified revised handoff.', { exact: true })).toBeVisible();
    await host.getByRole('button', { name: 'Back to projects', exact: true }).click();
  }
  await host.getByRole('button', { name: 'Production operations', exact: true }).click();
  await host.getByRole('button', { name: 'People 1', exact: true }).click();
  await host.getByRole('button', { name: /Casey Coordinator/ }).click();
  await expect(host.getByRole('combobox', { name: 'Filter by assignee' })).toHaveValue('person-1');
  await host.getByRole('button', { name: 'Notes 0', exact: true }).click();
  await host.getByLabel('New note', { exact: true }).fill('Handoff accepted.');
  await host.getByRole('button', { name: 'Add note', exact: true }).click();
  await expect(host.getByText('Handoff accepted.', { exact: true })).toBeVisible();
  // A narrow desktop window must adapt even when the browser viewport is wide.
  await host.evaluate(el => { el.style.width = '620px'; el.style.right = 'auto'; });
  await host.getByRole('button', { name: 'Tasks 1', exact: true }).click();
  const row = host.getByRole('button', { name: /^Review production handoff/ });
  await expect(row.locator('.row-owner')).toBeVisible();
  await expect(row.locator('.row-status')).toBeVisible();
  expect(await row.evaluate(el => el.scrollWidth <= el.clientWidth)).toBeTruthy();
  await page.screenshot({ path: '/private/tmp/project-narrow-window.png' });
  await row.click();
  await expect(host.locator('.task-list')).toBeHidden();
  await expect(host.getByRole('heading', { name: 'Acceptance criteria' })).toBeVisible();
  await host.evaluate(el => { el.style.width = ''; el.style.right = '0'; });
  await page.setViewportSize({ width: 390, height: 844 });
  await host.getByRole('button', { name: 'Close task details' }).click();
  await host.getByRole('button', { name: 'Tasks 1', exact: true }).click();
  await host.getByRole('button', { name: /^Review production handoff/ }).click();
  await expect(host.getByRole('heading', { name: 'Acceptance criteria' })).toBeVisible();
  const bounds = await host.getByRole('region', { name: 'Project task details' }).boundingBox();
  expect(bounds).toBeTruthy();
  expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(391);
  await page.screenshot({ path: '/private/tmp/project-mobile-tested.png' });
});
