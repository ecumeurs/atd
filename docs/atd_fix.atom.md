---
id: atd_fix
human_name: "ATD Fix"
type: MECHANIC
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, audit, fix, split]
parents:
  - [[atd_audit]]
dependents: []
---

# ATD Fix

## INTENT
To automatically split bloated atoms identified by `atd audit` into properly sized child atoms, each containing exactly one rule.

## THE RULE / LOGIC
Reads an audit report file, extracts all `[BLOATED]` entries, and for each bloated atom: reads its content, sends it to the LLM (task `fix_split`) requesting a JSON split proposal with `{parent_logic, splits[{id_suffix, human_name, intent, logic}]}`. If accepted, writes child `.atom.md` files and rewrites the original as a MODULE parent. Invalidates the SQLite audit cache for all modified files.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd fix --audit <report.txt> [--dry-run]`
- **LLM Task:** `fix_split`
- **Code Tag:** `@spec-link [[atd_fix]]`
