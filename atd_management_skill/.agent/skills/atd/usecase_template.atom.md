---
id: usecase_[slug]
human_name: [Human Readable Use Case Name]
type: USECASE
version: 1.0
status: DRAFT
priority: CORE
tags: [usecase, actor, workflow]
parents:
  - [[domain_or_module_atom_id]]
dependents:
  - [[mechanic_or_rule_that_implements_a_step]]
---

# [Use Case Name]

## INTENT
To describe the end-to-end workflow from the perspective of [Actor] achieving [Goal].

## WORKFLOW (Ordered Steps)
1. Actor [does X] → handled by [[mechanic_or_rule_atom_id]]
2. System validates per [[rule_atom_id]]
3. System responds with [[entity_atom_id]]

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[usecase_[slug]]]`
- **Related Issue:** `#NNN`
- **Test Names:** `TestUseCase[Name]`

## EXPECTATION (For Testing)
- [ ] Step 1 completes successfully given valid input.
- [ ] Step 2 rejects invalid state per the linked rule atom.
- [ ] Entire workflow produces [expected outcome].
