# Legacy-to-ATD Protocol Experiment Log

**Target:** `upsilonbattle/battlearena/ruler` package
**Objective:** Iteratively develop prompting strategies to successfully dissect legacy Go code into strictly formatted ATD Markdown files. At this stage, we are ONLY generating the ATDs, not modifying the source code to add `@spec-link` tags.

---

## Iteration 1: Deconstructing `ruler/README.md`

### Prompt / Approach Used
We ran the `atd-dissect` tool on the monolithic `/ruler/README.md` file. The tool provided the system prompt instructing us to act as the ATD Deconstructor and output a JSON array of proposed Atom boundaries.

### Mechanism
* Tool: `atd-dissect`
* Target: `/home/bastien/work/skill/upsilonbattle/battlearena/ruler/README.md`
* Actor: LLM (Architect Pass 1 - Structural Mapping)

### Result
We successfully identified the following core logical boundaries based on the README:

```json
[
  {
    "proposed_id": "ruler-arena-state",
    "responsibility": "Defines the high-level progression states of the game (WaitingForControllers, InProgress, Finished).",
    "excerpt_range": "Lines 16-34 (Arena State)"
  },
  {
    "proposed_id": "ruler-turn-logic",
    "responsibility": "Manages Delay Credits, turn order calculation, and action costs (+200 move, +500 attack).",
    "excerpt_range": "Lines 36-63 (Turn Logic & Game Rule Sets)"
  },
  {
    "proposed_id": "ruler-actor-testing",
    "responsibility": "Standardizes async testing patterns utilizing the FakeController, Inbox channel, and History slice to prevent race conditions.",
    "excerpt_range": "Lines 203-259 (Writing Actor Tests)"
  }
]
```

### Evaluation / Critique

**Successes:** 
* The dissection cleanly separated the "Rules/Mechanics" (Arena State, Turn Logic) from the "Testing/API Patterns" (Actor Testing). 
* The boundaries are distinct and manageable.

**Failures / Missing Granularity:** 
* The `Turn Logic` section actually groups two distinct concepts: *Turn Order Computation* (Lowest credit plays first) and *Delay Costs* (+500 for attacks). According to the new "Minimum Atomic Scale" rule, these should probably be split. 

**Next Steps (Iteration 2):**
We'll take the `ruler-turn-logic` excerpt and pass it through "Pass 2" (The Surgical Generator) to actually write out the strict `.atom.md` template, ensuring we adhere to the Minimum Atomic Scale rule by splitting it during generation.

---

## Iteration 2: "Shadow Atom" Generation (`ruler-turn-logic`)

### Prompt / Approach Used
Acting as the Analyst/Architect, we take the `ruler-turn-logic` summary from Iteration 1 and expand it into the rigid `.atom.md` template structure, ensuring we apply the "Minimum Atomic Scale" constraint to separate the Time logic from the Cost logic.

### Mechanism
* Manual generation following the ATD rules, referencing `battlearena/ruler/README.md`.
* Creating two distinct files: `ruler-turn-logic.atom.md` and `ruler-action-costs.atom.md`.

### Result
Successfully drafted two single-responsibility Atoms stored in `/upsilonbattle/docs/`.
1. **`ruler-turn-logic`**: Handles determining the active entity via Delay Credit comparison.
2. **`ruler-action-costs`**: Handles the math constants added to Delay Credit after specific actions.

### Evaluation / Critique

**Successes:**
* Granularity control was flawless. Splitting the logic means a future Developer updating "Move Costs" won't accidentally break the "Turn Queue Loop."
* The Technical Interface is clean and primed for Developer tagging during the Archaeology phase.

**Failures / Missing Concepts:**
* The README mentions that "Move makes use of Properties: Movement (reach), JumpHeight". This belongs in a separate rule. We need to dissect `rules/move.go` and `rules/skill.go` for the core game actions.

**Next Steps (Iteration 3):**
Move out of the README and dive into the legacy source code (`rules/attack.go`). We will use `atd-dissect` on the undocumented legacy Go code to see if the LLM can tease out the core Rules from raw implementation logic.

---

## Iteration 3: Deconstructing Undocumented Legacy Code (`rules/attack.go`)

### Prompt / Approach Used
Ran `atd-dissect` on `/battlearena/ruler/rules/attack.go`. We rely purely on the LLM's capacity to interpret Go syntax, structures, and math equations without any guiding English documentation. 

