---
name: atd-create-atom
description: Create new ATD atoms with proper parent verification, governance gates, and validation. Use this when asked to create documentation, specs, requirements, or add atoms for features.
---

# ATD Create Atom Skill

## When to Activate

Use this skill whenever the user asks you to:
- Create a new atom / spec / requirement
- Add documentation for a feature
- Define a new API, UI component, or service
- Author a user story, use case, or business rule
- "Write ATD docs for X" or "Document feature Y"

---

## Pre-flight Checklist

Before creating any atom, ensure:
- [ ] The `atd_audit` and `atd_search` tools are available and functioning (not silently returning zeros)
- [ ] You have access to the correct project context (run `atd_workspace_list` and `atd_workspace_use` if needed)
- [ ] The `atd_config` tool is accessible for granularity calibration

---

## Step 1 — Search for Existing/Overlapping Atoms

Search both via grep and semantic to avoid creating duplicates:

```bash
# Grep search (fast, deterministic)
mcp__atd__atd_search(grep="your feature concept")

# Semantic search (if available)
mcp__atd__atd_search(query="your feature description in natural language")
```

**If you find overlapping atoms:**
- Present them to the user with their IDs and summaries
- Ask whether to extend the existing atom(s) or create a new one
- If extending, switch to **Skill B (atd-alter-atom)** instead

---

## Step 2 — Verify Parent Atom Exists (The "No Parent, No Code" Rule)

Every atom except CONTRACT and VISION MUST have a parent BUSINESS or ARCHITECTURE atom.

1. Identify the appropriate layer for your new atom:
   - **BUSINESS**: Requirements, user stories, business rules, domain context
   - **ARCHITECTURE**: APIs, entities, modules, UI components, specifications
   - **IMPLEMENTATION**: Mechanics, algorithms, validation rules

2. For BUSINESS-layer atoms:
   - The parent must be BUSINESS-layer
   - You must read CONTRACT and VISION (governance gate)

3. For ARCHITECTURE-layer atoms:
   - The parent can be BUSINESS or ARCHITECTURE
   - Verify the ancestry chain reaches BUSINESS (no orphan subtrees)

4. For IMPLEMENTATION-layer atoms:
   - The parent MUST be ARCHITECTURE
   - **STOP** if you cannot trace a BUSINESS ancestor

5. If no suitable parent exists:
   - **STOP and propose the missing upstream atom to the user**
   - Ask for explicit confirmation before proceeding
   - Example: "To create this API atom, I first need a BUSINESS-layer parent. Shall I create `req_user_authentication` as a REQUIREMENT atom?"

---

## Step 3 — Read CONTRACT and VISION (For BUSINESS-Layer Atoms Only)

If creating a BUSINESS-layer atom (REQUIREMENT, USER_STORY, RULE, DOMAIN):

```bash
# Find CONTRACT atom
mcp__atd__atd_query(field="type", search="CONTRACT")

# Read the full CONTRACT atom file using the file_path from the query result
read(filePath="/path/from/query/output/contract_atd.atom.md")

# Find VISION atom
mcp__atd__atd_query(field="type", search="VISION")

# Read the full VISION atom file using the file_path from the query result
read(filePath="/path/from/query/output/vision_atd.atom.md")
```

