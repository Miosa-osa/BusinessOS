package schemahealth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Requirement is one runtime schema object the current server code expects.
type Requirement struct {
	Kind   string
	Name   string
	Detail string
}

func (r Requirement) String() string {
	if r.Detail == "" {
		return r.Kind + " " + r.Name
	}
	return r.Kind + " " + r.Name + " (" + r.Detail + ")"
}

// Report is the result of checking a database against the runtime schema
// contract.
type Report struct {
	Missing []Requirement
}

func (r Report) OK() bool {
	return len(r.Missing) == 0
}

func (r Report) Error() string {
	if r.OK() {
		return ""
	}
	items := make([]string, 0, len(r.Missing))
	for _, req := range r.Missing {
		items = append(items, req.String())
	}
	return "database schema is missing required runtime objects: " + strings.Join(items, ", ")
}

// ErrSchemaDrift marks schema contract failures so callers can distinguish DB
// connectivity from a connected but wrong-shaped database.
var ErrSchemaDrift = errors.New("database schema drift")

type SchemaError struct {
	Report Report
}

func (e *SchemaError) Error() string {
	return e.Report.Error()
}

func (e *SchemaError) Unwrap() error {
	return ErrSchemaDrift
}

// Check verifies the current database has the schema required by the current
// BusinessOS runtime. It intentionally checks concrete tables and columns used
// by module handlers, not every historical table in schema.sql.
func Check(ctx context.Context, pool *pgxpool.Pool) (Report, error) {
	if pool == nil {
		return Report{}, fmt.Errorf("schema health check requires a database pool")
	}

	report := Report{}

	for _, table := range RequiredTables {
		exists, err := tableExists(ctx, pool, table)
		if err != nil {
			return report, err
		}
		if !exists {
			report.Missing = append(report.Missing, Requirement{
				Kind: "table",
				Name: table,
			})
		}
	}

	for table, columns := range RequiredColumns {
		for _, column := range columns {
			exists, err := columnExists(ctx, pool, table, column)
			if err != nil {
				return report, err
			}
			if !exists {
				report.Missing = append(report.Missing, Requirement{
					Kind:   "column",
					Name:   table + "." + column,
					Detail: "module runtime dependency",
				})
			}
		}
	}

	for _, function := range RequiredFunctions {
		exists, err := functionExists(ctx, pool, function)
		if err != nil {
			return report, err
		}
		if !exists {
			report.Missing = append(report.Missing, Requirement{
				Kind: "function",
				Name: function,
			})
		}
	}

	if !report.OK() {
		return report, &SchemaError{Report: report}
	}

	return report, nil
}

func tableExists(ctx context.Context, pool *pgxpool.Pool, table string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = $1
		)
	`, table).Scan(&exists)
	return exists, err
}

func columnExists(ctx context.Context, pool *pgxpool.Pool, table string, column string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = $1
			  AND column_name = $2
		)
	`, table, column).Scan(&exists)
	return exists, err
}

func functionExists(ctx context.Context, pool *pgxpool.Pool, function string) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM pg_proc p
			JOIN pg_namespace n ON n.oid = p.pronamespace
			WHERE n.nspname = 'public'
			  AND p.proname = $1
		)
	`, function).Scan(&exists)
	return exists, err
}
