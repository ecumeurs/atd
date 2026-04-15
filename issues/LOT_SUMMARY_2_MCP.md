# Lot Summary: MCP & Automation Workflows

**Lot Objective:** Enhance the AI Agent's effectiveness by exposing high-level ATD pipelines as structured MCP tools and automating complex local deconstruction/generation tasks.

---

## Included Issues

| Ref | Title | Severity | Status | Component |
|---|---|---|---|---|
| [ISS-051](file:///home/bastien/work/skill/issues/ISS-051_20260325_atd_verify_mcp_git_diff_failure.md) | atd_verify MCP Git Diff Failure | High | Open | `scripts/cmd/atd/cmd/verify.go` |
| [ISS-050](file:///home/bastien/work/skill/issues/ISS-050_20260325_mcp_protocol_unification.md) | Unified Cold-Start & Audit as MCP Tools | High | Open | `scripts/cmd/atd/cmd/serve.go` |
| [ISS-045](file:///home/bastien/work/skill/issues/ISS-045_20260324_mcp_tools_refactor_and_cleanup.md) | MCP Tools Review & Parameter Cleanup | Medium | Open | `scripts/cmd/atd/cmd/mcp_tools.go` |
| [ISS-039](file:///home/bastien/work/skill/issues/ISS-039_20260324_reconcile_mcp.md) | Reconcile Tool via MCP | Medium | Open | `scripts/cmd/atd/cmd/serve.go` |
| [ISS-031](file:///home/bastien/work/skill/issues/ISS-031_20260324_cold_start_mcp.md) | Cold-Start & Full Audit via MCP | High | Open | `scripts/cmd/atd/cmd/serve.go` |
| [ISS-011](file:///home/bastien/work/skill/issues/ISS-011_20260304_atd_generation_orchestration.md) | ATD Gen Orchestration / Local Dissection | Medium | Open | `atd_management_skill` |

---

## Thematic Analysis

### Core Themes
1. **Tool Primitive Upgrades**: Moving from fine-grained manual orchestration to coarse-grained "composite" tools (e.g., `atd_cold_start` instead of 7 sequential calls).
2. **Protocol Encapsulation**: Removing internal configuration details (file paths, DB locations) from MCP schemas to provide a cleaner "Agent-centric" interface.
3. **Local Logic Processing**: Leveraging local LLMs (Ollama) more effectively for document dissection and atom generation to reduce cloud token consumption and context window pressure.

### System Risks
> [!IMPORTANT]
> **Fragility**: High-gravity tasks like cold-starting are currently locked in shell scripts. Forcing an AI agent to replicate this logic step-by-step via multiple MCP calls increases the risk of inconsistent results and state corruption.
>
> **Token Cost**: Over-reliance on cloud models for repetitive tasks (like deconstructing source code into boundaries) creates unnecessary overhead and limits the system's viability for large repositories.

---

## Strategic Roadmap

- **Short Term**: Refactor `mcp_tools.go` to remove redundant path parameters; fix the `atd_verify` git diff bug.
- **Medium Term**: Implement `atd_cold_start` and `atd_full_audit` as composite MCP tools in the Go backend.
- **Long Term**: Transition the entire Dissection -> Generation pipeline to local Ollama models with a unified `atd-gen` wrapper.
