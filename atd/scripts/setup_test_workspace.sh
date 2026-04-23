#!/bin/bash

# setup_test_workspace.sh
# Creates a fake monorepo structure for validating ATD workspace support.

set -e

WORKSPACE_DIR="test-workspace"

echo "Creating workspace structure in $WORKSPACE_DIR..."

mkdir -p "$WORKSPACE_DIR/project-a/docs"
mkdir -p "$WORKSPACE_DIR/project-a/src"
mkdir -p "$WORKSPACE_DIR/project-b/docs"
mkdir -p "$WORKSPACE_DIR/project-b/src"

# 1. Create Workspace Config
cat > "$WORKSPACE_DIR/.atd.workspace" <<EOF
{
  "workspace_name": "test-workspace",
  "projects": [
    {
      "name": "project-a",
      "path": "./project-a"
    },
    {
      "name": "project-b",
      "path": "./project-b"
    }
  ]
}
EOF

# 2. Project A Setup
cat > "$WORKSPACE_DIR/project-a/.atd" <<EOF
{
  "docs_path": "docs/",
  "code_paths": ["src/"]
}
EOF

cat > "$WORKSPACE_DIR/project-a/docs/atom-a.atom.md" <<EOF
---
id: atom-a
human_name: "Atom A"
type: RULE
layer: ARCHITECTURE
status: STABLE
---

# Atom A
## INTENT
Requirement for project A.
EOF

cat > "$WORKSPACE_DIR/project-a/src/main.go" <<EOF
package main

// @spec-link [[atom-a]]
func Main() {}
EOF

# 3. Project B Setup
cat > "$WORKSPACE_DIR/project-b/.atd" <<EOF
{
  "docs_path": "docs/",
  "code_paths": ["src/"]
}
EOF

cat > "$WORKSPACE_DIR/project-b/docs/atom-b.atom.md" <<EOF
---
id: atom-b
human_name: "Atom B"
type: MECHANIC
layer: IMPLEMENTATION
status: STABLE
parents:
  - [[project-a:atom-a]]
---

# Atom B
## INTENT
Implementation in project B that depends on A.
EOF

cat > "$WORKSPACE_DIR/project-b/src/lib.go" <<EOF
package lib

// @spec-link [[atom-b]]
func Lib() {}
EOF

echo "Done! Structure created in $WORKSPACE_DIR."
echo "Tree view:"
ls -R "$WORKSPACE_DIR"
