package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/knowledgefilter"
)

type document struct {
	path    string
	title   string
	body    string
	section string
}

type workspaceSpec struct {
	slug string
	name string
}

func main() {
	var databaseURL string
	var root string
	var ownerEmail string
	var workspaceList string
	var engineURL string
	var engineTenant string
	var dryRun bool

	flag.StringVar(&databaseURL, "database-url", os.Getenv("DATABASE_URL"), "BusinessOS PostgreSQL URL")
	flag.StringVar(&root, "root", os.Getenv("KNOWLEDGE_WORKSPACES_ROOT"), "OptimalOS workspaces directory")
	flag.StringVar(&ownerEmail, "owner-email", "", "BusinessOS owner email")
	flag.StringVar(&workspaceList, "workspaces", "", "comma-separated slug or slug=Display Name entries")
	flag.StringVar(&engineURL, "engine-url", "http://127.0.0.1:4200", "Optimal Engine base URL stored on each workspace")
	flag.StringVar(&engineTenant, "engine-tenant", "default", "Optimal Engine tenant prefix")
	flag.BoolVar(&dryRun, "dry-run", false, "inspect without changing PostgreSQL")
	flag.Parse()

	if databaseURL == "" || root == "" || ownerEmail == "" || workspaceList == "" {
		log.Fatal("database-url, root, owner-email, and workspaces are required")
	}

	specs, err := parseWorkspaceSpecs(workspaceList)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	var ownerID string
	if err := pool.QueryRow(ctx, `SELECT id FROM "user" WHERE lower(email)=lower($1)`, ownerEmail).Scan(&ownerID); err != nil {
		log.Fatalf("resolve owner %s: %v", ownerEmail, err)
	}

	for _, spec := range specs {
		dir := filepath.Join(root, spec.slug)
		if !isWorkspaceProjection(dir) {
			log.Fatalf("%s is not a durable OptimalOS workspace projection", dir)
		}
		docs, err := collectDocuments(dir)
		if err != nil {
			log.Fatalf("collect %s: %v", spec.slug, err)
		}
		fmt.Printf("%-28s %4d documents", spec.slug, len(docs))
		if dryRun {
			fmt.Println(" [dry run]")
			continue
		}
		if err := syncWorkspace(ctx, pool, ownerID, spec, engineURL, engineTenant, docs); err != nil {
			log.Fatalf("sync %s: %v", spec.slug, err)
		}
		fmt.Println(" [synced]")
	}
}

func parseWorkspaceSpecs(raw string) ([]workspaceSpec, error) {
	seen := map[string]bool{}
	var specs []workspaceSpec
	for _, entry := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), "=", 2)
		slug := strings.TrimSpace(parts[0])
		if slug == "" || seen[slug] {
			continue
		}
		name := titleFromName(slug)
		if len(parts) == 2 && strings.TrimSpace(parts[1]) != "" {
			name = strings.TrimSpace(parts[1])
		}
		if strings.ContainsAny(slug, "/\\") || slug == "." || slug == ".." {
			return nil, fmt.Errorf("invalid workspace slug %q", slug)
		}
		seen[slug] = true
		specs = append(specs, workspaceSpec{slug: slug, name: name})
	}
	if len(specs) == 0 {
		return nil, fmt.Errorf("no workspaces supplied")
	}
	return specs, nil
}

