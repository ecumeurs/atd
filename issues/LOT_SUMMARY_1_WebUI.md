# Lot Summary: WebUI & Visual Analytics

**Lot Objective:** Transform the ATD WebUI from a static documentation explorer into a powerful, interactive ATD management Integrated Development Environment (IDE).

---

## Included Issues

| Ref | Title | Severity | Status | Component |
|---|---|---|---|---|
| [ISS-069](file:///home/bastien/work/skill/issues/ISS-069_20260407_webui_explorer_global_health.md) | WebUI Explorer Global Health Dashboard | Medium | Open | `webui/static` |
| [ISS-068](file:///home/bastien/work/skill/issues/ISS-068_20260403_webui_atd_health_indicators.md) | Enhance WebUI ATD Health Indicators | Medium | Open | `webui/static` |
| [ISS-067](file:///home/bastien/work/skill/issues/ISS-067_20260403_webui_atd_tree_view.md) | Dedicated ATD Tree/Graph View | High | Open | `webui/static` |
| [ISS-066](file:///home/bastien/work/skill/issues/ISS-066_20260403_webui_atd_reformulate_modal.md) | ATD Content Reformulation Modal | High | Open | `webui/static` |
| [ISS-065](file:///home/bastien/work/skill/issues/ISS-065_20260403_webui_atd_force_weave.md) | Force Weaving from ATD Detail Panel | Medium | Open | `webui/static` |
| [ISS-064](file:///home/bastien/work/skill/issues/ISS-064_20260403_webui_search_regression_and_perf.md) | WebUI Search Regression and Performance | High | Open | `webui/static` |
| [ISS-063](file:///home/bastien/work/skill/issues/ISS-063_20260403_webui_spec_builder_enhancements.md) | WebUI Spec Builder Enhancements | High | Open | `webui/static` |
| [ISS-061](file:///home/bastien/work/skill/issues/ISS-061_20260402_webui_document_generation_bugs) | WebUI Document Generation Failures | High | Open | `webui/` |
| [ISS-055](file:///home/bastien/work/skill/issues/ISS-055_20260330_webui_issue_integration.md) | WebUI Issue Integration | Medium | Open | `webui/static` |
| [ISS-007](file:///home/bastien/work/skill/issues/ISS-007_20260304_webui_binary_link.md) | Link WebUI to Project Binaries | Medium | Open | `webui` |

---

## Thematic Analysis

### Core Themes
1. **Visual Intelligence**: Moving beyond simple lists to global health dashboards, detailed graph/tree views, and granular health indicators (ancestry, implementation status).
2. **Interactive Authoring**: Implementing specialized modals for spec building, content reformulation, and document generation directly within the UI.
3. **Internal Tools Integration**: Surfacing project issues (`ISS-NNN`) and CLI operations (weaving, auditing) directly in the web interface to reduce context switching.

### System Risks
> [!WARNING]
> **Logic Divergence**: Several issues ([ISS-064](file:///home/bastien/work/skill/issues/ISS-064_20260403_webui_search_regression_and_perf.md), [ISS-061](file:///home/bastien/work/skill/issues/ISS-061_20260402_webui_document_generation_bugs)) highlight regressions where the WebUI's internal search or modal logic has diverged from the expected ATD rules, leading to "not found" errors for existing atoms.
>
> **Friction**: The current gap between visualization and action forces developers to keep a terminal open alongside the WebUI for basic maintenance tasks.

---

## Strategic Roadmap

- **Short Term**: Fix search regressions and document generation bugs; unify the "Add Atom" dialogs.
- **Medium Term**: Implement the dedicated Tree/Graph view and the Global Health Dashboard.
- **Long Term**: Integrate the job queue for background CLI tasks and implement bidirectional linking between ATDs and Issues.