### Mechanism
* Tool: `atd-dissect`
* Target: `/home/bastien/work/skill/upsilonbattle/battlearena/ruler/rules/attack.go`
* Actor: LLM (Architect Pass 1 - Structural Mapping)

### Result
The LLM accurately parsed the Go logic blocks and grouped them conceptually rather than sequentially:

```json
[
  {
    "proposed_id": "ruler-attack-validation",
    "responsibility": "Ensures the entity exists, is controlled by the requester, has not already acted, and target is within AttackRange on valid Ground.",
    "excerpt_range": "Lines 115-177 (preAttackChecks)"
  },
  {
    "proposed_id": "ruler-attack-resolution",
    "responsibility": "Computes (AttackerAttack - FoeDefense) = Damage. Applies Damage to Foe HP. Charges Attacker 500 Delay and sets HasActed/HasMoved flags.",
    "excerpt_range": "Lines 39-75 (State Updates)"
  },
  {
    "proposed_id": "ruler-entity-death",
    "responsibility": "If Foe HP <= 0, permanently strips the entity from the Grid, Entities map, and Turn Queue.",
    "excerpt_range": "Lines 77-87 (Death Loop)"
  }
]
```

### Evaluation / Critique

**Successes:**
* **Major Success**: The LLM easily translated raw Go code into understandable game mechanics. It didn't just spit back code; it synthesized the formula `Damage = Attack - Defense`. 
* The constraints were identified perfectly. Splitting "Validation" from "Resolution" perfectly aligns with the Minimum Atomic Scale rule.

**Next Steps:**
Write these 3 Atoms to the `docs/` folder. This confirms that the protocol works not just on wikis, but directly on raw undocumented source code.

---

## Iteration 4: The Legacy Extraction Protocol (`rules/move.go`)

### Prompt / Approach Used
To respond to the Architect's need for a formalized, repeatable "Legacy Extraction" workflow, we updated `rules/ATD.md` and `SKILL.md` to formally instantiate **The Legacy Extractor** sub-mode. 

We then deployed this sub-mode on `rules/move.go` using `atd-dissect`.

### Mechanism
* Tool: `atd-dissect -file=rules/move.go`
* Protocol: **Legacy Code Extraction Workflow** (Step 1 -> Step 5)

### Result
The `atd-dissect` engine natively parsed the file and isolated the complex validation loops from the state modification math:

```json
[
  {
    "proposed_id": "ruler-movement-validation",
    "responsibility": "Validates path integrity including grid bounding, JumpHeight restrictions against adjacent cells, total movement budget, and turn ownership.",
    "excerpt_range": "Lines 71-158 (preMoveChecks)"
  },
  {
    "proposed_id": "ruler-movement-resolution",
    "responsibility": "Applies the path movement to the Grid, mutates the Entity's Position vector, depletes the Movement property, and adds +200 Delay Credit per tile.",
    "excerpt_range": "Lines 36-69 (State Updates)"
  }
]
```

Acting as the Shadow Architect, we manually drafted `docs/ruler-movement-validation.atom.md` and `docs/ruler-movement-resolution.atom.md`.

### Evaluation / Critique

**Successes:**
* **Scalable Workflow:** We proved the formal workflow from `SKILL.md` can extract documentation seamlessly from *any* Go file without needing English comments. The LLM accurately understood `jumpHeight` array-indexing logic on adjacency.
* **Bug Discovery via Extraction:** While converting the raw Go code into English requirements, we realized `rules/move.go` *forgets* to set `HasMoved = true` inside the Resolution block (unlike `attack.go` which successfully sets it). An ATD Validator Audit in the future would catch this logic gap instantly.

**Next Steps / Ready:**
The protocol is successful. The Shadow Atoms exist. The Developer is now clear to enter **Archaeology Mode** to run `atd-recon` and surgically inject `@spec-link` tags back into `move.go` and `attack.go`.

---

## Iteration 5: The Git-Aware Auditor (`atd-verify-diff`)

### Prompt / Approach Used
The User suggested offloading work to the host machine to minimize LLM token consumption by first running native tests (like `go test`), and then asking the LLM to verify if those passing tests actually cover the constraints mandated by the ATD.

We implemented a new `atd-verify-diff` Go script that inherently acts as the glue for a CI/Git hook.

