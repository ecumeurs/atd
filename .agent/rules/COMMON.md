---
trigger: always_on
---

# Upsilon Hub: Operational Standards & Guards

These rules are non-negotiable for all implementation work in Upsilon Hub.

## 1. Communication Layer Protocol
- **API Envelope**: Never modify the communication layer without ensuring compliance with `[[api_standard_envelope]]`.
- **Warning**: Altering serialization or the bridge API contract must trigger an explicit warning for the user to approve.

## 2. Error Handling Philosophy
- **Crash Early**: Defaulting hides critical errors. Avoid silent failures or catch-all default values in the core logic.
- **Fail Fast**: We are in development mode; clear panics or rejections are better than undefined behavior.

## 3. Testing Tool Usage
- **Selective Testing**: Do **NOT** use `trigger_all_ci_tests.sh`. It is too slow and destructive to the current session.
- **Precision Testing**: Always use `scripts/trigger_one_ci_test.sh` for targeted scenario validation.
- **Unit Tests**: `run_all_unit_tests.sh` is lightweight and should be run frequently during refactoring.
- **Isolation**: Testing is meant to test production code. Avoid adding "test-only" logic to production files unless absolutely necessary for observability.