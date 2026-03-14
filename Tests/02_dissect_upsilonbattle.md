# Task 02 — Dissect upsilonbattle: ruler + entity

## Objective
Use `atd dissect` to extract atomic boundaries from legacy documentation in `upsilonbattle/`. Verify the tool produces usable proposals for atom creation.

## Prerequisites
- Task 01 complete (`.atd` exists in `upsilonbattle/`)
- Ollama available **or** using IDE Agent passthrough (no `--llm` flag)

## Steps

### 1. Dissect the ruler README (IDE Agent passthrough mode)
```bash
cd /home/bastien/work/skill/upsilonbattle
atd dissect --file battlearena/ruler/README.md
```
> This prints the dissect prompt to stdout. The IDE Agent reads it and proposes atom boundaries.

### 2. Dissect with local Ollama (if available)
```bash
atd dissect --file battlearena/ruler/README.md --llm
```
> Routes through `qwen2.5-coder:14b` → `llama3.2` fallback → IDE Agent fallback (writes to `pipeline_output/`)

### 3. Dissect the entity README
```bash
atd dissect --file battlearena/entity/README.md
```

### 4. Dissect the property README
```bash
atd dissect --file battlearena/property/README.md
```

### 5. Dissect ruler.go (code file)
```bash
atd dissect --file battlearena/ruler/ruler.go --llm
```

## Expected Output

**Without `--llm`:** A structured prompt listing the document content and asking for atom boundary proposals. IDE Agent processes this and outputs a JSON list like:
```json
{"atoms": [{"id": "ruler_turn_flow", "responsibility": "...", "line_range": [10, 45]}]}
```

**With `--llm` (Ollama):** Same JSON output, written directly.

**With `--llm` (fallback):** A `pipeline_output/dissect_ruler.result` file path printed; the IDE Agent picks this up via `atd continue`.

## Acceptance Criteria
- [ ] Passthrough mode prints a non-empty prompt to stdout
- [ ] At least one dissect call produces valid atom boundary proposals (JSON with `id` + `responsibility`)
- [ ] `ruler/README.md` yields at least 3 proposed atoms
- [ ] `entity/README.md` yields at least 2 proposed atoms
