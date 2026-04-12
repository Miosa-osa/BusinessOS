import { Pool } from 'pg';

const pool = new Pool({
	connectionString: 'postgresql://rhl@localhost:5432/business_os'
});

export { pool };