func isWorkspaceProjection(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	for _, marker := range []string{".wiki", ".optimal", "workspace.yaml", "organization.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

func collectDocuments(root string) ([]document, error) {
	var docs []document
	ignoreRules := knowledgefilter.Load(root)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && strings.HasPrefix(entry.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(entry.Name(), ".") || !strings.EqualFold(filepath.Ext(entry.Name()), ".md") {
			return nil
		}
		lowerName := strings.ToLower(entry.Name())
		if strings.HasSuffix(lowerName, ".abstract.md") || strings.HasSuffix(lowerName, ".overview.md") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if knowledgefilter.Ignored(rel, ignoreRules) {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		section := "docs"
		if before, _, ok := strings.Cut(rel, "/"); ok {
			section = before
		}
		docs = append(docs, document{
			path: rel, title: titleFromName(entry.Name()), body: string(body), section: section,
		})
		return nil
	})
	sort.Slice(docs, func(i, j int) bool { return docs[i].path < docs[j].path })
	return docs, err
}

func titleFromName(name string) string {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	words := strings.Fields(strings.NewReplacer("-", " ", "_", " ").Replace(name))
	for i, word := range words {
		switch strings.ToLower(word) {
		case "ai":
			words[i] = "AI"
		case "miosa":
			words[i] = "MIOSA"
		case "osa":
			words[i] = "OSA"
		case "iq":
			words[i] = "IQ"
		default:
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
	}
	return strings.Join(words, " ")
}

func syncWorkspace(ctx context.Context, pool *pgxpool.Pool, ownerID string, spec workspaceSpec, engineURL, engineTenant string, docs []document) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var workspaceID string
	engineWorkspace := strings.TrimSuffix(engineTenant, ":") + ":" + spec.slug
	err = tx.QueryRow(ctx, `
		INSERT INTO workspaces (name, slug, description, plan_type, max_members, max_projects, max_storage_gb, settings, owner_id)
		VALUES ($1::varchar, $2::varchar, 'Connected to the matching OptimalOS workspace', 'free', 50, 1000, 25,
			jsonb_build_object(
				'module_profile', 'primitive',
				'optimal_engine', jsonb_build_object('enabled', true, 'base_url', $3::text, 'workspace', $5::text)
			), $4)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			owner_id = EXCLUDED.owner_id,
			settings = jsonb_set(
				CASE WHEN workspaces.settings ? 'module_profile' THEN workspaces.settings
				ELSE jsonb_set(COALESCE(workspaces.settings, '{}'::jsonb), '{module_profile}', '"primitive"'::jsonb, true) END,
				'{optimal_engine}',
				jsonb_build_object('enabled', true, 'base_url', $3::text, 'workspace', $5::text),
				true
			),
			updated_at = NOW()
		RETURNING id::text
	`, spec.name, spec.slug, engineURL, ownerID, engineWorkspace).Scan(&workspaceID)
	if err != nil {
		return fmt.Errorf("upsert workspace: %w", err)
	}

	var roleCount int
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM workspace_roles WHERE workspace_id=$1::uuid`, workspaceID).Scan(&roleCount); err != nil {
		return fmt.Errorf("count roles: %w", err)
	}
	if roleCount == 0 {
		if _, err := tx.Exec(ctx, `SELECT seed_default_workspace_roles($1::uuid)`, workspaceID); err != nil {
			return fmt.Errorf("seed roles: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role_id, role_name, role, status, joined_at)
		SELECT $1::uuid, $2, id, 'owner', 'owner', 'active', NOW()
		FROM workspace_roles WHERE workspace_id=$1::uuid AND name='owner'
		ON CONFLICT (workspace_id, user_id) DO UPDATE SET
			role_id=EXCLUDED.role_id, role_name='owner', role='owner', status='active', updated_at=NOW()
	`, workspaceID, ownerID); err != nil {
		return fmt.Errorf("upsert membership: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO user_workspace_profiles (workspace_id, user_id, display_name, title, work_email, timezone)
		SELECT $1::uuid, id, name, 'Owner', email, 'America/Chicago' FROM "user" WHERE id=$2
		ON CONFLICT (user_id, workspace_id) DO UPDATE SET
			display_name=EXCLUDED.display_name, title=EXCLUDED.title, work_email=EXCLUDED.work_email,
			timezone=EXCLUDED.timezone, updated_at=NOW()
	`, workspaceID, ownerID); err != nil {
		return fmt.Errorf("upsert owner profile: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO sync_policies (workspace_id, module, sync_mode, is_published)
		SELECT $1::uuid, module, 'workspace', TRUE
		FROM (VALUES ('knowledge'), ('contexts'), ('projects'), ('tasks')) AS modules(module)
		ON CONFLICT (workspace_id, module) DO UPDATE SET sync_mode='workspace', is_published=TRUE
	`, workspaceID); err != nil {
		return fmt.Errorf("upsert sync policies: %w", err)
	}

	var bytesUsed int64
	if len(docs) > 0 {
		if _, err := tx.Exec(ctx, `DELETE FROM knowledge_documents WHERE workspace_slug=$1`, spec.slug); err != nil {
			return fmt.Errorf("clear knowledge copy: %w", err)
		}
		for _, doc := range docs {
			if _, err := tx.Exec(ctx, `
			INSERT INTO knowledge_documents (workspace_slug, workspace_id, path, title, body, section, synced_by, synced_at)
			VALUES ($1, $2::uuid, $3, $4, $5, $6, $7, $8)
		`, spec.slug, workspaceID, doc.path, doc.title, doc.body, doc.section, ownerID, time.Now().UTC()); err != nil {
				return fmt.Errorf("insert knowledge %s: %w", doc.path, err)
			}
			bytesUsed += int64(len(doc.body))
		}
	} else if err := tx.QueryRow(ctx,
		`SELECT COALESCE(SUM(octet_length(body)), 0) FROM knowledge_documents WHERE workspace_slug=$1`,
		spec.slug,
	).Scan(&bytesUsed); err != nil {
		return fmt.Errorf("measure existing knowledge: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO workspace_storage (workspace_id, bytes_used, updated_at)
		VALUES ($1::uuid, $2, NOW())
		ON CONFLICT (workspace_id) DO UPDATE SET bytes_used=EXCLUDED.bytes_used, updated_at=NOW()
	`, workspaceID, bytesUsed); err != nil {
		return fmt.Errorf("update storage: %w", err)
	}

	return tx.Commit(ctx)
}
