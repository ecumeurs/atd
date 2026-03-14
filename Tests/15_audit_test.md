# Task 15 — Audit: Bloat Detection + Congruence

## Objective
Run `atd audit` to check all existing ATD atoms for documentation bloat and semantic collisions, then verify code compliance for key atoms.

> ⚠️ **Token-heavy**: Audit uses LLM for every atom. Run after tasks 02-06 or 14.

## Prerequisites
- At least 7 atoms exist in `upsilonbattle/docs/` (tasks 03-06 or task 14)
- Ollama available with `llama3.2` and `nomic-embed-text`

## Steps

### 1. Run full audit (bloat + collision detection)
```bash
cd /home/bastien/work/skill/upsilonbattle
atd audit
```
This runs:
- **Phase 1 (Bloat)**: Each atom's INTENT + LOGIC sections are evaluated by local LLM for compound-rule violations
- **Phase 2 (Collision)**: Cosine similarity between atom embeddings identifies semantic overlaps

### 2. Run audit with lower threshold (more sensitive)
```bash
atd audit --threshold 0.75
```

### 3. Code compliance check for a specific atom
```bash
atd audit \
  --atom docs/ruler_attack.atom.md \
  --code battlearena/ruler/ruler.go
```
**Expected:** `{"passed": true, "resolutionMessage": "..."}` or `{"passed": false, ...}` with explanation.

### 4. Run congruence check
```bash
atd congruence
```
> Validates that logically adjacent ATD rules don't contradict each other.

### 5. Review audit results
```bash
cat docs/.atd_audit.db  # Contains cached audit results
atd query --field status --search DRAFT  # Atoms that may need review
```

## Expected Output
- `=== ATD AUDIT PROTOCOL INITIATED ===` header
- Per-atom result: `PASS`, `BLOATED`, or `PENDING_IDE`
- Collision report: `[COLLISION]` warnings if semantic overlap detected
- Code compliance returns a `passed` boolean

## Acceptance Criteria
- [ ] Audit runs without fatal errors
- [ ] At least one atom receives a `PASS` result
- [ ] If any atom is `BLOATED`, the message explains which section violated atomicity
- [ ] Code audit for `ruler_attack` + `ruler.go` returns a result (pass or fail)
- [ ] No unhandled panics or error exits
