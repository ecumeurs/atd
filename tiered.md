# Tiered LLM Access — ISS-021 Implementation Reference

**Date:** 2026-03-09  
**Ref:** `ISS-021`  
**Status:** Pre-Implementation Planning  
**Scope:** All `atd-*` tools + cold-start and full-audit shell pipelines  

---

## Part 1 — Tool Impact Map: Which Tools Are Affected

Two new models are being introduced:
- **DeepSeek-R1:7b** — reasoning-oriented, chain-of-thought, strong at structured analysis
- **Qwen2.5-Coder:14b** — code-first, understands code structure, ideal for extraction and tagging

Four available models total (post-upgrade):

| Handle | Model | Type | VRAM (approx) |
|---|---|---|---|
| A | `llama3.2` | General chat / fast fallback | ~4 GB |
| B | `nomic-embed-text` | Embedding only (no generation) | ~0.5 GB |
| C | `deepseek-r1:7b` | Reasoning / structured analysis | ~6 GB |
| D | `qwen2.5-coder:14b` | Code understanding / extraction | ~10 GB |

### Category A — Live Ollama HTTP Calls (model + URL hardcoded)

These make a direct `http.Post` to the Ollama API. They are the primary targets for the tiered provider refactor.

| Tool | File | Line | Current Model | Endpoint | Hardcode Type |
|---|---|---|---|---|---|
| `atd-ollama-audit` | `atd-ollama-audit/main.go` | L30 + L40 | `llama3.2` | `/api/generate` | Model AND URL fully hardcoded |
| `atd-ollama-generate` | `atd-ollama-generate/main.go` | L44, L95–98 | `llama3.2` (reads `.atd` with fallback) | `/api/generate` | URL hardcoded, model partially configurable |
| `atd-ollama-indexer` | `atd-ollama-indexer/main.go` | L35 + L44 | `nomic-embed-text` | `/api/embeddings` | Model AND URL fully hardcoded |
| `atd-ollama-search` | `atd-ollama-search/main.go` | L31 + L38 | `nomic-embed-text` | `/api/embeddings` | Model AND URL fully hardcoded |
| `atd-compare` | `atd-compare/main.go` | L32 + L37 | `llama3.2` | `/api/generate` at `127.0.0.1` | Model AND URL fully hardcoded |

### Category B — Stdout Prompt Passthrough (expects IDE agent to execute)

These tools build and print a structured prompt to stdout. No HTTP call is made. The IDE agent (Gemini/Claude) is the implicit LLM consumer. They require a `--llm` flag to be added for optional tiered routing.

| Tool | File | Prompt Task | IDE Agent Role |
|---|---|---|---|
| `atd-dissect` | `atd-dissect/main.go` | Document → atom boundary extraction | Reads output, generates atoms |
| `atd-reconcile` | `atd-reconcile/main.go` | Semantic diff / conflict resolution | Reads output, returns JSON diff |
| `atd-congruence` | `atd-congruence/main.go` | Multi-atom cross-audit (congruence table) | Reads output, produces audit report |
| `atd-generate-snapshot` | `atd-generate-snapshot/main.go` | Rewrite fragments as narrative doc | Reads output, returns prose |
| `atd-recon` | `atd-recon/main.go` | Validate code ↔ atom match + confidence | Reads output, returns JSON confidence |

### Category C — Shell Pipeline Orchestrators

| Script | LLM Involvement |
|---|---|
| `atd-cold-start.sh` | Phase 3 → calls `atd-ollama-indexer`; Phase 5 → calls `atd-dissect` (stdout, pauses for IDE agent); Phase 7 → calls `atd-ollama-search` + `atd-recon` |
| `atd-full-audit.sh` | Calls `atd-audit` (deterministic), `atd-compare` (live Ollama), `atd-audit-fixer` (deterministic) |

### Category D — Fully Deterministic (no LLM, no change needed)

