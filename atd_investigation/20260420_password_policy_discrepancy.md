# ATD Tooling Discrepancy: Password Policy Rule

**Date:** 2026-04-20
**Atom involved:** `[[rule_password_policy]]`
**Context:** This file documents a false negative in ATD trace metrics where a fully implemented and tested rule was reported as having 0% coverage.

## Observed Behavior

When running `atd_trace rule_password_policy`, the system reported:
- **Implementation Rate:** 0%
- **Test Coverage Rate:** 0%
- **Warnings:** `Architecture atom has no Implementation dependents`

However, the reality in the codebase is:
- **Implementation:** Exists in `RegisterRequest.php` and `AuthController.php` via `@spec-link [[rule_password_policy]]`.
- **Tests:** Exist in `e2e_password_policy.js` and `edge_auth_password_policy_full.js`.

## Root Causes

1. **Test Tag Mismatch**: Tests were using `@spec-link` instead of the mandated `@test-link`. The ATD tool doesn't count `@spec-link` as test coverage.
2. **Layer Bias in Metrics**: The ATD trace logic currently penalizes `ARCHITECTURE` layer atoms if they lack `IMPLEMENTATION` layer dependents (e.g., `MECHANIC` atoms), even if they are directly linked to valid source code.

## User Policy Feedback

The user (Lead Architect) has clarified the following policy for the ATD development team:
- **Direct Architecture Linking is Legit**: It is perfectly acceptable (and often preferred) to use `@spec-link` at the `ARCHITECTURE` level for standard implementations.
- **Implementaton Layer is Selective**: `IMPLEMENTATION` layer atoms (Mechanics) should be reserved for **complex implementation techniques** or **special case handling**, not as a mandatory bridge for every rule.
- **Test Link Rigidity**: Tests MUST be branded with `@test-link` to be correctly categorized.

## Recommendations for ATD Dev Team

1. **Relax Orphan Detection**: Do not warn about missing implementation dependents if a stable `@spec-link` exists at the Architecture layer.
2. **Update Tooling Display**: In the IDE extension (code lens), differentiate between "Planned" coverage and "Direct" coverage to respect the documentation hierarchy.
