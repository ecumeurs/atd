---
id: atd_discover
human_name: "ATD Discover"
type: SERVICE
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, discover, links, intent]
parents:
  - [[atd_cli]]
dependents: []
---

# ATD Discover

## INTENT
To automatically discover which ATD atoms govern a given source code file by extracting the code's architectural intent via LLM and matching it semantically against the ATD docs index.

## THE RULE / LOGIC
Two-step process: (1) Send the code file to the LLM (task `intent_extract`) to get a 2-sentence architectural summary. (2) Embed that summary via Nomic and cosine-rank it against all atom doc embeddings in the SQLite index. Returns top 3 candidate atoms with similarity scores and a recommendation prompt for final validation.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd discover -file <code_path>`
- **LLM Tasks:** `intent_extract`, `embed`
- **Code Tag:** `@spec-link [[atd_discover]]`
