# ISS-023 — Unified `atd` CLI Binary

## Objective

Merge all 23 individual ATD Go tools (`scripts/atd-*`) into a single `atd` binary using `spf13/cobra` subcommands. This consolidation:

- Eliminates massive code duplication (6× `GenerateRequest` structs, 4× `EmbeddingRequest`, 3× `cosineSimilarity`, 3× atom parsers)
- Removes per-tool flag boilerplate — all config comes from `.atd` at project root  
- Introduces **task-based LLM routing**: tools request a *task type* (e.g. `dissect`, `audit_bloat`), the config resolves which model and provider to use
- Implements **`atd continue`** protocol for IDE Agent fallback when no Ollama is available
- Merges 6 tools with overlapping responsibilities (23 → 17 subcommands + 2 new)

## Related Issues

| Issue | Role |
|-------|------|
| **ISS-023** | Primary — binary unification |
| ISS-021 | Tiered LLM access (provider chain, task routing) |
| ISS-022 | Test traceability (`@test-link` convention) |
| ISS-024 | ATD indexing and research (Nomic) |
| ISS-011 | Generation orchestration (IDE Agent handoff) |

## How to Execute

Tasks are numbered sequentially in `task_list.md`. Each task has a dedicated file (`01_*.md` through `12_*.md`) with:
- **Context**: What exists today and why we're changing it
- **Exact steps**: What files to create/modify, with code patterns
- **Acceptance criteria**: How to verify the task is done
- **Dependencies**: Which tasks must be complete first

**Execute tasks in order.** Some tasks can be parallelized (noted in task_list.md).

## Architecture Summary

```
scripts/
├── cmd/atd/
│   ├── main.go              ← Cobra entry point
│   └── cmd/
│       ├── root.go           ← Global config loading + flags
│       └── <subcommand>.go   ← One per subcommand (19 files)
├── internal/
│   ├── atom/parse.go         ← Unified atom parser
│   ├── ollama/
│   │   ├── client.go         ← HTTP client (generate, embed, list)
│   │   ├── provider.go       ← Task→model→provider resolution
│   │   └── format.go         ← JSON format schemas
│   ├── cosine/similarity.go  ← Single cosine impl
│   ├── prompt/*.go           ← Per-task prompt templates
│   └── pipeline/continue.go  ← Task list parser + resume
├── config/config.go          ← Extended with LLM config
└── go.mod / go.work          ← Updated module setup
```

## Final Subcommands (19)

| Command | LLM Task | Source |
|---------|----------|--------|
| `atd dissect` | `dissect` | atd-dissect |
| `atd generate` | `dissect` | atd-ollama-generate |
| `atd audit` | `audit_bloat` + `embed` + `audit_code` | atd-audit + atd-ollama-audit |
| `atd fix` | `fix_split` | atd-audit-fixer |
| `atd compare` | `compare` | atd-compare |
| `atd index` | `embed` | atd-ollama-indexer |
| `atd search` | `embed` | atd-ollama-search + atd-tag-sweep |
| `atd update` | — | atd-update + atd-legacy-wrapper |
| `atd congruence` | `congruence` | atd-congruence |
| `atd reconcile` | `reconcile` | atd-reconcile |
| `atd recon` | `recon` | atd-recon |
| `atd crawl` | — | atd-crawl + atd-report-gaps |
| `atd assemble` | `snapshot` | atd-assemble + atd-generate-snapshot |
| `atd query` | — | atd-query |
| `atd weave` | — | atd-link-weaver |
| `atd verify` | — | atd-verify-diff |
| `atd roadmap` | — | atd-roadmap-builder |
| `atd discover` | `intent_extract` + `embed` | atd-discover-links |
| `atd test-links` | — | NEW (ISS-022) |
| `atd continue` | — | NEW (ISS-011) |
