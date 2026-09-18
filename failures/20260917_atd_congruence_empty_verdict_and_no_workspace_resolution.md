# atd congruence returns a bare `is_congruent: false` with no findings body

**Date:** 2026-09-17
**Context:** ISS-118 post-task ATD sync (Workflow B), verifying
`upsilonauth:service_gdpr_export_orchestrator` against its parents/dependents.

## Expectation

`atd congruence --target <id>` should return a verdict *plus an actionable
`audit_report` body* naming which related atom contradicts the target and in
which section (INTENT / THE RULE / LOGIC). A `false` verdict with no findings
is unusable: it asserts an inconsistency exists but gives the operator nothing
to act on, and cannot be distinguished from an LLM parse failure.

Secondary expectation: `--target` should resolve a workspace-qualified atom
(`upsilonauth:service_gdpr_export_orchestrator`) or search the workspace
projects, rather than only the root `docs/` directory. There is no
`--workspace` flag on `congruence` (nor on `trace`), so every multi-project
lookup requires a manual `--docs <submodule>/docs` override.

## Verbatim command

```
cd /home/bastien/work/upsilon/upsilon-hub && timeout 300 atd congruence --target service_gdpr_export_orchestrator --docs upsilonauth/docs
```

## Verbatim actual output

```
[LLM] Task=code_analysis Model=llama3.2:latest Provider=local (Fallback)
{
  "audit_report": "System Congruence Verification",
  "is_congruent": false
}

  															   
```

(Note the trailing line of stray tab/whitespace characters, and that
`audit_report` carries only a title string — no findings.)

## Prior related invocation (resolution failure)

```
cd /home/bastien/work/upsilon/upsilon-hub && timeout 300 atd congruence --target service_gdpr_export_orchestrator
```

```
Error: target atom 'service_gdpr_export_orchestrator' not found in /home/bastien/work/upsilon/upsilon-hub/docs
Usage:
  atd congruence [flags]

Flags:
      --docs string     Path to docs directory
  -h, --help            help for congruence
      --target string   The specific Atom ID to cross-audit

Global Flags:
  -p, --project string   active project name
  -v, --verbose          verbose output

target atom 'service_gdpr_export_orchestrator' not found in /home/bastien/work/upsilon/upsilon-hub/docs
```

Both invocations exited 0 despite failing.

## Impact on the task

The congruence verdict was discarded as unusable. Congruence was substituted
with a manual read of the atom's `THE RULE / LOGIC` against the shipped
implementation, plus deterministic `atd trace` / `atd check` / `atd lint` /
`atd crawl --gaps --workspace`. The real atom-vs-code gaps reported for ISS-118
were found by that manual read, not by `congruence`.