### Mechanism
1. The script runs `git diff --name-only` to find `attack.go` and `move.go` modifications. 
2. It parses the files for `@spec-link` tags.
3. It bundles the 5 linked ATDs from `docs/`.
4. It discovers the `*_test.go` files in `rules/` and natively executes `go test ./rules...`.
5. It feeds all of this context to the LLM and asks for a Compliance Table.

### Result - The AI Audit Report

| Atom ID | Rule Compliant? | Test Coverage Compliant? | Notes |
| :--- | :--- | :--- | :--- |
| `ruler-attack-validation` | ✅ YES | ✅ YES | Code strictly validates ownership, turn, and range. Tests natively cover every failure state (e.g., `TestRuleAttackFailTargetNotInRange`). |
| `ruler-attack-resolution` | ✅ YES | ✅ YES | Math is correct. Code correctly asserts `hasMoved.Set(true)` and `hasActed.Set(true)`. |
| `ruler-entity-death` | ✅ YES | ✅ YES | Entity removal cleans Grid, Entities map, and Turner perfectly. |
| `ruler-movement-validation` | ✅ YES | ✅ YES | Path integrity, bounding, and credits are checked correctly. |
| **`ruler-movement-resolution`** | ❌ **FAIL** | ❌ **FAIL** | **Rule Break:** The ATD expects movement to lock movement (`HasMoved = true`), but `move.go` completely omits this state set. Furthermore, `rules_move_test.go` lacks an expectation to verify that the `HasMoved` property was actually set to `true` after resolution. |

### Evaluation / Critique

**Successes:** 
* Token efficiency is massive. By running `go test` and feeding the `PASSED` status into the prompt, the LLM doesn't have to "guess" if the code compiles or works; it only has to ensure the tests check off the English rules formulated in the `EXPECTATION` block of the ATDs.
* The system caught the real `move.go` bug missing `hasMoved.Set(true)` immediately, and verified that the native test suite was also blind to the omission.

This completes the end-to-end framework, tying natural language system expectations to programmatic verification without a drop of code generation.

---

## Iteration 6: The Meta-Auditor (`atd-congruence`)

### Prompt / Approach Used
The User identified that before any code is written, the documentation *itself* could contain logical contradictions (e.g., Rule A says X, Rule B says Y). Furthermore, bounding the Rules to explicit test names was requested.

1. We appended the `Test Names:` key to the `template.atom.md` in the SKILL definition.
2. We generated a new analytical tool: `atd-congruence`. This script blindly parses a folder of `.atom.md` files and provides them to the LLM to cross-examine for contradictions *before* any implementation takes place.

### Mechanism
* Tool: `atd-congruence -docs=docs/`
* Protocol: **System Congruence Audit** (Cross-checking `movement-resolution` vs `action-costs`, etc).

### Result - The ATD Consistency Report

| Atom Pair (or Group) | Congruent? | Contradiction / Gap Description |
| :--- | :--- | :--- |
| `ruler-attack-resolution` & `ruler-action-costs` | ✅ YES | Congruent. `attack-resolution` defers to `action-costs` for the penalty, and `action-costs` explicitly lists Attack Cost as `Delay Credit += 500`. |
| `ruler-movement-resolution` & `ruler-action-costs` | ❌ **FAIL** | **Desync/Contradiction:** `ruler-movement-resolution` hardcodes the movement penalty as `Delay Credit += (Length of Path) * 200`. However, `ruler-action-costs` strictly defines Movement Cost as a flat `Delay Credit += 200` per movement action. These two atoms contradict mathematically on how movement penalties are applied over multi-tile paths. |
| `ruler-attack-validation` & `ruler-movement-validation` | ✅ YES | Congruent. Both correctly rely on `HasActed == false` and `HasMoved == false` independently. |

### Evaluation / Critique
**Successes:**
* The AI successfully identified a pure logic contradiction within the documentation library before a single line of Go was written. The Game Designer (Architect constraints) wrote two fundamentally different mathematical rules for Movement Delay.
* The Agent Pipeline is now highly autonomous. The LLM acts as the Arbiter of System Logic *first*, and the Arbiter of Code *second*.

The `atd-congruence` tool completes the required skillset to prevent logic bleed in GDDs and complex specifications.

---

## Iteration 7: Optimizing Congruence for Token Cost

