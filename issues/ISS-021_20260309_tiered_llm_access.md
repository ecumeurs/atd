# Issue: Tiered LLM Access for ATD Operations

**ID:** `20260309_tiered_llm_access`
**Ref:** `ISS-021`
**Date:** 2026-03-09
**Severity:** High
**Status:** Open
**Component:** `infra/llm`
**Affects:** `atd-dissect`, `audit`, `comparison`, `merging`

---

## Summary

Implement a tiered LLM access system to optimize token usage and cost. The system will prioritize discovery of remote Ollama instances (e.g., on a local network or internet), falling back to a local Ollama provider, and finally to the IDE's agent LLM as a last resort. This ensures higher-tier models (DeepSeek-R1:7b, Qwen2.5-Coder:14b) can be used for complex tasks like dissection, audit, and merging while minimizing reliance on paid IDE agent tokens.

---

## Technical Description

### Background
Currently, many ATD operations rely on the local Ollama provider or directly on the IDE's agent LLM. Using the IDE's agent LLM for every task is expensive. Local Ollama instances are limited by the host machine's resources.

### The Problem Scenario
A user has a secondary machine with a more powerful GPU running Ollama. The ATD system should automatically detect and use models on this remote machine if available.

```
Request Task (e.g., Dissect Code)
│
├── Step 1: Check Remote Ollama (via Network/VPN)
│   ├── Found DeepSeek-R1? -> Use Remote
│   └── Found Qwen2.5-Coder? -> Use Remote
│
├── Step 2: Fallback to Local Ollama
│   ├── Model available? -> Use Local
│   └── Otherwise -> Next Step
│
└── Step 3: Last Resort Fallback
    └── Use IDE Agent LLM (e.g., Gemini/Claude)
```

### Where This Pattern Exists Today
The existing LLM provider logic needs to be abstracted into a prioritized "Discovery & Tiering" service. Currently, `atd-dissect` and `audit` tools handle LLM calls somewhat independently or through basic local Ollama wrappers.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | High — reduces cost and improves model quality |
| Detectability | Medium — manifested by higher token costs or lower quality output when falling back |
| Current mitigant | Manual configuration or direct use of local Ollama |

---

## Recommended Fix

**Short term:** Implement a basic fallback chain in the configuration: `providers: [remote, local, ide_agent]`.

**Medium term:** Implement auto-discovery (mDNS or configured IPs) for remote Ollama instances and capability checks (querying `/api/tags` to see available models).

**Long term:** An intelligent orchestrator that routes tasks based on complexity: "Dissection" (High Logic) -> DeepSeek-R1/Qwen2.5-Coder; "Simple Tagging" -> Llama3.

---

## References

- [Ollama API Documentation](https://github.com/ollama/ollama/blob/main/docs/api.md)
- [ISS-001: Audit Performance Optimization](ISS-001_20260304_audit_performance.md)
- [ISS-015: ATD Status Update Token Usage](ISS-015_20260304_atd_status_update_bin.md)
- Ollama remote server: 192.168.1.10:11434 