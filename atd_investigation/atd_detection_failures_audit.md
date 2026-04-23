# ATD Detection Failures: False Orphans Audit
**Date**: 2026-04-19
**Scope**: 245 atoms on disk

## Summary
During the documentation hardening audit, I identified several **"False Orphans"**: atoms that are correctly implementation-tagged in the source code using the `@spec-link [[atom_id]]` format, but are reported as orphans by the ATD `stats` and `crawl` tools.

---

## Case 1: `mech_action_economy_timeout_penalty_rules`

**Status**: False Orphan
**Atom Type**: MECHANIC
**Filesystem Path**: `docs/mech_action_economy_timeout_penalty_rules.atom.md`

### Observations
The indexer fails to detect the tags in `upsilonbattle/battlearena/ruler/ruler.go`, despite them following the exact standard pattern.

### Implementation Evidence
```go
// upsilonbattle/battlearena/ruler/ruler.go

// @spec-link [[mech_action_economy_timeout_penalty_rules]]
func (r *Ruler) startShotClock() {
    // ...
}

// @spec-link [[mech_action_economy_timeout_penalty_rules]]
func (r *Ruler) timeout(ctx actor.NotificationContext) {
    // ...
}
```

---

## Case 2: `mech_game_state_versioning`

**Status**: False Orphan
**Atom Type**: MECHANIC
**Filesystem Path**: `docs/mech_game_state_versioning.atom.md`

### Observations
This atom is correctly tagged in multiple locations across the `ruler` logic but is still reported as having 0% implementation coverage.

### Implementation Evidence
```go
// upsilonbattle/battlearena/ruler/ruler.go
r.GameState.IncTurn() // @spec-link [[mech_game_state_versioning]]

// upsilonbattle/battlearena/ruler/rules/gamestate.go
// @spec-link [[mech_game_state_versioning]]
func (g *GameState) IncTurn() {
    // ...
}
```

---

## Technical Analysis of Failure
1. **Sub-Repository Depth**: Both cases occur within the `upsilonbattle` directory. It is possible the indexer's default `walk` depth or directory inclusion rules are skipping these sub-folders during the `@spec-link` extraction phase.
2. **Comment Parsing**: The tags are placed in trailing comments (e.g., `IncTurn() // @spec-link...`) or block comments. The parser may be prioritizing standalone comment lines.
3. **Index Stale State**: The presence of "ghost atoms" (610 total reported vs 245 real files) suggests that the indexer is failing to purge and refresh its internal graph correctly, leading to persistent stale coverage metrics.

## Recommendation for ATD Team
- **Root Audit**: Validate that the `@spec-link` extraction engine is processing the `upsilonbattle/` and `battleui/` directories with the same priority as the core.
- **Cache Invalidation**: Force a clean wipe of the `.atd_index.db` and `.atd_docs_index.db` and re-run discovery purely from logic boundaries.
