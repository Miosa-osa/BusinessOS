import { expect, test } from '@playwright/test';

test('task titles open details without changing status in board and list views', async ({ page }) => {
  let writes = 0;
  await page.route('**/api/**', async route => {
    const url = new URL(route.request().url());
    if (!url.pathname.startsWith('/api/')) return route.continue();
    if (route.request().method() !== 'GET') writes++;
    let body: unknown = {};
    if (url.pathname.endsWith('/dashboard/tasks')) body = [{ id: 'task-1', title: 'Map Atlas handoff', description: '## Outcome\n\n03:10 Ethan: Map the handoff. Human review required.\n\n[Meeting transcript](/packages/meeting.md)', status: 'todo', priority: 'high', project_id: 'atlas', assignee_id: 'ethan', due_date: '2026-09-09T00:00:00Z' }];
    if (url.pathname.endsWith('/projects')) body = [{ id: 'atlas', name: 'Atlas Roofing' }];
    if (url.pathname.endsWith('/team')) body = [{ id: 'ethan', name: 'Ethan Cole' }];
    if (url.pathname.endsWith('/knowledge/file')) {
      expect(url.searchParams.get('workspace')).toBe('northstar-growth');
      expect(url.searchParams.get('path')).toBe('packages/meeting.md');
      body = {content:'# Working Session\n\nThe receiving owner must acknowledge the handoff.\n\n<img src=x onerror="window.unsafe=true">'};
    }
    await route.fulfill({ json: body });
  });
  // Mount the real task window independently of server-side authentication.
  await page.goto('/login');
  await page.evaluate(async () => {
    const runtimePath: string = '/node_modules/.vite/deps/svelte.js';
    const windowPath: string = '/src/routes/(app)/tasks/+page.svelte';
    const { mount } = await import(runtimePath);
    const { default: TaskWindow } = await import(windowPath);
    const storePath: string = '/src/lib/stores/workspaces.ts';
    const { currentWorkspace } = await import(storePath);
    currentWorkspace.set({id:'test-workspace',slug:'northstar-growth',name:'Northstar Growth'});
    const host = document.createElement('div');
    host.style.cssText = 'position:fixed;inset:0;z-index:99999;background:white';
    document.body.append(host);
    mount(TaskWindow, { target: host });
  });
  await page.getByRole('combobox', { name: 'Filter by project' }).selectOption('atlas');
  for (const view of ['Board view', 'List view']) {
    await page.getByRole('button', { name: view, exact: true }).click();
    await page.getByRole('button', { name: 'Map Atlas handoff', exact: true }).click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toContainText('03:10 Ethan');
    await expect(dialog).toContainText('Ethan Cole');
    await expect(dialog).toContainText('September 9, 2026');
    await dialog.getByRole('link', {name:'Meeting transcript',exact:true}).click();
    await expect(dialog).toContainText('The receiving owner must acknowledge the handoff.');
    expect(await page.evaluate(() => Reflect.get(window,'unsafe'))).toBeUndefined();
    await dialog.getByRole('button', {name:'Back to task'}).click();
    await expect(dialog).toContainText('03:10 Ethan');
    await page.keyboard.press('Escape');
    await expect(dialog).toHaveCount(0);
    await expect(page.getByRole('combobox', { name: 'Filter by project' })).toHaveValue('atlas');
  }
  expect(writes).toBe(0);
});
