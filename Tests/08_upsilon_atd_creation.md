# Task 08 — Upsilon: Create ATD Atoms from Coldstart Docs

## Objective
Using the dissect proposals from Task 07, create `.atom.md` files in `upsilon/docs/`. Focus on DOMAIN and MECHANIC atoms that reflect the system's written specifications.

## Prerequisites
- Task 07 complete (dissect proposals from all coldstart docs)
- IDE Agent familiar with atom template

## Suggested Atom Set (from coldstart content)

### From `upsilon.md` (domain overview)
- `upsilon_domain` — DOMAIN: What the Upsilon project is and why it exists
- `upsilon_architecture` — MODULE: High-level architectural overview

### From `architecture.md`
- `upsilon_actor_model` — MECHANIC: Actor-based concurrency model
- `upsilon_message_queue` — MECHANIC: How actors communicate via message queues

### From `commerce.md`
- `upsilon_economy` — DOMAIN: Core economic loops
- At least 2 more MECHANIC atoms from commerce mechanics

### From `intrigue.md`
- At least 2 DOMAIN or MECHANIC atoms from intrigue/political system

### From `milestone_1.md`
- `upsilon_milestone_1` — SPECIFICATION: Goals and scope of milestone 1

## Steps

### 1. Create domain atom for the project
```bash
cd /home/bastien/work/skill/upsilon
atd update --file docs/upsilon_domain.atom.md \
  --set "id=upsilon_domain" \
  --set "type=DOMAIN" \
  --set "status=DRAFT" \
  --intent "Defines the purpose, scope, and high-level vision of the Upsilon meta-project."
```

### 2. Create remaining atoms
> For each proposed atom from Task 07, create the file and populate via `atd update`.
> Use the dissect proposals as the primary source for `--intent` and `--logic` content.

### 3. Verify all atoms
```bash
atd query --search ""
```

### 4. Weave parent links
```bash
atd weave
```

## Expected Output
- At least 8 `.atom.md` files in `upsilon/docs/`
- Atoms cover all 5 coldstart documents
- `atd query` returns all created atoms

## Acceptance Criteria
- [ ] At least 8 atoms created in `upsilon/docs/`
- [ ] Atoms span DOMAIN, MODULE, MECHANIC, and SPECIFICATION types
- [ ] `milestone_1` has a SPECIFICATION atom
- [ ] `atd weave` runs without error
