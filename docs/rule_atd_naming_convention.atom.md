---
id: rule_atd_naming_convention
human_name: "ATD Naming Convention"
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [atd, naming, convention, id, filename]
parents:
  - [[domain_atd_structure]]
dependents: []
---

# ATD Naming Convention

## INTENT
To enforce a deterministic, type-aware naming convention for ATD atom IDs and filenames.

## THE RULE / LOGIC
Every atom must follow this naming pattern:

```
ID:       <type_lowercase>_<descriptive_slug_in_snake_case>
Filename: <id>.atom.md
```

### Rules
1. **ID prefix:** The `id` field MUST start with the lowercase version of the atom's `type` (e.g., `service_`, `mechanic_`, `rule_`, `module_`).
2. **Slug:** After the type prefix, use a descriptive `snake_case` slug that reflects the atom's `human_name`.
3. **Filename:** The filename MUST be `<id>.atom.md` — no exceptions, no variations.
4. **No project prefix:** Do NOT add project-level prefixes (e.g., `atd_`) to the ID. The type prefix is sufficient for namespace differentiation.
5. **Case:** The type prefix is always lowercase. The slug is always `snake_case`.

### Examples

| Type | human_name | Correct ID | Filename |
|---|---|---|---|
| `SERVICE` | "ATD Search" | `service_atd_search` | `service_atd_search.atom.md` |
| `MECHANIC` | "Index Chunking" | `mechanic_index_chunking` | `mechanic_index_chunking.atom.md` |
| `RULE` | "ATD Naming Convention" | `rule_atd_naming_convention` | `rule_atd_naming_convention.atom.md` |
| `MODULE` | "ATD CLI" | `module_atd_cli` | `module_atd_cli.atom.md` |
| `DOMAIN` | "ATD Philosophy" | `domain_atd_philosophy` | `domain_atd_philosophy.atom.md` |

### Enforcement
- `atd update` automatically enforces this convention when an `id` is set alongside a `type`.
- `atd lint` should flag atoms whose filename does not match `<id>.atom.md`.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_atd_naming_convention]]`
- **Enforced by:** `update.go:runUpdate()` lines 232-250
- **Validated by:** `atd lint` (structural check)

## EXPECTATION (For Testing)
- Given type=SERVICE and id="atd_search", the enforced ID should be "service_atd_search".
- Given type=MECHANIC and id="index_chunking", the enforced ID should remain "mechanic_index_chunking" (already correct).
- All filenames in `docs/` must match the pattern `<id>.atom.md`.
