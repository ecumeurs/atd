---
id: api_webui_health_stats
status: REVIEW
human_name: WebUI Health Stats API
layer: ARCHITECTURE
version: 1.0
priority: 4
parents: [[module_webui]]
dependents:
  - [[mechanic_webui_ancestry_validator]]
  - [[mechanic_webui_coverage_mapper]]
  - [[mechanic_webui_health_stats]]
  - [[mechanic_webui_summary_aggregation]]
type: API
tags: [webui, api, stats]
---

# WebUI Health Stats API

## INTENT
Expose quantitative documentation health metrics to the frontend.

## THE RULE / LOGIC
Endpoint `/api/stats` returning global health metrics:
- **Total Atoms**: Count of all atoms in the project.
- **Coverage**: % of atoms with at least one `@spec-link` in code.
- **Testing**: % of atoms with at least one `@test-link` in tests.
- **Orphans**: Count of atoms with no ancestry to `CUSTOMER` or `DOMAIN`.

## TECHNICAL INTERFACE (The Bridge)
- **Endpoint**: `GET /api/stats`
- **Response Schema**: `HealthStats { Total, SpecCoverage, TestCoverage, Orphans }`
- **Code Tag**: `@spec-link [[api_webui_health_stats]]`

## EXPECTATION (For Testing)
`/api/stats` returns valid JSON with all required health metrics.