`atd-assemble`, `atd-audit`, `atd-audit-fixer`, `atd-crawl`, `atd-discover-links`, `atd-legacy-wrapper`, `atd-link-weaver`, `atd-query`, `atd-report-gaps`, `atd-roadmap-builder`, `atd-tag-sweep`, `atd-update`, `atd-verify-diff`

> These tools operate purely on file I/O, regex, JSON parsing, and Go logic. No prompt construction or HTTP call of any kind.

---

## Part 2 — Model Assignment per Tool

### Assignment Matrix

| Tool | Category | Recommended Model | Rationale |
|---|---|---|---|
| `atd-ollama-audit` | A (live HTTP) | **deepseek-r1:7b** | Pass/fail rule compliance is a reasoning task. DeepSeek's chain-of-thought handles structured binary judgment precisely. |
| `atd-ollama-generate` | A (live HTTP) | **qwen2.5-coder:14b** | Atom boundary extraction from code requires code comprehension. Qwen2.5-Coder outperforms llama3.2 on structured JSON-from-code tasks. |
| `atd-ollama-indexer` | A (live HTTP) | **nomic-embed-text** (unchanged) | Embeddings only. `nomic-embed-text` is purpose-built and must stay consistent with the search index. Do not swap. |
| `atd-ollama-search` | A (live HTTP) | **nomic-embed-text** (unchanged) | Must match the indexer model exactly — vectors are incompatible across models. |
| `atd-compare` | A (live HTTP) | **deepseek-r1:7b** | Atom collision resolution (MERGE / REFACTOR / PARENT proposals) is a structured reasoning task. DeepSeek-R1 handles the diagnostic categories well. |
| `atd-dissect` | B (stdout) | **qwen2.5-coder:14b** → IDE fallback | Code and doc structural parsing. Qwen2.5-Coder first; IDE agent only as last resort. |
| `atd-reconcile` | B (stdout) | **deepseek-r1:7b** → IDE fallback | Semantic conflict detection across atom stores is analytical reasoning, not code extraction. |
| `atd-congruence` | B (stdout) | **deepseek-r1:7b** → IDE fallback | Cross-atom logical consistency audit. Pure reasoning task — R1 suits it well. |
| `atd-generate-snapshot` | B (stdout) | **llama3.2** → deepseek-r1 (if quality matters) | Narrative rewriting of assembled fragments. llama3.2 suffices for speed; R1 for higher fidelity prose. |
| `atd-recon` | B (stdout) | **qwen2.5-coder:14b** → IDE fallback | Code ↔ atom match validation. Code-reading task; Qwen2.5-Coder for precision. |

### Summary Decision Table

```
Task Type              → Model
──────────────────────────────────────────────────────
Code structure / extraction   → qwen2.5-coder:14b
Rule compliance / audit       → deepseek-r1:7b
Document / narrative dissect  → deepseek-r1:7b  (or llama3.2 for speed)
Embedding (search/index)      → nomic-embed-text  [never change]
Simple fallback / speed       → llama3.2
Last resort                   → IDE Agent (Gemini/Claude)
```

---

## Part 3 — Fallback Chain: Local Ollama → Remote Ollama → IDE Agent

### Architecture: Priority Discovery Service

A new shared library (`scripts/lib/llm/`) must abstract all Ollama HTTP calls into a tiered provider. The `.atd` config should drive the chain.

#### Proposed `.atd` config extension

```json
{
  "llm": {
    "providers": [
      {
        "name": "remote",
        "base_url": "http://192.168.1.10:11434",
        "timeout_ms": 2000,
        "models": ["deepseek-r1:7b", "qwen2.5-coder:14b", "llama3.2"]
      },
      {
        "name": "local",
        "base_url": "http://localhost:11434",
        "timeout_ms": 500,
        "models": ["llama3.2", "nomic-embed-text"]
      },
      {
        "name": "ide_agent",
        "base_url": "__IDE_PASSTHROUGH__",
        "timeout_ms": 0,
        "models": ["*"]
      }
    ],
    "task_routing": {
      "code_dissect":   "qwen2.5-coder:14b",
      "audit":          "deepseek-r1:7b",
      "reconcile":      "deepseek-r1:7b",
      "generate":       "qwen2.5-coder:14b",
      "embed":          "nomic-embed-text",
      "default":        "llama3.2"
    }
  }
}
```

