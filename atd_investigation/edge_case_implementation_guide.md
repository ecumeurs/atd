# Edge Case Implementation Guide

This document provides guidance for implementing the edge case testing battery.

---

## File Organization

### Directory Structure
```
upsiloncli/tests/scenarios/
├── e2e_*.js              # Customer scenarios (existing)
└── edge_*.js              # Edge case tests (new)

atd_investigation/
├── edge_case_testing_battery.md  # Master plan (completed)
├── example_edge_*.js           # Example implementations (completed)
└── edge_case_implementation_guide.md  # This file
```

---

## Test Script Template

```javascript
// upsiloncli/tests/scenarios/edge_test_category_scenario.js
// @spec-link [[atom_id_primary]]
// @spec-link [[atom_id_secondary]]

const agentIndex = upsilon.getAgentIndex();
const botId = Math.floor(Math.random() * 10000) + "_" + agentIndex;
const accountName = "test_bot_" + botId;
const password = "VerySecurePassword123!";

upsilon.log(`[Bot-${agentIndex}] Starting EC-XX: Test Description`);

// 1. Setup
upsilon.bootstrapBot(accountName, password);
const matchData = upsilon.joinWaitMatch("1v1_PVE");

// 2. Test the edge case
try {
    // Invalid action that should fail
    upsilon.call("game_action", { /* invalid params */ });
    upsilon.assert(false, "ERROR: Invalid action was accepted!");
} catch (e) {
    // Verify correct error was thrown
    upsilon.assertEquals(e.error_key, "expected.error.key", "Wrong error key");
    upsilon.log(`[Bot-${agentIndex}] ✅ Test passed: ${e.message}`);
}

// 3. Verify valid action works (if applicable)
try {
    upsilon.call("game_action", { /* valid params */ });
    upsilon.log(`[Bot-${agentIndex}] ✅ Valid action succeeded`);
} catch (e) {
    // May fail due to other constraints (turn, range, etc.)
    upsilon.log(`[Bot-${agentIndex}] Valid action failed (may be expected): ${e.message}`);
}

upsilon.log(`[Bot-${agentIndex}] EC-XX: TEST DESCRIPTION PASSED.`);
```

---

## Error Key Reference

From ATD atoms and Go tests, the following error keys are expected:

| Error Key | Description | Source ATOM |
|---|---|---|
| `entity.turn.missmatch` | Action out of turn | `[[mech_move_validation_move_validation_turn_mismatch]]` |
| `entity.controller.missmatch` | Wrong controller/owner | `[[mech_move_validation_move_validation_controller_mismatch]]` |
| `entity.path.obstacle` | Move through obstacle | `[[mech_move_validation_move_validation_obstacle_collision]]` |
| `entity.path.occupied` | Move to occupied tile | `[[mech_move_validation_move_validation_entity_collision]]` |
| `entity.path.too.long` | Move exceeds movement credits | `[[mech_move_validation_move_validation_path_length_credits]]` |
| `entity.path.notadjascent` | Move path not adjacent | `[[mech_move_validation_move_validation_path_adjacency]]` |
| `entity.movement.already` | Move after attack | `[[mech_move_validation_move_validation_already_moved]]` |
| `entity.alreadyacted` | Attack already acted | `[[mech_skill_validation_action_state_verification]]` |
| `skill.target.outofgrid` | Attack outside grid | `[[mech_skill_validation_grid_boundaries_verification]]` |
| `skill.cooldown` | Skill on cooldown | `[[mech_skill_validation_economic_cost_verification_cooldown_check]]` |
| `rule.friendly_fire` | Attack on teammate | `[[rule_friendly_fire_team_validation]]` |
| `entity.notfound` | Entity doesn't exist | Various |
| `entity.attack.target.invalid` | Invalid attack target | Various |

---

## Multi-Agent Coordination

For tests requiring 2+ bots:

```javascript
// Bot 0
if (agentIndex === 0) {
    upsilon.setShared("test_value", someData);
}
upsilon.syncGroup("barrier_name", agentCount);

// All bots now have access
const sharedValue = upsilon.getShared("test_value");
```

---

## Error Handling Pattern

All edge case tests should follow this pattern:

```javascript
try {
    // Attempt invalid action
    upsilon.call("route_name", { params });
    // If we reach here, the test failed
    upsilon.assert(false, "ERROR: Expected error was not thrown");
} catch (e) {
    // Verify error structure
    upsilon.log(`Caught error: ${e.message}`);
    upsilon.assert(e.error_key !== undefined, "Error missing error_key");
    upsilon.assert(e.status_code !== undefined, "Error missing status_code");

    // Verify specific error key (if applicable)
    upsilon.assertEquals(e.error_key, "expected.key", "Wrong error key");
    upsilon.assert(e.status_code >= 400 && e.status_code < 600, "Expected 4xx/5xx status");
}
```

