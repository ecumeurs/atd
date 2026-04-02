---
id: service_atd_roadmap
human_name: "ATD Roadmap"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, roadmap, coverage, scanning]
parents:
  - [[module_atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Roadmap

## INTENT
To scan a codebase's structural elements (structs, interfaces, functions, classes) and produce a coverage roadmap showing which elements are linked to ATD atoms and which are pending documentation.

## THE RULE / LOGIC
Uses `git ls-files` to enumerate tracked files matching `supported_extensions`, then applies language-agnostic regex patterns to detect structural keywords (class, struct, interface, func, def, fn, type, etc.). For each detected element, checks if a preceding line contains `@spec-link` — marking it as DOCUMENTED or PENDING. Outputs a JSON roadmap with file, line, keyword, name, status, and ATD link.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd roadmap --dir <path> --out <roadmap.json>`
- **LLM Task:** None (deterministic)
- **Code Tag:** `@spec-link [[service_atd_roadmap]]`
