#!/usr/bin/env bash

# ATD Skill Initialization Script
# This runs when the skill is loaded to ensure necessary structures exist

echo "Initializing ATD Management Skill..."

# Dependency Checks
if ! command -v docker &> /dev/null; then
    echo "[!] FATAL: Docker is not installed or not in PATH."
    echo "Please install Docker to run the local Ollama LLM and Nomic Vector Indexer."
    exit 1
fi

if ! curl -s http://localhost:11434/api/tags &> /dev/null; then
    echo "[!] ERROR: Ollama API is not reachable at http://localhost:11434."
    echo "To start the local Ollama container, please run:"
    echo "  docker run -d -v ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama"
    exit 1
fi

models=$(curl -s http://localhost:11434/api/tags)
if [[ ! "$models" == *"llama3.2"* ]]; then
    echo "[!] WARNING: llama3.2 model not found locally."
    echo "Please run: docker exec -it ollama ollama pull llama3.2"
fi

if [[ ! "$models" == *"nomic-embed-text"* ]]; then
    echo "[!] WARNING: nomic-embed-text model not found locally."
    echo "Please run: docker exec -it ollama ollama pull nomic-embed-text"
fi

DOCS_DIR="./docs"

if [ -d "$DOCS_DIR" ]; then
    echo "ATD: $DOCS_DIR directory already exists. Preserving atoms."
elif [ -L "$DOCS_DIR" ]; then
    echo "ATD: $DOCS_DIR is a symlink. Preserving linkage."
else
    echo "ATD: $DOCS_DIR missing. Creating Atomic Documentation directory..."
    mkdir -p "$DOCS_DIR"
    
    # Optionally drop a master template so the user knows where to start
    cat << 'EOF' > "$DOCS_DIR/template.atom.md"
---
id: [[master_blueprint]]
human_name: Master Blueprint
type: CORE
version: 1.0
status: DRAFT
priority: CORE
tags: [architecture]
parents: []
dependents: []
---

# Master Blueprint

## 🎯 INTENT
Define the foundational logic for the system here.

## ⚙️ THE RULE / LOGIC
- Rule 1: ...

## 🔌 TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `N/A`
- **Code Tag:** `@spec-link [[master_blueprint]]`
EOF
    echo "ATD: Initialization complete. Seeded a template atom."
fi
