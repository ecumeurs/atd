# SKILL.md: Atomic Traceable Documentation (ATD)

---
name: atd_management
description: Manages the lifecycle, traceability, and verification of atomic documentation fragments and their relationship to source code.
---

## Skill Intent
This skill turns standard Markdown documentation into a **Development Governance System**. It allows an AI assistant to enforce logical constraints natively, preventing code implementation that contradicts the agreed-upon system blueprint.

## The Atom Structure
Every "Atom" is a Markdown file. To minimize LLM search costs, we use a **Strict Header** that allows for deterministic (regex/string) searching.

### `template.atom.md`

```markdown
---
id: [UNIQUE_SLUG]
human_name: [Human Readable Name]
type: [MECHANIC | API | UI | DATA | DOMAIN | RULE | USAGE | BUILD | SERVICE | ENTITY | MODULE | REQUIREMENT | SPECIFICATION | USECASE | USER_STORY]
version: [1.0]
status: [DRAFT | REVIEW | STABLE]
priority: [CORE | SECONDARY | EXPERIMENTAL | FLAVOR]
tags: [tag1, tag2]
parents: 
  - [[parent_atom_id]]
dependents:
  - [[dependency_atom_id]]
---

# NAME OF THE ATOM

## INTENT
[One sentence: Why does this exist?]

## THE RULE / LOGIC
[The "Meat". Use pseudo-code, formulas, or strict bullet points.]
- Formula: Value = X * Y
- Condition: If A then B

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /v1/example`
- **Code Tag:** `@spec-link [[UNIQUE_SLUG]]`
- **Related Issue:** `#123`
- **Test Names:** `TestMyLogic1`, `TestMyLogic2`

