---
id: mechanic_search_grep
human_name: "Keyword Search (Grep Mode)"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: DRAFT
priority: 2
tags: [atd, search, grep, keyword]
parents:
  - [[service_atd_search]]
dependents: []
---

# Keyword Search (Grep Mode)

## INTENT
To describe how `atd search --grep` performs literal keyword matching across project files.

## THE RULE / LOGIC
1. **Root discovery:** Uses `config.ProjectRoot()` to determine the project root directory.
2. **File walk:** Recursively walks the project directory using `filepath.Walk`.
3. **Path exclusions:** Skips directories containing `/.git/` or `/.atd`.
4. **Extension filter:** Only scans files matching `supported_extensions` from `.atd` config OR files ending in `.atom.md`.
5. **String match:** Performs a case-sensitive `strings.Contains()` check on file content.
6. **Output:** 
   - Default: Prints each matching file path (relative to project root) with a `Grep: Found match in <path>` prefix.
   - Paths Only: Prints unique absolute file paths, one per line.

This mode does NOT use the vector index or any LLM. It is a pure filesystem operation.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[mechanic_search_grep]]`
- **Implementation:** `search.go:runGrepSearch()`
- **Depends on:** `config.ProjectRoot()`, `config.ActiveConfig.SupportedExtensions`

## EXPECTATION (For Testing)
- Searching for a known string should return at least one match.
- Files in `.git/` and `.atd` directories must never appear in results.
- Binary files and unsupported extensions must be excluded.
- When `paths_only` is true, the response must contain only unique absolute file paths.
