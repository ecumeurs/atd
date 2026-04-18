#!/bin/bash
set -e

# Base directories
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPTS_DIR="$PROJECT_ROOT/atd"
SKILL_TOOLS_DIR="$PROJECT_ROOT/atd_management_skill/.agent/skills/atd/tools"
DEST_BIN_DIR="$SCRIPTS_DIR/bin"

echo "======================================"
echo "Compiling ATD Toolchain"
echo "======================================"

# Ensure destination directories exist
mkdir -p "$DEST_BIN_DIR/lib"
mkdir -p "$SKILL_TOOLS_DIR/lib"

# 1. Compile atd (single binary)
cd "$SCRIPTS_DIR"
echo "[Build] Compiling atd..."
go build -o "$DEST_BIN_DIR/atd" "./cmd/atd/main.go"

# Mirror to skill tools directory
cp "$DEST_BIN_DIR/atd" "$SKILL_TOOLS_DIR/atd"

# 2. Copy Shell Scripts and Libraries
echo "[Copy] Synchronizing shell scripts..."
for script in *.sh; do
    if [ -f "$script" ]; then
        echo "  [+] $script"
        cp "$script" "$DEST_BIN_DIR/$script"
        cp "$script" "$SKILL_TOOLS_DIR/$script"
    fi
done

# Copy logging library
if [ -f "lib/logging.sh" ]; then
    cp lib/logging.sh "$DEST_BIN_DIR/lib/"
    cp lib/logging.sh "$SKILL_TOOLS_DIR/lib/"
fi

# 3. Synchronize Config & Rules
echo "[Sync] Updating configuration and rules..."
cp "$PROJECT_ROOT/.atd" "$PROJECT_ROOT/atd_management_skill/" 2>/dev/null || true
cp "$PROJECT_ROOT/atd_management_skill/.agent/rules/ATD.md" "$PROJECT_ROOT/.agent/rules/ATD.md" 2>/dev/null || true

# 4. Install to System Path (handling "Text file busy")
echo "[Install] Installing atd binary to ~/.local/bin/atd..."
set +e
# Stop any running atd instances
killall atd 2>/dev/null
# Give it a second to die
sleep 1
# Remove old binary to avoid "Text file busy"
rm -f "$HOME/.local/bin/atd"
set -e
cp "$DEST_BIN_DIR/atd" "$HOME/.local/bin/atd"
chmod +x "$HOME/.local/bin/atd"

# 5. Fix Permissions
echo "[Permissions] Fixing permissions..."
chmod +x "$DEST_BIN_DIR"/atd* 2>/dev/null || true
chmod +x "$SKILL_TOOLS_DIR"/atd* 2>/dev/null || true

# 6. Extensions
echo "[Extensions] Installing VS Code / Antigravity extensions..."
VSC_EXT="$HOME/.vscode/extensions/local-dev.atd-linker"
AG_EXT="$HOME/.antigravity/extensions/local-dev.atd-linker"

rm -rf "$VSC_EXT" "$AG_EXT"
if [ -d "$PROJECT_ROOT/extension" ]; then
    cp -r "$PROJECT_ROOT/extension" "$VSC_EXT"
    cp -r "$PROJECT_ROOT/extension" "$AG_EXT"
fi

echo "======================================"
echo "[Done] Toolchain compiled and installed."
echo "  - Local Binary: scripts/bin/atd"
echo "  - System Binary: ~/.local/bin/atd"
echo "======================================"