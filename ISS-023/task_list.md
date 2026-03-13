# ISS-023 — Task List

## Execution Order

Tasks are numbered. Execute in order unless parallelism is noted.
Mark `[x]` when complete. Mark `[/]` when in progress.

---

### Phase 1 — Foundation (must be sequential)

- [x] **01** — [Go Module Setup](01_go_module_setup.md)  
  Create `scripts/cmd/atd/` module, add cobra dependency, set up `go.work`.

- [x] **02** — [Config Extension](02_config_extension.md)  
  Extend `scripts/config/config.go` with LLM providers, task→model routing. Update `.atd`.

- [x] **03** — [Shared Libraries](03_shared_libraries.md)  
  Create `internal/atom/`, `internal/cosine/`, `internal/ollama/client.go`.

- [x] **04** — [Tiered Provider](04_tiered_provider.md)  
  Implement `internal/ollama/provider.go` — task→model→provider resolution chain.

- [x] **05** — [Prompt Templates](05_prompt_templates.md)  
  Create `internal/prompt/*.go` — one per LLM task type with format schemas.

- [x] **06** — [Root Command & Continue Protocol](06_root_and_continue.md)  
  Cobra root command, `atd continue`, pipeline output logic.

### Phase 2 — Deterministic Subcommands (can be parallelized)

- [x] **07** — [Deterministic Subcommands](07_deterministic_subcommands.md)  
  Migrate: `update`, `crawl`, `query`, `weave`, `verify`, `roadmap`, `assemble`, `test-links`.

### Phase 3 — LLM Subcommands (sequential, each builds on provider)

- [x] **08** — [Dissect & Generate](08_dissect_generate.md)  
  Migrate: `dissect` (stdout + --llm), `generate` (LLM boundary extraction).

- [x] **09** — [Index & Search](09_index_search.md)  
  Migrate: `index` (Nomic embedder, --mode), `search` (semantic + --grep).

- [x] **10** — [Audit & Fix & Compare](10_audit_fix_compare.md)  
  Migrate: `audit` (bloat + collision + --code), `fix` (LLM split), `compare` (LLM resolution).

- [ ] **11** — [Remaining LLM Commands](11_remaining_llm_commands.md)  
  Migrate: `congruence`, `reconcile`, `recon`, `discover`.

### Phase 4 — Documentation & Verification

- [ ] **12** — [Documentation & Verification](12_docs_and_verification.md)  
  Create `docs/` ATD atoms per subcommand. Write help text. Run smoke + integration tests.

### Phase 5 — MCP HTTP Server (sequential)

- [ ] **13** — [MCP HTTP Server Infrastructure](13_mcp_http_server.md)  
  Create `internal/mcp/server.go` (HTTP handler, protocol types), `internal/mcp/registry.go` (tool registry), and `cmd/serve.go` (`atd serve` subcommand on port 7474). Also scaffold empty `cmd/mcp_tools.go`. No new Go dependencies — stdlib only.

- [ ] **14** — [MCP Tool Registrations](14_mcp_tool_registrations.md)  
  Populate `cmd/mcp_tools.go`: register all 8 deterministic subcommands (`query`, `crawl`, `weave`, `update`, `roadmap`, `verify`, `assemble`, `test-links`) as MCP tools. Requires refactoring each `cmd/X.go` to extract `runX(...)` helper functions that return `(string, error)` instead of printing to stdout.
