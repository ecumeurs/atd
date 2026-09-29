# 2026-09-20 — `atd weave` silently PRUNES `dependents:` entries it cannot resolve

## Title
`atd weave` (run in `upsilon-hub` as part of a post-task documentalist sync) deletes
`dependents:` list entries from atom frontmatter when the linked id doesn't resolve under
the current index — with no diff shown, no warning printed, and a clean/success exit — instead
of leaving the unresolved entry in place and reporting it.

## Expected
Running `atd weave` (or `atd weave --workspace`) should re-derive/refresh cross-atom link
metadata without destructively removing existing `dependents:`/`parents:` entries that the
tool merely fails to resolve in the moment. At minimum, an entry it intends to drop should be
reported (which id, which file, why unresolved) rather than silently vanishing from the file.

## Verbatim command actually run
```
atd weave --workspace
```
(run by a documentalist subagent as the standard post-task-sync step, per this project's ATD
gating protocol; no non-default flags)

## Verbatim output actually produced
Not captured verbatim by the invoking subagent (it reported success/no-error and moved on,
consistent with the tool's own signal — no indication in its stdout that anything had been
removed). The only hard evidence is the file's before/after state, confirmed independently by
the orchestrating session via direct diff/read both times:

**Before weave (restored and verified via `git diff`, clean 1-for-1 id rewrite):**
```
dependents:
  - [[upsilonauth:mech_sanctum_token_renewal]]
  - [[upsilonbattleui:mechanic_frontend_auth_bridge]]
  - [[upsilonbattleui:req_ui_session_timeout]]
```

**Immediately after weave (re-read by the orchestrating session):**
```
dependents:
  - [[upsilonbattleui:mechanic_frontend_auth_bridge]]
  - [[upsilonbattleui:req_ui_session_timeout]]
```

`[[upsilonauth:mech_sanctum_token_renewal]]` — and only that entry — is gone. The two
resolvable `upsilonbattleui:`-prefixed siblings are untouched. This happened twice in the same
session against `docs/req_security_token_ttl.atom.md` in `upsilon-hub` (a `STABLE`,
`BUSINESS`-layer atom), both times immediately following a `weave` invocation and at no other
point.

## Assessment: tool bug vs. invocation error
Confirmed tool-bug shape, not invocation error — this is a read/maintenance command, took no
target-specific arguments, and produced a destructive side effect on a file it wasn't told to
edit.

**Root cause, as far as can be established without ATD-team-side tooling access:** this
project has a pre-existing split-brain id-resolution problem — `upsilonauth`'s atom ids are
declared bare (e.g. `mech_sanctum_token_renewal`) in their own `id:` frontmatter, but every
in-repo code/doc reference to them is self-prefixed
(`[[upsilonauth:mech_sanctum_token_renewal]]`), and `atd` appears to resolve references by
literal string match rather than a normalized id. From the root project's index, the
prefixed form therefore resolves to nothing — `atd check --atom` on the real atom reports 0
code links while a phantom `upsilonauth:`-prefixed bucket separately holds the true counts
(6 impl / 3 test). `atd weave` appears to treat any `dependents:`/`parents:` entry it cannot
resolve as dead weight and prune it on write, with no log line marking the removal and a
successful/quiet exit either way.

**Why this is higher severity than a typical no-op bug:** it isn't idempotent-safe. A
correctly-restored, human-verified link was destroyed a second time by simply re-running the
same maintenance command in the same session, with nothing in the tool's own output flagging
that anything had changed. Any workflow that runs `weave` as a routine sync step (this
project's post-task documentalist sync does, by design) will silently erode STABLE atoms'
traceability links project-wide, specifically wherever cross-project self-prefixed ids are in
use — which, per the split-brain finding above, may be systemic rather than isolated to this
one atom.

## Suggested reproduction for the ATD team
1. In a multi-project ATD workspace, declare an atom's `id:` bare in its own project
   (`id: some_mechanic`), but reference it from another project's atom using a self-prefixed
   form (`[[that_project:some_mechanic]]`) in a `dependents:` (or `parents:`) list.
2. Confirm `atd check --atom` on `some_mechanic` shows 0 links while a separate
   `that_project:some_mechanic` bucket shows the real link counts (the split-brain).
3. Run `atd weave --workspace` (or scoped equivalent) and diff every atom file that referenced
   the prefixed form beforehand.
4. Expect: the prefixed `dependents:`/`parents:` entry is gone, with no corresponding message
   in weave's output.

## Remediation performed (local workaround, not a fix)
The entry was restored a second time by hand-editing
`upsilon-hub/docs/req_security_token_ttl.atom.md` directly — bypassing `atd` entirely — and
verified by `git diff`. This is expected to be pruned again by any future `atd weave` run
until either (a) the id-resolution split-brain is fixed tool-side so prefixed references
resolve correctly, or (b) this project settles and applies a single naming convention (strip
self-prefixes from in-project tags, or have every reference use a project-qualified id
consistently) and re-indexes. Neither has been decided yet; this report exists so the ATD
team can look at the tool-side prune behavior independently of that project-side convention
decision.

## Related prior reports
- `20260919_atd_update_force_dependents_silent_drop.md` — original filing suspected
  `atd update --force`; its Addendum 3 (added same day, after directly observing the second
  loss immediately follow a `weave` call) revises the root-cause attribution to `atd weave`
  and documents the connecting split-brain finding. This report is the clean, dedicated
  write-up of that revised finding, filed separately so it routes to the ATD team without the
  surrounding investigation narrative.