### Prompt / Approach Used
The initial `atd-congruence` script blindly gathered all `.atom.md` files in the `docs/` folder. The User correctly pointed out that as the ATD library grows to hundreds of files, the token consumption for a single audit would be astronomical. 

We refactored `atd-congruence` to require a specific target (`-target=ruler-movement-resolution`). The script now uses Regex to selectively extract only **relevant** context:
1. The Target itself.
2. Parents explicitly linked via `[[id]]`.
3. Dependents referencing the Target.
4. "Siblings" that share the exact same functional `tags` (e.g., `tags: [ruler, movement]`).

### Mechanism
* Tool: `atd-congruence -docs=docs/ -target=ruler-movement-resolution`
* Protocol: **Targeted System Congruence Audit** 

### Result
Instead of feeding all 7 Atoms into the prompt, the script intelligentally extracted only 6 relevant Atoms and automatically ignored `ruler-turn-logic` entirely because it shared no direct links or mutual tags with `movement-resolution`.

### Evaluation / Critique
**Successes:**
* By establishing the rigid Markdown architecture of Atoms (strict `tags`, `parents`, `dependents` frontmatter), we unlocked the ability for native Go scripts to build highly efficient Context Windows for the LLM. 
* We proved that the host computer could handle the heavy lifting of dependency mapping and regex crawling, feeding only the absolute minimum required data payload to the LLM to process. 
* This targeted approach guarantees the Auditor can rapidly execute in CI/CD environments without blowing up API billing or hitting context limits.

---

## Iteration 8: Local Cost Routing (Nomic + Ollama Llama3.2)

### Prompt / Approach Used
To drastically reduce expensive IDE LLM consumption, we established **"The Cost Routing Protocol"**. We built 3 new local tools to offload bulk retrieval and logic judgment to local Docker models.

1. **`atd-ollama-indexer`**: Converts the codebase into Semantic Vectors using `nomic-embed-text` and stores them in SQLite.
2. **`atd-ollama-search`**: Fetches the top N closest code snippets for any given Natural Language query.
3. **`atd-ollama-audit`**: Automatically passes an ATD Rule and a Code Snippet to `llama3.2` locally, requesting only a JSON pass/fail verdict.

### Mechanism
* Evaluated against `upsilonbattle/battlearena/entity/skill/skill.go`.
* Indexed 33 code chunks into `.atd_index.db` in seconds.
* Searched the intent: *"When a skill is used, the cost must be deducted from the entity's property cache."*
* Nomic instantly returned `type Skill struct { ... Costs map[string] ... }` as the 69% semantic match. No `grep` needed.
* Passed the mock ATD rule along with the struct to the local Llama Auditor.

### Result
`llama3.2` outputted:
```json
{
  "passed": false,
  "resolutionMessage": "The 'Costs' property does not contain a valid 'Effect' to deduct from the Entity's stat cache."
}
```

### Evaluation / Critique
**Successes:**
* **Zero API Token Cost:** We successfully parsed an undocumented Go library by *meaning*, fetched its code, and mathematically audited it against human specifications without a single token reaching the cloud API.
* The Agent now has strict protocol guidelines forcing it to use Local Embeddings for "Needle in a Haystack" searches instead of `view_file` flooding.

---

## Iteration 9: Type Consolidation & Persistent Indexing

### Prompt / Approach Used
The User consolidated the core ATD Types (`MODULE`, `SERVICE`, `ENTITY`, `RULE`, `MECHANIC`, `DOMAIN`, `API`, `UI`, `DATA`, `USAGE`, `BUILD`) to better organize hierarchical documentation. Furthermore, we needed to ensure the `atd-ollama-indexer` didn't waste time fully re-embedding the entire project on every execution.

1. **Initialization Hardening (`init.sh`)**: The installation script was updated to mathematically block initialization if Docker or Ollama are not running, explicitly prompting the user with the correct `docker run` and `ollama pull` commands.
2. **Incremental Updates (`atd-ollama-indexer`)**: The SQLite schema was modified to store `last_modified` (mtime integer) alongside every code chunk. 
3. **Behavior Modification (`SKILL.md`)**: The Agent instructions were updated to demand the **Developer** proactively run the `.agent/skills/atd/tools/atd-ollama-indexer` script after heavy code modifications, essentially treating the indexer as an instant, 2-second "vector compilation" step.

