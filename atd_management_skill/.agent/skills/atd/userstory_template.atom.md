---
id: userstory_[slug]
human_name: [User Story Short Name]
type: USER_STORY
version: 1.0
status: DRAFT
priority: SECONDARY
tags: [user_story, agile, requirements]
parents:
  - [[usecase_atom_id]]
dependents: []
---

# [User Story Short Name]

## INTENT
As a [role], I want [capability] so that [benefit].

## THE RULE / LOGIC
- **Precondition:** [State required before story begins]
- **Flow:** [Ordered, narrative-level steps]
- **Postcondition:** [Verifiable system state after story completes]

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[userstory_[slug]]]`
- **Related Issue:** `#NNN`
- **Test Names:** `TestStory[Name]`

## ACCEPTANCE CRITERIA
- [ ] Given [context], When [action], Then [outcome].
- [ ] Given [edge case], Then [fallback behavior].
