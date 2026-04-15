---
id: mechanic_webui_document_generation
status: DRAFT
tags: [webui, llm, generation, caching]
human_name: Document Generation Pipeline
version: 1.0
priority: 4
parents:
  - [[ui_webui_document_viewer]]
dependents: []
type: MECHANIC
layer: IMPLEMENTATION
---

# New Atom

## INTENT
To process a list of chosen ATDs from a semantic search into a single coherent AI-synthesized document based on the provided intent and caches it.

## THE RULE / LOGIC
1. A `search-document-context` endpoint proxies `atd search` queries to return the top 5 relevant atoms.
2. The user validates the selected IDs and provides an intent string.
3. The `generate-document` API proxies to `atd assemble --starts <ids> --intent <intent> --structured --json`.
4. The backend parses this result, saving the full text and linked metadata in an LRU (capacity 5).
5. If the capacity overflows, the oldest item will be evicted.

## TECHNICAL INTERFACE

## EXPECTATION
The document is structured and generated containing user-selected dependencies and correctly stores in the LRU up to 5 elements.
