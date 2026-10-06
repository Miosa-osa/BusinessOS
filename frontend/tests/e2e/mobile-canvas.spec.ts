import { expect, test } from '@playwright/test';

test.describe('Mobile Infinity Canvas', () => {
	test.use({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });

	test('keeps canvas controls clear of the dock and restores a readable window', async ({ page }) => {
		await page.route('**/api/auth/session', (route) =>
			route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					user: { id: 'mobile-user', name: 'Mobile User', email: 'mobile@example.com' },
					session: { id: 'mobile-session' }
				})
			})
		);
		await page.route('**/api/v1/workspaces', (route) =>
			route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ workspaces: [] }) })
		);

		await page.addInitScript(() => {
			const desktopId = 'mobile-infinity';
			const finderWindow = {
				id: 'finder-mobile',
				module: 'finder',
				title: 'Finder',
				x: 120,
				y: 90,
				width: 900,
				height: 620,
				minWidth: 320,
				minHeight: 220,
				minimized: false,
				maximized: false
			};
			localStorage.setItem('businessos-onboarded', 'true');
			localStorage.setItem('cookie_consent', 'essential_only');
			sessionStorage.setItem('businessos-booted', 'true');
			localStorage.setItem('businessos-canvas-views', JSON.stringify({
				[desktopId]: { zoom: 0.08, pan: { x: 130, y: 260 }, compact: false }
			}));
			localStorage.setItem('businessos_desktop_settings', JSON.stringify({
				version: '1.2.0',
				activeDesktopId: desktopId,
				desktopSpaces: [{
					id: desktopId,
					name: 'Infinity Desktop',
					kind: 'personal',
					desktopIcons: [],
					dockPinnedItems: ['finder'],
					hiddenModules: [],
					folders: [],
					windows: [finderWindow],
					windowOrder: [finderWindow.id],
					focusedWindowId: finderWindow.id,
					createdAt: new Date(0).toISOString(),
					updatedAt: new Date(0).toISOString()
				}]
			}));
		});

		await page.goto('/window');

		const toolbar = page.locator('.infinity-toolbar');
		const dock = page.locator('.dock');
		const canvasWindow = page.locator('.window').first();
		await expect(toolbar).toBeVisible();
		await expect(dock).toBeVisible();
		await expect(canvasWindow).toBeVisible();
		await expect(toolbar.locator('strong')).not.toHaveText('8%');

		const toolbarBox = await toolbar.boundingBox();
		const dockBox = await dock.boundingBox();
		expect(toolbarBox).not.toBeNull();
		expect(dockBox).not.toBeNull();
		expect(toolbarBox!.y + toolbarBox!.height).toBeLessThanOrEqual(dockBox!.y - 8);
		await expect.poll(async () => (await canvasWindow.boundingBox())?.width ?? 0).toBeGreaterThanOrEqual(300);
	});
});
