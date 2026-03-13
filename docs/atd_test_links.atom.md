---
id: atd_test_links
human_name: "ATD Test Links"
type: SPECIFICATION
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, test, traceability, test-link]
parents:
  - [[atd_cli]]
dependents: []
---

# ATD Test Links

## INTENT
To establish and verify bidirectional traceability between ATD atoms and their test coverage, using `@test-link [[ATOM_ID]]` tags in test files alongside the existing `@spec-link` convention.

## THE RULE / LOGIC
Scans source files for `@test-link [[ATOM_ID]]` annotations placed in test files (e.g. `_test.go`, `test_*.py`). When given a specific atom ID, walks the ATD hierarchy (parents and dependents) to collect all tests transitively linked to that atom's subtree. Outputs a JSON report: `[{atom_id, test_file, test_function, line}]`. This enables CI/CD checks that verify every STABLE atom has at least one linked test.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd test-links [--atom <id>] [--src <path>]`
- **LLM Task:** None (deterministic)
- **Convention:** `@test-link [[ATOM_ID]]` in test files
- **Code Tag:** `@spec-link [[atd_test_links]]`
