# Task 10 — Audit, Fix & Compare

**Depends on:** Task 03 (atom, cosine), Task 04 (provider), Task 05 (prompts for audit_bloat, audit_code, fix_split, compare)  
**Produces:** `atd audit`, `atd fix`, `atd compare` subcommands

## Context

These are the most complex tools. `atd-audit` (574 lines) is the largest single tool, doing both LLM-powered bloat detection and Nomic-powered collision mapping. `atd-audit-fixer` (322 lines) reads audit reports and uses LLM to split bloated atoms. `atd-compare` (316 lines) resolves collisions pairwise. `atd-ollama-audit` (118 lines) checks code against rules — merges into `atd audit --code`.

## Steps

### 10.1 Create `cmd/audit.go`

**Flags:** `--threshold <float>` (override similarity threshold), `--code <path> --atom <path>` (code compliance mode), `--docs <override>`

**Default mode (Bloat + Collision):**

```
Phase 1: Bloat Detection (per atom)
├── Walk docs/*.atom.md
├── For each atom:
│   ├── Parse with atom.Parse()
│   ├── Check mtime against SQLite cache
│   ├── If stale:
│   │   ├── Resolve provider for "audit_bloat"
│   │   │   ├── Ollama → two-query prompt (Intent + Logic separately)
│   │   │   │   └── Parse YES/NO response
│   │   │   └── IDE fallback →
│   │   │       ├── Write prompt to pipeline_output/audit_<atom>.prompt
│   │   │       └── Append to task_list
│   │   ├── Resolve provider for "embed" → embed intent text
│   │   └── Store {status, embedding} in SQLite
│   └── Output: "Auditing: file.atom.md ... [OK] / [BLOATED]"
│
Phase 2: Collision Detection (pairwise)
├── For each pair of atoms:
│   ├── Check collision cache
│   ├── Cosine similarity > threshold?
│   │   ├── YES → ancestry BFS walk
│   │   │   ├── Shared parent found → "[SOUND] Shared Parent: X"
│   │   │   └── No shared parent → "[COLLISION] a <--> b [MISSING ABSTRACTION]"
│   │   └── NO → skip
│   └── Cache result
└── Output: full audit report to stdout
```

**`--code` mode (absorbs atd-ollama-audit):**

```
Read atom file → Parse rule/logic
Read code snippet
├── Resolve provider for "audit_code"
│   ├── Ollama → POST with JSON format {passed, resolutionMessage}
│   └── IDE → Write prompt + task_list
└── Output: JSON {passed, resolutionMessage}
```

**Source references:**
- `atd-audit/main.go` (full 574 lines) — primary source
- `atd-ollama-audit/main.go` (118 lines) — for `--code` mode

### 10.2 Create `cmd/fix.go`

**Flags:** `--audit <report.txt>`, `--dry-run`

**Execution flow:**

```
Parse [BLOATED] lines from audit report
For each bloated atom:
├── Read full atom content
├── Read metadata via atom.ParseMeta()
├── Resolve provider for "fix_split"
│   ├── Ollama → POST with JSON format {parent_logic, splits[]}
│   │   ├── Parse JSON response
│   │   ├── Write child .atom.md files
│   │   └── Rewrite original as MODULE parent
│   └── IDE → Write prompt + task_list
├── Invalidate DB cache for modified files
└── Report: "✓ Rewrote X as MODULE parent, Y children created"
```

**Source reference:** `atd-audit-fixer/main.go` (322 lines)

### 10.3 Create `cmd/compare.go`

**Flags:** `-a <atom_a>`, `-b <atom_b>`, `--out <report.md>`

**Execution flow:**

```
Parse both atoms with atom.Parse()
Extract shared keywords (deterministic set intersection)
Build comparison prompt:
  <System_Context>: You are an ATD collision resolver
  <Atom A>: {intent, logic, tags}
  <Atom B>: {intent, logic, tags}
  <Shared Keywords>: [...]
├── Resolve provider for "compare"
│   ├── Ollama → POST, get plaintext resolution
│   └── IDE → Write prompt + task_list
Build markdown report with header + diagnosis
└── Output: stdout or --out file
```

**Source reference:** `atd-compare/main.go` (316 lines)

### 10.4 Port SQLite schema

The audit SQLite DB schema from `atd-audit/main.go`:

```sql
CREATE TABLE IF NOT EXISTS atom_docs_index (
    id TEXT PRIMARY KEY,
    file_path TEXT,
    intent_text TEXT,
    embedding BLOB,
    status TEXT,
    last_modified INTEGER
);
CREATE TABLE IF NOT EXISTS collision_cache (
    file_a TEXT,
    file_b TEXT,
    similarity REAL,
    result TEXT,
    last_checked INTEGER,
    PRIMARY KEY (file_a, file_b)
);
```

### 10.5 Write tests

**`cmd/audit_test.go`:**
- Test: default mode runs against a temp docs dir with 2 sample atoms (mock Ollama)
- Test: `--code` mode constructs correct prompt with rule + code
- Test: audit report format has expected `[OK]` / `[BLOATED]` markers

**`cmd/fix_test.go`:**
- Test: `--dry-run` prints proposed changes without writing
- Test: parse of audit report extracts [BLOATED] filenames correctly

**`cmd/compare_test.go`:**
- Test: determines shared keywords between two atom files
- Test: report markdown has expected headers

## Acceptance Criteria

- [ ] `atd audit` runs and produces report with [OK]/[BLOATED]/[COLLISION] markers
- [ ] `atd audit --code <snippet> --atom <atom>` outputs JSON pass/fail
- [ ] `atd fix --audit <report> --dry-run` shows proposed splits without writing
- [ ] `atd compare -a X -b Y` produces markdown report
- [ ] SQLite caching works (second audit run faster)
- [ ] Tests pass
