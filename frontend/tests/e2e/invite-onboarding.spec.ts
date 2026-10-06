import { expect, test } from '@playwright/test';

test.describe('Workspace invitation onboarding', () => {
	test('shows the invited workspace before sign-in', async ({ page }) => {
		await page.route('**/api/auth/session', (route) =>
			route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ error: 'Not authenticated' }) })
		);
		await page.route('**/api/auth/csrf', (route) =>
			route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ csrf_token: 'test-csrf' }) })
		);
		await page.route('**/api/**/workspaces/invites/validate', (route) =>
			route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					valid: true,
					workspace_name: 'Northstar Growth',
					email: 'invitee@example.com',
					role: 'member',
					expires_at: '2026-09-15'
				})
			})
		);

		await page.goto('/invite/northstar-token');

		await expect(page.getByRole('heading', { name: 'Join Northstar Growth' })).toBeVisible();
		await expect(page.locator('.invite-details span').filter({ hasText: 'invitee@example.com' })).toBeVisible();
	});

	test('redirects an unauthenticated desktop once instead of loading module login screens', async ({ page }) => {
		await page.route('**/api/auth/session', (route) =>
			route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ error: 'Not authenticated' }) })
		);

		await page.goto('/window');

		await expect(page).toHaveURL(/\/login\?redirect=%2Fwindow$/);
	});

	test('does not enter protected routes when the OAuth callback has no session', async ({ page }) => {
		await page.route('**/api/auth/session', (route) =>
			route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ user: null, session: null }) })
		);

		await page.goto('/auth/callback?redirect=%2Fwindow');

		await expect(page.getByRole('heading', { name: 'Sign in was not completed' })).toBeVisible();
		await expect(page).toHaveURL(/\/auth\/callback/);
		await expect(page.getByRole('link', { name: 'Return to sign in' })).toHaveAttribute('href', '/login?redirect=%2Fwindow');
	});
});
