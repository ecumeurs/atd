# Task 07 — Upsilon: Init and Dissect Coldstart Docs

## Objective
Initialize ATD in the `upsilon/` meta-project (which has documentation only, in `coldstart/`), then dissect all coldstart docs to extract atom proposals.

## Prerequisites
- `atd` binary on PATH
- `upsilon/coldstart/` contains documentation files

## Steps

### 1. Initialize ATD in upsilon
```bash
atd init --dir /home/bastien/work/skill/upsilon
```

### 2. List coldstart docs
```bash
ls /home/bastien/work/skill/upsilon/coldstart/
```
> Expected: `architecture.md`, `commerce.md`, `intrigue.md`, `milestone_1.md`, `upsilon.md`

### 3. Dissect architecture.md
```bash
cd /home/bastien/work/skill/upsilon
atd dissect --file coldstart/architecture.md
```

### 4. Dissect upsilon.md (domain overview)
```bash
atd dissect --file coldstart/upsilon.md
```

### 5. Dissect commerce.md
```bash
atd dissect --file coldstart/commerce.md
```

### 6. Dissect intrigue.md
```bash
atd dissect --file coldstart/intrigue.md
```

### 7. Dissect milestone_1.md
```bash
atd dissect --file coldstart/milestone_1.md
```

## Expected Output
- `.atd` created in `upsilon/`
- `upsilon/docs/` directory created (empty)
- Each dissect call returns a prompt (passthrough) or JSON atom proposals (--llm mode)
- `architecture.md` should yield at least 3 domain/module atoms
- `commerce.md` should yield at least 4 mechanic/domain atoms

## Notes
The `upsilon` project has only documentation and minimal code. The goal is to test:
1. Can `atd dissect` handle pure specification documents?
2. Do the proposals make reasonable ATD boundaries from rich prose documents?

## Acceptance Criteria
- [ ] `atd init` creates `.atd` in `upsilon/`
- [ ] Each coldstart doc produces a non-empty dissect output
- [ ] `architecture.md` proposals include at least one `MODULE` or `DOMAIN` atom
- [ ] `commerce.md` proposals include at least one `MECHANIC` atom
