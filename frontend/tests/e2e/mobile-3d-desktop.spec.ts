import { expect, test } from '@playwright/test';

test.describe('Mobile experimental 3D desktop', () => {
	test.use({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });

	test('loads only the focused module as a live iframe', async ({ page }) => {
		const pageErrors: string[] = [];
		page.on('pageerror', (error) => pageErrors.push(error.message));

		await page.route('**/api/auth/session', (route) =>
			route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					user: { id: 'mobile-3d-user', name: 'Mobile 3D User', email: 'mobile3d@example.com' },
					session: { id: 'mobile-3d-session' }
				})
			})
		);
		await page.route('**/api/v1/workspaces', (route) =>
			route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ workspaces: [] }) })
		);

		await page.addInitScript(() => {
			localStorage.setItem('businessos-onboarded', 'true');
			localStorage.setItem('cookie_consent', 'essential_only');
			sessionStorage.setItem('businessos-booted', 'true');
		});

		await page.goto('/window');
		await page.getByRole('button', { name: 'Desktop menu' }).click();
		await page.getByText('Experimental 3D Desktop').click();

		await expect(page.locator('.desktop-3d')).toBeVisible();
		await expect(page.locator('.desktop-3d canvas')).toBeVisible();
		await expect(page.locator('.desktop-3d .osa-orb')).toBeHidden();
		await expect(page.locator('.window-iframe')).toHaveCount(0);
		await page.locator('.desktop-3d .dock-item').first().click();
		await expect(page.locator('.window-iframe')).toHaveCount(1);
		await expect.poll(async () => (await page.locator('.window-iframe').boundingBox())?.width ?? 0)
			.toBeGreaterThanOrEqual(300);
		await page.waitForTimeout(3_000);
		expect(pageErrors).toEqual([]);
		await expect(page.locator('.desktop-3d')).toBeVisible();
	});
});