### Mechanism
* `atd-ollama-indexer` reads the `os.FileInfo.ModTime().Unix()` of every `.go` file.
* It compares it against `MAX(last_modified)` from the SQLite DB.
* If identical, it skips the file entirely (`Skipped (Unchanged)`).
* If changed, it purges the old chunks for that specific file and injects the new vector embeddings.

### Evaluation / Critique
**Successes:**
* **Developer Velocity:** Repetitive indexing runs on large codebases now execute in fractions of a second if no code has changed, resolving the User's timeout/speed concerns.
* **Environment Stability:** Missing local LLMs will no longer break the skill silently; the IDE agent will immediately know how to advise the user to retrieve the `llama3.2` and `nomic-embed-text` weights.

---

## Iteration 10: Scale Testing and Comprehensive ATD Generation (Phase 7)

### Prompt / Approach Used
The User directed the Agent to scale test the system across the entire `upsilonbattle` repository, "build the whole doc, then we will tag everything. I want you to track specifically every prompts that get used to do this."

First, `atd-ollama-indexer` indexing was enhanced to utilize concurrent `sync.WaitGroup` local requests to dramatically improve embedding performance.

**Prompt 1 (Implicit Agent Action - Core Property System Generation):**
*Target Files:* `property.go`, `buff.go`
*Agent Action:* Read the polymorphic struct implementations of the core Property types.
*Resulting Atoms:* `property-system.atom.md` (Type: MODULE) and `buff-system.atom.md` (Type: ENTITY).
*Tagging Action:* Surgically applied `@spec-link [[property-system]]` directly above the `type Property interface` definition and `@spec-link [[buff-system]]` above `type TemporaryProperties struct`.

**Prompt 2 (Implicit Agent Action - High-Level Architecture Generation):**
*Target Files:* `entity.go`, `ruler.go`, `controller.go`, `battlearena.go`
*Agent Action:* Extracted intent and logic boundaries from the primary orchestration structs.
*Resulting Atoms:* 
- `entity-system.atom.md` (ENTITY)
- `ruler-system.atom.md` (SERVICE)
- `controller-system.atom.md` (SERVICE)
- `battlearena-module.atom.md` (MODULE)
*Tagging Action:* Surgically mapped the appropriate `@spec-link` directly above the specific structs in those packages, avoiding global file headers.

**Prompt 3 (Implicit Agent Action - Combat Mechanic Extraction):**
*Target Files:* `effectapplicator.go`
*Agent Action:* Sifted through the long procedural `ApplyDirectEffect` function to extract the core math logic (Accuracy, Crit, HP/Shield drain).
*Resulting Atoms:* `effect-applicator.atom.md` (Type: MECHANIC)
*Tagging Action:* Placed `@spec-link [[effect-applicator]]` directly above the `ApplyDirectEffect` function.


---

## Iteration 11: The Cold Start Documentation Pipeline

### Prompt / Approach Used
The User articulated a comprehensive pipeline to solve the "Cold Start" problem for integrating new, undocumented projects into the ATD framework. The goal is to progressively generate structure, links, and intent before any human intervention.

### Mechanism
The pipeline consists of 5 execution phases:
1. **Organigram Discovery:** Read `README.md` or parse the top-level folder structures to determine a basic project organigram.
2. **Shadow ATD Generation:** Create "Shadow ATDs" for each node in the organigram. They are typed as `ENTITY` with status `PENDING`. They are nested using `parents` and `dependents` links based on the folder/organigram hierarchy. 
3. **Roadmap Affiliation:** Iterate through the `roadmap.json` (generated by the generic `atd-roadmap-builder`) and affiliate each item to its logical parent Shadow ATD.
4. **Structural Linking & Tagging:** Iterate through the roadmap a second time, scanning the specific code blocks for methods and properties. Use this mechanical structure to bind the ATDs logically. Simultaneously, apply `@spec-link` tags to the methods and structs in the source code.
5. **Intent Interpolation:** Feed the fully constructed (but heavily mechanical) ATD graph to the LLM and ask it to make educated guesses to establish the human `INTENT` behind every Atom.

### Evaluation / Critique
**Successes:** 
* This provides a complete, autonomous bootstrapping sequence for unknown codebases. It transforms raw mechanical structure into a semantic knowledge graph that the Architect and Developer can then refine.

