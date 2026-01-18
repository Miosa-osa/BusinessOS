#!/bin/bash

# Script to update old button patterns to btn-pill system
# Run from frontend directory

# Color definitions for btn-pill conversion
# bg-blue-600 hover:bg-blue-700 → btn-pill btn-pill-primary
# bg-red-600 hover:bg-red-700 → btn-pill btn-pill-danger
# bg-green-600 hover:bg-green-700 → btn-pill btn-pill-success
# bg-purple-600 hover:bg-purple-700 → btn-pill btn-pill-primary
# bg-gray-900 hover:bg-gray-800 → btn-pill btn-pill-primary

echo "Starting button style migration to btn-pill system..."

# Find all .svelte files
find src -name "*.svelte" -type f | while read -r file; do
  # Skip if file doesn't contain old button patterns
  if ! grep -q 'bg-\(blue\|red\|green\|purple\|gray\)-[5-9]00' "$file"; then
    continue
  fi

  echo "Processing: $file"

  # Create backup
  cp "$file" "${file}.bak"

  # Primary blue buttons - various formats
  # Full width with rounded-xl
  sed -i '' 's/class="\([^"]*\)w-full\([^"]*\)bg-blue-600 hover:bg-blue-700\([^"]*\)rounded-xl\([^"]*\)"/class="\1btn-pill btn-pill-primary btn-pill-block\4"/g' "$file"

  # Standard blue buttons with rounded-lg
  sed -i '' 's/class="\([^"]*\)bg-blue-600 hover:bg-blue-700\([^"]*\)rounded-lg\([^"]*\)"/class="\1btn-pill btn-pill-primary\3"/g' "$file"

  # Blue buttons with rounded-xl
  sed -i '' 's/class="\([^"]*\)bg-blue-600 hover:bg-blue-700\([^"]*\)rounded-xl\([^"]*\)"/class="\1btn-pill btn-pill-primary\3"/g' "$file"

  # Simple blue hover patterns
  sed -i '' 's/bg-blue-600 hover:bg-blue-700/btn-pill btn-pill-primary/g' "$file"
  sed -i '' 's/bg-blue-600 hover:bg-blue-500/btn-pill btn-pill-primary/g' "$file"
  sed -i '' 's/bg-blue-500 hover:bg-blue-600/btn-pill btn-pill-primary/g' "$file"

  # Danger red buttons
  sed -i '' 's/class="\([^"]*\)bg-red-600 hover:bg-red-700\([^"]*\)rounded-lg\([^"]*\)"/class="\1btn-pill btn-pill-danger\3"/g' "$file"
  sed -i '' 's/class="\([^"]*\)bg-red-500 hover:bg-red-600\([^"]*\)rounded-xl\([^"]*\)"/class="\1btn-pill btn-pill-danger\3"/g' "$file"
  sed -i '' 's/bg-red-600 hover:bg-red-700/btn-pill btn-pill-danger/g' "$file"
  sed -i '' 's/bg-red-500 hover:bg-red-600/btn-pill btn-pill-danger/g' "$file"

  # Success green buttons
  sed -i '' 's/class="\([^"]*\)bg-green-600 hover:bg-green-700\([^"]*\)rounded-lg\([^"]*\)"/class="\1btn-pill btn-pill-success\3"/g' "$file"
  sed -i '' 's/bg-green-600 hover:bg-green-700/btn-pill btn-pill-success/g' "$file"

  # Purple buttons (brand primary)
  sed -i '' 's/class="\([^"]*\)bg-purple-600 hover:bg-purple-700\([^"]*\)rounded-lg\([^"]*\)"/class="\1btn-pill btn-pill-primary\3"/g' "$file"
  sed -i '' 's/bg-purple-600 hover:bg-purple-700/btn-pill btn-pill-primary/g' "$file"

  # Gray dark buttons (primary in dark mode)
  sed -i '' 's/bg-gray-900 hover:bg-gray-800/btn-pill btn-pill-primary/g' "$file"
  sed -i '' 's/bg-gray-800 hover:bg-gray-700/btn-pill btn-pill-secondary/g' "$file"

  # Clean up redundant spacing and old sizing classes after conversion
  # Remove px-* py-* if btn-pill is present (btn-pill has its own padding)
  sed -i '' 's/class="\([^"]*\)btn-pill\([^"]*\)px-[0-9] py-[0-9]\([^"]*\)"/class="\1btn-pill\2\3"/g' "$file"
  sed -i '' 's/class="\([^"]*\)btn-pill\([^"]*\)px-[0-9]* py-[0-9.]*\([^"]*\)"/class="\1btn-pill\2\3"/g' "$file"

  # Remove rounded-* classes if btn-pill is present
  sed -i '' 's/class="\([^"]*\)btn-pill\([^"]*\)rounded-[a-z]*\([^"]*\)"/class="\1btn-pill\2\3"/g' "$file"

  # Check if file changed
  if ! diff -q "$file" "${file}.bak" > /dev/null 2>&1; then
    echo "  ✓ Updated: $file"
  else
    echo "  - No changes: $file"
    rm "${file}.bak"
  fi
done

echo ""
echo "Migration complete!"
echo "Review changes and run 'npm run build' to verify."
echo ""
echo "To restore from backups: find src -name '*.svelte.bak' -exec bash -c 'mv \"\$0\" \"\${0%.bak}\"' {} \;"
echo "To delete backups: find src -name '*.svelte.bak' -delete"
