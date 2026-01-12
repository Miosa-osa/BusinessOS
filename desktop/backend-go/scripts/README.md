# Development Scripts

This directory contains organized development and debugging scripts.

## Structure

```
scripts/
├── tests/          # Test scripts and verification tools
├── debug/          # Debug and inspection utilities
├── migrations/     # Migration runners and database tools
└── utils/          # General utilities
```

## Usage Guidelines

### Tests (`scripts/tests/`)
Scripts for testing functionality:
- `test_*.go` - Integration tests
- `verify_*.go` - Verification scripts
- `*.sh` - Shell-based test runners

**Example:**
```bash
go run scripts/tests/test_workspace_api.go
./scripts/tests/test_all_endpoints.sh
```

### Debug (`scripts/debug/`)
Scripts for debugging and inspection:
- `check_*.go` - Database inspection
- `debug_*.go` - Debug utilities
- `create_*.go` - Test data creation

**Example:**
```bash
go run scripts/debug/check_workspace_memories.go
go run scripts/debug/create_test_workspace.go
```

### Migrations (`scripts/migrations/`)
Database migration tools:
- `run_*.go` - Migration runners
- `run_*.sh` - Migration shell scripts

**Example:**
```bash
go run scripts/migrations/run_q1_migrations.go
./scripts/migrations/run_workspace_tests.sh
```

## Best Practices

1. **Never commit one-off scripts to root**
   - Keep root clean
   - Organize here or in `/tmp`

2. **Name consistently**
   - Tests: `test_<feature>.go`
   - Debug: `check_<what>.go` or `debug_<what>.go`
   - Migrations: `run_<migration_name>.go`

3. **Document complex scripts**
   - Add comments explaining what it does
   - Include usage examples

4. **Clean up when done**
   - Move reusable scripts here
   - Delete one-off debug scripts
   - Keep only what's valuable

## Temporary Scripts

For quick one-off scripts, use:
```bash
# Create in tmp/ (gitignored)
mkdir -p tmp
echo "package main ..." > tmp/quick_test.go
go run tmp/quick_test.go
# Delete when done
```
