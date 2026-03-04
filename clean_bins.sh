#!/bin/bash
# Description: Remove compiled binaries from scripts subfolders, keeping scripts/bin

# Find and remove executable files in scripts subdirectories,
# excluding the main bin folder and source/script files.
# -mindepth 2 ensures we look inside subdirectories of scripts, not scripts itself.
find scripts -mindepth 2 -type f -executable \
    -not -path "scripts/bin/*" \
    -not -name "*.sh" \
    -not -name "*.go" \
    -exec rm -v {} +
