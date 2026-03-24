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

### 2. The "Minimum Atomic Scale" Rule
* Each atom file must describe exactly ONE state-changing rule.
* If an `## INTENT` statement requires the words "and" or "also", you must split the logic into multiple atoms.
* **Always check tolerances:** Before creating a new atom, use the `atd_config` tool to retrieve the `bloating_factor` for that specific atom type. Keep `RULE` and `MECHANIC` types laser-focused (strict), while allowing slightly broader scopes only for naturally narrative types like `USECASE` or `API`.

### 3. File Modification & Tool Guardrails
* **Never rewrite an entire `.atom.md` file.** Always use the `atd_update` tool to surgically modify specific frontmatter fields or H2 sections.
* **Prioritize deterministic tools:** Use `atd_query`, `atd_crawl`, `atd_weave`, and `atd_update` for fast, token-free structural operations.
* **Delegate LLM tasks:** When semantic analysis, complex extraction, or auditing is required, do not do the analysis yourself. Instead, use the MCP's LLM-backed tools (`atd_search`, `atd_audit`, `atd_discover`, `atd_dissect`) to offload the work to ATD's configured models and save your own context window.

### 4. The Day-to-Day Workflow
When asked to build a feature, fix a bug, or update code, you must follow this lifecycle loop:
* **Plan:** Use `atd_query` or `atd_search` to find existing relevant atoms. Create new `DRAFT` atoms using `atd_update` to capture new requirements before writing code.
* **Specify:** Ensure every new atom links upward using the `parents` field in the frontmatter. Run `atd_weave` to establish the downward dependency graph (`dependents`).
* **Implement:** Write the code. You must annotate the source code with `@spec-link [[atom_id]]` to map it to the implementation. Annotate tests with `@test-link [[atom_id]]`.
* **Verify:** Run `atd_verify` to check if your code changes align with the specification. Use `atd_test_links` to confirm the atom has test coverage.
* **Evolve:** Before modifying any `STABLE` atom, you must run `atd_crawl` to assess the blast radius and impact on the rest of the system.

### 5. Surgical Traceability (Tag Placement)
* **No Global Headers:** Do not place `@spec-link` tags at the top of a source file unless the atom literally represents the entire architectural pattern of that file.
* **Target Logic Boundaries:** Place `@spec-link` tags directly above the specific class definition, function, decorator, or logical block that implements the rule.
* **Discovery:** If you are unsure where to place tags in undocumented code, use `atd_discover` to get placement recommendations.

### 6. Respect the Documentation Hierarchy
* **CUSTOMER Layer (`REQUIREMENT`, `USECASE`, etc.):** Treat these as low-volatility. Do not alter `STABLE` customer atoms without explicit human permission. **Requirement:** When requesting this permission from the user, you must proactively run `atd_crawl` and present the impact analysis/blast radius to them.
* **ARCHITECTURE Layer (`MODULE`, `API`, etc.):** Treat these as moderate-volatility. Always run an impact analysis (`atd_crawl`) before changing.
* **IMPLEMENTATION Layer (`MECHANIC`, `BUILD`, etc.):** Treat these as high-volatility. Update these freely as you refactor or write new code.
