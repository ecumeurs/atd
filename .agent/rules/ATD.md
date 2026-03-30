---
trigger: always_on
---

# IDE Agent Ruleset: Atomic Traceable Documentation (ATD)

**Core Mandate:** You are operating in a codebase governed by Atomic Traceable Documentation (ATD). Documentation and code are not separate entities; they co-evolve as a verifiable graph. You must maintain bidirectional traceability from requirements to code to tests.

### 1. The Atom Blueprint
Every atom (`.atom.md`) is a single-responsibility file with strict YAML frontmatter and four mandatory H2 sections. You must adhere to this exact structure when conceptualizing or updating atoms:

```markdown
---
id: unique_slug
human_name: "Human Readable Name"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: DRAFT
priority: 3
tags: [tag1, tag2]
parents:
  - [[parent_atom_id]]
dependents:
  - [[child_atom_id]]
---

# Human Readable Name

## INTENT
[One sentence: Why does this exist? No "and" or "also".]

## THE RULE / LOGIC
[The core specification. Use pseudo-code, formulas, or strict bullet points.]

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /v1/example` (if applicable)
- **Code Tag:** `@spec-link [[unique_slug]]`
- **Related Issue:** `#123`
- **Test Names:** `TestMyLogic1`, `TestMyLogic2`

## EXPECTATION (For Testing)
[Verifiable acceptance criteria for pass/fail testing. What must be true?]
```

### 2. Document Types & Bloat Factor Quick Reference

Use this table to determine the correct `type`, `layer`, and expected granularity when creating atoms. The **Bloat Factor** column shows the default `bloating_factor` from `.atd` config (1.0 = strictest, 0.1 = most relaxed).

| Type | Family | Typical Layer | Bloat Factor | Granularity |
|---|---|---|---|---|
| `MODULE` | Architectural | ARCHITECTURE | 0.3 | Broad — acts as a parent grouping |
| `SERVICE` | Architectural | ARCH / IMPL | 0.8 | Single responsibility |
| `ENTITY` | Architectural | ARCHITECTURE | 0.8 | Single data model |
| `RULE` | Logic | CUSTOMER / ARCH | 0.8 | One rule per atom. Strict. |
| `MECHANIC` | Logic | IMPLEMENTATION | 0.8 | One algorithm per atom |
| `DOMAIN` | Logic | CUSTOMER | 0.8 | Narrative-driven context |
| `API` | Interface | ARCH / IMPL | 0.1 | One contract, include payloads |
| `UI` | Interface | CUSTOMER / ARCH | 0.8 | One screen or flow |
| `DATA` | Operations | IMPLEMENTATION | 0.8 | Per config domain |
| `USAGE` | Operations | CUSTOMER | 0.1 | Per tutorial/workflow |
| `BUILD` | Operations | IMPLEMENTATION | 0.8 | Per pipeline stage |
| `REQUIREMENT` | Requirements | CUSTOMER | 0.3 | Less granular than RULE |
| `SPECIFICATION` | Requirements | CUSTOMER / ARCH | 0.3 | Can reference multiple atoms |
| `USECASE` | Requirements | CUSTOMER | 0.1 | Multi-step workflow. Relaxed. |
| `USER_STORY` | Requirements | CUSTOMER | 0.1 | Agile story. Relaxed. |

> **Rule:** Before creating an atom, check the bloat factor for its type. High factor (≥0.7) = laser-focused on ONE rule. Low factor (≤0.3) = broader scope is acceptable.

### 3. The "Minimum Atomic Scale" Rule
* Each atom file must describe exactly ONE state-changing rule.
* If an `## INTENT` statement requires the words "and" or "also", you must split the logic into multiple atoms.
* **Always check tolerances:** Use the table above or the `atd_config` tool with `bloating_factor` to retrieve the tolerance for a specific atom type.

### 4. File Modification & Tool Guardrails
* **Never rewrite an entire `.atom.md` file.** Always use the `atd_update` tool to surgically modify specific frontmatter fields or H2 sections.
* **Prioritize deterministic tools:** Use `atd_query`, `atd_trace`, `atd_crawl`, `atd_weave`, and `atd_update` for fast, token-free structural operations.
* **Delegate LLM tasks:** When semantic analysis, complex extraction, or auditing is required, do not do the analysis yourself. Instead, use the MCP's LLM-backed tools (`atd_search`, `atd_audit`, `atd_discover`, `atd_dissect`) to offload the work to ATD's configured models and save your own context window.

### 5. The Day-to-Day Workflow
When asked to build a feature, fix a bug, or update code, you must follow this lifecycle loop:
* **Plan:** Use `atd_query` or `atd_search` to find existing relevant atoms. Create new `DRAFT` atoms using `atd_update` to capture new requirements before writing code.
* **Specify:** Ensure every new atom links upward using the `parents` field in the frontmatter. Run `atd_weave` to establish the downward dependency graph (`dependents`).
* **Implement:** Write the code. You must annotate the source code with `@spec-link [[atom_id]]` to map it to the implementation. Annotate tests with `@test-link [[atom_id]]`.
* **Verify:** Run atd_trace to get a structured Health Snapshot JSON for the specific atom. Ensure the atom's implementation and test coverage metrics meet the required standards. Always ensure that a new atom has a link toward the upper layers ( Customer <- Architecture <- Implementation). If none are present that fits the need, adresse the issue to the user. 
* **Evolve:** Before modifying any `STABLE` atom, you must run `atd_crawl` to assess the blast radius and impact on the rest of the system.

### 6. Surgical Traceability (Tag Placement)
* **No Global Headers:** Do not place `@spec-link` tags at the top of a source file unless the atom literally represents the entire architectural pattern of that file.
* **Target Logic Boundaries:** Place `@spec-link` tags directly above the specific class definition, function, decorator, or logical block that implements the atom.
* **Test Logic Boundaries:** Place `@test-link` tags directly above the specific function, decorator, or logical block that test the atom.
* **Discovery:** If you are unsure where to place tags in undocumented code, use `atd_discover` to get placement recommendations.

### 7. Respect the Documentation Hierarchy
* **CUSTOMER Layer (`REQUIREMENT`, `USECASE`, etc.):** Treat these as low-volatility. Do not alter `STABLE` customer atoms without explicit human permission. **Requirement:** When requesting this permission from the user, you must proactively run `atd_crawl` and present the impact analysis/blast radius to them.
* **ARCHITECTURE Layer (`MODULE`, `API`, etc.):** Treat these as moderate-volatility. Always run an impact analysis (`atd_crawl`) before changing.
* **IMPLEMENTATION Layer (`MECHANIC`, `BUILD`, etc.):** Treat these as high-volatility. Update these freely as you refactor or write new code.

### 8. Pragmatic Traceability & Health (The Trace Rule)
When using `atd_trace`, treat the resulting health metrics as a guide rather than a strict blocker. Apply the following logic:

**Top-Down Design is Expected:** It is perfectly acceptable for CUSTOMER and ARCHITECTURE layer atoms to have a 0% implementation_rate or test_coverage_rate. Missing code/tests at this stage simply mean the feature is "on the to-do list." Do not stubbornly attempt to generate tests or code unless the user explicitly asks you to build the implementation.

**Implementations Require Roots:** The only strict warning you must act upon is missing ancestry. If you are creating or modifying an IMPLEMENTATION atom and atd_trace reports has_customer_origin: false, you must stop and ask the user for clarification. Code should not exist without a reason. Let the user define the missing upstream requirement before you proceed.