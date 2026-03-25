# ATD System Architect & Tooling

## Intent of the Project
The ATD (Atomic Traceable Document) project provides a framework and an associated AI agent skill to strictly manage software architecture, rules, and requirements. It ensures that code acts as a perfect reflection of atomic specifications. By structuring documentation into highly focused, single-responsibility files (Atoms), it bridges the gap between high-level game design/architecture and low-level code implementation. The system is designed to provide AI agents with a deterministic, searchable source of truth that enforces rules before code is ever written or modified.

## What are ATDs?
ATDs are **Atomic Traceable Documents**. Each ATD is a single Markdown file representing a unique, isolated piece of logic, domain knowledge, or specification. To minimize context windows and allow for automated parsing, ATDs employ a strict YAML header that defines metadata such as ID, type (e.g., `DOMAIN`, `MECHANIC`, `API`, `RULE`), status, and dependency relationships.

Because an ATD is atomic, it describes only one primary rule or concept. This strict granularity prevents ambiguous requirements and ensures that every rule can be individually tested, verified, and linked directly to the application's source code.

## How it Works
1. **The ATD Framework**: Developers and architects write ATDs referencing systems, mechanics, and entities.
2. **The Tools**: A unified `atd` CLI binary (built from `scripts/cmd/atd/`) provides all ATD operations as subcommands. The tools can validate formatting, extract legacy logic into new Atoms, weave dependency graphs, and verify congruence between the documentation and the codebase.
3. **The Link (@spec-link)**: Code objects (functions, classes) are annotated with `@spec-link [[ATOM_ID]]`. The ATD agents and tools trace these links to verify that code implementations align with current architecture definitions. If a developer or an AI agent attempts to violate an ATD rule, the discrepancy is flagged.
4. **Agent Integration**: The AI assistant (having the ATD skill) operates under specific modes (Architect, Developer, Analyst) to either create/manage Atoms, write compliant code, or audit the system respectively.
5. **MCP Server**: `atd serve` exposes all 19 ATD operations as [MCP](https://modelcontextprotocol.io) tools over JSON-RPC 2.0, enabling IDE agents (VS Code, Claude Desktop) to invoke ATD commands directly without shell access.

## Setup & Tooling
The ATD system relies on the `atd` unified binary compiled from the Go source.

### Compiling the Toolchain
Build and install the `atd` binary from the project root:
```bash
cd scripts/cmd/atd && go build -o /usr/local/bin/atd .
```
Or use the compilation script:
```bash
./compile_tools.sh
```

### Initializing a Project
Before using any `atd` subcommand on a new or legacy project, bootstrap the `.atd` config file:
```bash
# In your project root
atd init

# With custom paths or model
atd init --docs specs/ --model deepseek-r1:7b

# Force overwrite an existing config
atd init --force
```
This creates `.atd` (full provider chain + model routing config) and the `docs/` directory.

### Ollama LLM Setup
Many ATD tools (e.g., `atd dissect`, `atd audit`, `atd index`) use a local Ollama instance for LLM processing.

#### 1. Start the Ollama Container
Ensure you have Docker installed and run the following command to start the Ollama service:
```bash
docker run -d -v ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama
```

#### 2. Install Required Models
The system requires specific models for text generation and embeddings. Pull them using these commands:
```bash
docker exec -it ollama ollama pull llama3.2
docker exec -it ollama ollama pull nomic-embed-text
```

### MCP Integration (VS Code / Claude Desktop)
`atd serve` starts a JSON-RPC 2.0 MCP server exposing **19 tools** — all ATD subcommands including LLM-backed operations (e.g., `atd_dissect`, `atd_search`) and deterministic diagnostics (e.g., `atd_stats`, `atd_lint`, `atd_trace`).

**stdio transport (recommended):** Add to `.mcp.json` in your project root:
```json
{
  "servers": {
    "atd": {
      "type": "stdio",
      "command": "/path/to/atd",
      "args": ["serve"]
    }
  }
}
```

**HTTP transport:** `atd serve --http --port 7474` then point your MCP client at `http://localhost:7474/mcp`.

## Reference Project
**`upsilonbattle`** serves as the primary reference project used to test and validate this skill. It demonstrates how ATD mechanics, API routes, and domain elements interact in a real-world scenario, acting as the testbed for the ATD toolchain's extraction, auditing, and generation capabilities.

## Open Issues

| Name | Date | Status | Severity | Oneliner |
|---|---|---|---|---|
| [Proof Test Trace with Complex Graph](issues/ISS-047_20260325_trace_proof_test_case.md) | 2026-03-25 | Open | Medium | Proof Test Trace with complex graph and health violations. |
| [ATD Graph Visualization Improvements](issues/ISS-048_20260325_atd_graph_visualization_improvements.md) | 2026-03-25 | Open | Medium | The current ATD graph visualization in the VS Code extension (`atd.showFullGr... |
| [Weave Regex Corrupts Dependents Field](issues/ISS-046_20260324_weave_dependents_corruption.md) | 2026-03-24 | Open | Critical | The `atd weave` command corrupts the `dependents` YAML field on every run, pr... |
| [MCP Tools Review and Parameter Cleanup](issues/ISS-045_20260324_mcp_tools_refactor_and_cleanup.md) | 2026-03-24 | Open | Medium | The current MCP tools expose internal implementation details (like file paths... |
| [ATD Verify Tool is Hardcoded to Go Testing](issues/ISS-044_20260324_verify_go_dependency.md) | 2026-03-24 | Open | High | The `atd verify` tool currently has a hard dependency on Go, specifically exe... |
| [WebUI Specification](issues/ISS-041_20260324_webui_specification.md) | 2026-03-24 | Open | Medium | The WebUI currently exists only as the kernel of an idea — basic rendering wi... |
| [Crawl Summary and Mermaid Graph Export](issues/ISS-040_20260324_crawl_summary_mermaid.md) | 2026-03-24 | Open | Medium | The dependency graph from `atd crawl` is a raw JSON blob. There is no human-r... |
| [Reconcile Tool via MCP](issues/ISS-039_20260324_reconcile_mcp.md) | 2026-03-24 | Open | Medium | When new external requirements arrive that semantically overlap with existing... |
| [Type-Specific Atom Templates](issues/ISS-038_20260324_type_specific_templates.md) | 2026-03-24 | Open | Medium | All atoms use the same generic template regardless of type. `API` atoms would... |
| [Cross-Project Atom Sharing with Destination Selection](issues/ISS-035_20260324_cross_project_sharing.md) | 2026-03-24 | Open | Low | ATD is currently single-project. There is no mechanism for sharing atoms betw... |
| [Change History Sidecar per Atom](issues/ISS-034_20260324_changelog_sidecar.md) | 2026-03-24 | Open | Medium | Atoms have a `version` field but no change log. When an atom is modified, the... |
| [Atom Deprecation and Archival Statuses](issues/ISS-033_20260324_atom_deprecation_archival.md) | 2026-03-24 | Open | Medium | There is no `DEPRECATED` or `ARCHIVED` status for atoms. When a feature is re... |
| [Cold-Start and Full Audit via MCP](issues/ISS-031_20260324_cold_start_mcp.md) | 2026-03-24 | Open | High | The cold-start pipeline (`roadmap` → `index` → `dissect` → `weave` → `discove... |
| [Implement `map-impact` sub-mode for `atd crawl`](issues/ISS-030_20260323_crawl_map_impact_submode.md) | 2026-03-23 | Open | Medium | The `atd_map_impact` tool is mentioned in the `ATD.md` rules as a "Ripple che... |
| [VSCode Extension for Atomic Link Following](issues/ISS-029_20260319_vscode_atomic_link.md) | 2026-03-19 | Open | Medium | Currently, developers using the ATD system in VS Code cannot easily navigate ... |
| [Insufficient Dissection Granularity for Complex Files](issues/ISS-027_20260314_atd_dissect_quality.md) | 2026-03-14 | Open | Medium | The `atd dissect` tool fails to identify a sufficient number of atomic bounda... |
| [Low ATD Content Verbosity](issues/ISS-026_20260314_atd_low_verbosity.md) | 2026-03-14 | Open | Low | Atoms generated during the initial creation phase (Task 03) are often sparse,... |
| [Research efficient API logic tracking for ATD atoms](issues/ISS-020_20260306_api_logic_tracking_research.md) | 2026-03-06 | Open | Medium | Current API-typed Atoms use free-form text or simplified summaries that often... |
| [API typed atd aren't capturing full payload/contract details](issues/ISS-019_20260306_api_atd_payload_capture_shortcoming.md) | 2026-03-06 | Open | Medium | When working on projects to test ATDs, instructions for API expectations and ... |
| [Exclude User Stories and Use Cases from Bloat Checks](issues/ISS-018_20260305_exclude_usage_atoms_from_bloat.md) | 2026-03-05 | Open | Medium | Atoms that represent User Stories and Use Cases (typically typed as `USAGE` o... |
| [Replace Cold Start Mass Generative Step with Audit Loop](issues/ISS-017_20260304_cold_start_audit_replacement.md) | 2026-03-04 | Open | Medium | The final step of the cold start pipeline (`atd-cold-start.sh`) instructs the... |
| [Integrate Issue Management with ATD Management](issues/ISS-014_20260304_issue_atd_integration.md) | 2026-03-04 | Open | Medium | Integrate the `issue_management` skill as a side-skill for `atd_management`. ... |
| [Lack of Project Documentation and ATD](issues/ISS-013_20260304_lack_of_atd_documentation.md) | 2026-03-04 | Open | High | This project lacks comprehensive documentation, including the Atomic Technica... |
| [ATD Generation Orchestration and Local Dissection](issues/ISS-011_20260304_atd_generation_orchestration.md) | 2026-03-04 | Open | Medium | There is a lack of orchestration between the IDE agent and the local ATD gene... |
| [ATD Status Management and Workflow](issues/ISS-010_20260304_atd_status_management.md) | 2026-03-04 | Open | Medium | The `status` attribute is currently ignored. Implementing status-based logic ... |
| [ATD Version Management Implementation](issues/ISS-009_20260304_atd_version_management.md) | 2026-03-04 | Open | Medium | The `version` attribute in ATD YAML frontmatter is currently ignored. The sys... |
| [WebUI LLM Integration for Content Iteration](issues/ISS-008_20260304_webui_llm_integration.md) | 2026-03-04 | Open | Medium | Link the WebUI to a "true" LLM (via Ollama or an API) to allow for content it... |
| [Link WebUI to Project Binaries](issues/ISS-007_20260304_webui_binary_link.md) | 2026-03-04 | Open | Medium | Integrate the WebUI with the project's heavy-duty binaries and scripts (e.g.,... |
| [WebUI ATD Editing and ID Propagation](issues/ISS-006_20260304_webui_edit_propagation.md) | 2026-03-04 | Open | High | The WebUI needs to allow altering ATDs (all fields). Crucially, changing an A... |
| [WebUI ATD Preview HTML Rendering](issues/ISS-005_20260304_webui_html_rendering.md) | 2026-03-04 | Open | Low | The ATD preview in the WebUI is not rendered as HTML. It shows plain, trimmed... |
| [WebUI Navigation and Exploration Improvements](issues/ISS-004_20260304_webui_navigation.md) | 2026-03-04 | Open | High | The WebUI currently fails to allow full exploration of all ATDs. It only show... |
| [ATD Dissection Granularity Enforcement](issues/ISS-002_20260304_atd_granularity.md) | 2026-03-04 | Open | Medium | Ensure that the dissection of documents and general ATD creation strictly fol... |
| [Audit Performance Optimization](issues/ISS-001_20260304_audit_performance.md) | 2026-03-04 | Open | Medium | The current auditing process is too slow. It requires access to a more perfor... |

