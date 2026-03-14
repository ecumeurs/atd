# Task 14 — Cold Start Pipeline: upsilonbattle

## Objective
Run the full ATD cold start pipeline on `upsilonbattle/` to auto-generate the initial ATD framework from scratch, without any manually pre-created atoms.

> ⚠️ **Token-heavy**: This task involves multiple LLM calls (dissect, recon, intent extraction). Run after all previous tests are verified.

## Prerequisites
- Task 01 complete (`.atd` exists)
- Ollama available with `qwen2.5-coder:14b`, `llama3.2`, and `nomic-embed-text`
- Clean slate: remove manually created atoms from tasks 02-06 (or use a fresh clone)

## Steps

### Option A: Full script pipeline (recommended)
```bash
cd /home/bastien/work/skill
bash scripts/atd-cold-start.sh upsilonbattle/
```

This orchestrates:
1. **Roadmap**: `atd roadmap` → identifies high-density files
2. **Index**: `atd index --mode code` → vectorizes all source files
3. **Domain extraction**: IDE Agent reads existing `.md` files → writes DOMAIN atoms
4. **Dissect**: `atd dissect --llm` on high-density files → proposes atoms
5. **Weave**: `atd weave` → connects parent/dependent links
6. **Recon**: Search-then-recon → tags matching source files with `@spec-link`

### Option B: Step-by-step manual
```bash
cd /home/bastien/work/skill/upsilonbattle

# Step 1: Roadmap
atd roadmap --dir battlearena/ --out pipeline_output/roadmap.json

# Step 2: Index code
atd index --dir battlearena/ --mode code

# Step 3: Dissect top files from roadmap
# Read roadmap.json, pick top 5 files, dissect each
atd dissect --file battlearena/ruler/ruler.go --llm
atd dissect --file battlearena/entity/entity.go --llm
atd dissect --file battlearena/property/property.go --llm

# Step 4: Weave
atd weave

# Step 5: Search-then-recon for each atom
atd discover --file battlearena/ruler/rulermethods/rulermethods.go
```

## Expected Output
- `pipeline_output/roadmap.json` — complexity ranking
- Multiple `.atom.md` files in `docs/`
- `@spec-link` tags in top source files
- `atd crawl` shows no STABLE atoms without implementations

## Acceptance Criteria
- [ ] Pipeline completes (script or manual) without fatal errors
- [ ] At least 10 atoms generated in `docs/`
- [ ] `atd weave` runs after atom creation
- [ ] At least one source file has `@spec-link` injected
- [ ] `atd crawl --gaps` shows fewer unimplemented STABLE atoms than before
