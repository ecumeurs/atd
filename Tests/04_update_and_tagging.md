# Task 04 — Update & Tagging: spec-link injection

## Objective
Update atom metadata fields using `atd update`, then use `atd discover` to recommend and apply `@spec-link` tags to upsilonbattle source files.

## Prerequisites
- Task 03 complete (atoms exist in `upsilonbattle/docs/`)
- Ollama available for `atd discover` (uses embed + intent_extract) — or IDE Agent fallback

## Steps

### 1. Update an atom status to STABLE
```bash
cd /home/bastien/work/skill/upsilonbattle
atd update --file docs/ruler.atom.md --set "status=STABLE"
```

### 2. Update the logic section of an atom
```bash
atd update --file docs/ruler_attack.atom.md \
  --logic "Damage = attacker.Attack - defender.Defense (minimum 0). Range check: attacker must be within [min_range, max_range] cells of defender. Attack counter decremented per use."
```

### 3. Discover @spec-link tags for ruler.go
```bash
atd discover --file battlearena/ruler/ruler.go
```
> IDE Agent (or local Ollama) extracts intent, searches the docs index, recommends atom IDs.

### 4. Discover @spec-link tags for entity.go
```bash
atd discover --file battlearena/entity/entity.go
```

### 5. Inject @spec-link into ruler.go
Based on discover recommendations, inject tags above the relevant functions:
```bash
atd update --file docs/ruler.atom.md \
  --spec-link "ruler" \
  --spec-link-file battlearena/ruler/ruler.go
```

### 6. Verify tags were injected
```bash
grep "@spec-link" battlearena/ruler/ruler.go
```

## Expected Output
- `ruler.atom.md` status updated to STABLE
- `atd discover` returns JSON with recommended atom IDs (e.g., `["ruler", "ruler_turn_flow", "ruler_attack"]`)
- `ruler.go` contains at least one `// @spec-link [[ruler]]` comment

## Acceptance Criteria
- [ ] `atd update --set status=STABLE` runs without error
- [ ] `atd update --logic` modifies only the logic section, leaving other sections intact
- [ ] `atd discover` returns at least one recommendation
- [ ] At least one `@spec-link` tag appears in `ruler.go` or `entity.go`
