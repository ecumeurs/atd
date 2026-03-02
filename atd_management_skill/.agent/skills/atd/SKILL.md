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
type: [MECHANIC | API | UI | DATA | DOMAIN | RULE | USAGE | BUILD | SERVICE | ENTITY | MODULE | REQUIREMENT | SPECIFICATION]
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
| **Requirements** | `REQUIREMENT` | Requirements for the system. Either hard, external constraint or a soft imposition.| 
| | `SPECIFICATION` | Specifications for the system. |


## Toolset Ingestion List
When utilizing this skill, the Agent has access to the following operational tools:

### Read/Crawl Tools
1. **`atd-query(search_term)`**: Deterministic search through `@id` and `@links` headers. No LLM used.
2. **`atd-crawl(docs_path, src_path)`**: Crawls the repo for `@spec-link [[atom_id]]` in code, and parent/dependent tags in other atoms. Generates the Dependency Graph JSON.
3. **`atd-report-gaps(graph_json)`**: Scans the dependency graph to find `STABLE` Atoms containing zero source code implementations.
4. **`atd-congruence(docs_path)`**: **(Meta-Auditor / Architect)** Cross-validates ATDs against their related dependencies. Identifies logically adjacent rules and uses the LLM to verify there are no inherent system contradictions *before* any code is written. Outputs an analysis report.
5. **`atd-verify-diff(docs_path)`**: **(Git Integration / CI)** Runs `git diff`, extracts impacted `@spec-link` tags, automatically executes native tests (e.g., `go test`), and builds the ultimate Auditor prompt for the IDE Assistant.
6. **`atd-dissect(document_text)`**: Uses structural boundary mapping to propose multiple independent `.atom.md` fragments from legacy documentation.
7. **`atd-link-weaver(docs_dir)`**: Automatically populates the `dependents: []` array in Markdown headers by scanning `parents` references, establishing the bi-directional graph.
8. **`atd-cold-start.sh <project_dir>`**: **The Master Pipeline.** Orchestrates reading domain documentation, generating the structural roadmap, indexing the code to `.atd_index.db`, queuing dense files, dissecting logic, weaving links, and tagging the codebase via the Search-Then-Recon protocol. Used heavily for initiating a new repository into the ATD framework.

### High-Volume / Local Auditing (The Cost Routing Protocol)
To prevent the Primary Agent (IDE) from wasting expensive API tokens on high-volume analysis or brute-force code scanning, use the Local LLM toolchain:
7. **`atd-ollama-indexer(dir_path, db_path)`**: Runs `nomic-embed-text` locally against all `.go` files in a directory to chunk and store their Semantic Vectors into a persistent SQLite DB. **(Developer Note: This index is persistent and checks file modification times. Proactively run this after making significant codebase changes or right before an audit so the LLM's vector view is perfectly synced!)**
8. **`atd-ollama-search(db_path, query)`**: Searches the local `.atd_index.db` using the semantic intent of an Atom and returns the top 3 nearest code blocks. **NEVER use IDE tokens to brute-force read unfamiliar code directories when searching for an Atom implementation. Search the index.**
9. **`atd-ollama-audit(atom_path, code_string)`**: Passes a Rule and a Code Snippet to a local `llama3.2` model to rapidly judge logic congruence. Returns `{"passed": bool, "resolutionMessage": string}`.

## Legacy Code Extraction Workflow (The Cold Start)
When operating on undocumented legacy projects, the Architect should run the following baseline protocol natively:

1. **The Automated Pipeline:** Execute `scripts/atd-cold-start.sh <repo_target>`.
2. **Phase 1-3 (Mapping):** The script automatically vectorizes the code, identifies high-density structs/funcs via the roadmap, and extracts existing human `.md` files.
3. **Phase 4-5 (LLM Action):** The IDE Assistant must then read the `pipeline_output/domain_*.md.txt` to write `DOMAIN` specifications, followed by reading `pipeline_output/dissect_*.json` to write `MECHANIC` specifications in the `docs/` folder.
4. **Phase 6 (Weaving):** The pipeline auto-runs `atd-link-weaver` to connect the parents mathematically.
5. **Phase 7 (Search-Then-Recon):** The pipeline auto-tags the codebase in the background by searching the Nomic index for the Atom's intent, and then passing the matched file to the local LLM.
6. **`atd-reconcile(new_text_block)`**: Match new inbound spec requirements against the populated library.
7. **`atd-audit(atom_logic, code_snippet)`**: **LLM-Powered.** The core "Auditor". Compares documentation logic against hard source code.

### Legacy Bridging Tools
7. **`atd-discover-links(source_file, docs_path)`**: Recommends which ATD tags to apply by cross-referencing file logic against the known ATD registry.
   > **Constraint: Surgical Proximity**
   > * **NO Global Headers:** Do not place `@spec-link` tags in the file header unless the atom represents the entire architectural pattern of the file.
   > * **Logic Boundaries:** Place tags immediately above class definitions, decorators, or major logical blocks (e.g., above a group of related FastAPI endpoints).
   > * **Granularity Match:** If an atom describes "Version Control," the tag must be placed only at the start of the version control section of the code.
8. **`atd-recon(atom_id, src_candidate)`**: Semantic archaeology. Validates if a specific candidate file is an implementation of a target Atom.
9. **`atd-tag-sweep(atom_id, search_folder, keyword)`**: Attempts to locate potential implementations for specific atoms inside an untagged `/src` tree based on keyword matching.
10. **`atd-legacy-wrapper(target_file, atom_id)`**: Automatically injects a `@spec-link` tag directly into a target source code file.

### Generation Tools
11. **`atd-assemble(start_ids, purpose)`**: Combines fragments sequentially into a temporary readable document. Follows dependency links.
12. **`atd-generate-snapshot(theme, text_file)`**: Utilizes the aggregated assembly text alongside an LLM to generate narrative flowing documents, ignoring raw metadata.
