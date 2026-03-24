# Issue: Document Hierarchy (Layer Field) and Numeric Priority

**ID:** `20260324_layer_field_priority`
**Ref:** `ISS-042`
**Date:** 2026-03-24
**Severity:** High
**Status:** Resolved
**Component:** `scripts/cmd/atd`, `.atd` config, all atom parsers
**Affects:** All ATD tools, MCP server, atom templates, existing atoms

---

## Summary

ATD is introducing two breaking frontmatter changes: (1) a new `layer` field (`CUSTOMER` / `ARCHITECTURE` / `IMPLEMENTATION`) that classifies atoms into a documentation hierarchy governing volatility and human oversight, and (2) changing `priority` from an enum (`CORE` / `SECONDARY` / `EXPERIMENTAL` / `FLAVOR`) to a numeric scale (`1`–`5`). All parsers, validators, MCP tools, templates, and existing atoms must be updated.

---

## Technical Description

### Background
The ATD framework previously had no formal hierarchy distinguishing customer requirements from architecture designs from implementation details. All atoms were treated equally in terms of governance. The `priority` field used an enum that provided no granularity for ranking.

### The New `layer` Field
- `CUSTOMER`: Requirements, specifications, use cases, domain context. Low volatility. Near-immutable once `STABLE`. Heavy human sign-off required for changes.
- `ARCHITECTURE`: Modules, services, entities, API contracts, UI designs. Moderate volatility. Impact analysis required for changes.
- `IMPLEMENTATION`: Mechanics, data schemas, build pipelines, developer guides. High volatility. Evolves with code.

The field is named `layer` (not `domain`) to avoid collision with the `DOMAIN` document type.

### The New `priority` Format
- Old: `CORE | SECONDARY | EXPERIMENTAL | FLAVOR`
- New: Integer `1` (low priority) through `5` (highest priority)

### Migration Required
1. **YAML parsers**: Accept `layer` field (required), validate against enum. Accept numeric `priority`.
2. **MCP tools**: `atd_update`, `atd_query`, `atd_lint` must understand the new fields.
3. **Templates**: All templates updated to include `layer`.
4. **Existing atoms**: All 53 atoms in `docs/` need `layer` field added and `priority` converted.
5. **`.atd` config**: May need `layer` to influence bloat factor or audit behavior.

### Where This Pattern Exists Today
- `docs/*.atom.md` — all 53 atoms lack `layer` and use enum `priority`.
- `scripts/cmd/atd/` — parser code expects old format.
- `scripts/pkg/mcp/` — MCP tool schemas.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | Certain — this is a planned breaking change |
| Impact if triggered | High — all tools and atoms affected |
| Detectability | High — parsers will reject unknown fields or wrong formats |
| Current mitigant | Documentation in ATD.md specifies the new format |

---

## Recommended Fix

**Short term:** Update Go parsers to accept both old and new formats during transition (backward-compatible parsing). Add `layer` with sensible defaults to all existing atoms.
**Medium term:** Write a migration script to batch-update all existing atoms: add `layer` based on type heuristics, convert `priority` enum to numeric values (`CORE`→5, `SECONDARY`→3, `EXPERIMENTAL`→2, `FLAVOR`→1).
**Long term:** Remove backward compatibility for old format. `atd lint` (ISS-032) should validate the new fields are present and correct.

---

## References

- [ATD.md §1.2](file:///home/bastien/work/skill/ATD.md) (Template and Frontmatter Fields)
- [ATD.md §1.4](file:///home/bastien/work/skill/ATD.md) (Document Hierarchy & Layers)
- [atd_structure.atom.md](file:///home/bastien/work/skill/docs/atd_structure.atom.md)