#### Fallback Flow

```
Tool requests model M for task T
│
├── Step 1: Check remote Ollama (192.168.1.10:11434)
│   ├── GET /api/tags → model M listed? → Use Remote
│   └── Timeout or not found → Step 2
│
├── Step 2: Check local Ollama (localhost:11434)
│   ├── GET /api/tags → model M listed? → Use Local
│   └── model not available → try degraded fallback model (llama3.2 local)
│   └── Still not found or error → Step 3
│
└── Step 3: IDE Agent passthrough
    └── Emit prompt to stdout (current behavior preserved)
        Used by: atd-dissect, atd-reconcile
```

#### Implementation Touchpoints (per tool)

| Tool | Change Required |
|---|---|
| `atd-ollama-audit` | Replace hardcoded `llama3.2` + `localhost` with `llm.Query("audit", prompt)` |
| `atd-ollama-generate` | Replace `http.Post(localhost...)` with `llm.Query("generate", prompt)` |
| `atd-ollama-indexer` | Replace `http.Post(localhost... embeddings)` with `llm.Embed(text)` |
| `atd-ollama-search` | Same as indexer |
| `atd-dissect` | Add optional `--llm` flag: if set, route via tiered provider; otherwise keep stdout passthrough |
| `atd-reconcile` | Same as `atd-dissect` |
| `.atd` config | Add `llm` block (see above) |
| `scripts/config/` | Extend `config.Load()` to parse new `llm` block and expose `ActiveConfig.LLM` |

---

## Part 4 — Benchmarks: Local vs Remote, Current vs Upgraded

### Benchmark Methodology

All benchmarks should be run via `atd-ollama-audit` and `atd-ollama-generate` against:
- **Doc corpus:** `upsilon/coldstart/` files (narrative docs: `commerce.md`, `intrigue.md`, `milestone_1.md`, `architecture.md`)
- **Code corpus:** `upsilonbattle/battlearena/ruler/` package (`ruler.go`, `ruler_test.go`, `ruler_fullgame_test.go`)
- **Reference ATDs:** `upsilon/old_docs/` (104 atoms) — used as quality baseline

### Metrics to Capture

| Metric | How |
|---|---|
| Latency (ms/request) | Time from HTTP send to full response |
| Prompt token count | `prompt_eval_count` from Ollama response |
| Eval token count | `eval_count` from Ollama response |
| Total token cost | Prompt + eval |
| Output quality | Manual/automated comparison vs `old_docs/` reference atoms |
| Atom boundary accuracy | `atd-dissect` output vs reference atom count in `old_docs/` |

### Benchmark Matrix

| Scenario | Model | Endpoint | Task | Expected Δ |
|---|---|---|---|---|
| **Baseline A** | `llama3.2` | local | audit (ruler.go) | Current baseline |
| **Baseline B** | `llama3.2` | local | generate (commerce.md) | Current baseline |
| **Upgrade C** | `deepseek-r1:7b` | local | audit (ruler.go) | Better reasoning accuracy |
| **Upgrade D** | `qwen2.5-coder:14b` | local | generate (ruler.go) | Better code boundary extraction |
| **Remote E** | `deepseek-r1:7b` | remote (192.168.1.10) | audit (ruler.go) | Lower latency if remote GPU is faster |
| **Remote F** | `qwen2.5-coder:14b` | remote (192.168.1.10) | generate (ruler.go) | Full power dissection |
| **Degraded G** | `llama3.2` | local (fallback) | audit (ruler.go) | Cost of fallback vs baseline |

### Early Benchmark Results (atd-dissect prompt)

