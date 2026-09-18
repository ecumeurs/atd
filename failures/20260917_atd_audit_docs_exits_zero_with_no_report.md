# 2026-09-17 — `atd audit --docs` exits 0 but prints no report at all (silent no-op, distinct from the hang-forever recurrence)

## Expectation
`atd audit --docs <dir>` should complete a bloat/collision scan over the
target project's atom set and print a findings report — a table of atoms
with bloat/collision flags plus a summary line — as requested by the
post-task papertrail sync (Workflow B) for ISS-106's matchmaking
zombie-match fix, run to sanity-check `upsilonbattle/docs` (specifically
`specification_arena_lifecycle`, just bumped 1.0 -> 1.1 by an
architecture-capture agent) after `atd check`/`atd lint`/`atd weave`/`atd
trace` had already come back clean.

## Command (verbatim) and actual result (verbatim)

Run from `/home/bastien/work/upsilon/upsilon-hub` (the umbrella/workspace
root):
```
atd audit --docs upsilonbattle/docs --threshold 0.8 2>&1 | tail -30
```
Did not complete within the harness's 120s foreground timeout; auto-
backgrounded (task id `b3tei3d0n`). Final output file contents once the
task notified as `completed` (exit code 0, per the notification):
```
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=embed Model=nomic-embed-text:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
```
No table, no `[BLOATED]`/collision flags, no summary line, no error — the
process ran 26 LLM dispatch rounds (13 pairs of text_analysis + embed calls,
one per atom in `upsilonbattle/docs` at a guess) and then simply stopped,
reporting exit code 0 as if it had succeeded.

## Context
This is a different failure shape from the two prior `atd audit` reports
(`20260916_atd_audit_workspace_no_return.md`,
`20260917_atd_audit_never_returns_docs_and_atom_code_scopes.md`), which both
documented the process hanging indefinitely (never exiting, output frozen
mid-dispatch). Here the process *did* exit, and exited *cleanly* (code 0)
— but produced zero reportable content. Deterministic substitutes run in
the same session (`atd check --file`, `atd check` diff mode, `atd trace
<id> --summary`, `atd weave`, `atd lint`) all completed normally and
returned real, usable output on the same atom set and code, so this is not
an environment-wide outage — it is specific to `atd audit` in `--docs`
mode.

## Assessment
Combined with the two prior reports, `atd audit` now has two independently
observed broken modes on this local Ollama backend: (1) hangs forever with
no output past the first dispatch line, and (2) — this report — completes
all its LLM dispatch rounds, exits 0, and still emits no findings report.
Neither mode gives a caller (human or agent) any way to distinguish
"clean, zero findings" from "the reporting step itself is broken" — a
0-line report and a 0-line report are indistinguishable whether the atom
set is genuinely unflagged or the tool failed to render its own output
after finishing the scoring.

## Impact
Could not obtain a deterministic-tool-independent bloat/collision cross-
check for `upsilonbattle/docs` (13 atoms, including
`specification_arena_lifecycle`) during ISS-106's post-task sync. Per
Token Economy guidance this was treated as non-blocking — `atd check`,
`atd lint`, `atd weave`, and `atd trace` had already independently
confirmed the touched atom's link coverage and graph consistency — but the
one thing only `audit` covers (bloat/collision) went unchecked, and the
sync's final report to coding-leader had to note this as an unresolved gap
rather than a clean pass.

## Suggested follow-up
- Same root-cause suspicion as the two prior reports: whatever the local
  Ollama backend (`llama3.2:latest` / `nomic-embed-text:latest`) is doing
  after the dispatch lines print, `atd audit` should not be able to reach
  exit code 0 without also printing at least a summary line ("N atoms
  scanned, 0 flagged") — an empty successful run and a broken one must be
  distinguishable in the output.
- Given this is now three reports in two days
  (`20260916_atd_audit_workspace_no_return.md`,
  `20260917_atd_audit_never_returns_docs_and_atom_code_scopes.md`, this
  one) all pointing at `atd audit`'s LLM dispatch path never reliably
  producing a usable report, this likely needs the independent tooling
  review already flagged in memory (`atd-tooling-paused-pending-review`)
  before `atd audit` is relied on again for anything beyond a best-effort,
  non-blocking attempt.
