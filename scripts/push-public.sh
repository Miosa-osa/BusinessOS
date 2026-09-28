#!/bin/bash
# Push BusinessOS to the public repo, EXCLUDING:
# - Admin panel (platform administration)
# - Platform admin backend handler
# - Platform admin migrations (superadmin, audit log)
# - Personal .env files
# - Engine index database

set -e

PUBLIC_REMOTE="public"
EXCLUDE_FILES=(
    "frontend/src/routes/(app)/admin/"
    "desktop/backend-go/internal/handlers/platform_admin.go"
    "desktop/backend-go/internal/database/migrations/103_platform_admin.sql"
    "desktop/backend-go/internal/database/migrations/104_seed_superadmin.sql"
    "desktop/backend-go/.env"
    ".system/index.db"
)

echo "=== Preparing public push ==="
echo "Excluded files:"
for f in "${EXCLUDE_FILES[@]}"; do
    echo "  - $f"
done

echo ""
echo "This will force-push to the public BusinessOS repo."
echo "Press Ctrl+C to cancel, Enter to continue."
read

# Create a temporary branch for the filtered push
git checkout -b public-push-temp

# Remove excluded files from the index (not from disk)
for f in "${EXCLUDE_FILES[@]}"; do
    git rm -r --cached "$f" 2>/dev/null || true
done

git commit -m "Public release — admin module excluded" --allow-empty

# Push to public remote
git remote add "$PUBLIC_REMOTE" https://github.com/Miosa-osa/businessos.git 2>/dev/null || true
git push "$PUBLIC_REMOTE" public-push-temp:main --force

# Go back to main and delete temp branch
git checkout main
git branch -D public-push-temp

echo "=== Pushed to public repo (admin excluded) ==="
