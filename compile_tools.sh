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

# Compile all Go tools in scripts/cmd/
for tool_dir in cmd/*/; do
    # Remove trailing slash
    tool_dir="${tool_dir%/}"
    tool_name=$(basename "$tool_dir")
    
    # Check if a main.go exists in the directory
    if [ -f "$tool_dir/main.go" ]; then
        echo "[Build] Compiling $tool_name..."
        (go build -o "$DEST_DIR/$tool_name" "./$tool_dir" && go build -o "bin/$tool_name" "./$tool_dir")
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

echo "[Copy] Copying library scripts..."
mkdir -p "$DEST_DIR/lib"
cp lib/logging.sh "$DEST_DIR/lib/logging.sh"
mkdir -p bin/lib
cp lib/logging.sh bin/lib/logging.sh

echo "[Copy] Copying .atd configuration to skill folder..."
cp "$PROJECT_ROOT/.atd" "$PROJECT_ROOT/atd_management_skill/"

set +e
killall atd
set -e
cp bin/atd ~/.local/bin/atd

echo "======================================"
echo "[Permissions] Adding execute permissions to all tools..."
chmod +x "$DEST_DIR"/atd-*
chmod +x ~/.local/bin/atd

echo "[Done] All tools compiled and installed to: $DEST_DIR"
echo "[Done] atd binary installed to: ~/.local/bin/atd"

echo "======================================"
echo "Installing extensions..."

rm -rf ~/.vscode/extensions/local-dev.atd-linker
rm -rf ~/.antigravity/extensions/local-dev.atd-linker

cp -r ../extension ~/.vscode/extensions/local-dev.atd-linker
cp -r ../extension ~/.antigravity/extensions/local-dev.atd-linker

echo "[Done] All extensions installed to: ~/.vscode/extensions/local-dev.atd-linker and ~/.antigravity/extensions/local-dev.atd-linker"