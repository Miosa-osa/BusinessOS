#!/usr/bin/env python3
import os
import re
from pathlib import Path

def check_file(filepath):
    """Check if file uses slog but doesn't import log/slog"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()

        # Check if file uses slog. or slog.Logger or slog.Default()
        uses_slog = bool(re.search(r'\bslog\.', content))

        # Check if file imports log/slog
        has_import = bool(re.search(r'"log/slog"', content))

        if uses_slog and not has_import:
            return True
    except Exception as e:
        print(f"Error reading {filepath}: {e}")
    return False

def main():
    missing_imports = []
    internal_dir = Path('internal')

    for go_file in internal_dir.rglob('*.go'):
        if check_file(go_file):
            missing_imports.append(str(go_file))

    if missing_imports:
        print("Files using slog without importing log/slog:")
        for file in missing_imports:
            print(f"  - {file}")
        print(f"\nTotal: {len(missing_imports)} files")
    else:
        print("All files have correct slog imports!")

    return len(missing_imports)

if __name__ == '__main__':
    exit(main())
