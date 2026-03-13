---
id: atd_verify
human_name: "ATD Verify"
type: SERVICE
version: 1.0
status: DRAFT
priority: CORE
tags: [atd, cli, verify, git, compliance]
parents:
  - [[atd_cli]]
dependents: []
---

# ATD Verify

## INTENT
To audit modified files against their linked ATD specifications by combining git diff analysis, native test execution, and LLM-powered compliance verification.

## THE RULE / LOGIC
Runs `git diff --name-only` to find modified files, extracts `@spec-link` tags from each, reads the corresponding atoms from docs, discovers and runs test files in changed directories (via `go test`), and assembles a comprehensive audit prompt with: ATD specifications, modified source code, test files, and test execution results. Outputs the prompt to stdout for IDE Agent processing.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd verify`
- **LLM Task:** None (stdout passthrough — IDE Agent processes the prompt)
- **Code Tag:** `@spec-link [[atd_verify]]`
