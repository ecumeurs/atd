# Issue: API typed atd aren't capturing full payload/contract details

**ID:** `20260306_api_atd_payload_capture_shortcoming`
**Ref:** `ISS-019`
**Date:** 2026-03-06
**Severity:** Medium
**Status:** Open
**Component:** `atd_management_skill`
**Affects:** Accuracy of API documentation in ATDs

---

## Summary

When working on projects to test ATDs, instructions for API expectations and implementations (including full payloads and detailed actions) are only partially captured by the ATD generation logic. This leads to incomplete documentation that misses critical data structures and functional side-effects.

---

## Technical Description

### Background
The ATD system should extract "Atoms" from descriptions or code. For `type: API`, it is expected to capture the full technical contract, including nested payload structures, specific headers, and the logical "Action" or "Side Effect" associated with the endpoint.

### The Problem Scenario
The user provided a detailed API contract during instruction:

**API Contract (Ingest):**
- `POST /internal/arena/start`
  - Payload: `{ match_id: string, players: [ {id, entities...} ], callback_url: string }`
  - Action: Spawns a new battle arena, maps the `callback_url` to the game state. Starts the 30s [[rule_turn_clock]].
  - Returns: `{ arena_id: string, initial_state: object }`
- `POST /internal/arena/{id}/action`
  - Payload: `{ player_id: string, type: string, target_coords: {x,y} }`
  - Action: Translates REST payload to native [[api_ruler_methods]] messages.
  - Returns: `200 Accepted` immediately.

The resulting ATD failed to document the full payload and action details, capturing only a subset of the provided specification.

### Where This Pattern Exists Today
This manifests during the extraction/generation phase of API-typed Atoms when an LLM or script processes high-density technical instructions but simplifies them too much in the output.

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High |
| Impact if triggered | Medium |
| Detectability | Medium — manifests as incomplete ATD files compared to input intent |
| Current mitigant | Manual review and manual correction of generated ATDs |

---

## Recommended Fix

**Short term:** Update the system prompt or extraction scripts for `type: API` to prioritize verbatim capture of payload schemas and "Action" bullets. 
**Medium term:** Implement a validation step that checks if all fields mentioned in the "Contract" section of the prompt made it into the `THE RULE / LOGIC` section of the Atom.
**Long term:** Use a more structured schema (e.g., OpenAPI fragments) as the source of truth for API Atoms instead of free-form text.

---

## References

- [atd_management_skill](file:///home/bastien/work/skill/atd_management_skill/)
- [ATD.md](file:///home/bastien/work/skill/.agent/skills/atd/SKILL.md) (Logic for API atom extraction)
