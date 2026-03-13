# ATD System Architect & Tooling

## Intent of the Project
The ATD (Atomic Traceable Document) project provides a framework and an associated AI agent skill to strictly manage software architecture, rules, and requirements. It ensures that code acts as a perfect reflection of atomic specifications. By structuring documentation into highly focused, single-responsibility files (Atoms), it bridges the gap between high-level game design/architecture and low-level code implementation. The system is designed to provide AI agents with a deterministic, searchable source of truth that enforces rules before code is ever written or modified.

## What are ATDs?
ATDs are **Atomic Traceable Documents**. Each ATD is a single Markdown file representing a unique, isolated piece of logic, domain knowledge, or specification. To minimize context windows and allow for automated parsing, ATDs employ a strict YAML header that defines metadata such as ID, type (e.g., `DOMAIN`, `MECHANIC`, `API`, `RULE`), status, and dependency relationships.

Because an ATD is atomic, it describes only one primary rule or concept. This strict granularity prevents ambiguous requirements and ensures that every rule can be individually tested, verified, and linked directly to the application's source code.

## How it Works
1. **The ATD Framework**: Developers and architects write ATDs referencing systems, mechanics, and entities.
2. **The Tools**: A suite of Go-based CLI tools (e.g., `atd-audit`, `atd-crawl`, `atd-dissect`, `atd-link-weaver`) processes these ATDs. The tools can validate formatting, extract legacy logic into new Atoms, weave dependency graphs, and verify congruence between the documentation and the codebase.
3. **The Link (@spec-link)**: Code objects (functions, classes) are annotated with `@spec-link [[ATOM_ID]]`. The ATD agents and tools trace these links to verify that code implementations align with current architecture definitions. If a developer or an AI agent attempts to violate an ATD rule, the discrepancy is flagged.
4. **Agent Integration**: The AI assistant (having the ATD skill) operates under specific modes (Architect, Developer, Analyst) to either create/manage Atoms, write compliant code, or audit the system respectively.

## Setup & Tooling
The ATD system relies on a suite of tools that must be compiled and configured to function correctly.

### Compiling the Toolchain
To build and install all ATD CLI tools and scripts into the skill's utility directory, run the compilation script from the root of the project:
```bash
./compile_tools.sh
```
This script will:
1. Compile all Go tools located in the `scripts/` directory.
2. Copy necessary shell scripts to the toolchain destination.
3. Set the required execution permissions.

### Ollama LLM Setup
Many ATD tools (e.g., `atd-ollama-audit`, `atd-dissect`) use a local Ollama instance for LLM processing.

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

## Reference Project
**`upsilonbattle`** serves as the primary reference project used to test and validate this skill. It demonstrates how ATD mechanics, API routes, and domain elements interact in a real-world scenario, acting as the testbed for the ATD toolchain's extraction, auditing, and generation capabilities.

## Open Issues

| Name | Date | Status | Severity | Oneliner |
|---|---|---|---|---|
| [Add a tool to handle atds indexing and research through atds (nomic)](issues/ISS-024_20260313_atd_indexing_nomic.md) | 2026-03-13 | Open | Medium | While the project has `atd-ollama-indexer` and `atd-ollama-search` using Nomi... |
| [Merge all tools into ONE atd go binary](issues/ISS-023_20260313_atd_binary_unification.md) | 2026-03-13 | Open | Medium | Currently, the project consists of dozens of small, independent Go binaries l... |
| [Formalize Test Handling and Traceability in ATD System](issues/ISS-022_20260311_test_handling_shortcoming.md) | 2026-03-11 | Open | Medium | Currently, test files and test functions are not handled in a specific way by... |
| [Tiered LLM Access for ATD Operations](issues/ISS-021_20260309_tiered_llm_access.md) | 2026-03-09 | Open | High | Implement a tiered LLM access system to optimize token usage and cost. The sy... |
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

