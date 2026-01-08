package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	fmt.Println("=== ADDING DEBUG LOGGING TO FUNCTION ===\n")

	// Recreate function with RAISE NOTICE for debugging
	_, err = pool.Exec(ctx, `
CREATE OR REPLACE FUNCTION get_accessible_memories(
    p_workspace_id UUID,
    p_user_id TEXT,
    p_memory_type TEXT DEFAULT NULL,
    p_limit INT DEFAULT 100
)
RETURNS TABLE (
    id UUID,
    title TEXT,
    content TEXT,
    memory_type TEXT,
    visibility TEXT,
    importance NUMERIC,
    tags TEXT[],
    metadata JSONB,
    is_owner BOOLEAN,
    access_count INT,
    created_at TIMESTAMPTZ
) AS $$
DECLARE
    v_member_exists BOOLEAN;
    v_row_count INT;
BEGIN
    RAISE NOTICE '=== FUNCTION START ===';
    RAISE NOTICE 'p_workspace_id: %', p_workspace_id;
    RAISE NOTICE 'p_user_id: %', p_user_id;
    RAISE NOTICE 'p_memory_type: %', p_memory_type;
    RAISE NOTICE 'p_limit: %', p_limit;

    -- Check membership
    SELECT EXISTS (
        SELECT 1 FROM workspace_members
        WHERE workspace_id = p_workspace_id
        AND user_id = p_user_id
        AND status = 'active'
    ) INTO v_member_exists;

    RAISE NOTICE 'Membership check: %', v_member_exists;

    IF NOT v_member_exists THEN
        RAISE NOTICE 'User is not a member, returning empty';
        RETURN;
    END IF;

    -- Count matching rows
    SELECT COUNT(*) INTO v_row_count
    FROM workspace_memories wm
    WHERE wm.workspace_id = p_workspace_id
    AND wm.is_active = true
    AND (
        wm.visibility = 'workspace' OR wm.visibility IS NULL
        OR
        (wm.visibility = 'private' AND wm.owner_user_id = p_user_id)
        OR
        (wm.visibility = 'shared' AND (wm.owner_user_id = p_user_id OR p_user_id = ANY(COALESCE(wm.shared_with, ARRAY[]::TEXT[]))))
    )
    AND (p_memory_type IS NULL OR wm.memory_type = p_memory_type);

    RAISE NOTICE 'Matching rows: %', v_row_count;

    RETURN QUERY
    SELECT
        wm.id,
        wm.title,
        wm.content,
        wm.memory_type,
        wm.visibility,
        wm.importance_score as importance,
        wm.tags,
        wm.metadata,
        (wm.owner_user_id = p_user_id OR wm.owner_user_id IS NULL) as is_owner,
        wm.access_count,
        wm.created_at
    FROM workspace_memories wm
    WHERE wm.workspace_id = p_workspace_id
    AND wm.is_active = true
    AND (
        wm.visibility = 'workspace' OR wm.visibility IS NULL
        OR
        (wm.visibility = 'private' AND wm.owner_user_id = p_user_id)
        OR
        (wm.visibility = 'shared' AND (wm.owner_user_id = p_user_id OR p_user_id = ANY(COALESCE(wm.shared_with, ARRAY[]::TEXT[]))))
    )
    AND (p_memory_type IS NULL OR wm.memory_type = p_memory_type)
    ORDER BY wm.importance_score DESC NULLS LAST, wm.created_at DESC
    LIMIT p_limit;

    RAISE NOTICE '=== FUNCTION END ===';
END;
$$ LANGUAGE plpgsql;
	`)

	if err != nil {
		log.Fatalf("Failed to update function: %v", err)
	}

	fmt.Println("✓ Function updated with debugging\n")
	fmt.Println("Now calling the function...\n")

	// Get a connection to listen for notices
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Release()

	// Set up notice handler
	conn.Conn().Config().OnNotice = func(pc *pgconn.PgConn, n *pgconn.Notice) {
		fmt.Printf("[NOTICE] %s\n", n.Message)
	}

	// Call it
	rows, err := conn.Query(ctx, "SELECT * FROM get_accessible_memories($1, $2)",
		uuid.MustParse("064e8e2a-5d3e-4d00-8492-df3628b1ec96"), "ZVtQRaictVbO9lN0p-csSA")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	fmt.Printf("\nFunction returned: %d rows\n", count)
}
