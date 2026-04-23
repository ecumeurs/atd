#!/bin/bash
set -e

ATD_BIN="$(pwd)/bin/atd"
TEST_DIR="$(pwd)/test-fixtures/workspace-workflow"

echo "Running Workspace E2E Tests..."

cd "$TEST_DIR"

# 1. Initialize workspace
"$ATD_BIN" workspace init --name "e2e-ws"

# 2. Add projects
"$ATD_BIN" workspace add --name "frontend" --path ./frontend
"$ATD_BIN" workspace add --name "backend" --path ./backend

# 3. Verify workspace list
"$ATD_BIN" workspace list | grep "frontend"
"$ATD_BIN" workspace list | grep "backend"

# 4. Verify project isolation (stats)
cd frontend
"$ATD_BIN" stats | grep "\"project\": \"frontend\""
"$ATD_BIN" stats | grep "\"total_atoms\": 2"

cd ../backend
"$ATD_BIN" stats | grep "\"project\": \"backend\""
"$ATD_BIN" stats | grep "\"total_atoms\": 1"

# 5. Verify cross-project reference via crawl (optional if crawl --workspace is implemented)
# For now, let's just check if stats --workspace works
cd ..
"$ATD_BIN" stats --workspace | grep "\"total_atoms\": 3"

echo "Workspace E2E Tests PASSED!"
