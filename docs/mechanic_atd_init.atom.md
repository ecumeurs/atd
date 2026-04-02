---
id: mechanic_atd_init
human_name: "ATD Init Command"
type: MECHANIC
version: 1.0
status: STABLE
priority: 5
tags: [atd, cli, init, bootstrap]
parents:
  - [[module_atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Init Command

## INTENT
Bootstrap a `.atd` configuration file in any project directory, enabling all other `atd` subcommands. The single required first step for any new or legacy project.

## THE RULE / LOGIC
- Accepts `--dir <path>` (default: CWD), `--docs <path>` (default `docs/`), `--model <name>` (default `llama3.2`), `--force` (overwrite)
- Walks up from `--dir` to detect an existing `.atd`; errors if found unless `--force` is set
- Writes a complete `.atd` JSON config with: docs path, bloating factors, tiered provider chain (`remote → local → ide_agent`), model-to-task routing, and logging config
- Creates `--docs` directory if it does not exist
- Deterministic: no LLM, no network calls
- Is **not** exposed as an MCP tool (one-time bootstrap operation)

## TECHNICAL INTERFACE (The Bridge)
- **Binary:** `scripts/cmd/atd/cmd/init.go`
- **Usage:** `atd init [--dir PATH] [--docs PATH] [--model NAME] [--force]`
- **Code Tag:** `@spec-link [[mechanic_atd_init]]`

## EXPECTATION (For Testing)
- Running `atd init --dir /tmp/testproject` creates `/tmp/testproject/.atd` and `/tmp/testproject/docs/`
- Running again without `--force` returns an error
- Running with `--force` overwrites without error
