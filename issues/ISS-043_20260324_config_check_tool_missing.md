# Issue: Config Check Tool Missing in ATD Toolkit

**ID:** `20260324_config_check_tool_missing`
**Ref:** `ISS-043`
**Date:** 2026-03-24
**Severity:** Medium
**Status:** Resolved
**Component:** `scripts/cmd/atd/cmd`
**Affects:** User, Developers, CI/CD pipelines

---

## Summary

The ATD toolkit currently lacks a dedicated tool to validate the `.atd` configuration file and verify the availability of configured models on LLM providers. This leads to silent failures, confusing errors during execution of other tools (like `atd audit` or `atd index`), and difficulty in troubleshooting environment setup issues.

---

## Technical Description

### Background

The `.atd` file is the central configuration for ATD tools, defining document paths and LLM provider configurations. Tools rely on these settings to connect to local or remote Ollama services and use specific models for tasks like embedding, auditing, and reconciliation.

### The Problem Scenario

1. A user updates their `.atd` file to use a new remote provider or a different model.
2. The user runs an ATD tool (e.g., `atd index`).
3. If the provider is unreachable or the model is not pulled on the server, the tool fails deep in the execution pipeline with a generic error.
4. There is no simple, unified command to "smoke test" the environment and reveal exactly what is wrong (e.g., "Remote provider at 192.168.1.10 is unreachable" or "Model 'qwen2.5-coder:14b' is missing on local provider").

### Where This Pattern Exists Today

- `scripts/config/config.go`: Loads configuration but does not validate field presence or connectivity.
- `scripts/pkg/ollama/client.go`: Contains `ListModels` and `Generate` functions but lacks a unified health check orchestrator.
- `atd_management_skill/.agent/skills/atd/init.sh`: Performs some hardcoded checks for `localhost` but isn't extensible or tool-integrated.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium (Operational friction, setup confusion) |
| Detectability | High (Errors during tool usage) |
| Current mitigant | Manual `curl` commands or reading log files. |

---

## Recommended Fix

**Short term:** Implement an `atd check` (or `atd status`) subcommand that:
1. Validates the `.atd` file structure and required fields.
2. Iterates through all configured `llm.providers`.
3. Checks connectivity for each provider.
4. Lists all available models on each reachable provider.
5. Cross-references available models with the ones required by the `llm.models` and `fallback_model` configuration.
6. Reports a clear "Ready" or "Missing" status for each task-model pair.

**Medium term:** Expose this tool as an MCP command `atd_check` so IDE agents can verify the environment automatically.

**Long term:** Integrate the check into the `atd` initialization and cold-start workflows.

---

## References

- [scripts/config/config.go](file:///home/bastien/work/skill/scripts/config/config.go)
- [scripts/pkg/ollama/client.go](file:///home/bastien/work/skill/scripts/pkg/ollama/client.go)
- [.atd](file:///home/bastien/work/skill/.atd)
