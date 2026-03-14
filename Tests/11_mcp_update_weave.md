# Task 11 — MCP: Update + Weave via MCP Tools

## Objective
Test `atd_update` and `atd_weave` through the MCP interface. These are write operations that modify files — verify the changes are persistent on disk.

## Prerequisites
- Task 09 complete (MCP connected)
- Task 03-04 complete (atoms exist in `upsilonbattle/docs/`)

## Steps

### 1. Update an atom's status via MCP
Ask the IDE Agent:
> *"Use atd_update via MCP to set the status of the ruler atom to REVIEW."*

**MCP call:**
```json
{
  "name": "atd_update",
  "arguments": {
    "file": "/home/bastien/work/skill/upsilonbattle/docs/ruler.atom.md",
    "set": ["status=REVIEW"]
  }
}
```

**Verify on disk:**
```bash
grep "status:" /home/bastien/work/skill/upsilonbattle/docs/ruler.atom.md
```
**Expected:** `status: REVIEW`

### 2. Update an atom's intent section via MCP
```json
{
  "name": "atd_update",
  "arguments": {
    "file": "/home/bastien/work/skill/upsilonbattle/docs/ruler_attack.atom.md",
    "intent": "Validates attack range and computes damage as attacker.Attack minus defender.Defense, minimum 0."
  }
}
```

**Verify:** Open the file and confirm only the INTENT section changed.

### 3. Add a parent reference and run weave via MCP
First, update a mechanic atom to add a parent:
```json
{
  "name": "atd_update",
  "arguments": {
    "file": "/home/bastien/work/skill/upsilonbattle/docs/ruler_attack.atom.md",
    "set": ["parents=[[ruler]]"]
  }
}
```

Then run weave:
```json
{
  "name": "atd_weave",
  "arguments": {}
}
```

**Verify:** Check that `ruler.atom.md` now lists `ruler_attack` in its `dependents:`.

## Expected Output
- Status change is persistent on disk
- Intent section replaced cleanly, other sections unchanged
- Weave populates dependents correctly

## Acceptance Criteria
- [ ] `atd_update set=["status=REVIEW"]` changes status on disk
- [ ] `atd_update intent=` replaces only the INTENT section
- [ ] `atd_weave` returns a summary string (e.g., "Weaved N links")
- [ ] Parent atom's `dependents:` is updated after weave
