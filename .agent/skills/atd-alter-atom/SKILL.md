---
name: atd-alter-atom
description: Modify existing ATD atoms with proper blast radius analysis, governance enforcement, and traceability preservation. Use this when asked to change, update, refine, rename, or reparent atoms, especially STABLE or BUSINESS/ARCHITECTURE layer atoms.
---

# ATD Alter Atom Skill

## When to Activate

Use this skill whenever the user asks you to:
- Change / update / refine an existing atom
- Rename or reparent an atom
- Modify a STABLE atom
- Edit BUSINESS or ARCHITECTURE layer atoms
- "Fix the atom X" or "Update the spec for Y"

**This skill has higher impact than Skill A** because it's where governance actually matters.

---

## Pre-flight Checklist

Before modifying any atom, ensure:
- [ ] The `atd_trace`, `atd_crawl`, `atd_check`, and `atd_lint` tools are available and functioning
- [ ] Tools are returning consistent results (trace/check/query agree on coverage)
- [ ] You have access to the correct project context

---

## Step 1 — Analyze Blast Radius (BEFORE Any Editing)

Always understand what you're about to break:

```bash
# 1. Get blast radius — what depends on this atom?
mcp__atd__atd_crawl(atom="target_atom_id")

# 2. Get vertical context — what does this atom govern?
#    This also returns the absolute file path in context[target_id].file_path
mcp__atd__atd_trace(atom="target_atom_id", summary=true)

# 3. Read the full atom file to understand complete context
#    Use the file_path from the trace output (e.g., from context[target_id].file_path)
read(filePath="/path/from/trace/output/target_atom.atom.md")

# 4. Get blast radius — what does it impact?
mcp__atd__atd_trace(atom="target_atom_id")
```

**Review the output:**
- How many downstream atoms depend on this one?
- How many code files have `@spec-link` or `@test-link` to this atom?
- What is the implementation coverage rate?

**If the blast radius is large:**
- Warn the user about potential impact
- Consider whether this change should be split into smaller, safer edits
- Proceed only with explicit user confirmation

---

## Step 2 — Governance Gate: STABLE and BUSINESS Layer Protection

Check the atom's current state:

```bash
mcp__atd__atd_query(field="id", search="target_atom_id")
```

**If `status=STABLE` OR `layer=BUSINESS`:**

You MUST require explicit user confirmation before proceeding.

**Present this information to the user:**
1. Current status and layer
2. Blast radius (from Step 1)
3. The proposed change
4. Potential downstream impact

**Example confirmation prompt:**
> The atom `req_user_auth` is currently STABLE and has 12 downstream dependents and 8 implementation links. You want to change the INTENT from "Users must authenticate with username and password" to "Users must authenticate with MFA". This will require updating all 12 downstream atoms and verifying 8 code locations still comply. **Do you want to proceed? (Respond with "yes" to confirm)**

**Do NOT proceed without explicit user confirmation.**

---

## Step 3 — Edit Surgically via `atd_update`

NEVER rewrite the entire file. Make targeted updates:

```bash
# For content changes (INTENT, LOGIC, TECHNICAL INTERFACE, EXPECTATION)
mcp__atd__atd_update(
  file="docs/target_atom.atom.md",
  intent="Updated intent text",
  logic="Updated logic text"
)

# For metadata changes (status, type, layer, parents)
mcp__atd__atd_update(
  file="docs/target_atom.atom.md",
  set=[
    "status=REVIEW",
    "parents=[[new_parent_atom_id]]"
  ]
)

# For combined changes
mcp__atd__atd_update(
  file="docs/target_atom.atom.md",
  set=["status=REVIEW"],
  intent="Updated intent",
  logic="Updated logic"
)
```

**Critical Rules:**
- Use `set` only for metadata changes (status, type, layer, parents)
- Use named parameters (intent, logic, technical_interface, expectation) for content changes
- Never use both `set` and content parameters for the same field
- Make the smallest change that achieves the goal

---

## Step 4 — Verify Reference Propagation (If ID/Type Changed)

If you changed the atom's `id` or `type`, verify that references were updated:

```bash
# Check for broken references
mcp__atd__atd_lint()

# Verify downstream atoms still have correct parent references
mcp__atd__atd_crawl(atom="new_atom_id")
```

**What to look for:**
- Any `[[old_atom_id]]` references that should be `[[new_atom_id]]`
- Type mismatches in parent-child relationships
- Orphaned atoms that lost their parent

**If you find broken references:**
- The `atd_update` tool should have auto-propagated them (per `mcp_tools.go:444`)
- If not, you may need to manually update affected atoms using `atd_update` on each
- Re-run `atd_weave()` to rebuild the dependency graph

---

## Step 5 — Re-Weave and Verify

Run the verification sequence:

```bash
# 1. Rebuild dependency graph
mcp__atd__atd_weave()

# 2. Structural check on affected files
mcp__atd__atd_check(file="path/to/touched/code/file.go")

# 3. Lint for taxonomy and reference issues
mcp__atd__atd_lint()
```

**Address any issues:**
- If `lint` reports broken references, fix them (Step 4)
- If `check` finds problems, the change may have broken implementation compliance
- Re-run the sequence after any fixes

