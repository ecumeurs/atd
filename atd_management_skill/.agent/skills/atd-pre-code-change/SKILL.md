---
name: atd-pre-code-change
description: Ensure code changes don't break spec intent by tracing @spec-link atoms before editing. Use this whenever editing, refactoring, or fixing code that carries (or should carry) @spec-link tags.
---

# ATD Pre-Code Change Skill

## When to Activate

Use this skill whenever the user asks you to:
- Edit / refactor / fix code that carries `@spec-link` tags
- Modify code that should have `@spec-link` but doesn't
- "Change the authentication logic" or "Refactor the player manager"
- Work on any file in `upsilonapi/`, `upsilonbattle/`, `battleui/`, or `upsiloncli/`

**This is the bridge between "doc-code co-evolution" and daily reality.**

---

## Pre-flight Checklist

Before modifying any code, ensure:
- [ ] You can run `atd_heatmap_code` or grep for `@spec-link`
- [ ] `atd_trace` with `summary=true` is working and narrating the correct atom
- [ ] `atd_trace` (without `summary`) is working and get the blast radius
- [ ] You can read files using the Read tool to access atom content from file_path
- [ ] `atd_check` is available for post-edit verification

---

## Step 1 — Identify @spec-link Atoms in the Target File(s)

Prefered method of linking an Atom to code is above function name, so checks the comments above functions to find `@spec-link`. If none are directly present:

**Use grep:**
```bash
grep -n "@spec-link \[\[" path/to/your/file.go
```

**Collect all atom IDs** referenced in the file. You'll need to trace each one.

**Example output:**
```
path/to/file.go:45:// @spec-link [[api_auth_login]]
path/to/file.go:78:// @spec-link [[mech_action_economy_action_cost_rules]]
```

---

## Step 2 — Load Spec Context via `atd_trace`

For each atom ID found in Step 1, get the narrative context:

```bash
mcp__atd__atd_trace(atom="api_auth_login", summary=true)
mcp__atd__atd_trace(atom="mech_action_economy_action_cost_rules", summary=true)
```

Summary will provide insight on the why this code is as it is.

Then, read the full atom file using the file_path from the trace output:

```bash
# Extract file_path from context[atom_id].file_path in the trace output
# Example: if trace returned context["api_auth_login"].file_path = "/home/user/docs/api_auth_login.atom.md"
read(filePath="/home/user/docs/api_auth_login.atom.md")
```

Also get the blast radius without the summary option:

```bash
mcp__atd__atd_trace(atom="api_auth_login")
mcp__atd__atd_trace(atom="mech_action_economy_action_cost_rules")
```

Without `summary` option, will indicate blast radius.


---

## Step 3 — Verify Planned Change Aligns with Spec Intent

Before making any code changes, ask yourself:

**Does this change still satisfy the atom's INTENT?**
- If YES → Proceed, but verify EXPECTATION still holds
- If NO → **STOP**. Route through **Skill B (atd-alter-atom)** to update the atom first

**Does this change alter the atom's EXPECTATION?**
- If YES → The atom's EXPECTATION section needs updating after the code change
- Note this for post-edit verification (Step 5)

**Example decision tree:**

```
User: "Refactor the auth logic to use JWT instead of session tokens"

Step 2 reveals:
- Atom: api_auth_login (ARCHITECTURE, STABLE)
- INTENT: "Users authenticate with username and password to receive a session token"
- EXPECTATION: "System returns a session token valid for 15 minutes"

Decision:
- This changes the INTENT (JWT ≠ session token)
- This changes the EXPECTATION (JWT structure ≠ session token structure)
- **ACTION:** Route to Skill B first — update api_auth_login to reflect JWT approach
- Only then proceed with code change
```

---

## Step 4 — Execute the Code Change

Now that you've verified alignment (or updated the spec via Skill B), make the code change.

**Follow surgical placement rules for @spec-link tags:**
- Place the tag directly above the specific function/class being implemented
- NEVER place file-level tags
- Use specific atoms, not generic ones

**Example:**
```go
// GOOD: Specific function
// @spec-link [[mech_combat_standard_attack_computation]]
func computeDamage(attacker, defender) int {
    return max(1, attacker.attack - defender.defense)
}

// BAD: File-level tag
// @spec-link [[mech_combat_standard_attack_computation]]
package combat
```

**If the file has no @spec-link but should:**
- Identify which atoms this code implements (use `atd_search` to find candidates)
- Add appropriate `@spec-link` tags following surgical placement
- This is how you fix missing traceability

---

## Step 5 — Post-Edit Verification

After making the code change, verify structural compliance:

```bash
mcp__atd__atd_check(file="path/to/your/file.go")
```

**Address any issues:**
- Missing `@spec-link` tags → Add them (follow surgical placement)
- Unlinked code → Either add links or explain why it's truly orphan code
- Lint issues → Fix them

**If you noted in Step 3 that EXPECTATION changed:**

Update the atom's EXPECTATION section:

```bash
mcp__atd__atd_update(
  file="docs/affected_atom.atom.md",
  expectation="Updated expectation reflecting new code behavior"
)
```

Then re-weave:
```bash
mcp__atd__atd_weave()
```

---

## Step 6 — Verify No Behavior Shifts (Unless Intentional)

