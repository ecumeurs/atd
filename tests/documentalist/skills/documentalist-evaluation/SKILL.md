---
name: documentalist-evaluation
description: This skill must be activated when an agent needs to run, interpret, or report the documentalist evaluation harness.
---

# Documentalist Evaluation

Use this skill for the repository's documentalist behavioral evaluation. Do not use it for ordinary ATD checks or to change the canonical `upsilon-hub` source.

Run deterministic checks first:

```bash
python3 -m unittest tests/documentalist/test_runner.py
python3 -m py_compile tests/documentalist/runner.py tests/documentalist/test_runner.py
```

Live evaluation invokes the installed OpenCode `documentalist` agent and configured model/provider. It can cost money and produce nondeterministic results. Obtain explicit authorization before a live run unless the request explicitly asks for one. Require Python 3, `make`, installed `atd` and `opencode` CLIs, the installed `documentalist` agent, and configured provider access.

Choose the narrowest scenario: `preflight`, `drift`, `missing-link`, or `tool-fallback`. Run all live scenarios only when authorized:

```bash
make -C atd documentalist-eval
python3 tests/documentalist/runner.py --scenario <id> --live
```

Use `--keep` only when a successful disposable fixture must be inspected. Failures automatically retain their `/tmp/opencode/documentalist-*` fixture with `documentalist-result.json` and `documentalist-transcript.json`. Interpret the JSON `status`, `failures`, `changes`, `fixture`, and command exit status before diagnosing the agent or harness. Preserve canonical `upsilon-hub`; never push.

`missing-link` currently fails because the harness catches the known file-scope placement defect in `atd update --spec-link`. The all-live Make target is therefore expected to return nonzero until the defect is fixed. Treat that expected scenario failure as evidence that the harness works, not as a harness failure.