---

## Step 6 — Flag Downstream Impact (If Intent Changed)

If your change altered the atom's INTENT or EXPECTATION (not just LOGIC or TECHNICAL INTERFACE):

You MUST flag downstream `@spec-link` and `@test-link` sites for review.

```bash
# Find all downstream code links
mcp__atd__atd_crawl(atom="target_atom_id", links=true)

# Get test coverage
mcp__atd__atd_test_links(atom="target_atom_id")
```

**Present a summary to the user:**

> The change to `req_user_auth` affects the following implementations that may need review:
> - `upsilonapi/handlers/auth.go:45` (@spec-link)
> - `upsilonbattle/auth/manager.go:112` (@spec-link)
> - `tests/integration/auth_test.go:23` (@test-link)
>
> Please verify these still comply with the updated INTENT/EXPECTATION.

---

## Step 7 — Update Status (If Appropriate)

Consider whether the atom's status should change:

```bash
# For STABLE atoms that were modified
mcp__atd__atd_update(
  file="docs/target_atom.atom.md",
  set=["status=REVIEW"]
)
```

**Status progression rules:**
- DRAFT → REVIEW → STABLE
- STABLE → REVIEW (if modified significantly)
- REVIEW → STABLE (after review and implementation verification)
- Never skip status levels without explicit user direction

---

## Step 8 — Notify the User

Tell the user:
1. What was changed
2. The blast radius (number of downstream dependents and code links)
3. Any downstream sites flagged for review
4. The new status (if changed)
5. Next steps (e.g., "Review flagged implementations before promoting to STABLE")

Example:
> Updated atom `req_user_auth` (REQUIREMENT, BUSINESS layer). Changed INTENT to add MFA requirement. This affects 12 downstream atoms and 8 code locations. Flagged `upsilonapi/handlers/auth.go:45` and `upsilonbattle/auth/manager.go:112` for review. Status changed to REVIEW. Please verify implementations comply before promoting to STABLE.

---

## Quick Reference Checklist

```
[ ] Ran atd_crawl for blast radius
[ ] Ran atd_trace for vertical context
[ ] Verified STABLE/BUSINESS governance gate (got explicit confirmation)
[ ] Edited surgically via atd_update (never rewrote file)
[ ] Verified reference propagation if id/type changed
[ ] Re-ran atd_weave to rebuild graph
[ ] Ran atd_check on affected files
[ ] Ran atd_lint for taxonomy/issues
[ ] Flagged downstream @spec-link/@test-link sites if intent changed
[ ] Updated status appropriately
[ ] Notified user with change summary, blast radius, flagged sites, and next steps
```

---

## Error Recovery

**If `atd_trace(summary=true)` narrates the wrong atom:**
- Fall back to raw JSON: `atd_trace(atom="target_id", summary=false)`
- Use the `node` field directly for context
- Report this as a tool reliability issue (see investigation §3.3)

**If `atd_crawl` and `atd_check` disagree on coverage:**
- Trust `atd_check` (it uses the live crawl)
- Use `atd_check` numbers for blast radius assessment
- Report the disagreement to the user as a tool consistency issue

**If `atd_lint` reports broken references after id/type change:**
- The auto-propagation may have failed
- Manually update each affected atom using `atd_update`
- Re-run `atd_weave()` after all fixes

**If the tool refuses to modify a STABLE+BUSINESS atom:**
- This is the new governance guard (ISS-010 fix)
- Ask the user to use the CLI with `--force` if they're certain
- Or remind them they can use MCP with `Force:true` if wired

---

## Common Pitfalls

❌ **DON'T:**
- Skip blast radius analysis ("it's a small change")
- Modify STABLE or BUSINESS atoms without explicit confirmation
- Rewrite the entire file instead of surgical edits
- Forget to re-run `atd_weave()` after changes
- Change INTENT/EXPECTATION without flagging downstream sites
- Use both `set` and content parameters for the same field

✅ **DO:**
- Always analyze blast radius before editing
- Require explicit confirmation for STABLE/BUSINESS atoms
- Edit surgically via `atd_update`
- Verify reference propagation after id/type changes
- Re-weave, check, and lint after every change
- Flag downstream sites when intent changes
- Update status appropriately
- Notify user with full context

---

## Governance Enforcement Notes

**STABLE+BUSINESS Protection:**
- Per ISS-010 fix, `atd_update` now refuses to modify STABLE+BUSINESS atoms without `--force` (CLI) or `Force:true` (MCP)
- This guard makes the "require confirmation" step tool-backed, not just prose
- If you hit this guard, explain it to the user and ask for explicit direction

**Ancestry Rules:**
- Every IMPLEMENTATION atom must have a BUSINESS ancestor
- `atd_trace`'s `has_customer_origin` flag is the canonical check
- Treat this as a hard blocker if false (the ruleset does, and so should you)

**Type/Layer Canonicalization:**
- `atd_lint` now validates types against the canonical list
- `atd_lint` now enforces CONTRACT/VISION presence
- Non-canonical types/layers will be flagged; fix them or justify the exception