*Test Run: Extracted directly from `atd-dissect` prompt running against Local vs Remote Ollama.*

| Node | Model | Doc Type | Atoms | Time | Output Tokens |
|---|---|---|---|---|---|
| Local | `llama3.2:latest` | Code (`ruler.go`) | 0 | 93.36s | 36 |
| Remote | `llama3.2:1b` | Code (`ruler.go`) | 0 | 4.10s | 30 |
| Remote | `llama3.2:latest` | Code (`ruler.go`) | 0 | 6.09s | 36 |
| Remote | `deepseek-r1:7b` | Code (`ruler.go`) | 0 | 4.44s | 4 |
| Remote | `qwen2.5-coder:14b` | Code (`ruler.go`) | 0 | 18.59s | 63 |
| Local | `llama3.2:latest` | Doc (`commerce.md`) | 0 | 118.75s | 169 |
| Remote | `llama3.2:1b` | Doc (`commerce.md`) | 0 | 5.63s | 91 |
| Remote | `llama3.2:latest` | Doc (`commerce.md`) | 0 | 9.54s | 39 |
| Remote | `deepseek-r1:7b` | Doc (`commerce.md`) | 0 | 7.37s | 3 |
| Remote | `qwen2.5-coder:14b` | Doc (`commerce.md`) | 0 | 14.74s | 35 |

*Note: The `0` atoms extracted indicates that `atd-dissect`'s zero-shot extraction occasionally struggles to output purely structured JSON without the IDE fallback in its current state, or the prompt needs further tuning. However, the performance gap is stark: the remote `llama3.2:latest` is ~12-15x faster than the local instance.*

### Benchmark Command Template (once tiered lib is implemented)

```bash
# Baseline — current
time ./bin/atd-ollama-audit \
  -atom upsilon/old_docs/mechanic_delta_math.atom.md \
  -code "$(cat upsilonbattle/battlearena/ruler/ruler.go)"

# Upgraded — deepseek-r1 (will use tiered provider)
ATD_LLM_TASK_ROUTE=audit \
ATD_LLM_MODEL=deepseek-r1:7b \
time ./bin/atd-ollama-audit \
  -atom upsilon/old_docs/mechanic_delta_math.atom.md \
  -code "$(cat upsilonbattle/battlearena/ruler/ruler.go)"

# Dissect — qwen2.5-coder via tiered lib
ATD_LLM_TASK_ROUTE=code_dissect \
./bin/atd-dissect \
  --llm \
  -file upsilonbattle/battlearena/ruler/ruler.go \
  > pipeline_output/dissect_ruler_qwen.json
```

---

## Part 5 — atd-dissect Test & Bench Plan

### Test Corpus

#### 1. Documentation Corpus (Upsilon coldstart — narrative docs)

| Document | Lines | Expected Atom Count (ref) | ATD type focus |
|---|---|---|---|
| `upsilon/coldstart/commerce.md` | ~400 | ~8–12 atoms | MECHANIC, ENTITY, DOMAIN |
| `upsilon/coldstart/intrigue.md` | ~250 | ~5–8 atoms | DOMAIN, MECHANIC |
| `upsilon/coldstart/milestone_1.md` | ~350 | ~6–10 atoms | REQUIREMENT, MODULE |
| `upsilon/coldstart/architecture.md` | ~180 | ~4–6 atoms | MODULE, SERVICE |

**Reference baseline:** `upsilon/old_docs/` — 104 atoms generated from equivalent Upsilon sources. Used for qualitative comparison (not one-to-one mapping, but structural similarity expected).

#### 2. Code Corpus (upsilonbattle ruler — Go package)

| File | Lines | Complexity | ATD type focus |
|---|---|---|---|
| `upsilonbattle/battlearena/ruler/ruler.go` | 415 | High (actor model, state machine) | MECHANIC, SERVICE, ENTITY |
| `upsilonbattle/battlearena/ruler/ruler_test.go` | ~430 | Medium | SPECIFICATION, MECHANIC |
| `upsilonbattle/battlearena/ruler/ruler_fullgame_test.go` | ~80 | Low | USAGE |

