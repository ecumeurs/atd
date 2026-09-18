# 2026-09-16 — `atd audit --workspace` never returns (LLM-backed scan)

## Expectation
`atd audit --workspace` should complete a bloat/collision scan across the
umbrella's atom set and return findings, as a complement to the deterministic
`atd lint` / `atd crawl --gaps` checks.

## Command (verbatim)
```
timeout 20 atd audit --workspace
```

## Actual result (verbatim)
```
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
[LLM] Task=text_analysis Model=llama3.2:latest Provider=local
EXIT:124
```

It had already been auto-backgrounded once before this, after exceeding a
120s foreground timeout while producing no output at all.

## Context
Encountered during a delegated repo-wide ATD structural audit of the Upsilon
umbrella (upsilon-hub), run alongside `atd lint <docs>` per project and
`atd crawl --gaps --workspace`. Both deterministic commands completed
normally and answered the question that had been asked; only the LLM-backed
`audit` stalled.

## Assessment
Most likely an unreachable or very slow local Ollama backend in this
environment (`Provider=local`, `Model=llama3.2:latest`) rather than a defect
in `atd` itself — the two `[LLM] Task=text_analysis` lines show it dispatched
requests and then simply never got a response. Not conclusively separated
from a genuine `atd` hang, because no timeout/retry/backoff diagnostic is
emitted by the tool: it gives no indication of whether it is waiting on the
backend or spinning, which is itself the actionable finding here.

## Suggested follow-up
`atd audit` would be materially easier to diagnose with a bounded per-request
LLM timeout and an explicit error on backend unreachability, instead of
silently blocking forever after printing the dispatch line.