## EXPECTATION (For Testing)
[What must be true for this to be "Passed"?]
- Input 10 -> Output 20.
```

| Category | Type | Purpose |
| :--- | :--- | :--- |
| **Architectural** | `MODULE` | High-level grouping (e.g., "The Auth System"). |
| | `SERVICE` | Logical orchestrator/Manager (e.g., DocumentManager). |
| | `ENTITY` | Data structures and State (e.g., SpecFile). |
| **Logic/Rules** | `RULE` | Strict constraints (e.g., "Max file size is 5MB"). |
| | `MECHANIC` | Functional procedures (e.g., "How the diff is calculated"). |
| | `DOMAIN` | Business context, intent, and "The Why." |
| **Interface** | `API` | External/Internal technical contracts (FastAPI routes). |
| | `UI` | Visual requirements and user interaction flows. |
| **Operations** | `DATA` | Static data, configuration, or database schemas. |
| | `USAGE` | Examples, tutorials, and "How-to-use" snippets. |
| | `BUILD` | CI/CD, environment setup, and deployment logic. |
| **Requirements** | `REQUIREMENT` | Requirements for the system. Either hard, external constraint or a soft imposition. |
| | `SPECIFICATION` | Specifications for the system. |
| | `USECASE` | End-to-end workflow narrative (multi-step). Links to child `MECHANIC`/`RULE` atoms via `dependents`. Bloat-check auto-passed. |
| | `USER_STORY` | Agile story: *"As a [role], I want [X] so that [Y]"*. Links to parent `USECASE` and tests via `TECHNICAL INTERFACE`. Bloat-check auto-passed. |


## Toolset Ingestion List
All `atd` subcommands automatically discover the project context by looking for a `.atd` configuration file in the current directory or its parents. This file defines the `docs_path`, LLM provider chain, and model routing.

When utilizing this skill, the Agent has access to the following operational tools:

### Read/Crawl Tools
0. **`atd init([--dir], [--docs], [--model], [--force])`**: Bootstrap a `.atd` configuration file in a target directory. **Must be run once per project.** Creates the docs folder, writes full default config with provider chain and model-to-task routing. Deterministic, no LLM.
1. **`atd query(--search <term>, [--field <id|type|tags|...>)`**: Deterministic search through `@id` and metadata headers. No LLM used.
2. **`atd crawl(--src <path>, [--gaps], [--docs <path>])`**: Crawls the repo for `@spec-link [[atom_id]]` in code and parent/dependent tags in atoms. Generates a Dependency Graph JSON. If `--gaps` is set, identifies STABLE atoms with zero implementations.
3. **`atd congruence([--docs <path>])`**: **(Meta-Auditor / Architect)** Cross-validates ATDs against their related dependencies. Identifies logically adjacent rules and uses the LLM to verify there are no inherent system contradictions *before* any code is written.
4. **`atd verify`**: **(Git Integration / CI)** Runs `git diff`, extracts impacted `@spec-link` tags, and builds the Auditor prompt for the IDE Assistant.
5. **`atd dissect(--file <path>, [--llm])`**: Uses structural boundary mapping to propose multiple independent `.atom.md` fragments from legacy documentation. Returns a prompt to stdout (passthrough) or JSON boundaries (--llm).
6. **`atd generate(--dissect <path>)`**: Generates atom boundaries by sending a `dissect` prompt file to the LLM for structured JSON extraction.
7. **`atd weave`**: Automatically populates the `dependents: []` array in Markdown headers by scanning `parents` references, establishing the bi-directional graph.
8. **`atd roadmap(--dir <path>, [--out <path>])`**: Generates a structural roadmap of the codebase, identifying high-density files for prioritized documentation.
9. **`atd-cold-start.sh <project_dir>`**: **The Master Pipeline.** Bash script that orchestrates `roadmap` -> `index` -> `dissect` -> `weave` -> `discover` -> `recon` to initiate a new repository.

### Granularity Control ("Minimum Atomic Scale")
When generating or deconstructing Atoms, you must adhere to the "Minimum Atomic Scale" to prevent overly broad definitions:
* **The "One Rule" Rule:** If a section of text contains more than one "State-Changing Rule" (e.g., a tax calculation AND a cooldown timer), it **must** be split into two atoms. Refer to `bloating_factor` in `.atd` configuration file (ratio between 1 and 0, 1 meaning exactly one rule per atom, with 0.3 meaning a dozen rules per atom, 0 meaning no limit)
* **Intent Clarity:** If an intent statement requires the word "and" or "also," the granularity is likely too low. Split until the intent is a single, focused objective.

### High-Volume / Local Auditing (The Cost Routing Protocol)
To prevent the Primary Agent (IDE) from wasting expensive API tokens, use the local/remote Ollama backend via the unified CLI:
10. **`atd index(--dir <path>, [--mode code|docs|all], [--db <path>])`**: Runs `nomic-embed-text` locally against the codebase or docs to chunk and store Semantic Vectors in a persistent SQLite DB.
11. **`atd search(--query <term>, [--scope code|docs|all], [--limit <int>], [--grep <term>])`**: Performs semantic search using the local index OR a keyword grep search. Top matches are returned with similarity scores.
12. **`atd audit([--threshold <float>], [--code <path> --atom <path>])`**: **LLM-Powered.** The structural "Auditor". Analyzes ATDs for documentation bloat or missing abstractions. In compliance mode (`--code`), validates snippet against a specific atom.

## Legacy Code Extraction Workflow (The Cold Start)
When operating on undocumented legacy projects, the Architect should run the following baseline protocol:

1. **The Automated Pipeline:** Execute `scripts/atd-cold-start.sh <repo_target>`.
2. **Phase 1-3 (Mapping):** The script runs `atd roadmap` and `atd index` to vectorize the code and identify documentation gaps.
3. **Phase 4-5 (LLM Action):** The IDE Assistant uses `atd dissect` on target files to propose boundaries, and then creates the `.atom.md` files.
4. **Phase 6 (Weaving):** Run `atd weave` to connect the parents mathematically.
5. **Phase 7 (Search-Then-Recon):** The pipeline uses `atd discover` and `atd recon` to auto-tag the codebase with `@spec-link`.
6. **`atd reconcile(--file <path>, --intent <text>)`**: Match new inbound spec requirements against the populated library.

### Legacy Bridging Tools
13. **`atd discover(--file <path>, [--docs <dir>])`**: Recommends which ATD tags to apply by cross-referencing file logic against the known ATD registry.
    > **Constraint: Surgical Proximity**
    > * **NO Global Headers:** Do not place `@spec-link` tags in the file header unless the atom represents the entire architectural pattern of the file.
    > * **Logic Boundaries:** Place tags immediately above class definitions, decorators, or logical blocks.
14. **`atd recon(--atom <path>, --candidate <path>)`**: Semantic archaeology. Validates if a specific candidate file is an implementation of a target Atom. Confidence score returned.
15. **`atd test-links(--src <path>, [--atom <id>])`**: Audits `@test-link [[ATOM_ID]]` tags in source code to map atoms to their verification tests.

### Generation Tools
16. **`atd assemble(--starts <ids>, [--purpose <text>], [--snapshot], [--theme <text>])`**: Stitches atoms together into a cohesive document. If `--snapshot` is set, uses the LLM to generate a narrative executive summary.

### Write / Edit Tools

> [!IMPORTANT]
> **MANDATORY USAGE:** When modifying any field of an existing `.atom.md` file, you MUST use `atd update` instead of rewriting the file via LLM.

17. **`atd update --file <path> [--set key=value ...] [--intent <text>] [--logic <text>] [--spec-link <id> --spec-link-file <path>]`**: Surgically modifies an ATD file or injects `@spec-link` into source code.
    - **Frontmatter edits**: `--set status=STABLE`, `--set parents=[[parent_id]]`.
    - **Body section edits**: `--intent "..."`, `--logic "..."`.
    - **ID / type rename**: When `--set id=new_id` is passed, the file is automatically renamed and references are updated.
    - **Tag Injection**: Passes `--spec-link` and `--spec-link-file` to insert tags into source code without manual editing.

## MCP Server Mode

`atd serve` starts a JSON-RPC 2.0 / MCP 2025-11-25 server exposing all ATD tools to IDE agents (VS Code, Claude Desktop) and any MCP host.

### Transport
- **stdio** (default, recommended): `atd serve` — host launches as subprocess, communicates via stdin/stdout
- **HTTP** (optional): `atd serve --http --port 7474` — single `/mcp` endpoint

### Registered Tools (14)
| Name | LLM | Description |
|---|---|---|
| `atd_query` | No | Atom frontmatter search |
| `atd_crawl` | No | Dependency graph + gap report |
| `atd_weave` | No | Bi-directional link weaving |
| `atd_update` | No | Surgical field edits |
| `atd_roadmap` | No | Source complexity map |
| `atd_verify` | No | Git-diff audit prompt |
| `atd_assemble` | No/Yes | Stitch atoms into document |
| `atd_test_links` | No | @test-link audit |
| `atd_dissect` | Yes | Dissect file into atom boundaries |
| `atd_index` | Yes (embed) | Build semantic index |
| `atd_search` | Yes (embed) | Semantic + grep search |
| `atd_audit` | Yes | Bloat + collision audit |
| `atd_recon` | Yes | Validate code implements atom |
| `atd_discover` | Yes | Recommend @spec-link tags |

> **Note:** `atd init` is **not** a MCP tool — it is a one-time filesystem bootstrap.

### VS Code Configuration (`.mcp.json`)
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
