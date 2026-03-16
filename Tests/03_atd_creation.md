# Task 03 — ATD Creation: upsilonbattle Atoms

## Objective
Using the dissect proposals from Task 02, create the first set of `.atom.md` files in `upsilonbattle/docs/`. At minimum, create atoms for the Ruler, Entity, and BattleArena modules.

## Prerequisites
- Task 02 complete (dissect proposals available)
- IDE Agent understands the atom template (see `SKILL.md`)

## Atom Template Reference
```markdown
---
id: [UNIQUE_SLUG]
human_name: [Human Readable Name]
type: [MODULE|MECHANIC|RULE|ENTITY|DOMAIN|...]
version: 1.0
status: DRAFT
priority: CORE
tags: [tag1, tag2]
parents:
  - [[parent_atom_id]]
dependents: []
---
# Name
## INTENT
## THE RULE / LOGIC
## TECHNICAL INTERFACE (The Bridge)
## EXPECTATION (For Testing)
```

## Steps

### 1. Create atds

Note: these are example, you need to follow directives from the dissect proposals and granularity guidelines

Using `atd update` to create from scratch (or write the file directly and update with atd update):
```bash
cd /home/bastien/work/skill/upsilonbattle
# Create the file first, then use atd update to finalize
atd update --file docs/battle_arena.atom.md \
  --set "id=battle_arena" \
  --set "human_name=BattleArena Module" \
  --set "type=MODULE" \
  --set "status=DRAFT" \
  --intent "Top-level game session container. Owns the grid, controller registry, and delegates turn arbitration to the Ruler."
```

```bash
atd update --file docs/ruler.atom.md \
  --set "id=ruler" \
  --set "human_name=Ruler Service" \
  --set "type=SERVICE" \
  --set "status=DRAFT" \
  --intent "Arbitrates all game rules: movement range, attack range, damage computation, turn end detection, and death resolution."
```

### 2. Verify atoms are well-formed
```bash
atd query --search ""
```
> Should list all atoms created in `upsilonbattle/docs/`

## Expected Output
- At least 7 `.atom.md` files in `upsilonbattle/docs/`
- `atd query` returns them all as valid JSON
- Each atom has valid frontmatter (`id`, `type`, `status`, `intent`)

## Acceptance Criteria
- [ ] `docs/battle_arena.atom.md` exists and is valid
- [ ] `docs/ruler.atom.md` exists and is valid
- [ ] At least 5 MECHANIC or RULE atoms created
- [ ] `atd query --search ""` returns 7+ atoms
