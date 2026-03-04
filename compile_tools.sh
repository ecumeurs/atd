#!/bin/bash
set -e

# Base directories
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"
DEST_DIR="$PROJECT_ROOT/atd_management_skill/.agent/skills/atd/tools"

echo "======================================"
echo "Compiling ATD Toolchain"
echo "======================================"

# Ensure destination directory exists
mkdir -p "$DEST_DIR"

cd "$SCRIPTS_DIR"

# Compile all Go tools in scripts/
for tool_dir in atd-*/; do
    # Remove trailing slash
    tool_name="${tool_dir%/}"
    
    # Check if a main.go exists in the directory
    if [ -f "$tool_name/main.go" ]; then
        echo "[Build] Compiling $tool_name..."
        (cd "$tool_name" && go build -o "$DEST_DIR/$tool_name" . && go build -o "../bin/$tool_name" .)
    fi
done

# Copy any shell scripts directly into the destination
for script in *.sh; do
    if [ -f "$script" ]; then
        echo "[Copy] Copying script $script..."
        cp "$script" "$DEST_DIR/$script"
    fi
done

cp atd-cold-start.sh bin/atd-cold-start.sh
cp atd-full-audit.sh bin/atd-full-audit.sh

echo "======================================"
echo "[Permissions] Adding execute permissions to all tools..."
chmod +x "$DEST_DIR"/atd-*

echo "[Done] All tools compiled and installed to: $DEST_DIR"
