# Task 11 — Remaining LLM Commands

**Depends on:** Task 04 (provider), Task 05 (prompts), Task 06 (continue), Task 03 (atom, cosine)  
**Produces:** `atd congruence`, `atd reconcile`, `atd recon`, `atd discover`

## Context

These are "Category B" tools that currently emit prompts to stdout. With the tiered provider, they gain the ability to also call Ollama directly when available, and fall back to the IDE Agent pipeline when not.

## Steps

### 11.1 Create `cmd/congruence.go`

**Flags:** `-target <atom_id>`  
**Source:** `atd-congruence/main.go` (143 lines)

**Execution flow:**
```
Load all atoms from config.DocsDir()
Find target atom by ID
Regex-extract related atoms:
  - Parents (linked via [[id]])
  - Dependents (who reference target)
  - Tag siblings (shared tags)
Build congruence prompt with only related atoms
├── Resolve provider for "congruence"
│   ├── Ollama → POST, get markdown table
│   └── IDE → Write prompt + task_list
└── Output
```

**Help text:**
```
Audit logical consistency between a target atom and its related atoms.

Checks parents, dependents, and tag-siblings for contradictions
in their INTENT and LOGIC sections.

Example:
  atd congruence -target ruler-movement-resolution
```

### 11.2 Create `cmd/reconcile.go`

**Flags:** `-new <path>`, `-store <path>`  
**Source:** `atd-reconcile/main.go` (64 lines)

**Execution flow:**
```
Read -new file (inbound edits / new documentation)
Read -store file (existing atom store)
Build reconcile prompt
├── Resolve provider for "reconcile"
│   ├── Ollama → POST with JSON format [{proposed_id, relationship, change_context}]
│   └── IDE → Write prompt + task_list
└── Output
```

### 11.3 Create `cmd/recon.go`

**Flags:** `-atom <path>`, `-candidate <code_path>`  
**Source:** `atd-recon/main.go` (65 lines)

**Execution flow:**
```
Read atom file → extract INTENT + LOGIC
Read candidate code file
Build recon prompt: "Does this code implement this atom?"
├── Resolve provider for "recon"
│   ├── Ollama → POST with JSON format {Confidence: int, Mismatches: str}
│   └── IDE → Write prompt + task_list
└── Output
```

### 11.4 Create `cmd/discover.go`

**Flags:** `-file <code_path>`  
**Source:** `atd-discover-links/main.go` (229 lines)

**Execution flow:**
```
Read code file
│
Step 1: Extract code intent
├── Resolve provider for "intent_extract"
│   ├── Ollama → POST, get 2-sentence summary
│   └── IDE → Write intent prompt + task_list (pause here)
│
Step 2: Embed intent
├── Resolve provider for "embed" → Nomic
│   └── If unavailable → error (no IDE fallback for embed)
│
Step 3: Query atom docs index
├── Open docs DB (same as atd audit DB)
├── Cosine rank against stored atom embeddings
├── Take top 3 matches
│
Step 4: Build recommendation prompt
├── "Which of these atoms define this code?"
├── Resolve provider for "intent_extract" (reuse)
│   ├── Ollama → POST, get recommended IDs
│   └── IDE → Write prompt + task_list
└── Output: recommended atom links
```

**Note:** `discover` has two LLM touchpoints (intent extraction + recommendation). If either falls back to IDE, the task_list should contain both steps in order. `atd continue` will handle multi-step resumption.

### 11.5 Add `--snapshot --llm` to `atd assemble`

Update `cmd/assemble.go` (created in Task 07) to handle `--snapshot` with LLM:

```go
if snapshot {
    // Build snapshot prompt from prompt.SnapshotBuild(assembled, theme)
    res, err := ollama.Query("snapshot", snapshotPrompt, nil)
    if errors.Is(err, ollama.ErrIDEFallback) {
        pipeline.WritePromptFile("snapshot_"+startIDs[0], snapshotPrompt)
        pipeline.WriteTaskList("atd assemble --snapshot", ...)
    } else {
        fmt.Println(res.Response)
    }
}
```

### 11.6 Write tests

**Per command:**
- Test flag validation (missing required flags → error)
- Test prompt construction contains expected content
- Test IDE fallback creates files in pipeline_output/

**`cmd/discover_test.go` (extra):**
- Test multi-step task_list generation contains both intent + recommendation tasks

## Acceptance Criteria

- [ ] `atd congruence -target X` with valid atoms produces congruence prompt
- [ ] `atd reconcile -new A -store B` builds reconcile prompt
- [ ] `atd recon -atom A -candidate B` builds recon prompt
- [ ] `atd discover -file <code>` (with Nomic) shows top 3 atom matches
- [ ] `atd assemble --snapshot --starts X` (fallback) creates pipeline_output
- [ ] All 4 commands show in `atd --help`
- [ ] Tests pass
