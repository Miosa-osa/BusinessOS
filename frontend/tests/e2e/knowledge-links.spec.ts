import { expect, test } from '@playwright/test';

test('relative document links stay inside the Knowledge workspace', async ({ page }) => {
  page.on('pageerror', error => console.error(error.message));
  const root = 'packages/current-review/';
  const docs: Record<string, string> = {
    [root + 'start.md']: '# Review hub\n[Read scripts](scripts.md)\n[Other package](../other/examples.md)\n[Website](https://example.com)',
    [root + 'scripts.md']: '# Spoken scripts\n[Back to hub](./start.md)\n[Missing](missing.md)',
    'packages/other/examples.md': '# Example collection\n[Back to hub](../current-review/start.md)'
  };
  const requested: string[] = [];
  await page.route('**/api/**', async route => {
    const url = new URL(route.request().url());
    if (!url.pathname.startsWith('/api/')) return route.continue();
    let body: unknown = {};
    let status = 200;
    if (url.pathname.endsWith('/auth/session')) body = { user: { id: 'test-user', name: 'Reviewer', email: 'review@example.test' }, session: { id: 'test-session' } };
    if (url.pathname.endsWith('/knowledge/workspaces')) body = { workspaces: [{ slug: 'review-workspace', name: 'Review workspace' }] };
    if (url.pathname.endsWith('/knowledge/tree')) body = { tree: [{ type: 'dir', name: 'packages', path: 'packages', children: [{ type: 'dir', name: 'current-review', path: root.slice(0,-1), children: [{ type: 'file', name: 'start.md', title: 'Review hub', path: root + 'start.md' }] }] }] };
    if (url.pathname.endsWith('/knowledge/file')) {
      expect(url.searchParams.get('workspace')).toBe('review-workspace');
      const path = url.searchParams.get('path')!;
      requested.push(path);
      body = { content: docs[path] };
      if (!docs[path]) status = 404;
    }
    await route.fulfill({ status, json: body });
  });
  await page.goto('/knowledge?embed=true');
  if ((page.viewportSize()?.width ?? 1440) < 769) await page.getByRole('button', { name: 'Toggle pages panel', exact: true }).last().click();
  await page.getByRole('button', { name: 'Review hub', exact: true }).click();
  await page.getByRole('link', { name: 'Read scripts' }).click();
  await expect(page.locator('.kb-prose')).toContainText('Spoken scripts');
  await expect(page).toHaveURL(/\/knowledge\?embed=true$/);
  await page.getByRole('link', { name: 'Missing', exact: true }).click();
  await expect(page.locator('.kb-prose')).toContainText('Could not load this document');
  if ((page.viewportSize()?.width ?? 1440) < 769) await page.getByRole('button', { name: 'Toggle pages panel', exact: true }).last().click();
  await page.getByRole('button', { name: 'Review hub', exact: true }).click();
  await page.getByRole('link', { name: 'Other package' }).click();
  await expect(page.locator('.kb-prose')).toContainText('Example collection');
  await page.getByRole('link', { name: 'Back to hub' }).click();
  await expect(page.getByRole('link', { name: 'Website' })).toHaveAttribute('href', 'https://example.com');
  expect(requested).toContain('packages/other/examples.md');
});
