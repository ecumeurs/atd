# Atomic Traceable Documentation (ATD) — Reference Manual

> **Audience:** IDE agents, developers, and CI tooling.
> **Purpose:** Establish the canonical rules for **when** and **how** to use ATD within a development environment.
> **Primary objective:** Maintain high-level traceability from **Customer requirements** → **Architecture** → **Code** → **Tests** (`@spec-link` + `@test-link`).

---

## Table of Contents

1. [How ATD Works](#1-how-atd-works)
   - [Philosophy](#11-philosophy)
   - [The Atom](#12-the-atom)
   - [Document Types](#13-document-types)
   - [Document Hierarchy & Layers](#14-document-hierarchy--layers)
   - [Relationships & Graph Model](#15-relationships--graph-model)
   - [Lifecycle](#16-lifecycle)
   - [Configuration](#17-configuration)
2. [MCP Toolset Reference](#2-mcp-toolset-reference)
   - [Deterministic Tools](#21-deterministic-tools)
   - [LLM-Backed Tools](#22-llm-backed-tools)
   - [Tool Decision Matrix](#23-tool-decision-matrix)
3. [Gap Analysis & Recommendations](#3-gap-analysis--recommendations)

---

## 1. How ATD Works

### 1.1 Philosophy

ATD turns Markdown documentation into a **Development Governance System**. Five principles govern the framework:

| # | Principle | Rule |
|---|---|---|
| 1 | **Minimum Atomic Scale** | Each atom describes exactly ONE state-changing rule. If an intent needs "and" or "also", split it. |
| 2 | **Bidirectional Traceability** | Every atom links to code via `@spec-link [[atom_id]]` tags and to tests via `@test-link [[atom_id]]` tags; every code module links back to its governing atom. This creates a verifiable chain: Customer requirement → Architecture → Implementation → Test. |
| 3 | **Doc-Code Co-evolution** | During **cold-start** (bootstrapping an undocumented codebase), atoms are extracted FROM existing implementations — the code is the initial source of truth. Once the initial ATD base is established, **documentation and code evolve together**: new features begin as `DRAFT` atoms (requirements, specs, design) before implementation, and implementation feeds back into atom refinement. Neither side is subordinate; they are kept in sync through the verification loop. |
| 4 | **LLM-Assisted, Human-Governed** | Local/remote LLMs handle bulk extraction and auditing. The IDE Agent handles high-intelligence tasks. Humans govern final architecture. |
| 5 | **Token Economy** | Every LLM interaction is metered by task type. Cheap tasks (embedding, pass/fail) run locally. Expensive generation tasks run on capable models. Deterministic tasks (weaving, crawling, updating) never touch an LLM. |

### 1.2 The Atom

An **Atom** is a single-responsibility Markdown file (`.atom.md`) with a strict YAML frontmatter followed by four mandatory sections.

#### Template

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
[One sentence: Why does this exist?]

## THE RULE / LOGIC
[The core specification. Use pseudo-code, formulas, or strict bullet points.]

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /v1/example`
- **Code Tag:** `@spec-link [[unique_slug]]`
- **Related Issue:** `#123`
- **Test Names:** `TestMyLogic1`, `TestMyLogic2`

## EXPECTATION (For Testing)
[What must be true for this to be "Passed"?]
- Input 10 -> Output 20.
```

#### Frontmatter Fields

| Field | Required | Format | Description |
|---|---|---|---|
| `id` | ✅ | `snake_case` slug | Unique identifier. Convention: `<type_lowercase>_<descriptive_slug>`. |
| `human_name` | ✅ | Quoted string | Human-readable title. |
| `type` | ✅ | Enum (see §1.3) | Categorization of the atom's role. |
| `layer` | ✅ | `BUSINESS` / `ARCHITECTURE` / `IMPLEMENTATION` | Which layer of the documentation hierarchy this atom belongs to (see §1.4). |
| `version` | ✅ | Semver (`1.0`) | Document version (currently advisory, see ISS-009). |
| `status` | ✅ | `DRAFT` / `REVIEW` / `STABLE` | Maturity level. |
| `priority` | ✅ | Integer `1`–`5` | Importance level. `1` = low priority, `5` = highest priority. |
| `tags` | ✅ | YAML list | Free-form keywords for search. |
| `parents` | ✅ | `[[atom_id]]` list | Upward dependency links. |
| `dependents` | ✅ | `[[atom_id]]` list | Downward dependency links (auto-populated by `atd weave`). |

#### Mandatory Sections (H2)

| Section | Purpose |
|---|---|
| `## INTENT` | Single-sentence "Why?" — must not contain "and" or "also". |
| `## THE RULE / LOGIC` | The core technical or functional specification. |
| `## TECHNICAL INTERFACE` | Linking information: API endpoints, `@spec-link` tags, test names, issue refs. |
| `## EXPECTATION` | Verifiable acceptance criteria for pass/fail testing. |

#### Granularity & Bloat Control

ATD enforces the **"Minimum Atomic Scale"** to prevent overly broad atoms. This is governed by two mechanisms:

**1. The "One Rule" Rule (Structural)**
- If a section of text contains more than one "state-changing rule" (e.g., a tax calculation AND a cooldown timer), it **must** be split into two atoms.
- If an `## INTENT` statement requires the word "and" or "also", the atom is too broad — split until the intent is a single, focused objective.

**2. Bloat Factor (Configurable)**

The `bloating_factor` in `.atd` config controls how strictly the automated auditor (`atd audit`) enforces granularity per type. It is a ratio between `0.0` and `1.0`:

| Factor | Meaning | Use When |
|---|---|---|
| `1.0` | **Strictest.** The atom must describe exactly one rule, one procedure, or one constraint. Any compound logic is flagged as bloat. | `RULE`, `MECHANIC` — pure logic atoms that must be surgical. |
| `0.7–0.8` | **Default.** Allows minor supporting context around the core rule. A small preamble or two closely related sub-points are tolerated. | Most types, general-purpose default. |
| `0.3–0.5` | **Relaxed.** The atom can contain several related rules or a broader narrative. Suitable for grouping types. | `MODULE`, `REQUIREMENT`, `SPECIFICATION` — naturally broader. |
| `0.1` | **Very relaxed.** Almost no bloat enforcement. The atom is expected to be long and narrative. | `USECASE`, `USER_STORY`, `API` — multi-step workflows or detailed contracts. |
| `0.0` | **No enforcement.** Bloat checking is effectively disabled. | Not recommended for production use. |

> **Rule of thumb for agents:** When creating an atom, check the `bloating_factor` for its type in `.atd` config. If the factor is high (≥0.7), keep the atom laser-focused on a single rule. If the factor is low (≤0.3), you have room for a richer, multi-point specification — but still aim for coherence around a single *topic*.

The per-type overrides are set in `.atd`:
```json
"bloating_factor": {
  "default": 0.8,
  "type_overrides": {
    "MODULE": 0.3,
    "REQUIREMENT": 0.3,
    "SPECIFICATION": 0.3,
    "USECASE": 0.1,
    "USER_STORY": 0.1,
    "API": 0.1
  }
}
```

### 1.3 Document Types

Atoms are grouped into **13 consolidated types** across three functional families. The **Bloat Factor** column maps to the default `bloating_factor` per type in `.atd` config (1.0 = strictest, 0.1 = most relaxed).

| Type | Family | Typical Layer | Bloat Factor | Granularity | `@spec-link` Placement |
|---|---|---|---|---|---|
| `CONTRACT` | Governance | BUSINESS | 0.1 | **Unique**; project-wide mandatory rules | Governance only; no graph links |
| `VISION` | Governance | BUSINESS | 0.1 | **Unique**; project-wide scope/philosophy | Governance only; no graph links |
| `REQUIREMENT` | Requirements | BUSINESS | 0.3 | High-level external contract or constraint | Integration test suites |
| `USER_STORY` | Requirements | BUSINESS | 0.1 | User-facing workflow (synonym: `USECASE`, `WORKFLOW`) | Bloat-check relaxed |
| `RULE` | Logic | BUSINESS / ARCHITECTURE | 0.8 | Single business constraint or boolean check | At the validation point |
| `DOMAIN` | Logic | BUSINESS | 0.8 | Narrative-driven context: "The Why" | Documentation or orchestration files |
| `MECHANIC` | Logic | IMPLEMENTATION | 0.8 | One algorithm or procedural step | Above the implementation function |
| `MODULE` | Architectural | ARCHITECTURE | 0.3 | High-level grouping; broad scope is acceptable | Package/directory level or main entry point |
| `SERVICE` | Architectural | ARCHITECTURE | 0.3 | Command surface / service orchestration exposed by the system (CLI subcommand, MCP tool wrapper, provider) | Above the command handler or service entry point |
| `ENTITY` | Architectural | ARCHITECTURE | 0.8 | Single data structure or state model | Above type/struct definitions |
| `API` | Interface | ARCHITECTURE | 0.1 | One contract per atom; include sample payloads | Above route handler |
| `UI` | Interface | ARCHITECTURE | 0.8 | One screen or interaction flow | Near component definition |
| `SPECIFICATION` | Interface | ARCHITECTURE | 0.3 | Config schema, data schema, or interface specification (e.g. `.atd` config shape, payload schema) | Above the schema/config definition |

### 1.4 Project Governance: CONTRACT & VISION

`CONTRACT` and `VISION` are specialized, unique atoms that govern the evolution of the entire project. Both are **read for governance, never linked as structural ancestry**: no atom may list `contract_atd`/`vision_atd` (or a project's equivalently-named pair) in its own `parents:`. They gate change from the side; they are not trace origins and must remain graph-isolated.

1. **Uniqueness**: There must be exactly ONE `CONTRACT` atom and ONE `VISION` atom per project.
2. **Roles**:
    - **`VISION`**: What the project is *meant to become* — its scope and philosophy, with room for legitimate evolution. It MUST be read whenever a `BUSINESS` layer atom is added or updated. It prevents adding atoms that fall outside the project's intended purview (scope-creep protection).
    - **`CONTRACT`**: What the project *currently guarantees* to anything outside it — a versioned, semver-style surface, analogous to a public API contract. It is not an enumeration of every `BUSINESS` atom; most atoms are internal detail the contract never tracks. It MUST be read whenever a change touches an atom the Contract Surface Grid below classifies as surface.
3. **Contract Surface Grid** — whether a change to an atom requires touching `CONTRACT` is decided by reading the atom's existing fields, not by any new relationship:
    - **Status gate**: only `STABLE` atoms are eligible. `DRAFT`/`REVIEW` atoms are pre-guarantee — changing or removing them never touches `CONTRACT`.
    - **Layer gate**: only `BUSINESS`/`ARCHITECTURE` atoms are eligible. `IMPLEMENTATION` atoms (`MECHANIC`) are always internal and free to change as long as the guarantee above them still holds.
    - **Family/Type default** (per the families in §1.3):

      | Family | Types | Surface? |
      |---|---|---|
      | Requirements | `REQUIREMENT`, `USER_STORY` | Yes |
      | Interface | `API`, `UI`, `SPECIFICATION` | Yes |
      | Architectural | `SERVICE` | Yes |
      | Architectural | `MODULE`, `ENTITY` | Conditional — needs a human/agent judgment call, not automatic |
      | Logic | `RULE` at `BUSINESS` layer | Yes |
      | Logic | `RULE` at `ARCHITECTURE` layer, `DOMAIN` | No |
      | Governance | `CONTRACT`, `VISION` | Excluded (self) |
4. **Change classification** for a surface atom:
    - **Removed, or altered such that a prior guarantee no longer holds** → breaking → `CONTRACT`'s `version` MUST take a MAJOR bump in the same change, with explicit human confirmation.
    - **Newly promoted to `STABLE`, entering the surface for the first time** → additive → `CONTRACT`'s `version` takes a MINOR bump in the same change.
    - Anything the grid marks not-surface, or still `DRAFT`/`REVIEW` → `CONTRACT` is untouched.
5. **Overrides**: a breaking or additive surface change landing without the corresponding `CONTRACT` bump is a governance violation — reject it, or bump `CONTRACT` in the same change. Same principle for `VISION`: a `BUSINESS` atom outside its stated scope requires `VISION` to be updated to match, never silently allowed through.
6. **Bootstrapping**: If either is missing, the Agent MUST propose a definition based on the existing documentation and code — for `CONTRACT`, that means running the grid against whatever `STABLE` atoms already exist to produce its initial surface enumeration.

### 1.5 Operational-Root Exception

There are exactly two valid trace origins: a `BUSINESS` atom or an approved operational-root `MODULE`. An operational root is the sole non-governance parentless exception: it MUST be `type: MODULE`, `layer: ARCHITECTURE`, and `tags: [operational-root]`; it MUST be created as `DRAFT` with explicit human approval.

Use it only for a stable interface or compatibility obligation to a named external technical consumer or operational system. Do not use it for generic quality preferences, individual log statements, CI steps, dashboards, implementation notes, or tool choices without an external compatibility contract. Its normal sections MUST name the consumer/system, owned interface, acceptance criteria, and verification/owner. Descendants use normal parent rules and may be `API`, `SERVICE`, `SPECIFICATION`, `ENTITY`, `RULE`, or `MECHANIC`; their lineage must ultimately reach the operational root.

Example OpenTelemetry lineage: `module_telemetry_export` -> `spec_otel_signal_contract` -> `rule_otel_context_propagation` / `service_otel_exporter`. Document endpoint, protocol, and authentication ownership; resource attributes; signal names, units, and cardinality; trace-context propagation; sampling and redaction; and integration-test or collector-compatibility verification. Verification may use contract, integration, schema, or configuration tests; end-to-end tests are not required.

### 1.6 Document Hierarchy & Layers

Atoms are organized into three **layers** that reflect the documentation's relationship to change and human oversight. This hierarchy is the backbone of ATD's traceability model.

```
  BUSINESS                    ARCHITECTURE                IMPLEMENTATION
  ─────────────────           ─────────────────           ─────────────────
  REQUIREMENT                 MODULE                      MECHANIC
  USER_STORY                  ENTITY                      RULE (technical)
  RULE (business)             API / UI
  DOMAIN

  Volatility: LOW ◄──────────────────────────────────────────────► HIGH
  Human gate: HEAVY                MODERATE                    LIGHT
```

#### BUSINESS Layer

Atoms that capture **what the business wants and why**. These include requirements, user stories, business rules, and domain context.

- **Volatility:** Low. Once `STABLE`, these atoms are near-immutable. Alterations require high precautions and heavy human involvement (stakeholder sign-off, formal change requests).
- **Typical types:** `REQUIREMENT`, `USER_STORY`, `RULE` (business constraints), `DOMAIN`.
- **Traceability role:** One valid origin of the traceability chain. Every descendant must trace to a `BUSINESS` atom or approved operational-root `MODULE`.

#### ARCHITECTURE Layer

Atoms that capture **how the system is designed** to fulfill customer requirements. These include high-level modules, service orchestrators, entity definitions, API contracts, and UI flow designs.

- **Volatility:** Moderate. May be revised when unforeseen technical problems arise, but once stabilized they tend to remain fixed. Changes should trigger impact analysis (`atd crawl`).
- **Typical types:** `MODULE`, `SERVICE`, `ENTITY`, `API`, `UI`, `RULE` (architectural constraints).
- **Traceability role:** The **bridge** between customer intent and code reality.

#### IMPLEMENTATION Layer

Atoms that capture **how the code works in practice**. These include mechanics, data schemas, build pipelines, and developer-facing guides. These atoms evolve as the code is implemented, tested, and refactored.

- **Volatility:** High. Expected to change frequently during development. Subject to the doc-code co-evolution principle (§1.1, point 3).
- **Typical types:** `MECHANIC`, `DATA`, `BUILD`, `USAGE` (dev guides), `RULE` (technical constraints).
- **Traceability role:** The **leaf nodes** — linked directly to source code via `@spec-link` and to tests via `@test-link`. Must have a `BUSINESS` or approved operational-root `MODULE` ancestor (enforced by the pre-commit hook via `parents:` field).

### 1.7 Relationships & Graph Model

ATD builds a **bidirectional dependency graph** between atoms and between atoms and source code.

```
                    ┌──────────────┐
                    │   MODULE     │  (high-level)
                    └──────┬───────┘
               parents: ▲  │  ▼ :dependents
           ┌───────────────┼───────────────┐
     ┌─────┴──────┐  ┌─────┴──────┐  ┌────┴───────┐
     │  SERVICE    │  │   RULE     │  │  ENTITY    │
     └─────┬──────┘  └─────┬──────┘  └────┬───────┘
           │               │               │
     @spec-link       @spec-link      @spec-link
           │               │               │
     ┌─────┴──────┐  ┌─────┴──────┐  ┌────┴───────┐
     │  handler.go │  │ validate.go│  │  model.go  │  (source code)
     └─────┬──────┘  └────────────┘  └────────────┘
           │
     @test-link
           │
     ┌─────┴──────┐
     │handler_test │  (test code)
     └────────────┘
```

#### Link types

| Tag | Direction | Purpose |
|---|---|---|
| `parents: [[id]]` | Atom → Atom (upward) | "This atom is governed by…" |
| `dependents: [[id]]` | Atom → Atom (downward) | "These atoms depend on me…" (auto-populated by `weave`) |
| `@spec-link [[id]]` | Code → Atom | "This code implements…" |
| `@test-link [[id]]` | Test → Atom | "This test verifies…" |

#### Surgical Attachment Rules (for `@spec-link`)

1. **NO Global Headers:** Do not place `@spec-link` in the file header unless the atom represents the ENTIRE architectural pattern of the file.
2. **Logic Boundaries:** Place tags immediately above class definitions, decorators, or major logical blocks.
3. **Granularity Match:** If an atom describes a specific sub-feature, the tag goes only at the start of that specific section, NOT at the top of the file.

### 1.6 Lifecycle

ATD operates across two distinct lifecycle contexts: the **cold-start** (bootstrapping) and the **day-to-day** development cycle.

#### Cold Start (Bootstrapping an undocumented codebase)

Used once, at project inception or when onboarding ATD to a legacy codebase.

```
roadmap → index → author atoms → weave → discover → recon → audit
```

| Step | Tool | Description |
|---|---|---|
| 1. Complexity scan | `roadmap` | Identify high-density files to prioritize. |
| 2. Semantic index | `index` | Build the vector DB for search and matching. |
| 3. Authoring | (agent) | Author atoms for the prioritized source files. |
| 4. Graph weave | `weave` | Establish bidirectional parent/dependent links. |
| 5. Auto-tagging | `discover` | Recommend `@spec-link` placements in code. |
| 6. Validation | `recon` | Confirm that discovered links are correct. |
| 7. Full audit | `audit` | Check for bloat, collisions, and orphans. |

> The `atd-cold-start.sh <project_dir>` script orchestrates this full pipeline.

#### Day-to-Day Development Lifecycle

Once the ATD base is established, atoms and code co-evolve through recurring stages:

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                     │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐      │
│  │ PLAN     │───►│ SPECIFY  │───►│IMPLEMENT │───►│ VERIFY   │──┐   │
│  │          │    │          │    │          │    │          │  │   │
│  │Brainstorm│    │Create/   │    │Write code│    │audit     │  │   │
│  │Explore   │    │Update    │    │Add       │    │verify    │  │   │
│  │Ideate    │    │atoms     │    │@spec-link│    │test-links│  │   │
│  │          │    │(DRAFT)   │    │@test-link│    │          │  │   │
│  └──────────┘    └──────────┘    └──────────┘    └──────────┘  │   │
│       ▲                                                │        │   │
│       │                ┌──────────┐                    │        │   │
│       │                │ EVOLVE   │◄───────────────────┘        │   │
│       │                │          │                              │   │
│       │                │Impact    │                              │   │
│       │                │analysis  │                              │   │
│       │                │(crawl)   │                              │   │
│       │                └─────┬────┘                              │   │
│       └──────────────────────┘                                   │   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

| Stage | Context | Tools | Description |
|---|---|---|---|
| **Plan** | Feature request, brainstorm, early iteration | `query`, `search`, `assemble` | Explore existing atoms, identify related specs, brainstorm new features. Create `DRAFT` atoms for ideas still "in the air" — these document what's *being considered* without committing to implementation. Use `assemble` to pull together context for planning sessions. |
| **Specify** | Requirements & design | `update` (create), `weave` | Create or update atoms: `REQUIREMENT`, `USECASE`, `SPECIFICATION`, `MODULE`, `SERVICE`, etc. Set `status: DRAFT`. Link to parent atoms. Run `weave` to maintain the graph. |
| **Implement** | Investigation, Coding & testing | `trace`, `update`, `discover` | **Investigation Phase:** Before modifying any code, the agent MUST run `atd_trace(atom=..., summary=true)` to get a 500-word narrative assessment of the atom's vertical context and impact. Write code, annotate with `@spec-link [[atom_id]]` and `@test-link [[atom_id]]`. Update `IMPLEMENTATION`-layer atoms (`MECHANIC`, etc.) as code solidifies. Promote atoms to `REVIEW` when implementation is ready. |
| **Verify** | Validation | `audit`, `verify`, `test-links`, `crawl --gaps` | Run `verify` to check code-spec alignment. Run `audit` for bloat and collisions. Run `test-links` to confirm test coverage per atom. Run `crawl --gaps` to find orphaned atoms. |
| **Evolve** | Change management | `crawl`, `trace`, `update` | Before modifying an atom, run `atd_crawl` for blast radius (structural) and `atd_trace(summary=true)` for vertical context (semantic). Apply surgical edits via `update`. Promote stable atoms to `STABLE`. |

#### Status Transitions

Atoms move through maturity levels as they are refined:

```
DRAFT ──► REVIEW ──► STABLE ──► (DEPRECATED) ──► (ARCHIVED)
  │          │          │
  └──────────┴──────────┘  (can demote if spec changes)
```

- **DRAFT:** Initial ideas, brainstorming artifacts, early specs. Low validation overhead.
- **REVIEW:** Ready for peer or stakeholder validation. Subject to audit checks.
- **STABLE:** Approved and enforced. Code must comply. Changes require impact analysis.
- **DEPRECATED / ARCHIVED:** End-of-life (see ISS-033).

> **Layer-specific governance:** `BUSINESS` layer atoms at `STABLE` require heavy human sign-off to modify. `ARCHITECTURE` atoms require impact analysis. `IMPLEMENTATION` atoms can be updated more freely as code evolves.

### 1.7 Configuration

All tools read from a **`.atd`** JSON file at project root (created by `atd init`).

| Key | Purpose |
|---|---|
| `docs_path` | Relative path to the ATD docs folder. |
| `diff_similarity_threshold` | Cosine similarity threshold for collision detection (0.0–1.0). |
| `bloating_factor.default` | Default bloat tolerance. 1.0 = strict one-rule, 0.0 = no limit. |
| `bloating_factor.type_overrides` | Per-type overrides (e.g., `USECASE: 0.1` = very relaxed). |
| `llm.providers` | Ordered provider chain: `remote` → `local` → `ide_agent` passthrough. |
| `llm.models` | Map of model names to their assigned task types (e.g., `audit_code`, `embed`, `audit_bloat`). |
| `llm.fallback_model` | Default model when no task-specific model is found. |
| `logging.log_path` | Path for centralized ATD tool logging. |

> **Note:** `atd init` is a one-time filesystem bootstrap and is NOT exposed as an MCP tool.

---

## 2. MCP Toolset Reference

The `atd serve` command starts a JSON-RPC 2.0 / MCP (spec 2025-11-25) server via two transports:

- **stdio** (default): `atd serve` — host launches as subprocess.
- **HTTP**: `atd serve --http --port 7474` — single `/mcp` endpoint.

The server currently exposes **25 tools**.

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

---

### 2.1 Deterministic Tools

These tools NEVER call an LLM. They are safe, cheap, and fast.

---

#### `atd_query`

**Purpose:** Search atoms by frontmatter field value.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `search` | string | ✅ | Value to match (case-insensitive substring). |
| `field` | string | ❌ | Frontmatter field to search (e.g., `type`, `status`, `id`, `layer`, `tags`). Omit to search all fields. |

**Output:** JSON array of matching atoms with full frontmatter.

**When to use:** During **Plan** stage to find existing atoms before creating new ones, or to locate all atoms matching a criteria.

---

#### `atd_crawl`

**Purpose:** Build a dependency graph of ATD atoms and their `@spec-link` connections to source code.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `gaps` | boolean | ❌ | If `true`, return only `STABLE` atoms with zero code implementations (orphan detection). |

**Output:** Full dependency graph JSON (atoms → code links, parent/dependent edges). With `gaps=true`, only orphan atoms.

**When to use:**
- During **Evolve** stage before modifying a high-level atom, to understand ripple effects (blast radius analysis).
- During **Verify** stage to find `STABLE` atoms with no code implementations.

---

#### `atd_weave`

**Purpose:** Populate the `dependents[]` array in all atoms by scanning `parents` references. Establishes the bidirectional graph.

| Parameter | Type | Required | Description |
|---|---|---|---|
| *(none)* | — | — | Operates on the full docs directory. |

**Output:** Confirmation of weaving results (number of links created).

**When to use:**
- After creating or modifying atoms (especially `parents` fields).
- As part of any creation pipeline to ensure the graph is consistent.

---

#### `atd_update`

**Purpose:** Surgically modify an ATD file without LLM rewriting. Also injects `@spec-link` tags into source files.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `file` | string | ❌ | Path to the `.atom.md` file. Either `file` or `filter` must be provided. |
| `filter` | string | ❌ | Filter atoms to update natively (e.g., `'type=RULE,status=DRAFT'`). Mutually exclusive with `file`. |
| `set` | string[] | ❌ | Frontmatter edits as `key=value` pairs (e.g., `["status=STABLE", "priority=CORE"]`). |
| `intent` | string | ❌ | New `## INTENT` section text. |
| `logic` | string | ❌ | New `## THE RULE / LOGIC` section text. |
| `interface` | string | ❌ | New `## TECHNICAL INTERFACE` section text. |
| `expectation` | string | ❌ | New `## EXPECTATION` section text. |
| `spec_link` | string | ❌ | Atom ID to prepend as `@spec-link` in a source file (requires `spec_link_file`). |
| `spec_link_file` | string | ❌ | Source file path for `@spec-link` injection. |

**Output:** Confirmation of fields updated. If `id` or `type` is changed, file is renamed and all `[[old_id]]` references are propagated.

**When to use:**
- **MANDATORY** for editing existing `.atom.md` files. Do NOT rewrite the whole file.
- Changing status, priority, or any frontmatter field.
- **Batch Processing:** Use `filter` instead of `file` to apply the same update to all matching files atomically via a single call.
- Injecting `@spec-link` tags into source code without manual editing.
- Creating new atoms (provide the file path and all required fields).

---

#### `atd_roadmap`

**Purpose:** Scan a source directory and produce a complexity map ranking files by density.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `dir` | string | ✅ | Directory to scan. |
| `out` | string | ❌ | Optional output file path for `roadmap.json`. |

**Output:** JSON roadmap with files ranked by complexity (lines, cyclomatic density, function count).

**When to use:** During cold-start **Plan** stage to prioritize which files to document first.

---

#### `atd_stats`

**Purpose:** Produce quantitative documentation health metrics: total atoms, atoms by type, status, domain, coverage ratio, and orphan count.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `workspace` | boolean | ❌ | If `true`, aggregate stats from all projects in the workspace. |

**Output:** JSON report with quantitative health metrics (per-project when workspace=true).

**When to use:** During **Verify** stage to assess overall documentation quality, or in CI to generate health badges. Use `workspace:true` for a monorepo-wide view.

---

#### `atd_check`

**Purpose:** Unified coverage report: lists `@spec-link` (impl) and `@test-link` (test) coverage for every atom touched by the current diff or the full project.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `full` | boolean | ❌ | If `true`, audit the entire project regardless of changed files. |
| `file` | string | ❌ | Target a specific file for verification. |
| `line` | integer | ❌ | Target a specific line for verification (requires `file`). |
| `base` | string | ❌ | Base commit/ref to compare from (e.g. `HEAD~5`). |
| `target` | string | ❌ | Target commit/ref to compare to (defaults to working tree). |
| `semantic` | boolean | ❌ | Add LLM compliance check per `@spec-link` link. Returns PASS/FAIL per link. **Uses LLM — consumes tokens.** |

**Output:** Structured coverage report with impl/test link status per atom. With `semantic:true`, each impl link also gets a PASS/FAIL compliance verdict from the configured LLM.

**When to use:** During **Verify** stage (pre-commit or CI). Use without `semantic` for fast structural checks; add `semantic:true` for full spec compliance verification before promoting an atom to `STABLE`.

---

#### `atd_assemble`

**Purpose:** Stitch atoms together into a cohesive narrative document by walking the dependency graph.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `starts` | string | ✅ | Comma-separated list of root atom IDs to begin assembly from. |
| `purpose` | string | ❌ | Wraps output in `<System Objective>` tags for LLM consumption. |
| `snapshot` | boolean | ❌ | If `true`, delegate narrative generation to the IDE Agent (returns a task ID). **Uses LLM.** |
| `theme` | string | ❌ | Theme for snapshot narrative (default: "Executive Summary"). |

**Output:** Concatenated atom contents following the dependency graph. With `--snapshot`, an LLM-generated narrative.

**When to use:** During **Plan** stage for onboarding documents, architecture overviews, or executive summaries.

---

#### `atd_test_links`

**Purpose:** Audit `@test-link [[ATOM_ID]]` tags in source code to map atoms to their verification tests.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `atom` | string | ❌ | Filter for a specific atom ID. |

**Output:** JSON mapping of atom IDs to their associated test files and functions.

**When to use:** During **Verify** stage to confirm test coverage per atom, or before modifying an atom to identify which tests need re-running.

---

#### `atd_trace`

**Purpose:** Get a structured Health Snapshot JSON for a specific atom by traversing its graph ancestry and descendants. Optionally includes a narrative contextual summary via LLM.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `atom` | string | ✅ | Target ID of the atom to trace. |
| `summary` | boolean | ❌ | If `true`, adds a `summary` field with a ~500-word narrative contextual summary via LLM to the JSON output. **Uses LLM.** |

**Output:** JSON Health Snapshot with these fields:
- `target_id`: The canonical atom ID
- `layer`: Atom's layer (BUSINESS/ARCHITECTURE/IMPLEMENTATION)
- `health_summary`: Health metrics (ancestry_complete, has_business_origin, implementation_rate, test_coverage_rate)
- `metrics`: Detailed counts (total_dependents, implemented_dependents, total_code_files, total_tests)
- `graph_slice`: Lists of parents, dependents, code_links, test_links
- `context`: Map of atom IDs to AtomBrief objects (includes `file_path` for each atom)
- `warnings`: Layer compliance warnings
- `summary`: (optional) Narrative summary if `summary: true`

**When to use:** 
- **MANDATORY** during investigation phase prior to updating any code. 
- To diagnose an atom's health, check its dependency chain, or understand its vertical impact.
- To get the absolute file path to an atom definition (from `context[atom_id].file_path`)
- *Note: Narrative summary can be slow on lower-capability hardware but is essential for semantic context.*

---

#### `atd_env`

**Purpose:** Unified environment smoke test. Validates `.atd` config, checks provider connectivity, and verifies model availability. CLI equivalent: `atd env`.

| Parameter | Type | Required | Description |
|---|---|---|---|
| *(none)* | — | — | Automatically checks all configured providers. |

**Output:** Status list of providers (with available models) and a task resolution map (which model/provider will be used for each task type).

**When to use:** To diagnose "Connection Refused" or "Model Not Found" errors, or to verify a new provider/model is correctly configured.

---

#### `atd_config`

**Purpose:** View or modify the `.atd` configuration file. Supports granular queries for settings like bloating factors.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `list` | boolean | ❌ | If `true`, returns the full configuration as JSON. |
| `bloating_factor` | string | ❌ | Atom type to query for its bloating factor (e.g., `RULE`, `USECASE`). Check this before creating atoms. |
| `task` | string | ❌ | Task name to reassign (requires `model`). E.g., `audit_code`, `embed`, `audit_bloat`. |
| `model` | string | ❌ | Model name to assign to the task (requires `task`). |

**Output:** JSON configuration or a confirmation message of the update.

**When to use:** Checking bloating factors before creating atoms, or reassigning LLM task-to-model mappings.

---

#### `atd_lint`

**Purpose:** Perform fast, deterministic structural validation on all ATD atoms.

| Parameter | Type | Required | Description |
|---|---|---|---|
| *(none)* | — | — | Scans the documented project directory. |

**Output:** List of structural violations (missing fields, broken links, empty sections).

**When to use:** During **Verify** stage as a cheap first-pass before running the heavier `atd_audit`.

---

#### `atd_workspace_list`

**Purpose:** List all projects registered in the current `.atd.workspace` file.

| Parameter | Type | Required | Description |
|---|---|---|---|
| *(none)* | — | — | Reads the workspace config from the project root. |

**Output:** JSON array of project names and their paths.

**When to use:** At the start of any workspace task to discover available projects before switching context.

---

#### `atd_workspace_use`

**Purpose:** Switch the active project context to a specific project in the workspace.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `project` | string | ✅ | The project name to activate (must match a name in `.atd.workspace`). |

**Output:** Confirmation of the active project switch.

**When to use:** Before any ATD operation targeting a specific sub-project. Must be called explicitly — tools do not auto-switch context.

---

#### `atd_workspace_stats`

**Purpose:** Aggregate documentation health metrics across all projects in the workspace.

| Parameter | Type | Required | Description |
|---|---|---|---|
| *(none)* | — | — | Reads all registered workspace projects. |

**Output:** Per-project and workspace-wide atom counts, status breakdowns, and coverage summaries.

**When to use:** For workspace-level health dashboards and cross-project comparison. Equivalent to running `atd_stats` on each project and aggregating the results.

---

#### `atd_heatmap`

**Purpose:** Get heat map metrics for a specific atom across three layers: dependency coupling, code implementation density, and update instability.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `atom` | string | ✅ | Atom ID or path to the `.atom.md` file. |

**Output:** JSON heat metrics: dependency score, code link count, change instability score.

**When to use:** To identify "hot" atoms — highly connected, frequently changed, or over-implemented — before modifying or splitting them.

---

#### `atd_heatmap_code`

**Purpose:** Get heat map metrics for a specific source file based on `@spec-link` density and implementation pressure.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `file` | string | ✅ | Path to the source file. |

**Output:** JSON heat metrics for the file: number of linked atoms, spec-link density, instability score.

**When to use:** To spot over-documented or under-documented source files, and to prioritize review effort.

---

#### `atd_heatmap_project`

**Purpose:** Get a project-wide heat map summary ranked by documentation hotspots.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `layer` | string | ❌ | Heat layer to focus on: `dependency`, `code`, `updates`, or `all` (default). |

**Output:** Ranked list of atoms and files by heat score per selected layer.

**When to use:** During architecture reviews or sprint planning to identify which parts of the system carry the most documentation risk.

---

### 2.2 LLM-Backed Tools

These tools require an Ollama provider (local or remote) or fall back to IDE Agent passthrough. They consume tokens.

---

#### `atd_index`

**Purpose:** Build or refresh the semantic vector index of all source code and ATD documents.

| Parameter | Type | Required | Description |
|---|---|---|---|
| *(none)* | — | — | Indexes the entire project using `.atd` configuration. |

**Output:** Confirmation of indexing results (chunks processed, vectors stored, files skipped).

**When to use:**
- Before running semantic search (`atd_search`).
- After significant code or documentation changes to refresh the index.
- Uses `nomic-embed-text` for embeddings. Files unchanged since last indexing are automatically skipped (mtime-based caching).
- **Stale Entry Cleanup:** Automatically removes index entries for files that have been deleted from disk.
- **Chunk Length Handling:** Implements robust splitting to prevent context-length overflows on large files.

---

#### `atd_search`

**Purpose:** Search the project semantically or by keyword.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `query` | string | ❌ | Semantic search query (uses Nomic embeddings). Provide this OR `grep`, not both. |
| `grep` | string | ❌ | Literal keyword search across project files. Provide this OR `query`, not both. |
| `scope` | string | ❌ | Search scope: `code` (source files only), `docs` (ATD atoms only), or `all` (both). Default: `all`. |
| `limit` | integer | ❌ | Number of semantic results to return (default: 5). |

**Output:** Ranked results with similarity scores and file context.

**When to use:**
- During **Plan** stage to find related code or atoms by meaning, or to locate implementations when atom IDs are unknown.
- **Model:** `nomic-embed-text` for query embedding.

---

#### `atd_audit`

**Purpose:** Audit ATD atoms for documentation quality issues.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `atom` | string | ❌ | Path to atom file for **compliance mode** (requires `code`). |
| `code` | string | ❌ | Path to code file for **compliance mode** (requires `atom`). |

**Output:** Bloat warnings, collision pairs with similarity scores, and/or compliance verdict.

**When to use:**
- **Default mode:** Global quality audit — run during **Verify** stage to detect duplicate or bloated atoms.
- **Compliance mode:** Validate that a specific code file conforms to a specific atom's specification.
- After creating new atoms, to check for semantic overlap with existing ones.

---

#### `atd_recon`

**Purpose:** Semantic archaeology — validate whether a candidate source file implements a specific ATD atom.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `atom` | string | ✅ | Path to the target `.atom.md` file. |
| `candidate` | string | ✅ | Path to the candidate source code file to validate. |

**Output:** Confidence score and rationale for whether the code implements the atom.

**When to use:**
- Cold-start tagging: verifying a discovered file before applying `@spec-link`.
- Auditing existing links: confirming that an `@spec-link` is still valid after refactoring.
- **Model:** `qwen2.5-coder:14b`.

---

#### `atd_map`

**Purpose:** Three-mode tool for linking source code to ATD atoms. CLI equivalent: `atd map`.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `file` | string | ✅ | Path to the source file to analyse. |
| `atom` | string | ❌ | **Confirm mode:** atom ID to validate against the file. Returns a confidence score and rationale. |
| `new` | boolean | ❌ | **Propose mode:** return a new atom skeleton (id, type, layer, intent, logic) ready to pass to `atd_update`. |

**Modes:**
- **Default** (`file` only): Searches existing atoms for candidates and recommends `@spec-link` placements following surgical attachment rules.
- **Confirm** (`file` + `atom`): Validates whether the file implements the given atom. Equivalent to `atd_recon` but accepts an atom ID instead of a file path.
- **Propose** (`file` + `new: true`): Extracts architectural intent and proposes a brand-new atom skeleton when no existing atom fits.

**When to use:**
- **Default:** During **Implement** stage to link new or legacy files to the right atoms.
- **Confirm:** Before applying `@spec-link` tags during cold-start or after refactoring to verify the match.
- **Propose:** When adding a file that has no existing atom — use this to bootstrap a new DRAFT atom.

---

### 2.3 Tool Decision Matrix

| Scenario | Tool(s) | LLM? |
|---|---|---|
| "Find all atoms of type RULE" | `atd_query` (field=type) | No |
| "What atoms does this code implement?" | `atd_crawl` | No |
| "Find STABLE atoms with no code links" | `atd_crawl` (gaps=true) | No |
| "Synchronize parent/dependent graph" | `atd_weave` | No |
| "Change an atom's status to STABLE" | `atd_update` (set=["status=STABLE"]) | No |
| "Batch promote matching atoms" | `atd_update` (filter="type=RULE,status=DRAFT") | No |
| "What files are most complex?" | `atd_roadmap` | No |
| "Check impl/test link coverage" | `atd_check` | No |
| "Check if code still complies with spec" | `atd_check` (semantic=true) | Yes |
| "Generate a project overview" | `atd_assemble` | No (unless snapshot) |
| "Which tests cover this atom?" | `atd_test_links` | No |
| "Which atoms is this file linked to?" | `atd_heatmap_code` | No |
| "Which atoms are riskiest to modify?" | `atd_heatmap_project` | No |
| "Is this provider configured correctly?" | `atd_env` | No |
### 2.3 Tool Decision Matrix

| Task | Tool | LLM? |
|---|---|---|
| "List workspace projects" | `atd_workspace_list` | No |
| "Switch active project" | `atd_workspace_use` | No |
| "Workspace-wide health report" | `atd_workspace_stats` | No |
| "Build a searchable index" | `atd_index` | Yes (embed) |
| "Find code related to authentication" | `atd_search` (query=...) | Yes (embed) |
| "Is this documentation bloated?" | `atd_audit` | Yes |
| "Does this code match this atom?" | `atd_audit` (atom=X, code=Y) | Yes |
| "Find matching atoms for this file" | `atd_map` | Yes |
| "Confirm this file implements this atom" | `atd_map` (atom=id) | Yes |
| "Propose a new atom for this file" | `atd_map` (new=true) | Yes |
| "Inject @spec-link without manual edit" | `atd_update` (spec_link=id) | No |

---

## 3. Agent Behavioral Rules (Intransigence Guidelines)

To maintain graph integrity, all agents MUST follow these "hard behavioral triggers":

### 3.1 The "No Parent, No Code" Rule
If the user asks to implement a feature or mechanic:
1. **Search**: You MUST first execute `atd_search` to find relevant atoms.
2. **Verify Ancestry**: Check that the feature traces to a `BUSINESS` atom or an approved operational-root `MODULE`.
3. **STOP**: If neither exists, **STOP**. Do not write code or an `IMPLEMENTATION` atom. The sole non-governance parentless exception is an approved `MODULE` created as `DRAFT` and tagged `operational-root` for a named external compatibility obligation.
4. **Interview**: Propose the missing `BUSINESS`/`ARCHITECTURE` atom, or strictly-qualified operational root, and ask for approval before creation.

### 3.2 Governance Check
Before adding or modifying any `BUSINESS` layer atom, you MUST read the project's `CONTRACT` and `VISION` atoms.
- If the change violates the `CONTRACT` (removal of mandatory features) or `VISION` (scope creep), you must warn the user and require explicit confirmation.

---

## 4. Gap Analysis & Recommendations

The following gaps and improvement opportunities are identified from the current backlog, system analysis, and real-world usage patterns.

### 3.1 Already Tracked in Backlog

These issues are already filed and represent known gaps:

| Ref | Gap | Impact |
|---|---|---|
| **ISS-009** | **Version management is ignored.** The `version` field is parsed but never used for conflict resolution or latest-version selection. | Multiple versions of the same atom can silently coexist. |
| **ISS-010** | **Status has no enforcement.** `DRAFT`, `REVIEW`, `STABLE` are labels without workflow logic. No gated transitions, no stricter validation for `STABLE`. | All atoms are treated equally regardless of maturity. |
| **ISS-025** | **Creation produces sparse atoms.** `atd update` only populates `intent`; `parents`, `version`, `logic`, and `expectation` are left empty. | New atoms are incomplete and require manual cleanup. |
| **ISS-019** | **API atoms lose payload details.** `type: API` atoms created from detailed contracts capture only a subset of the specification. | API documentation is unreliable for contract verification. |
| **ISS-030** | **No targeted impact analysis.** `atd crawl` returns the entire graph; no `--target` flag for scoped ripple-effect analysis. | Architects cannot quickly assess the blast radius of a change. |
| **ISS-014** | **Issues and atoms are disconnected.** No machine-readable link between `issues/` and `docs/` atoms. | Technical debt is invisible when viewing a spec, and vice versa. |
| **ISS-013** | **Lack of self-documentation.** The ATD project does not dogfood its own framework comprehensively. | New contributors face a steep onboarding curve. |

### 3.2 New Gaps Identified

Beyond the backlog, the following capabilities would round out ATD as a comprehensive documentation governance tool:

#### 3.2.1 Atom Deprecation & Archival

**Problem:** There is no `DEPRECATED` or `ARCHIVED` status. When a feature is removed, the corresponding atom lingers as `STABLE` with no formal lifecycle end.

**Recommendation:** Add `DEPRECATED` and `ARCHIVED` statuses. `DEPRECATED` atoms should trigger warnings in `atd audit` if code still links to them. `ARCHIVED` atoms should be excluded from all tool outputs unless explicitly queried.

#### 3.2.2 Change History / Changelog per Atom

**Problem:** Atoms have a `version` field but no change log. When an atom is modified, there is no record of what changed, when, or why.

**Recommendation:** Implement a sidecar `<atom_id>.changelog.json` file (co-located next to the `.atom.md`). `atd update` should auto-append an entry (timestamp, field changed, old → new value, user/agent) whenever a field is modified. This is critical for `STABLE` atoms where traceability of spec changes is a governance requirement. The sidecar approach avoids polluting the atom file itself and allows tooling to parse change history independently.

#### 3.2.3 Cross-Project Atom Sharing

**Problem:** ATD is currently single-project. There is no mechanism for sharing atoms between repositories (e.g., a shared `DOMAIN` atom for company-wide business rules).

**Recommendation:** Support a `refs` or `imports` section in `.atd` config that points to external atom repositories (git URLs or local paths). `atd crawl` and `atd search` should resolve cross-project links. This also requires upgrading the tooling to handle **destination selection** — when creating or linking atoms, the user must be able to specify which project/repository the atom belongs to. Multi-repository projects would benefit significantly from this capability.


#### 3.2.7 Atom Templates per Type

**Problem:** All atoms use the same generic template. `API` atoms would benefit from built-in request/response schema sections; `USECASE` atoms need a `## WORKFLOW` section; `USER_STORY` atoms need `## ACCEPTANCE CRITERIA`.

**Recommendation:** `atd update` should apply type-specific templates when creating new atoms. Templates for `USECASE` and `USER_STORY` already exist in the skill but are not integrated into the CLI/MCP tooling.

#### 3.2.8 Conflict Resolution During Reconciliation

**Problem:** When new external requirements arrive that semantically overlap with existing atoms, there is no structured merge/conflict resolution protocol beyond manual review.

**Recommendation:** Formalize the Reconciler sub-mode as an MCP tool (`atd_reconcile`) that: (1) searches for semantically similar existing atoms, (2) classifies the overlap as `MERGE`, `UPDATE`, or `NEW`, and (3) presents the options to the agent/user. This tool currently exists as a CLI concept (`atd reconcile`) but is not exposed via MCP.

#### 3.2.9 Crawl Summary & Graph Visualization

**Problem:** The dependency graph from `atd crawl` is a raw JSON blob. There is no standard way to render it visually for architecture reviews, and no way to get a human-readable summary of the findings.

**Recommendation:**
1. **Summary output:** `atd crawl` should produce a structured summary of all extracted information alongside the graph — total atoms, link counts, orphans, coverage stats, and key findings. This summary should be included in the default output, not behind a flag.
2. **Mermaid export:** Add a `--format mermaid` flag to `atd crawl` that outputs the graph as a Mermaid diagram, directly embeddable in Markdown documents and renderable by most Markdown viewers. Mermaid format is sufficient; Graphviz/DOT is not required.
3. **WebUI integration:** The WebUI should consume this graph data for interactive exploration.

#### 3.2.10 WebUI Specification

**Problem:** The WebUI currently exists only as a kernel of an idea — basic rendering with no formal specification, no connection to the ATD binary tools, and no defined user experience.

**Recommendation:** Create a comprehensive specification for the WebUI as a set of ATD atoms (`MODULE`, `UI`, `SERVICE` types). This should cover: graph visualization, atom CRUD operations, audit dashboards, cold-start orchestration, and integration with the `atd` CLI/MCP tooling. The specification should be built before further UI development begins.

---

### 3.3 Priority Matrix

| Gap | Impact | Effort | Suggested Priority |
|---|---|---|---|
| Cold-start + audit via MCP (ISS-031) | High — unblocks MCP-only workflows | Medium | **P0** |
| Targeted impact analysis (ISS-030) | High — unblocks Architect mode | Low | **P0** |
| Status enforcement (ISS-010) | High — governance foundation | Medium | **P1** |
| Atom deprecation (§3.2.1) | Medium — lifecycle completeness | Low | **P1** |
| Batch operations (§3.2.5) | Medium — MCP efficiency | Medium | **P1** |
| Change history sidecar (§3.2.2) | Medium — audit trail | Medium | **P2** |
| Type-specific templates (§3.2.7) | Medium — atom quality | Low | **P2** |
| Version management (ISS-009) | Medium — dedup safety | Medium | **P2** |
| Crawl summary + Mermaid (§3.2.9) | Medium — architecture reviews | Low | **P2** |
| Reconcile via MCP (§3.2.8) | Medium — onboarding workflow | Medium | **P2** |
| WebUI specification (§3.2.10) | Medium — blocks further UI dev | Medium | **P2** |
| Cross-project sharing (§3.2.3) | Low — multi-repo orgs | High | **P3** |
