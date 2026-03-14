# ATD Test Suite — Task List

## Phase 0 — Setup

| # | File | Description | Deps | Token Cost |
|---|---|---|---|---|
| 01 | [01_setup_config.md](01_setup_config.md) | `atd init` in upsilonbattle, smoke test binary | None | None |

## Phase 1 — upsilonbattle (CLI / Local Model)

| # | File | Description | Deps | Token Cost |
|---|---|---|---|---|
| 02 | [02_dissect_upsilonbattle.md](02_dissect_upsilonbattle.md) | Dissect `ruler/README.md` and `entity/README.md` | 01 | Low |
| 03 | [03_atd_creation.md](03_atd_creation.md) | Create atoms from dissect output | 02 | Medium |
| 04 | [04_update_and_tagging.md](04_update_and_tagging.md) | Update atom fields, discover + apply @spec-link | 03 | Low |
| 05 | [05_weave_and_index.md](05_weave_and_index.md) | Weave links, index source code + docs | 04 | Low (Ollama embed) |
| 06 | [06_search_doc_and_code.md](06_search_doc_and_code.md) | Query atoms + semantic search | 05 | Low |

## Phase 2 — upsilon (Coldstart Docs)

| # | File | Description | Deps | Token Cost |
|---|---|---|---|---|
| 07 | [07_upsilon_coldstart.md](07_upsilon_coldstart.md) | `atd init` + dissect all coldstart docs | None | Low |
| 08 | [08_upsilon_atd_creation.md](08_upsilon_atd_creation.md) | Create atoms from dissect output | 07 | Medium |

## Phase 3 — MCP Round Tests

> **Prerequisite**: User configures `.mcp.json` and enables MCP in IDE Agent before task 09.

| # | File | Description | Deps | Token Cost |
|---|---|---|---|---|
| 09 | [09_mcp_setup.md](09_mcp_setup.md) | Handshake, tools/list verification | 01 | None |
| 10 | [10_mcp_dissect_query.md](10_mcp_dissect_query.md) | `atd_dissect`, `atd_query` via MCP | 09 | Low |
| 11 | [11_mcp_update_weave.md](11_mcp_update_weave.md) | `atd_update`, `atd_weave` via MCP | 09 | None |
| 12 | [12_mcp_index_search.md](12_mcp_index_search.md) | `atd_index`, `atd_search` via MCP | 09 | Low (Ollama embed) |
| 13 | [13_mcp_feature_coverage_check.md](13_mcp_feature_coverage_check.md) | Full 14-tool coverage audit, document gaps | 09–12 | None |

## Phase 4 — Cold Start Pipeline (Token-Heavy)

| # | File | Description | Deps | Token Cost |
|---|---|---|---|---|
| 14 | [14_cold_start_upsilonbattle.md](14_cold_start_upsilonbattle.md) | Full pipeline: roadmap → index → dissect → weave → recon | 01 | High |

## Phase 5 — Audit Test (Token-Heavy)

| # | File | Description | Deps | Token Cost |
|---|---|---|---|---|
| 15 | [15_audit_test.md](15_audit_test.md) | `atd audit` + `atd congruence` after ATDs exist | 03–06 | High |
