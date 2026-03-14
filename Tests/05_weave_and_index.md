# Task 05 — Weave & Index

## Objective
Populate bi-directional parent/dependent links between atoms (`atd weave`) and build a semantic vector index of both source code and ATD docs (`atd index`).

## Prerequisites
- Task 04 complete (atoms exist with parent references)
- Ollama with `nomic-embed-text` available (required for `atd index`)

## Steps

### 1. Confirm parent relationships exist in atoms
```bash
cd /home/bastien/work/skill/upsilonbattle
grep -r "parents:" docs/
```
> At least some atoms should reference parents (e.g., mechanic atoms under ruler).

### 2. Run atd weave to populate dependents[]
```bash
atd weave
```

### 3. Verify dependents were populated
```bash
grep -r "dependents:" docs/ | grep -v "\[\]"
```
> Parent atoms (e.g., `ruler.atom.md`) should now have `dependents` properly populated.

### 4. Index source code
```bash
atd index --dir battlearena/ --mode code
```

### 5. Index ATD docs
```bash
atd index --dir . --mode docs
```

### 6. Verify index file was created
```bash
ls -la docs/.atd_index.db
```

## Expected Output
- `atd weave` outputs a summary like `Weaved N links across M atoms`
- Parent atoms have non-empty `dependents:` arrays after weave
- `.atd_index.db` exists in `docs/` after indexing
- Index run reports chunks indexed (e.g., `Indexed 45 chunks across 8 files`)

## Acceptance Criteria
- [ ] `atd weave` completes without error
- [ ] At least one atom has non-empty `dependents:` after weave
- [ ] `atd index --mode code` completes and reports >0 chunks
- [ ] `atd index --mode docs` completes and reports >0 chunks
- [ ] `docs/.atd_index.db` exists and is non-zero bytes