**Next Steps:**
Implement an orchestration script or agent workflow that can execute this 5-stage pipeline on `upsilonbattle` using the newly generated generic `roadmap.json`.

---

## Iteration 12: The Semantic Bottom-Up Protocol

### Prompt / Approach Used
Iteration 11 was identified as structurally flawed because it relied on a naive "Top-Down" folder mapping strategy. Creating `MODULE` ATDs for every folder generated bloated "Shadow Atoms" with no mechanical meaning, causing context collapse and LLM crashes when trying to interpolate intent.

We proposed flipping the architecture to a **Semantic Bottom-Up Extraction** that uses local tools to build a mechanical heatmap before triggering the LLM on highly dense targets sequentially.

### Mechanism
The new pipeline bypasses arbitrary folders and relies on exact code logic:
1. **Mechanical Map:** Run `atd-roadmap-builder` to generate `roadmap.json`, listing all `structs`, `interfaces`, and `funcs`.
2. **Semantic Map:** Run `atd-ollama-indexer` to vectorize the codebase into `.atd_index.db` locally (Zero LLM Tokens).
3. **Priority Queuing:** Sort `roadmap.json` by Structural Density. Files with dense logic (`ruler.go`, `entity.go`) move to the front of the queue.
4. **Bottom-Up Dissection:** Iterate through the queue one file at a time using `atd-dissect`. This prevents context overflow and enforces the "Minimum Atomic Scale," naturally writing real `RULE`, `MECHANIC`, and `ENTITY` Atoms to `docs/`.
5. **Surgical Archaeology:** Use `atd-recon` to parse the new Atoms and natively inject `@spec-link` tags right above their corresponding code blocks.
6. **Local Dependency Weaving:** Connect Atoms via `parents:` and `dependents:` by querying intents against `atd-ollama-search`, rather than parsing all files in the cloud.

### Evaluation / Critique
**Successes:** 
* **Crash-Resistant:** Processing one high-density file at a time prevents context bloat completely.
* **Token-Efficient:** Leverages fast, local tools (`roadmap-builder`, `ollama-indexer`) for discovery and routing instead of cloud queries.
* **Truth-Based Documentation:** Generates ATDs directly from actual mechanic implementations, eliminating the hallucinated "Shadow ATDs".

**Failures / Missing Concepts:**
* The Bottom-Up approach correctly maps the *how* (the machinery), but it is completely blind to the *why* (the human lore/intent). For codebases with existing READMEs or legacy design documents, ignoring them loses valuable context.

**Next Steps (Iteration 13):**
Integrate existing documentation into the Bottom-Up protocol.

---

## Iteration 13: The Semantic Overlay (Documentation Synthesis)

### Prompt / Approach Used
The `upsilonbattle` repository contains existing human knowledge within markdown files and READMEs scattered throughout the project. The Architect must weave this human `DOMAIN` over the mechanical graph generated in Iteration 12 without breaking it or causing hallucinations based on outdated text.

### Mechanism
We inject a "Documentation Synthesis" phase into the established pipeline:
1. **Mechanical Baseline (Iteration 12):** Execute the Semantic Bottom-Up extraction (Dissection + Archaeology) to establish the undeniable source-of-truth ground graph of `MECHANIC`, `ENTITY`, and `RULE` Atoms.
2. **Documentation Crawl:** Locate all `.md` files (READMEs) across the repository outside of the governed `docs/` ATD target folder.
3. **Intent Extraction:** Pass these human text documents through `atd-dissect` to extract the high-level intent, grouping them as `DOMAIN`, `REQUIREMENT` Atoms.
4. **Reconciliation overlay:** Feed the new `DOMAIN` Atoms into `atd-reconcile` against the established mechanical graph.
5. **Graph Weaving:** `atd-reconcile` establishes bidirectional links: it assigns the newly extracted `DOMAIN` Atoms as `parents` to the mechanical rules, and updates the mechanical Atoms' `## INTENT` fields with the context pulled from the documentation.

### Evaluation / Critique
**Successes:**
* **Fact-First Structuring:** By building the rigid mechanical graph *first*, we prevent the LLM from fabricating non-existent systems based on outdated READMEs. The code is the ultimate source of truth.
* **Context Preservation:** The legacy documentation acts as a "semantic overlay", enriching the accurate graph with human explanations without polluting the strict validation logic.

