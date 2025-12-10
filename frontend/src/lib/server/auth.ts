import { betterAuth } from 'better-auth';
import { Pool } from 'pg';

export const auth = betterAuth({
	database: new Pool({
		connectionString: 'postgresql://rhl@localhost:5432/business_os'
	}),
	emailAndPassword: {
		enabled: true
	},
	session: {
		expiresIn: 60 * 60 * 24 * 7, // 7 days
		updateAge: 60 * 60 * 24 // 1 day
	},
	advanced: {
		defaultCookieAttributes: {
			path: '/', // Ensure cookies are sent to all routes
			sameSite: 'lax' as const,
			httpOnly: true
		}
	}
});
