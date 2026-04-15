# Lot Summary: Core Engine & Technical Debt

**Lot Objective:** Optimize the ATD core tooling for high-performance auditing, deterministic exploration, and language-agnostic verification across large-scale codebases.

---

## Included Issues

| Ref | Title | Severity | Status | Component |
|---|---|---|---|---|
| [ISS-062](file:///home/bastien/work/skill/issues/ISS-062_20260402_webui_atd_cli_integration.md) | Integrate WebUI into ATD CLI | Medium | Open | `scripts/cmd/atd` |
| [ISS-060](file:///home/bastien/work/skill/issues/ISS-060_20260402_refactor_atd_exploration_commands.md) | Refactor to Unified Exploration Package | Medium | Open | `scripts/cmd/atd` |
| [ISS-059](file:///home/bastien/work/skill/issues/ISS-059_20260402_refactor_atd_exploration_gitignore.md) | Update Exploration to Use .gitignore | Medium | Open | `scripts/pkg/exploration` |
| [ISS-052](file:///home/bastien/work/skill/issues/ISS-052_20260325_atd_stats_coverage_ancestry.md) | `atd stats` Coverage & Ancestry | Medium | Open | `scripts/cmd/atd/cmd/stats.go` |
| [ISS-049](file:///home/bastien/work/skill/issues/ISS-049_20260325_audit_trace_integration.md) | Audit Should Request Trace for Coverage | Medium | Open | `scripts/cmd/atd/cmd/audit.go` |
| [ISS-048](file:///home/bastien/work/skill/issues/ISS-048_20260325_atd_graph_visualization_improvements.md) | ATD Graph Viz Improvements | Medium | Open | `extension/extension.js` |
| [ISS-044](file:///home/bastien/work/skill/issues/ISS-044_20260324_verify_go_dependency.md) | ATD Verify hardcoded to Go Testing | High | Open | `scripts/cmd/atd/cmd/verify.go` |
| [ISS-040](file:///home/bastien/work/skill/issues/ISS-040_20260324_crawl_summary_mermaid.md) | Crawl Summary & Mermaid Export | Medium | Open | `scripts/cmd/atd/cmd/crawl.go` |
| [ISS-030](file:///home/bastien/work/skill/issues/ISS-030_20260323_crawl_map_impact_submode.md) | `map-impact` sub-mode for `atd crawl` | Medium | Open | `scripts/cmd/atd/cmd/crawl.go` |
| [ISS-027](file:///home/bastien/work/skill/issues/ISS-027_20260314_atd_dissect_quality.md) | Insufficient Dissection Granularity | Medium | Open | `scripts/cmd/atd/dissect` |
| [ISS-026](file:///home/bastien/work/skill/issues/ISS-026_20260314_atd_low_verbosity.md) | Low ATD Content Verbosity | Low | Open | `scripts/cmd/atd/update` |
| [ISS-020](file:///home/bastien/work/skill/issues/ISS-020_20260306_api_logic_tracking_research.md) | Research API Logic Tracking | Medium | Open | `atd_management_skill` |
| [ISS-019](file:///home/bastien/work/skill/issues/ISS-019_20260306_api_atd_payload_capture_shortcoming.md) | API Atoms payload capture shortcoming | Medium | Open | `atd_management_skill` |
| [ISS-018](file:///home/bastien/work/skill/issues/ISS-018_20260305_exclude_usage_atoms_from_bloat.md) | Exclude User Stories from Bloat Checks | Medium | Open | `scripts/atd-audit` |
| [ISS-017](file:///home/bastien/work/skill/issues/ISS-017_20260304_cold_start_audit_replacement.md) | Replace mass gen with audit loop | Medium | Open | `scripts/atd-cold-start.sh` |
| [ISS-002](file:///home/bastien/work/skill/issues/ISS-002_20260304_atd_granularity.md) | ATD Dissection Granularity Enforcement | Medium | Open | `scripts/atd-ollama-generate` |
| [ISS-001](file:///home/bastien/work/skill/issues/ISS-001_20260304_audit_performance.md) | Audit Performance Optimization | Medium | Open | `scripts/atd-audit` |

---

## Thematic Analysis (Updated 2026-04-14)

### Theme 1: Performance & Scalability
> [!NOTE]
> **Status: COMPLETED**
> - **Parallelism**: `atd index` now implements a worker pool (10 concurrent threads) for embedding generation.
> - **Smart Caching**: `atd audit` uses a SQLite-based fallback (`.atd_audit.db`) that skips unmodified atoms based on `mtime`.

### Theme 2: Language Agnostic Tooling
> [!IMPORTANT]
> **Status: EVOLVED (Pending Verification Fix)**
> - **Universal Discovery**: The Go engine now uses `git ls-files` and project-level `.gitignore` detection, making file discovery language-agnostic.
> - **Blocker**: `atd verify` remains hardcoded to `go test`. We need to implement a configurable `test_command` in `.atd` config to support non-Go projects.

### Theme 3: Auditing, Verify & WebUI Integration
> [!WARNING]
> **Status: PENDING**
> - **UI Visibility**: Audit reports (bloat/collisions) and `verify` results are currently CLI-only.
> - **CI/CD Integration**: Need to finalize machine-readable outputs for `atd verify` to allow blocking merges on spec violations.
> - **Stats Gap**: WebUI `/api/stats` currently lacks implementation for Test Coverage and Orphan detection.

### Theme 4: Ollama JSON Schema Migration
> [!TIP]
> **Status: IN PROGRESS**
> - **Strict JSON**: The core Ollama client now supports the `format: "json"` property.
> - **Required Fields**: Most tools (dissect, recon, generate) have been migrated to strict JSON schemas.
> - **To Do**: Migrate `audit_bloat` and `intent_extract` (used in `atd discover`) to use structured JSON schemas instead of raw string parsing.

---

## Strategic Roadmap

- **Short Term**: Refactor `audit_bloat` to use JSON schemas; implement configurable `verify_command` in `config.go`.
- **Medium Term**: Implement the `TestLinks` lookup in the Go `Explorer` to surface real metrics in the WebUI.
- **Long Term**: Deliver the `map-impact` (blast radius) analysis and full Mermaid graph exports.
