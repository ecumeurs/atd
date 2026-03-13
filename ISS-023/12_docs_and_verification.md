# Task 12 — Documentation & Verification

**Depends on:** All previous tasks  
**Produces:** ATD docs for each subcommand, verified build, smoke + integration tests

## Context

Each subcommand needs its own ATD atom in `docs/`. The unified binary needs end-to-end verification.

## Steps

### 12.1 Create `docs/` directory

```bash
mkdir -p docs/
```

### 12.2 Write per-tool ATD atoms

Create one `.atom.md` file per subcommand, plus supporting atoms. Use this template:

```markdown
---
id: atd_<subcommand>
human_name: ATD <Subcommand Name>
type: <TYPE>
version: 1.0
status: STABLE
priority: CORE
tags: [atd, cli, <category>]
parents:
  - [[atd_cli]]
dependents: []
---

# ATD <Subcommand Name>

## INTENT
<One sentence: WHY this tool exists and WHAT single responsibility it fulfills.>

## THE RULE / LOGIC
<The single behavioral rule this tool enforces. What it does, when it runs, what it produces.>

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd <subcommand> [flags]`
- **LLM Task:** `<task_type>` (or "None" for deterministic tools)
- **Code Tag:** `@spec-link [[atd_<subcommand>]]`
```

**Files to create (23 atoms):**

| File | Type | Parent |
|------|------|--------|
| `atd_cli.atom.md` | MODULE | (root) |
| `atd_config.atom.md` | SPECIFICATION | `atd_cli` |
| `atd_tiered_provider.atom.md` | SERVICE | `atd_config` |
| `atd_dissect.atom.md` | MECHANIC | `atd_cli` |
| `atd_generate.atom.md` | MECHANIC | `atd_cli` |
| `atd_audit.atom.md` | SERVICE | `atd_cli` |
| `atd_fix.atom.md` | MECHANIC | `atd_audit` |
| `atd_compare.atom.md` | MECHANIC | `atd_audit` |
| `atd_index.atom.md` | SERVICE | `atd_cli` |
| `atd_search.atom.md` | SERVICE | `atd_index` |
| `atd_update.atom.md` | MECHANIC | `atd_cli` |
| `atd_congruence.atom.md` | MECHANIC | `atd_cli` |
| `atd_reconcile.atom.md` | MECHANIC | `atd_cli` |
| `atd_recon.atom.md` | MECHANIC | `atd_cli` |
| `atd_crawl.atom.md` | SERVICE | `atd_cli` |
| `atd_assemble.atom.md` | MECHANIC | `atd_cli` |
| `atd_query.atom.md` | SERVICE | `atd_cli` |
| `atd_weave.atom.md` | MECHANIC | `atd_cli` |
| `atd_verify.atom.md` | SERVICE | `atd_cli` |
| `atd_roadmap.atom.md` | SERVICE | `atd_cli` |
| `atd_discover.atom.md` | SERVICE | `atd_cli` |
| `atd_test_links.atom.md` | SPECIFICATION | `atd_cli` |
| `atd_continue.atom.md` | MECHANIC | `atd_cli` |

### 12.3 Run `atd weave` to link dependents

After creating all atoms:
```bash
atd weave
```
This will auto-populate `dependents:` fields based on `parents:`.

### 12.4 Build verification

```bash
cd /home/bastien/work/skill/scripts
go build -o bin/atd ./cmd/atd/
```
Must compile with zero errors.

### 12.5 Smoke tests

Run each command with `--help`:
```bash
for cmd in dissect generate audit fix compare index search update congruence reconcile recon crawl assemble query weave verify roadmap discover test-links continue; do
    echo "--- $cmd ---"
    ./bin/atd $cmd --help
done
```
Every command must print help text.

### 12.6 Functional tests

```bash
# Deterministic tools (no Ollama):
./bin/atd dissect -file ../upsilonbattle/battlearena/ruler/ruler.go
./bin/atd crawl --src ../upsilonbattle/battlearena/ --gaps
./bin/atd query -search "movement"
./bin/atd roadmap --dir ../upsilonbattle/ --out /tmp/roadmap.json

# Integration (requires nomic-embed-text):
./bin/atd index --dir ../upsilonbattle/ --db /tmp/test.db
./bin/atd search --query "entity movement" --db /tmp/test.db

# IDE fallback (no Ollama):
./bin/atd dissect -file ../upsilonbattle/battlearena/ruler/ruler.go --llm
# Should create pipeline_output/task_list.md
cat pipeline_output/task_list.md
```

### 12.7 Run unit tests

```bash
cd /home/bastien/work/skill/scripts
go test ./internal/atom/
go test ./internal/cosine/
go test ./internal/prompt/
go test ./internal/pipeline/
go test ./internal/ollama/
go test ./cmd/atd/cmd/
```

### 12.8 Update shell scripts

Modify `atd-cold-start.sh` and `atd-full-audit.sh` to use the unified binary:
- Replace `"$BIN_DIR/atd-dissect"` → `"$BIN_DIR/atd" dissect`
- Replace `"$BIN_DIR/atd-audit"` → `"$BIN_DIR/atd" audit`
- etc.

### 12.9 Clean up old modules

Once all tests pass, remove old individual tool directories from `go.work`:
```
go 1.24.4
use (
    .
    ./cmd/atd
)
```
The old `scripts/atd-*/` directories can be kept for reference or deleted.

## Acceptance Criteria

- [ ] 23 ATD atom files in `docs/`
- [ ] `atd weave` successfully links all dependents
- [ ] Binary compiles cleanly
- [ ] All `--help` commands produce output
- [ ] Dissect works against upsilonbattle
- [ ] Index + search work (with nomic)
- [ ] IDE fallback creates task_list.md
- [ ] All unit tests pass
- [ ] Shell scripts updated