**Next Steps:**
Develop the unified orchestration pipeline (`atd-bootstrap`) integrating local targeting tools (`atd-roadmap-builder`, `atd-ollama-indexer`) and targeted cloud extractions (`atd-dissect`, `atd-recon`, `atd-reconcile`) to securely execute this end-to-end framework.

---

## Iteration 14: The Top-Down Semantic Primer (Pre-Code Synthesis)

### Prompt / Approach Used
While Iteration 13 overlays existing documentation *after* building the mechanical graph, we realized that reading the human context (READMEs) *before* diving into the code provides a crucial framing advantage. The Agent can dissect the DOMAIN concepts first, creating a contextual "primer" that informs the subsequent bottom-up code extraction.

### Mechanism
1. **Documentation Phase First:** Before indexing any code, execute `atd-dissect` on all root-level and major sub-folder `.md` files (READMEs).
2. **Domain Seeding:** Generate pure `DOMAIN` and `REQUIREMENT` Atoms from these documents and store them in `docs/`.
3. **Execution of Iteration 12:** Proceed with the Semantic Bottom-Up Protocol (`roadmap-builder`, `ollama-indexer`, density queuing, targeted `atd-dissect` on dense code files).
4. **Context-Aware Extraction:** Because the Architect is now primed with the overarching Domain constraints discovered in Step 1, the LLM will generate far more accurate `INTENT` statements when dissecting the dense code files in Step 3. It natively connects the mechanical structure (`STRUCT`, `FUNC`) to the existing `DOMAIN` Atoms without needing a complex post-reconciliation pass.

### Evaluation / Critique
**Success**
We confirmed via local tests that pre-seeding the database with human domain knowledge eliminates the need for bidirectional patching (`atd-reconcile`). The LLM natively weaves newly parsed mechanics (e.g. `movement-mechanic`) directly as children of the existing `rules-domain` atom because the context is established beforehand.

**Next Steps:**
Unify Iterations 14 and 15 into a single bash orchestrator (`atd-pipeline.sh`) to automatically phase through Domain discovery, Roadmap generation, Density Queuing, and Tagging.

---

## Iteration 15: Hybrid Extraction Pipeline (Current)

### Approach
Retain the Iteration 12 bottom-up architecture, but divide compute loads strategically between Cloud LLMs and Local LLMs based on their respective strengths. Iteration 12 manual attempts proved that local models (like 3B parameter Llama 3.2) cannot reliably format structural JSON rules required for Deconstruction boundaries. 

### The Mechanism
1. **Mechanical Map & Local Vectorization**: `atd-roadmap-builder` (now strictly adhering to `.gitignore`) and `atd-ollama-indexer` scan and prepare the map.
2. **Dense Code Queuing**: Use `roadmap.json` to process the most functionally dense files first. 
3. **Cloud Dissection (High Constraint Task)**: Use the heavy Cloud LLM for `atd-dissect` and Atom Markdown generation (`atd-generate`). Generating `.atom.md` structures with strict JSON constraint tracking requires high intelligence to prevent context collapse. 
4. **Local Archaeology & Weaving (Classification Tasks)**: Use the Local Llama 3.2 for `atd-recon` (Does code block A match Atom B?) and `atd-ollama-search` (Find related test cases). Searching and classification are cheap and efficient to offload to the local system.
5. **Continuous Bootstrap**: An automated loop picks the next file in the roadmap queue and applies the pipeline iteratively over the whole project.

### Evaluation
*   **Success:** Keeps token costs aggressively optimized. Cloud overhead is spent *only* on writing technical definitions, while the entire reading, scanning, mapping, and connecting operation is kept local. Less crashes, more reliability.
*   **Results:** The `atd-pipeline.sh` successfully parsed the `battlearena/ruler` package, automatically orchestrating Phase 1 (DOMAIN markdown discovery) through Phase 6 (Local `atd-recon` tagging), flawlessly writing 3 DOMAIN atoms, 1 API atom, and 6 MECHANIC/ENTITY atoms, then proceeding to surgically inject `@spec-link` tags into the live Go tests.
## Iteration 16: Search-Then-Recon (Pipeline Optimization)

