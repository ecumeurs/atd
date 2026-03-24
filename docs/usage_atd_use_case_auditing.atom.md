---
id: usage_atd_use_case_auditing
human_name: "Use Case: Full System Auditing"
type: USAGE
version: 1.0
status: STABLE
priority: 5
tags: [atd, usecase, auditing, health]
parents:
  - [[domain_atd_usage_protocol]]
dependents: []
layer: CUSTOMER
---

# Use Case: Full System Auditing

## INTENT
To periodically check the health and integrity of the ATD ecosystem.

## THE RULE / LOGIC
1. **Semantic Audit**: Run `atd audit` to detect "Semantic Collisions" (two atoms describing the same thing) using cosine similarity on embeddings.
2. **Bloat Check**: Detect atoms that are too long or complex, suggesting they need to be split.
3. **Gap Analysis**: Run `atd crawl --gaps` to find `STABLE` atoms that have no linked code (orphans).
4. **Link Audit**: Run `atd test-links` to ensure all `@spec-link` tags in source code point to existing valid atoms.

## TECHNICAL INTERFACE
- **Command:** `atd audit`, `atd crawl --gaps`, `atd test-links`.

## EXPECTATION
- Similarity score between atoms remains below the threshold (default 0.85).
- Zero orphan `STABLE` atoms.
- 100% valid `@spec-link` references.