---

## Cleanup Pattern

Use `onTeardown` for cleanup:

```javascript
upsilon.onTeardown(() => {
    try {
        upsilon.call("auth_delete", {});
    } catch (e) {
        // Ignore errors during cleanup
    }
});
```

---

## Running Tests Individually

```bash
# Single bot test
cd upsiloncli
./bin/upsiloncli --farm ../atd_investigation/example_edge_movement_obstacle_collision.js --timeout 60

# Multi-bot test
./bin/upsiloncli --farm ../atd_investigation/example_edge_attack_friendly_fire.js ../atd_investigation/example_edge_attack_friendly_fire.js --timeout 120
```

---

## Integration with CI

1. **Add tests to CI report generator** (`tests/ci_report.sh`):
   ```bash
   check_brd "EC-01" "edge_movement_obstacle_collision"
   check_brd "EC-02" "edge_movement_entity_collision"
   # ... etc
   ```

2. **Tests are auto-discovered** by `run_all_scenarios.sh` if they start with `edge_`

3. **Success marker**: `[SCENARIO_RESULT: PASSED]` appended to log file

---

## Unit Tests vs CLI Tests

| Type | Scope | Location | When to Use |
|---|---|---|---|
| CLI e2e | Full API flow validation | `upsiloncli/tests/scenarios/` | API contracts, multi-agent coordination |
| CLI edge | Error code validation, boundaries | `upsiloncli/tests/scenarios/edge_*.js` | Specific rule violations |
| Go unit | Internal logic validation | `*_test.go` files | Algorithm correctness, edge cases difficult to test via CLI |

---

## Priority Implementation Order

### Phase 1: P0 Movement & Authentication (Quick Wins)
1. EC-01: Movement on Obstacle Tiles
2. EC-20: Password Policy Full Coverage (extend existing)
3. EC-21: Invalid Credentials
4. EC-22: Session Timeout

### Phase 2: P0 Attack Validation
5. EC-10: Attack Out of Turn
6. EC-12: Friendly Fire (extend existing)
7. EC-13: Target Not in Range
8. EC-17: Already Acted

### Phase 3: P1 Character & Matchmaking
9. EC-25: Reroll Limit (extend existing)
10. EC-27: Progression Without Wins (extend existing)
11. EC-31: Queue While Already Queued
12. EC-32: Queue While in Match

---

## Common Pitfalls

1. **Not checking error_key**: Always verify the specific error key, not just that an error was thrown
2. **State not verified**: After failed action, verify the state hasn't changed
3. **Missing ATD links**: Always include `@spec-link` tags for traceability
4. **Unclear assertions**: Use descriptive error messages in `upsilon.assert`
5. **Assuming turn order**: Don't assume which bot gets which turn; use `waitNextTurn()` and check `currentCharacter()`
6. **Hardcoded positions**: Grids are generated randomly; always search for required elements (obstacles, enemies)
7. **No cleanup**: Always use `onTeardown` to clean up test accounts

---

## Testing Against the CLI Error Throwing Enhancement

When CLI is enhanced to properly throw on 4xx/5xx:

1. **Before enhancement**: Errors might be silent or require manual checking of `success` field
2. **After enhancement**: Errors throw structured JSON with:
   ```javascript
   {
       message: "Error description",
       success: false,
       error_key: "entity.path.obstacle",
       status_code: 400,
       request_id: "uuid"
   }
   ```

3. **Update tests**: Ensure all edge case tests use try-catch pattern

---

## Checklist for Each Test

- [ ] File starts with `edge_` prefix
- [ ] `@spec-link` tags for relevant atoms
- [ ] Uses try-catch for expected failures
- [ ] Verifies `error_key` matches expected
- [ ] Verifies state unchanged after failure
- [ ] Tests both failure and success cases (where applicable)
- [ ] Includes cleanup via `onTeardown`
- [ ] Uses descriptive log messages
- [ ] Works in both single-agent and multi-agent modes (if applicable)
- [ ] Documented in `edge_case_testing_battery.md`

---

## Next Steps

1. Review `edge_case_testing_battery.md` for full test list
2. Review example scripts for implementation pattern
3. Start with Phase 1 tests (quick wins)
4. Update `tests/ci_report.sh` as tests are implemented
5. Run full test suite to verify all pass
6. Update ATD atom status to STABLE for fully covered rules