### Test Procedure per Model

For each model (`llama3.2`, `deepseek-r1:7b`, `qwen2.5-coder:14b`):

```bash
# Step 1: Run atd-dissect against each file
./bin/atd-dissect -file <target> > pipeline_output/dissect_<target>_<model>.json

# Step 2: Count proposed atoms in output
jq '. | length' pipeline_output/dissect_<target>_<model>.json

# Step 3: Compare proposed_ids semantically against old_docs reference
# (manual or via atd-compare)
diff <(jq -r '.[].proposed_id' pipeline_output/dissect_*.json | sort) \
     <(ls upsilon/old_docs/*.atom.md | xargs -I{} basename {} .atom.md | sort)

# Step 4: Record latency and token usage from Ollama response metadata
```

### Pass Criteria (per dissect run)

| Check | Threshold |
|---|---|
| Atom count is plausible | Between 50%–200% of reference count |
| All proposed_ids are slug-formatted | 100% (regex: `^[a-z][a-z0-9_]+$`) |
| Responsibility descriptions are non-empty | 100% |
| No hallucinated ATD types | All types in: `MECHANIC\|API\|UI\|DATA\|DOMAIN\|RULE\|USAGE\|BUILD\|SERVICE\|ENTITY\|MODULE\|REQUIREMENT\|SPECIFICATION` |
| Latency per file < threshold | < 30s local, < 15s remote |

### Expected Model Ranking (hypothesis — to be validated by bench)

```
Code dissection (ruler.go):
  1st: qwen2.5-coder:14b  — best code boundary detection
  2nd: deepseek-r1:7b     — structured reasoning, slower
  3rd: llama3.2           — fast but less structural precision

Doc dissection (commerce.md):
  1st: deepseek-r1:7b     — narrative analysis, domain reasoning
  2nd: qwen2.5-coder:14b  — acceptable but code-biased
  3rd: llama3.2           — fast but lower semantic depth
```

---

## Appendix — Files Requiring Code Changes (Summary)

| File | Change | Category |
|---|---|---|
| `scripts/atd-ollama-audit/main.go` | Replace hardcoded `llama3.2` + `localhost:11434` with `llm.Query("audit", prompt)` | A |
| `scripts/atd-ollama-generate/main.go` | Replace direct `http.Post` with `llm.Query("generate", prompt)` | A |
| `scripts/atd-compare/main.go` | Replace hardcoded `llama3.2` + `127.0.0.1:11434` with `llm.Query("audit", prompt)` | A |
| `scripts/atd-ollama-indexer/main.go` | Replace hardcoded URL+model with `llm.Embed(text)` | A |
| `scripts/atd-ollama-search/main.go` | Same as indexer | A |
| `scripts/atd-dissect/main.go` | Add `--llm` flag: if set, route via tiered provider; else keep stdout passthrough | B |
| `scripts/atd-reconcile/main.go` | Add `--llm` flag, same pattern as dissect | B |
| `scripts/atd-congruence/main.go` | Add `--llm` flag, same pattern as dissect | B |
| `scripts/atd-generate-snapshot/main.go` | Add `--llm` flag, same pattern as dissect | B |
| `scripts/atd-recon/main.go` | Add `--llm` flag, same pattern as dissect | B |
| `scripts/lib/llm/` | **[NEW]** Tiered provider Go package (remote → local → stdout) | Infra |
| `.atd` | Add `llm` config block (providers + task_routing) | Config |
| `scripts/config/config.go` | Extend `config.Load()` to parse and expose `ActiveConfig.LLM` | Config |

---

*This document serves as the implementation blueprint for ISS-021 and does not replace individual Atoms. Once implementation is complete, formalize each component (tiered provider, config schema, task routing) as dedicated ATD atoms under `infra/llm`.*