If the change was supposed to be behavior-preserving (e.g., refactoring):

```bash
# Verify implementation coverage is still correct
mcp__atd__atd_trace(atom="affected_atom_id")

# Run tests if available
# (This depends on project-specific test infrastructure)
```

**If behavior shifted but wasn't supposed to:**
- The refactor introduced a bug
- Roll back or fix to align with original INTENT/EXPECTATION

**If behavior shifted and was intentional:**
- Update the atom's INTENT and/or EXPECTATION (Skill B)
- Document the behavior change in the atom's LOGIC section

---

## Step 7 — Add @test-link Tags (If Writing Tests)

If you're adding or modifying tests:

```go
// @test-link [[api_auth_login]]
func TestPlayerLogin(t *testing.T) {
    // test implementation
}
```

**Rules:**
- Place `@test-link` directly above the test function
- Link to the atom(s) the test verifies
- One test can link to multiple atoms

---

## Step 8 — Notify the User

Tell the user:
1. What atoms were traced and their key INTENT/EXPECTATION points
2. Whether the change aligned with specs or required spec updates
3. Any @spec-link/@test-link tags added or moved
4. Verification results from `atd_check`
5. Any atoms whose EXPECTATION was updated

Example:
> Traced 2 atoms for `upsilonapi/handlers/auth.go`:
> - `api_auth_login`: INTENT is JWT-based authentication (aligned with change)
> - `req_user_authentication`: EXPECTATION updated to reflect 15-minute JWT expiration
>
> Added `@spec-link [[api_auth_login]]` to `loginHandler` function. Verified with `atd_check` — all links valid. Updated `req_user_authentication` EXPECTATION to match new JWT behavior.

---

## Quick Reference Checklist

```
[ ] Identified all @spec-link atoms in target file(s) (heatmap or grep)
[ ] Loaded context for each atom via atd_trace(summary=true)
[ ] Read full atom files using file_path from trace output
[ ] Verified planned change aligns with INTENT of each atom
[ ] Routed to Skill B if INTENT would be violated
[ ] Noted if EXPECTATION would change
[ ] Executed code change
[ ] Added/moved @spec-link tags following surgical placement
[ ] Ran atd_check for post-edit verification
[ ] Updated atom EXPECTATION if behavior changed
[ ] Verified no unintended behavior shifts
[ ] Added @test-link tags if writing tests
[ ] Notified user with traced atoms, alignment status, and verification results
```

---

## Error Recovery

**If `atd_trace(summary=true)` narrates the wrong atom:**
- Fall back to raw JSON: `atd_trace(atom="target_id", summary=false)`
- Use the `node.intent` and `node.expectation` fields directly
- Report this as a tool reliability issue (see investigation §3.3)

**If `atd_heatmap_code` is not available:**
- Fall back to grep: `grep -r "@spec-link \[\[" path/to/file.go`
- Parse the output manually to extract atom IDs

**If the file has no @spec-link but should:**
- Search for candidate atoms: `atd_search(grep="relevant keywords")`
- Ask the user to confirm which atoms the code implements
- Add the links following surgical placement rules

**If `atd_check` reports missing links after your change:**
- You may have removed or moved code without updating tags
- Restore the tags following surgical placement
- Or add new tags if the code implements new atoms

---

## Common Pitfalls

❌ **DON'T:**
- Skip tracing @spec-link atoms before editing
- Assume the code change aligns with specs without verification
- Place file-level @spec-link tags
- Forget to update atom EXPECTATION when behavior changes
- Edit code governed by a STABLE atom without tracing its INTENT first
- Remove @spec-link tags without replacing them (unless deleting the code)

✅ **DO:**
- Always identify and trace @spec-link atoms before editing
- Verify alignment with INTENT before making changes
- Route to Skill B if the change violates INTENT
- Place tags surgically (above specific functions)
- Update EXPECTATION when behavior changes
- Run `atd_check` after edits
- Add @test-link tags to tests
- Notify user with full context

---

## Doc-Code Co-Evolution Loop

This skill enforces the core ATD principle:

```
Business Need
     ↓
Design Solution (Architecture) ←-- atoms define this
     ↓
Code Implementation ←-- @spec-link links trace back
     ↓
Test Verification ←-- @test-link links trace back
```

**Every code change should either:**
1. Align with existing atom INTENT/EXPECTATION (most common)
2. Trigger atom updates via Skill B (behavior changes)

**Never make code changes in isolation** without understanding which atoms govern that code. The @spec-link tags are your traceability backbone — treat them as essential, not optional.

---

## Working Across Projects

If the ATD workspace contains multiple projects:

1. Check which project you're working in:
   ```bash
   mcp__atd__atd_workspace_list()
   ```

2. Switch to the correct project if needed:
   ```bash
   mcp__atd__atd_workspace_use(project="upsilonapi")
   ```

3. Trace atoms with cross-project references:
   ```bash
   mcp__atd__atd_trace(atom="upsilonapi:api_auth_login")
   ```

4. When adding @spec-link tags, use the `project:atom` format:
   ```go
   // @spec-link [[upsilonapi:api_auth_login]]
   ```

**Never assume you're in the correct project context.** Always verify and switch explicitly.