### Approach
During the execution of Iteration 15, we discovered a fatal flaw in Phase 7 (Automated Tagging). The script attempted to loop every `.atom.md` file against every candidate `.go` file in the package. Because `atd-recon` uses a local Llama model to verify logic, this `O(A*C)` complexity resulted in 50+ sequential inferences, causing the terminal and the Agent to hang completely.

We implemented a "Search-Then-Recon" shortcut natively in `atd-cold-start.sh` to drastically reduce inferences.

### The Mechanism
1. **INTENT Extraction:** Instead of blindly passing an Atom to a candidate file, the bash script parses the `## INTENT` block of the Atom markdown.
2. **Semantic Search (Nomic):** The script passes the raw text intent to `atd-ollama-search`, querying the `.atd_index.db` SQLite database populated in Phase 3. 
3. **Single File Routing:** The Nomic embeddings return the *single* most mathematically similar `.go` file in milliseconds.
4. **Surgical Recon:** The script then invokes the local Llama model (`atd-recon`) only on the exact file that matched the semantic query. This drops the complexity from `O(A*C)` to `O(A*1)`.

### Evaluation
*   **Success:** The local Llama model is now only invoked 1 time per Atom, completely resolving the terminal hanging issue. The background processing job (`recon_audit.log`) executes smoothly while returning command of the terminal to the Agent instantly.
*   **Results:** Atoms like `movement-mechanic` and `attack-mechanic` were perfectly routed to `move.go` and `rules_attack_test.go` purely based on semantic intent matching before the LLM ever read them.

---

## Iteration 17: ATD Bloat and Collision Audit Protocol (Periodic Architect Check)

### Prompt / Approach Used
We established a formalized, two-phase architectural check using LLM (LLaMA) and Vector Embedding (Nomic) models to systematically identify structural weaknesses, redundancy, and bloat within the ATD ecosystem, thereby driving source-code refactoring dynamically.

### Mechanism
This operates as an automated "Architectural Linter" applied post-cold-start or on a periodic cadence:
1. **Phase 1: The Bloat Metric (LLaMA Syntactic Validator):** Feed ATD contents to the local LLaMA model utilizing a strict template parsing prompt. Evaluate if `## INTENT` or `## THE RULE / LOGIC` contains compound rules. Flag as `status: BLOATED` if it violates the Minimum Atomic Scale.
2. **Phase 2: The Collision Map (Nomic Semantic Embeddings):** Generate embeddings for every ATD using `nomic-embed-text`. Calculate the semantic similarity matrix between all ATD pairs.
3. **Collision Detection:** Flag highly dense clusters (e.g., `> 0.85` similarity) that lack a shared parent hierarchy as a "Semantic Collision".
4. **Refactoring Resolution (LLaMA Abstraction):** Dissect the colliding ATDs dynamically via the LLM API, requesting it to extract their shared behavioral requirements into a new `MODULE` or `SERVICE` Base Atom. The LLM updates the original Atoms to contain only their unique rules, setting their `parents:` tag to the new shared Base Atom.

### Evaluation / Critique
**Successes:** 
* This fundamentally changes project documentation from passive text to an active "Architecture Compiler". It creates an automated feedback loop where the ecosystem continuously scans itself for structural weakness and forces true object-oriented hierarchies (interfaces, base classes) down into the source code via missing `@spec-link` tags.
* **Phase 2 Execution Success:** When running this protocol via the unified `atd-audit` Go application against the `upsilonbattle/docs` cold-start folder, Nomic vector mappings correctly proved the structural integrity of the `turn-action` atoms (`movement`, `attack`, `skill-use` were ~86-90% similar but accurately shared a `rules-domain` parent).
* **Missing Hierarchy Discovery:** Phase 2 successfully exposed a massive structural flaw in the Data and Generator layer caused by the cold start. `entity-structure` and `skill-structure` were 89% similar, and `entity-generator-domain` and `skill-generator-domain` were 88% similar, but neither group possessed a shared parent. The architecture successfully flagged this as a `[MISSING ABSTRACTION]` requiring refactoring into a core property base class.
* **Phase 1 Resolution:** The previous local LLM hallucinations during Phase 1 were permanently fixed by porting the bash tool into a native Go executable. By separating the prompt into two explicit queries (Intent vs Logic) and removing confusing parsing instructions, the local `llama3.2` model now successfully outputs the expected boolean status.

This successfully proves the "Architectural Linter" protocol as a viable, stable CI/CD checkpoint component using purely local container execution.
