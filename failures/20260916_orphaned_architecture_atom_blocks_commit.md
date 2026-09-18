# ATD failure: orphaned ARCHITECTURE atom aborts commit

**Date:** 2026-09-16
**Context:** ISS-118 (GDPR export) — committing atom realignment in `upsilonapi`

## Expectation

`git commit` of `upsilonapi/docs/*.atom.md` should succeed. The staged change
only removes a stale `dependents` entry (`[[api_profile_export]]`) from
`api_laravel_gateway.atom.md`, because that atom migrated to the Auth-owned
`/api/v1/auth/export` route.

## Verbatim command

```
cd /home/bastien/work/upsilon/upsilon-hub/upsilonapi && git add docs/ && git commit -q -F - <<'MSG'
docs(iss-118): realign GDPR export atoms with the service-owned model
...
MSG
```

## Verbatim actual result

```
🔍 Running ATD Structural Integrity Check...
❌ ERROR: Orphaned Atom Detected -> docs/api_laravel_gateway.atom.md
   Reason: This is an ARCHITECTURE atom but has no parents defined.
   Fix 1 : Add a parent business/design requirement -> parents: [[req_your_parent]]
   Fix 2 : Use the escape hatch -> parents: [[req_tech_debt_backlog]]

🛑 ATD Check Failed. Commit aborted. Please fix the above atoms.
```

## Analysis

The orphan is PRE-EXISTING, not introduced by this change. Verified on a
tracked file with a sound method:

```
git show HEAD:docs/api_laravel_gateway.atom.md | head -20
```

shows `parents: []` already empty at HEAD. The hook evidently validates only
atoms touched by the commit, so unrelated edits to a long-orphaned atom
surface the debt at an arbitrary later time.

Two observations worth acting on:

1. The check is commit-scoped, so orphaned atoms accumulate invisibly until an
   unrelated commit happens to touch them. A repo-wide audit would surface the
   real backlog instead of ambushing whoever edits next.
2. `api_laravel_gateway` describes the Laravel gateway + Reverb WebSockets,
   which Phase 6 DECOMMISSIONED. The correct resolution is probably
   deprecation/retirement of the atom, not attaching a parent to keep a
   dead-subsystem atom structurally valid. Using the `req_tech_debt_backlog`
   escape hatch would silence the error while preserving a misleading atom.

Resolution routed to `documentalist` rather than resolved inline, per the rule
that atoms are not silently rewritten to satisfy tooling.
