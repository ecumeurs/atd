---
id: usage_atd_use_case_code_sync
human_name: "Use Case: Syncing Code and Documentation"
type: USAGE
version: 1.0
status: STABLE
priority: 5
tags: [atd, usecase, sync, linkage]
parents:
  - [[domain_atd_usage_protocol]]
dependents: []
layer: IMPLEMENTATION
---

# Use Case: Syncing Code and Documentation

## INTENT
To ensure that code implementations are perfectly synchronized with their atomic specifications.

## THE RULE / LOGIC
1. **Developer Change**: A developer modifies logic in a file (e.g., `pkg/auth/login.go`).
2. **Link Check**: The IDE agent or the developer runs `atd verify` on that file.
3. **Congruence Validation**: `atd verify` extracts the `@spec-link` (e.g., `[[auth_rule_password_strength]]`), reads the corresponding atom, and uses an LLM to check if the new code deviates from the rule.
4. **Correction loop**: If a mismatch is found, the tool flags it. The developer must either revert the code change or switch to Architect Mode to update the atom.

## TECHNICAL INTERFACE
- **Command:** `atd verify`, `atd-congruence`.
- **Tag:** `@spec-link [[ATOM_ID]]`.

## EXPECTATION
- Zero congruence violations on `STABLE` code.
- Mismatches are proactively flagged before commit.