Verify your atom aligns with:
- CONTRACT invariants (the project's governing rules)
- VISION scope boundaries (what's in-scope vs out-of-scope)

If your atom would violate either, **STOP** and discuss with the user.

---

## Step 4 — Pick Type and Layer

Use the canonical types from ATD.md:

**BUSINESS Layer:**
- `REQUIREMENT` — High-level business constraints
- `USER_STORY` — "As a [role], I want [X]" (synonyms: `USECASE`, `WORKFLOW`)
- `RULE` — Single business or technical constraint
- `DOMAIN` — Narrative context, "The Why"

**ARCHITECTURE Layer:**
- `MODULE` — High-level grouping / service boundary
- `SERVICE` — Command surface / service orchestration
- `API` — Interface contracts (endpoints, RPC)
- `ENTITY` — Data models and state structures
- `UI` — Screen or interaction flow
- `SPECIFICATION` — Config schema, data schema, or interface specification

**IMPLEMENTATION Layer:**
- `MECHANIC` — Algorithms and procedural logic

**Special Types (layer-independent):**
- `CONTRACT` — Unique; project-wide mandatory rules
- `VISION` — Unique; project-wide scope/philosophy

---

## Step 5 — Calibrate Granularity

Check the bloating factor configuration to ensure appropriate atomicity:

```bash
mcp__atd__atd_config(bloating_factor="<chosen_type>")
```

Enforce the **"no 'and' in INTENT"** rule:
- If you catch yourself using "and" in the intent, split into multiple atoms
- One atom = one state-changing rule or concept
- Ask the user to confirm the split if you're uncertain

---

## Step 6 — Author the Four H2 Sections

Every atom must have these four H2 sections:

```markdown
## INTENT
[What this atom defines or governs, single sentence, no "and"]

## LOGIC
[The rule, algorithm, or constraint, in clear prose]

## TECHNICAL INTERFACE
[API signatures, data structures, or implementation details]

## EXPECTATION
[What should be true when this is satisfied or implemented]
```

**Type-Specific Templates:**

**For `API` atoms:**
```markdown
## TECHNICAL INTERFACE
**Endpoint:** `POST /api/v1/auth/login`
**Request Body:**
```json
{
  "username": "string",
  "password": "string"
}
```
**Response:** `200 OK` with JWT token
```

**For `USECASE`/`USER_STORY` atoms:**
```markdown
## WORKFLOW
1. User initiates action
2. System validates preconditions
3. System performs operation
4. System returns result

## ACCEPTANCE CRITERIA
- [ ] Scenario A works
- [ ] Scenario B works
- [ ] Error case C handled correctly
```

**For `ENTITY` atoms:**
```markdown
## TECHNICAL INTERFACE
```go
type Player struct {
    ID       string
    Username string
    Level    int
    XP       int
}
```
```

---

## Step 7 — Create the Atom via `atd_update`

NEVER hand-write the `.atom.md` file. Always use the tool:

```bash
mcp__atd__atd_update(
  file="docs/your_atom_name.atom.md",
  set=[
    "id=your_atom_name",
    "type=REQUIREMENT",
    "layer=BUSINESS",
    "parents=[[parent_atom_id]]",
    "status=DRAFT",
    "priority=5"
  ],
  intent="Your single-sentence intent",
  logic="Your logic section content",
  technical_interface="Your TECHNICAL INTERFACE section",
  expectation="Your EXPECTATION section"
)
```

**Critical Rules:**
- Always start with `status=DRAFT`
- Always set at least one parent (except CONTRACT/VISION)
- ID should be `layer_type_name` format (e.g., `req_user_auth`, `api_player_join`)

---

## Step 8 — Verify and Integrate

Run the verification sequence:

```bash
# 1. Build dependency graph
mcp__atd__atd_weave()

# 2. Check for collisions or conflicts
mcp__atd__atd_audit()

# 3. Verify structural integrity on any files the atom might reference
mcp__atd__atd_check(file="path/to/related/code/file.go")
```

**Address any issues:**
- If `audit` reports conflicts, resolve with the user
- If `check` finds problems, the atom may need refinement
- Re-run `weave` after any fixes

---

## Step 9 — Notify the User

Tell the user:
1. The atom ID and file path
2. The type, layer, and parent(s)
3. The intent (one-line summary)
4. The next steps (e.g., "Ready for implementation" or "Needs review before promoting to REVIEW")

Example:
> Created atom `req_user_authentication` at `docs/req_user_authentication.atom.md` (REQUIREMENT, BUSINESS layer, parent: `contract_atd`). Intent: "Users must authenticate with username and password before accessing protected resources." Status: DRAFT, ready for implementation planning.

---

## Quick Reference Checklist

```
[ ] Searched for overlapping atoms (grep + semantic)
[ ] Verified parent atom exists (BUSINESS/ARCHITECTURE)
[ ] Read CONTRACT/VISION if BUSINESS-layer atom
[ ] Chose canonical type and layer
[ ] Checked bloating factor config
[ ] Verified "no and" in INTENT rule
[ ] Authored four H2 sections (INTENT, LOGIC, TECHNICAL INTERFACE, EXPECTATION)
[ ] Created via atd_update (never hand-wrote file)
[ ] Set status=DRAFT and parent(s)
[ ] Ran atd_weave to build graph
[ ] Ran atd_audit for collisions
[ ] Ran atd_check on touched files
[ ] Notified user with ID, path, intent, and next steps
```

---

## Error Recovery

**If `atd_search` returns silently empty:**
- The index may be missing. Try `atd_index` first, then retry the search.
- Fall back to `atd_query(field="type", search="...")` for deterministic search.

**If `atd_audit` reports conflicts:**
- Review the conflicting atoms with the user
- Decide whether to merge, rename, or split based on business intent
- Use `atd_update` to make adjustments, never edit files directly

**If parent atom verification fails:**
- Propose the missing parent atom to the user
- Offer to create the parent first (this skill) before proceeding with the child
- Never create orphan atoms (atoms without BUSINESS ancestry)

---

## Common Pitfalls

❌ **DON'T:**
- Skip parent verification ("I'll add it later")
- Use "and" in the INTENT section (violates atomicity)
- Hand-write `.atom.md` files (breaks traceability)
- Create IMPLEMENTATION atoms without ARCHITECTURE parents
- Set status to REVIEW or STABLE immediately (always start DRAFT)
- Forget to run `atd_weave` after creation

✅ **DO:**
- Search before creating (avoid duplicates)
- Verify ancestry traces to BUSINESS layer
- Read CONTRACT/VISION for BUSINESS-layer atoms
- Use `atd_update` for all atom creation
- Start with status=DRAFT
- Run the full verification sequence
- Notify the user with clear next steps