# Task 06 — Search: Doc and Code

## Objective
Validate that `atd query` (deterministic) and `atd search` (semantic + grep) work correctly on the upsilonbattle index built in Task 05.

## Prerequisites
- Task 05 complete (index built, atoms in `docs/`)

## Steps

### 1. Deterministic query by ID
```bash
cd /home/bastien/work/skill/upsilonbattle
atd query --field id --search ruler
```
> Should return the `ruler` atom JSON.

### 2. Deterministic query by type
```bash
atd query --field type --search MECHANIC
```
> Should return all MECHANIC atoms.

### 3. Deterministic query by tag
```bash
atd query --field tags --search battle
```

### 4. Semantic search in docs
```bash
atd search --query "movement and range constraints" --scope docs --limit 3
```
> Should return chunks from `ruler_movement.atom.md` and similar.

### 5. Semantic search in code
```bash
atd search --query "damage calculation attack minus defense" --scope code --limit 3
```
> Should return chunks near the damage formula in `ruler.go`.

### 6. Grep search
```bash
atd search --grep "@spec-link"
```
> Should list all files containing `@spec-link` tags.

### 7. Cross-scope search
```bash
atd search --query "entity health points" --scope all --limit 5
```

## Expected Output
- `atd query --search ruler` returns a JSON object with `id: "ruler"`
- Semantic searches return ranked results with file paths and chunk text
- Grep search lists at least one file containing `@spec-link`

## Acceptance Criteria
- [ ] `atd query --field id --search ruler` returns the ruler atom
- [ ] `atd query --field type --search MECHANIC` returns at least 3 atoms
- [ ] `atd search --query "..." --scope docs` returns ranked results from `.atom.md` files
- [ ] `atd search --query "..." --scope code` returns ranked results from `.go` files
- [ ] `atd search --grep "@spec-link"` lists files correctly
