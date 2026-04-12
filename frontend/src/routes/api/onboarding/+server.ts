import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { pool } from '$lib/server/db';

interface OnboardingData {
	business_type: string;
	business_name: string;
	role: string;
	team_size: string;
	use_cases: string[];
	workspace_name: string;
	theme: string;
}

export const POST: RequestHandler = async ({ request, cookies }) => {
	try {
		// Get current session
		const sessionToken = cookies.get('better-auth.session_token');
		if (!sessionToken) {
			return json({ error: 'Not authenticated' }, { status: 401 });
		}

		// Get user from session
		const result = await pool.query(
			`SELECT s.*, u.id as user_id, u.name, u.email
			 FROM session s
			 JOIN "user" u ON s."userId" = u.id
			 WHERE s.token = $1 AND s."expiresAt" > NOW()`,
			[sessionToken]
		);

		if (result.rows.length === 0) {
			return json({ error: 'Invalid session' }, { status: 401 });
		}

		const userId = result.rows[0].user_id;
		const data: OnboardingData = await request.json();

		// Upsert user_profile
		await pool.query(
			`INSERT INTO user_profile (
				user_id,
				business_type,
				business_name,
				role,
				team_size,
				use_cases,
				workspace_name,
				theme,
				onboarding_completed,
				created_at,
				updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, NOW(), NOW())
			ON CONFLICT (user_id) DO UPDATE SET
				business_type = $2,
				business_name = $3,
				role = $4,
				team_size = $5,
				use_cases = $6,
				workspace_name = $7,
				theme = $8,
				onboarding_completed = true,
				updated_at = NOW()`,
			[
				userId,
				data.business_type,
				data.business_name,
				data.role,
				data.team_size,
				JSON.stringify(data.use_cases),
				data.workspace_name,
				data.theme
			]
		);

		return json({ success: true });
	} catch (error) {
		console.error('Onboarding save error:', error);
		return json({ error: 'Failed to save onboarding data' }, { status: 500 });
	}
};

export const GET: RequestHandler = async ({ cookies }) => {
	try {
		// Get current session
		const sessionToken = cookies.get('better-auth.session_token');
		if (!sessionToken) {
			return json({ error: 'Not authenticated' }, { status: 401 });
		}

		// Get user from session
		const sessionResult = await pool.query(
			`SELECT s.*, u.id as user_id
			 FROM session s
			 JOIN "user" u ON s."userId" = u.id
			 WHERE s.token = $1 AND s."expiresAt" > NOW()`,
			[sessionToken]
		);

		if (sessionResult.rows.length === 0) {
			return json({ error: 'Invalid session' }, { status: 401 });
		}

		const userId = sessionResult.rows[0].user_id;

		// Get user profile
		const profileResult = await pool.query(
			`SELECT * FROM user_profile WHERE user_id = $1`,
			[userId]
		);

		if (profileResult.rows.length === 0) {
			return json({ data: null, onboarding_completed: false });
		}

		const profile = profileResult.rows[0];
		return json({
			data: {
				business_type: profile.business_type,
				business_name: profile.business_name,
				role: profile.role,
				team_size: profile.team_size,
				use_cases: typeof profile.use_cases === 'string' ? JSON.parse(profile.use_cases) : profile.use_cases,
				workspace_name: profile.workspace_name,
				theme: profile.theme
			},
			onboarding_completed: profile.onboarding_completed
		});
	} catch (error) {
		console.error('Onboarding get error:', error);
		return json({ error: 'Failed to get onboarding data' }, { status: 500 });
	}
};
