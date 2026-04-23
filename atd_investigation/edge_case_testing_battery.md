# Edge Case Testing Battery for Upsilon Battle

**Purpose**: Comprehensive edge case test scenarios for validating API boundaries, validation rules, and error handling via upsiloncli bot scripts
**Based on**: ATD atoms analysis and existing customer scenarios (2026-04-19)
**Integration**: Designed for `upsiloncli/tests/scenarios/` (CLI-based e2e tests) and unit tests

---

## Test Design Principles

1. **Boundary Testing**: Each scenario tests a specific rule boundary or constraint
2. **Error Validation**: Each scenario verifies that 4xx/5xx errors are properly thrown and caught
3. **ATD-First**: Each scenario links to specific ATOM IDs for traceability
4. **Deterministic Outcomes**: Clear pass/fail assertions based on expected error codes
5. **Self-Cleaning**: Each scenario handles its own cleanup (account deletion, match forfeiture)

---

## Category 1: Movement Validation Edge Cases

### EC-01: Movement on Obstacle Tiles
**Objective**: Verify that movement commands are rejected when path contains obstacle tiles
**Primary ATOM**: `[[mech_move_validation_move_validation_obstacle_collision]]`
**Secondary ATOMs**: `[[mech_board_generation_terrain_obstacles]]`, `[[entity_grid]]`
**File**: `edge_movement_obstacle_collision.js`
**Scenario Type**: Single Agent (PvE)

#### User Journey
1. Player registers and joins PvE match
2. Board generates with obstacles at specific coordinates
3. Player attempts to move character through obstacle tile
4. System rejects the move with appropriate error
5. Player moves around obstacle successfully

#### Validation Points
- ✅ Move to obstacle-occupied tile returns `entity.path.obstacle` error
- ✅ Character position unchanged after failed move
- ✅ Error message clearly indicates obstacle collision
- ✅ Alternative path around obstacle succeeds
- ✅ No turn credits consumed on failed move

#### Failure Conditions
- ❌ Move through obstacle succeeds (CRITICAL)
- ❌ No error thrown or unclear error message
- ❌ Turn credits consumed on failed move
- ❌ Character position changes despite failed move

---

### EC-02: Movement on Entity Collision
**Objective**: Verify that movement commands are rejected when final destination is occupied
**Primary ATOM**: `[[mech_move_validation_move_validation_entity_collision]]`
**Secondary ATOMs**: `[[entity_character]]`
**File**: `edge_movement_entity_collision.js`
**Scenario Type**: Multi-Agent (2v1 PvE)

#### User Journey
1. Two players join match (or player + AI)
2. Player A attempts to move to tile occupied by Player B's character
3. System rejects the move with appropriate error
4. Player A moves to adjacent empty tile successfully

#### Validation Points
- ✅ Move to entity-occupied tile returns `entity.path.occupied` error
- ✅ Attacker position unchanged after failed move
- ✅ Error message clearly indicates tile occupancy
- ✅ Adjacent empty tile movement succeeds
- ✅ No turn credits consumed on failed move

#### Failure Conditions
- ❌ Move to occupied tile succeeds (CRITICAL)
- ❌ Collision allows overlap of entities
- ❌ Turn credits consumed on failed move

---

### EC-03: Movement Already Attacked (Action Economy)
**Objective**: Verify that movement is blocked after character has attacked in current turn
**Primary ATOM**: `[[mech_move_validation_move_validation_already_moved]]`
**Secondary ATOMs**: `[[mech_action_economy]]`, `[[mech_action_economy_action_cost_rules]]`
**File**: `edge_movement_already_attacked.js`
**Scenario Type**: Single Agent (PvE)

#### User Journey
1. Player joins PvE match
2. Player character attacks enemy
3. Player attempts to move character
4. System rejects the move with appropriate error
5. Next turn, player can move again

#### Validation Points
- ✅ Move after attack returns `entity.movement.already` error
- ✅ Character position unchanged after failed move
- ✅ `has_attacked` flag is set to true after attack
- ✅ Movement allowed in subsequent turn
- ✅ `has_attacked` flag resets on turn change

#### Failure Conditions
- ❌ Move allowed after attack (CRITICAL)
- ❌ Action economy not enforced
- ❌ `has_attacked` flag not set/reset correctly

---

### EC-04: Movement Path Too Long
**Objective**: Verify that movement exceeds available movement credits is rejected
**Primary ATOM**: `[[mech_move_validation_move_validation_path_length_credits]]`
**Secondary ATOMs**: `[[mech_action_economy]]`, `[[entity_character]]`
**File**: `edge_movement_path_too_long.js`
**Scenario Type**: Single Agent (PvE)

#### User Journey
1. Player joins PvE match
2. Player character has limited movement credits (e.g., 3 tiles)
3. Player attempts to move 4 tiles
4. System rejects the move with appropriate error
5. Player moves 3 tiles successfully

#### Validation Points
- ✅ Move exceeding credits returns `entity.path.too.long` error
- ✅ Character position unchanged after failed move
- ✅ Movement credits properly tracked (move: 3, max_move: 3)
- ✅ Valid length move succeeds
- ✅ Move reduces movement credits correctly

#### Failure Conditions
- ❌ Move exceeding credits succeeds (CRITICAL)
- ❌ Movement credits not enforced
- ❌ Movement credits not reduced after valid move

---

### EC-05: Movement Path Not Adjacent
**Objective**: Verify that movement path with non-adjacent tiles is rejected
**Primary ATOM**: `[[mech_move_validation_move_validation_path_adjacency]]`
**Secondary ATOMs**: `[[entity_grid]]`
**File**: `edge_movement_path_not_adjacent.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player joins PvE match
2. Player attempts to move with path containing non-adjacent tiles (e.g., 0,0 → 0,2)
3. System rejects the move with appropriate error
4. Player moves with proper adjacent path

#### Validation Points
- ✅ Non-adjacent path returns `entity.path.notadjascent` error
- ✅ Character position unchanged after failed move
- ✅ Adjacent path succeeds
- ✅ Path validation checks each step

#### Failure Conditions
- ❌ Non-adjacent path accepted (CRITICAL)
- ❌ Path validation bypassed

---

### EC-06: Movement Out of Turn
**Objective**: Verify that movement is rejected when not entity's turn
**Primary ATOM**: `[[mech_move_validation_move_validation_turn_mismatch]]`
**Secondary ATOMs**: `[[mech_initiative]]`, `[[mech_action_economy]]`
**File**: `edge_movement_out_of_turn.js`
**Scenario Type**: Multi-Agent (2 bots)

#### User Journey
1. Two bots join PvP match
2. Bot A attempts to move during Bot B's turn
3. System rejects the move with appropriate error
4. Bot B moves successfully
5. Bot A moves successfully when turn arrives

#### Validation Points
- ✅ Move out of turn returns `entity.turn.missmatch` error
- ✅ Only current turn entity can move
- ✅ Turn order enforced by initiative
- ✅ Movement allowed when turn arrives

#### Failure Conditions
- ❌ Move allowed out of turn (CRITICAL)
- ❌ Turn order not enforced

---

### EC-07: Movement Wrong Controller
**Objective**: Verify that movement is rejected when controller doesn't own entity
**Primary ATOM**: `[[mech_move_validation_move_validation_controller_mismatch]]`
**Secondary ATOMs**: `[[entity_player]]`
**File**: `edge_movement_wrong_controller.js`
**Scenario Type**: Multi-Agent (2 bots)

#### User Journey
1. Two bots join PvP match
2. Bot A attempts to move Bot B's character
3. System rejects the move with appropriate error
4. Bot A moves their own character successfully

#### Validation Points
- ✅ Move wrong controller returns `entity.controller.missmatch` error
- ✅ Only entity owner can move their character
- ✅ Controller identity enforced
- ✅ Bot can move their own characters

#### Failure Conditions
- ❌ Bot can control opponent's character (CRITICAL)
- ❌ Controller identity not enforced

---

### EC-08: Movement Grid Boundaries
**Objective**: Verify that movement to coordinates outside grid is rejected
**Primary ATOM**: `[[mech_skill_validation_grid_boundaries_verification]]` (shared logic)
**Secondary ATOMs**: `[[entity_grid]]`, `[[mech_board_generation_terrain_obstacles]]`
**File**: `edge_movement_grid_boundaries.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player joins PvE match
2. Player attempts to move to negative coordinate (-1, 0)
3. Player attempts to move to coordinate beyond grid (e.g., 20, 20 on 10x10 grid)
4. System rejects both moves with appropriate errors
5. Player moves to valid coordinate successfully

#### Validation Points
- ✅ Move to negative coordinates returns error
- ✅ Move beyond grid bounds returns error
- ✅ Valid coordinate move succeeds
- ✅ Grid boundary validation enforced

#### Failure Conditions
- ❌ Move outside grid succeeds (CRITICAL)
- ❌ Grid boundaries not enforced

---

### EC-09: Movement Jump Limitations
**Objective**: Verify that movement doesn't allow jumping over gaps or non-adjacent tiles
**Primary ATOM**: `[[mech_move_validation_move_validation_jump_limitations]]`
**Secondary ATOMs**: `[[entity_grid]]`, `[[mech_board_generation_terrain_obstacles]]`
**File**: `edge_movement_jump_limitations.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player joins PvE match with water/obstacle tiles
2. Player attempts to move across water (non-walkable)
3. System rejects the move with appropriate error
4. Player moves around water successfully

#### Validation Points
- ✅ Move across non-walkable tiles returns error
- ✅ Pathfinding respects terrain type
- ✅ Alternative path around terrain succeeds
- ✅ Jump mechanics not allowed without specific ability

#### Failure Conditions
- ❌ Move across non-walkable tiles succeeds (CRITICAL)
- ❌ Terrain validation bypassed

---

## Category 2: Attack Validation Edge Cases

### EC-10: Attack Out of Turn
**Objective**: Verify that attack is rejected when not entity's turn
**Primary ATOM**: `[[mech_skill_validation_turn_controller_identity_verification]]`
**Secondary ATOMs**: `[[mech_initiative]]`, `[[mech_action_economy]]`
**File**: `edge_attack_out_of_turn.js`
**Scenario Type**: Multi-Agent (2 bots)

#### User Journey
1. Two bots join PvP match
2. Bot A attempts to attack during Bot B's turn
3. System rejects the attack with appropriate error
4. Bot B attacks successfully
5. Bot A attacks successfully when turn arrives

#### Validation Points
- ✅ Attack out of turn returns `entity.turn.missmatch` error
- ✅ Only current turn entity can attack
- ✅ Turn order enforced
- ✅ Attack allowed when turn arrives

#### Failure Conditions
- ❌ Attack allowed out of turn (CRITICAL)
- ❌ Turn order not enforced for attacks

---

### EC-11: Attack Wrong Controller
**Objective**: Verify that attack is rejected when controller doesn't own entity
**Primary ATOM**: `[[mech_skill_validation_turn_controller_identity_verification]]`
**Secondary ATOMs**: `[[entity_player]]`
**File**: `edge_attack_wrong_controller.js`
**Scenario Type**: Multi-Agent (2 bots)

#### User Journey
1. Two bots join PvP match
2. Bot A attempts to attack using Bot B's character
3. System rejects the attack with appropriate error
4. Bot A attacks with their own character successfully

#### Validation Points
- ✅ Attack wrong controller returns `entity.controller.missmatch` error
- ✅ Only entity owner can attack with their character
- ✅ Controller identity enforced for attacks

#### Failure Conditions
- ❌ Bot can attack with opponent's character (CRITICAL)
- ❌ Controller identity not enforced for attacks

---

### EC-12: Attack Friendly Fire (Same Team)
**Objective**: Verify that attacks on teammates are rejected
**Primary ATOM**: `[[rule_friendly_fire]]`, `[[rule_friendly_fire_team_validation]]`
**Secondary ATOMs**: `[[entity_character]]`, `[[rule_friendly_fire_match_type]]`
**File**: `edge_attack_friendly_fire.js`
**Scenario Type**: Multi-Agent (2v2 PvP)

#### User Journey
1. Four bots join 2v2 PvP match
2. Bot A attempts to attack Bot B (teammate)
3. System rejects the attack with appropriate error
4. Bot A attacks enemy successfully

#### Validation Points
- ✅ Attack on teammate returns friendly fire error
- ✅ Target team validation enforced
- ✅ Attack on enemy succeeds
- ✅ Team identification is accurate
- ✅ Match type (PvP vs PvE) correctly determines teams

#### Failure Conditions
- ❌ Attack on teammate succeeds (CRITICAL - allows griefing)
- ❌ Friendly fire not enforced
- ❌ Team identification incorrect

---

### EC-13: Attack Target Not in Range
**Objective**: Verify that attacks beyond skill range are rejected
**Primary ATOM**: `[[mech_skill_validation_range_limit_verification]]`
**Secondary ATOMs**: `[[mech_combat_attack_computation]]`, `[[entity_character]]`
**File**: `edge_attack_target_not_in_range.js`
**Scenario Type**: Single Agent (PvE)

#### User Journey
1. Player joins PvE match
2. Player attempts to attack enemy beyond range (e.g., range 1, distance 3)
3. System rejects the attack with appropriate error
4. Player moves closer and attacks successfully

#### Validation Points
- ✅ Attack beyond range returns range error
- ✅ Distance calculation is accurate (Manhattan or Euclidean)
- ✅ Valid range attack succeeds
- ✅ Range validation per skill (different skills have different ranges)

#### Failure Conditions
- ❌ Attack beyond range succeeds (CRITICAL)
- ❌ Range validation bypassed
- ❌ Distance calculation incorrect

---

### EC-14: Attack Target Out of Grid
**Objective**: Verify that attacks on coordinates outside grid are rejected
**Primary ATOM**: `[[mech_skill_validation_grid_boundaries_verification]]`
**Secondary ATOMs**: `[[entity_grid]]`
**File**: `edge_attack_target_out_of_grid.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player joins PvE match
2. Player attempts to attack coordinate outside grid (e.g., 20, 20)
3. System rejects the attack with appropriate error
4. Player attacks enemy at valid coordinate successfully

#### Validation Points
- ✅ Attack outside grid returns `skill.target.outofgrid` error
- ✅ Valid coordinate attack succeeds
- ✅ Grid boundary validation for attacks

#### Failure Conditions
- ❌ Attack outside grid succeeds (CRITICAL)
- ❌ Grid boundaries not enforced for attacks

---

### EC-15: Attack Target Not on Valid Cell Type
**Objective**: Verify that attacks on invalid terrain types are rejected
**Primary ATOM**: `[[mech_combat_attack_computation]]`
**Secondary ATOMs**: `[[entity_grid]]`, `[[mech_board_generation_terrain_obstacles]]`
**File**: `edge_attack_target_invalid_cell.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player joins PvE match with water tiles
2. Enemy positioned on water tile (if allowed) or player attempts to attack water
3. System rejects the attack with appropriate error if invalid
4. Player attacks enemy on valid terrain successfully

#### Validation Points
- ✅ Attack on invalid terrain returns `entity.attack.celltype` error
- ✅ Valid terrain attack succeeds
- ✅ Terrain validation for attacks

#### Failure Conditions
- ❌ Attack on invalid terrain succeeds (CRITICAL)
- ❌ Terrain validation bypassed

---

### EC-16: Attack Target No Entity
**Objective**: Verify that attacks on empty tiles are rejected
**Primary ATOM**: `[[mech_combat_attack_computation]]`
**Secondary ATOMs**: `[[entity_character]]`
**File**: `edge_attack_target_no_entity.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player joins PvE match
2. Player attempts to attack empty tile with no entity
3. System rejects the attack with appropriate error
4. Player attacks enemy on occupied tile successfully

#### Validation Points
- ✅ Attack on empty tile returns `entity.attack.noentity` error
- ✅ Attack on entity succeeds
- ✅ Entity existence validation

#### Failure Conditions
- ❌ Attack on empty tile succeeds (CRITICAL)
- ❌ Entity validation bypassed

---

### EC-17: Attack Already Acted (Cooldown)
**Objective**: Verify that attack is blocked when entity has already attacked
**Primary ATOM**: `[[mech_skill_validation_action_state_verification]]`
**Secondary ATOMs**: `[[mech_action_economy]]`, `[[mech_action_economy_action_cost_rules]]`
**File**: `edge_attack_already_acted.js`
**Scenario Type**: Single Agent (PvE)

#### User Journey
1. Player joins PvE match
2. Player character attacks enemy
3. Player attempts to attack again in same turn
4. System rejects the attack with appropriate error
5. Next turn, player can attack again

#### Validation Points
- ✅ Second attack returns `entity.alreadyacted` error
- ✅ Only one attack per turn (unless skill allows multiple)
- ✅ Attack allowed in subsequent turn
- ✅ Action state flag properly set/reset

#### Failure Conditions
- ❌ Multiple attacks per turn allowed (CRITICAL)
- ❌ Action economy not enforced for attacks

---

### EC-18: Attack Skill Cooldown
**Objective**: Verify that skills with cooldown cannot be used while on cooldown
**Primary ATOM**: `[[mech_skill_validation_economic_cost_verification_cooldown_check]]`
**Secondary ATOMs**: `[[mech_combat_attack_computation]]`, `[[domain_skill_system]]`
**File**: `edge_attack_skill_cooldown.js`
**Scenario Type**: Single Agent (PvE with skill-equipped character)

#### User Journey
1. Player joins PvE match with skill-equipped character
2. Player uses skill with cooldown
3. Player attempts to use same skill again in same turn (if applicable) or next turn
4. System rejects the skill use with appropriate error if on cooldown
5. After cooldown expires, player uses skill successfully

#### Validation Points
- ✅ Skill on cooldown returns `skill.cooldown` error
- ✅ Cooldown timer properly tracked
- ✅ Skill use succeeds after cooldown expires
- ✅ Different skills can be used in same turn (if economy allows)

#### Failure Conditions
- ❌ Skill use allowed during cooldown (CRITICAL)
- ❌ Cooldown not enforced
- ❌ Cooldown timer incorrect

---

### EC-19: Attack Entity Targeting Rules
**Objective**: Verify that attack respects skill targeting rules (self, enemy, ally, tile)
**Primary ATOM**: `[[mech_skill_validation_entity_targeting_rules_verification]]`
**Secondary ATOMs**: `[[rule_friendly_fire]]`, `[[entity_character]]`
**File**: `edge_attack_targeting_rules.js`
**Scenario Type**: Multi-Agent (2v2 PvP with skill-equipped character)

#### User Journey
1. Four bots join 2v2 PvP match
2. Bot A uses "self-only" skill on enemy (should fail)
3. Bot A uses "enemy-only" skill on ally (should fail)
4. Bot A uses "ally-only" skill on enemy (should fail)
5. Bot A uses skill with correct targeting (should succeed)

#### Validation Points
- ✅ Invalid targeting returns appropriate error
- ✅ Each targeting type enforced correctly
- ✅ Valid targeting succeeds
- ✅ Target type validation per skill

#### Failure Conditions
- ❌ Invalid targeting allowed (CRITICAL)
- ❌ Targeting rules not enforced

---

## Category 3: Authentication & Session Edge Cases

### EC-20: Password Policy Enforcement (Full Coverage)
**Objective**: Verify all password policy requirements are enforced
**Primary ATOM**: `[[rule_password_policy]]`
**Secondary ATOMs**: `[[req_security]]`, `[[uc_player_registration]]`
**File**: `edge_auth_password_policy_full.js` (extends existing e2e_password_policy.js)
**Scenario Type**: Single Agent

#### User Journey
1. Player attempts registration with various non-compliant passwords:
   - Less than 15 characters
   - No uppercase letter
   - No numeric digit
   - No special symbol
   - Valid password but password_confirmation doesn't match
2. All non-compliant registrations rejected
3. Player registers with compliant password

#### Validation Points
- ✅ Password < 15 characters rejected
- ✅ Password without uppercase rejected
- ✅ Password without number rejected
- ✅ Password without symbol rejected
- ✅ Password with mismatched confirmation rejected
- ✅ Compliant password accepted
- ✅ Clear error messages for each violation

#### Failure Conditions
- ❌ Weak password accepted (CRITICAL - security issue)
- ❌ Unclear error messages
- ❌ Compliant password rejected

---

### EC-21: Authentication with Invalid Credentials
**Objective**: Verify that login fails with incorrect credentials
**Primary ATOM**: `[[api_auth_login]]`, `[[uc_player_login]]`
**Secondary ATOMs**: `[[req_security_authorization]]`
**File**: `edge_auth_invalid_credentials.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player registers account
2. Player attempts login with wrong password
3. Player attempts login with wrong account name
4. All failed logins return appropriate errors
5. Player logs in with correct credentials

#### Validation Points
- ✅ Wrong password returns 401 Unauthorized error
- ✅ Wrong account name returns error
- ✅ No token issued on failed login
- ✅ Correct credentials issue valid token
- ✅ Clear error messages for authentication failures

#### Failure Conditions
- ❌ Login succeeds with wrong credentials (CRITICAL - security issue)
- ❌ No error on failed login
- ❌ Valid login rejected

---

### EC-22: Session Timeout / Expired Token
**Objective**: Verify that expired JWT tokens are rejected
**Primary ATOM**: `[[requirement_req_ui_session_timeout]]`, `[[req_security_token_ttl]]`
**Secondary ATOMs**: `[[mechanic_mech_frontend_auth_bridge]]`
**File**: `edge_auth_session_timeout.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player logs in and receives JWT token
2. Player attempts API call with manually expired token
3. System rejects the request with 401 Unauthorized
4. Player logs in again and receives new token
5. New token allows API calls

#### Validation Points
- ✅ Expired token returns 401 Unauthorized error
- ✅ No access to protected routes with expired token
- ✅ New token issued on fresh login
- ✅ New token allows API access
- ✅ Token expiration enforced (15-minute TTL)

#### Failure Conditions
- ❌ Expired token accepted (CRITICAL - security issue)
- ❌ Session timeout not enforced
- ❌ No clear error for expired session

---

### EC-23: Authentication with Missing Token
**Objective**: Verify that API calls without token are rejected
**Primary ATOM**: `[[req_security_authorization]]`, `[[mechanic_mech_frontend_auth_bridge]]`
**Secondary ATOMs**: `[[req_security_public_access]]`
**File**: `edge_auth_missing_token.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player attempts to access protected endpoint without token
2. System rejects the request with 401 Unauthorized
3. Player logs in and receives token
4. Protected endpoint accessed successfully with token

#### Validation Points
- ✅ Protected route returns 401 without token
- ✅ Public routes work without token (login, register)
- ✅ Token in Authorization header properly validated
- ✅ Valid token allows access

#### Failure Conditions
- ❌ Protected route accessed without token (CRITICAL - security issue)
- ❌ Token validation bypassed

---

### EC-24: Admin Authentication with Non-Admin Account
**Objective**: Verify that non-admin accounts cannot access admin endpoints
**Primary ATOM**: `[[uc_admin_login]]`, `[[req_admin_experience]]`
**Secondary ATOMs**: `[[rule_admin_access_restriction]]`
**File**: `edge_auth_non_admin_access.js`
**Scenario Type**: Single Agent

#### User Journey
1. Regular player account is created
2. Player attempts to access admin endpoint (e.g., `/admin/users`)
3. System rejects the request with 403 Forbidden
4. Admin logs in and accesses admin endpoint successfully

#### Validation Points
- ✅ Non-admin accessing admin returns 403 Forbidden
- ✅ Admin login requires admin role
- ✅ Admin endpoints require authentication
- ✅ Admin can access admin endpoints

#### Failure Conditions
- ❌ Non-admin can access admin endpoints (CRITICAL - security issue)
- ❌ Admin role not enforced
- ❌ Admin endpoints accessible without authentication

---

## Category 4: Character & Progression Edge Cases

### EC-25: Character Reroll Limit (3 Max)
**Objective**: Verify that reroll limit of 3 is enforced
**Primary ATOM**: `[[mech_character_reroll_limit]]`, `[[us_character_reroll_reroll_counter]]`
**Secondary ATOMs**: `[[uc_player_registration]]`, `[[us_character_reroll]]`
**File**: `edge_char_reroll_limit.js` (extends existing e2e_character_reroll.js)
**Scenario Type**: Single Agent

#### User Journey
1. Player registers and receives initial character roster
2. Player rerolls character (reroll_count = 1)
3. Player rerolls character (reroll_count = 2)
4. Player rerolls character (reroll_count = 3)
5. Player attempts 4th reroll → rejected
6. Reroll count displays correctly

#### Validation Points
- ✅ Reroll succeeds up to 3 times
- ✅ 4th reroll returns error (limit exceeded)
- ✅ Reroll count accurately tracked (0, 1, 2, 3)
- ✅ Each character can be individually rerolled
- ✅ Error message clearly indicates limit

#### Failure Conditions
- ❌ More than 3 rerolls allowed (CRITICAL)
- ❌ Reroll count not tracked correctly

---

### EC-26: Character Reroll After Match Participation
**Objective**: Verify that rerolls are blocked after first match
**Primary ATOM**: `[[mech_character_reroll_limit]]` (post-match constraint)
**Secondary ATOMs**: `[[us_character_reroll]]`
**File**: `edge_char_reroll_post_match.js`
**Scenario Type**: Single Agent (PvE)

#### User Journey
1. Player registers with initial character roster
2. Player joins and completes a match (or forfeits)
3. Player attempts to reroll character
4. System rejects the reroll with appropriate error
5. Reroll count displays as locked

#### Validation Points
- ✅ Reroll before match succeeds
- ✅ Reroll after match participation rejected
- ✅ Match participation flag set correctly
- ✅ Clear error message for post-match reroll restriction
- ✅ New accounts can reroll (verification)

#### Failure Conditions
- ❌ Reroll allowed after match (CRITICAL)
- ❌ Match participation not tracked correctly

---

### EC-27: Progression Stat Allocation Without Wins
**Objective**: Verify that stat upgrades are rejected without match wins
**Primary ATOM**: `[[rule_progression]]`, `[[uc_progression_stat_allocation]]`
**Secondary ATOMs**: `[[us_win_progression]]`
**File**: `edge_prog_allocation_no_wins.js` (partial coverage in e2e_progression_constraints.js)
**Scenario Type**: Single Agent

#### User Journey
1. New player registers (total_wins = 0)
2. Player attempts to upgrade character stats
3. System rejects the upgrade with appropriate error
4. Player wins a match (total_wins = 1)
5. Player can now allocate 1 stat point

#### Validation Points
- ✅ Upgrade without wins rejected
- ✅ Clear error message about win requirement
- ✅ One win grants exactly 1 allocation point
- ✅ Point can be allocated to any valid stat
- ✅ Upgrade fails again if no remaining points

#### Failure Conditions
- ❌ Upgrades allowed without wins (CRITICAL)
- ❌ Win tracking incorrect
- ❌ Point allocation not granted on win

---

### EC-28: Progression Attribute Cap Violation
**Objective**: Verify that upgrades beyond (10 + total_wins) cap are rejected
**Primary ATOM**: `[[rule_progression]]`
**Secondary ATOMs**: `[[entity_character]]`
**File**: `edge_prog_attribute_cap.js` (partial coverage in e2e_progression_constraints.js)
**Scenario Type**: Single Agent (PvE)

#### User Journey
1. Player with wins wins a match and receives point
2. Player attempts to upgrade beyond cap: HP + Attack + Defense + Movement > 10 + total_wins
3. System rejects the upgrade with appropriate error
4. Player upgrades within cap successfully

#### Validation Points
- ✅ Upgrade exceeding cap rejected
- ✅ Cap calculation: 10 + total_wins enforced
- ✅ Valid upgrades within cap succeed
- ✅ Clear error message about cap violation
- ✅ Cap updates correctly after each win

#### Failure Conditions
- ❌ Upgrade beyond cap succeeds (CRITICAL)
- ❌ Cap calculation incorrect
- ❌ Valid upgrades rejected

---

### EC-29: Progression Movement Gate (Every 5 Wins)
**Objective**: Verify that movement upgrades only allowed every 5 wins
**Primary ATOM**: `[[rule_progression]]`, `[[us_win_progression_movement_locked]]`
**Secondary ATOMs**: `[[us_win_progression]]`
**File**: `edge_prog_movement_gate.js` (partial coverage in e2e_progression_constraints.js)
**Scenario Type**: Single Agent (requires rigged win scenario)

#### User Journey
1. Player has 3 wins
2. Player attempts to upgrade movement
3. System rejects upgrade with appropriate error
4. Player achieves 5 wins
5. Player can now upgrade movement

#### Validation Points
- ✅ Movement upgrade rejected at 1-4 wins
- ✅ Movement upgrade allowed at 5 wins
- ✅ Movement upgrade blocked at 6-9 wins
- ✅ Next movement upgrade allowed at 10 wins
- ✅ Gate formula: floor(total_wins / 5) enforced

#### Failure Conditions
- ❌ Movement upgrade allowed before 5 wins (CRITICAL)
- ❌ Movement gate not enforced correctly
- ❌ Gate formula incorrect

---

### EC-30: Progression Negative Stat Value
**Objective**: Verify that negative stat upgrades are rejected
**Primary ATOM**: `[[rule_progression]]` (Non-Negativity constraint)
**Secondary ATOMs**: `[[entity_character]]`
**File**: `edge_prog_negative_value.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player wins a match and receives allocation point
2. Player attempts to allocate negative value (e.g., hp: -1)
3. System rejects the upgrade with appropriate error
4. Player allocates positive value successfully

#### Validation Points
- ✅ Negative upgrade rejected
- ✅ Zero upgrade rejected (unless that's valid behavior)
- ✅ Positive upgrade succeeds
- ✅ Clear error message about negative values
- ✅ No attribute can have negative value

#### Failure Conditions
- ❌ Negative upgrade allowed (CRITICAL)
- ❌ Stats can go below zero

---

## Category 5: Matchmaking Edge Cases

### EC-31: Join Queue While Already in Queue
**Objective**: Verify that player cannot join multiple queues simultaneously
**Primary ATOM**: `[[rule_matchmaking_single_queue]]`
**Secondary ATOMs**: `[[api_matchmaking]]`, `[[usecase_api_flow_matchmaking]]`
**File**: `edge_match_queue_while_queued.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player joins queue for 1v1_PVP
2. Player attempts to join queue for 2v2_PVP
3. System rejects second queue join with 409 Conflict
4. Player leaves first queue
5. Player can now join second queue

#### Validation Points
- ✅ Second queue join returns 409 Conflict
- ✅ Clear error message about already queued
- ✅ Player can only be in one queue at a time
- ✅ Leave queue clears all queue entries
- ✅ Can join different queue after leaving

#### Failure Conditions
- ❌ Multiple queues allowed (CRITICAL)
- ❌ Queue status not tracked correctly
- ❌ Leave queue doesn't clear entries

---

### EC-32: Join Queue While in Active Match
**Objective**: Verify that player cannot join queue while in active match
**Primary ATOM**: `[[rule_matchmaking_single_queue]]`
**Secondary ATOMs**: `[[api_matchmaking]]`, `[[uc_matchmaking]]`
**File**: `edge_match_queue_while_in_match.js`
**Scenario Type**: Multi-Agent (2 bots)

#### User Journey
1. Two bots join PvP match
2. Bot A attempts to join another queue while match is active
3. System rejects queue join with 409 Conflict
4. Bot A leaves match (forfeits)
5. Bot A can now join queue

#### Validation Points
- ✅ Queue join while in match returns 409 Conflict
- ✅ Clear error message about active match
- ✅ Match status correctly tracked
- ✅ Can join queue after match ends

#### Failure Conditions
- ❌ Queue join allowed during match (CRITICAL)
- ❌ Match status not enforced

---

### EC-33: Invalid Game Mode
**Objective**: Verify that invalid game modes are rejected
**Primary ATOM**: `[[api_matchmaking]]`, `[[req_matchmaking_matchmaking_queue]]`
**Secondary ATOMs**: `[[spec_match_format]]`
**File**: `edge_match_invalid_game_mode.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player attempts to join queue with invalid game mode (e.g., "3v3_PVP")
2. System rejects the request with 400 Bad Request
3. Player joins queue with valid game mode (1v1_PVP)

#### Validation Points
- ✅ Invalid game mode returns 400 Bad Request
- ✅ Valid game modes accepted (1v1_PVP, 2v2_PVP, 1v1_PVE, 2v2_PVE)
- ✅ Clear error message about invalid mode
- ✅ Game mode validation enforced

#### Failure Conditions
- ❌ Invalid game mode accepted (CRITICAL)
- ❌ Game mode not validated

---

### EC-34: Leave Queue When Not Queued
**Objective**: Verify graceful handling of leave queue when not queued
**Primary ATOM**: `[[api_matchmaking]]`
**Secondary ATOMs**: `[[usecase_api_flow_matchmaking]]`
**File**: `edge_match_leave_not_queued.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player attempts to leave queue when not in any queue
2. System handles request gracefully (200 OK or specific error)
3. Player joins queue successfully
4. Player leaves queue successfully

#### Validation Points
- ✅ Leave queue when not queued doesn't crash
- ✅ Appropriate response (200 OK if idempotent, or 404/409 if strict)
- ✅ Can join queue after failed leave
- ✅ Normal leave queue works

#### Failure Conditions
- ❌ Leave queue when not queued causes crash/error
- ❌ System state corrupted

---

## Category 6: Match Resolution Edge Cases

### EC-35: Forfeit Out of Turn
**Objective**: Verify that forfeit is only allowed during player's turn
**Primary ATOM**: `[[rule_forfeit_battle]]`
**Secondary ATOMs**: `[[uc_match_resolution]]`, `[[mech_initiative]]`
**File**: `edge_match_forfeit_out_of_turn.js`
**Scenario Type**: Multi-Agent (2 bots)

#### User Journey
1. Two bots join PvP match
2. Bot A attempts to forfeit during Bot B's turn
3. System rejects forfeit or only allows during any player's turn
4. Bot A forfeits successfully

#### Validation Points
- ✅ Forfeit enforced correctly (may allow anytime or only own turn)
- ✅ Forfeit ends match immediately
- ✅ Opponent declared winner
- ✅ Match status recorded correctly

#### Failure Conditions
- ❌ Forfeit doesn't end match (CRITICAL)
- ❌ Wrong winner declared

---

### EC-36: Action After Match End
**Objective**: Verify that no actions can be taken after match concludes
**Primary ATOM**: `[[uc_match_resolution]]`
**Secondary ATOMs**: `[[mech_action_economy]]`
**File**: `edge_match_action_after_end.js`
**Scenario Type**: Multi-Agent (2 bots)

#### User Journey
1. Two bots join PvP match
2. One bot forfeits, match ends
3. Other bot attempts to take action (move/attack)
4. System rejects the action with appropriate error

#### Validation Points
- ✅ Action after match end rejected
- ✅ Match status correctly set to finished
- ✅ Winner declared before action attempt
- ✅ Clear error message about match being finished

#### Failure Conditions
- ❌ Actions allowed after match ends (CRITICAL)
- ❌ Match status not enforced

---

## Category 7: API & Communication Edge Cases

### EC-37: Missing Request ID Header
**Objective**: Verify that API calls without X-Request-ID are handled appropriately
**Primary ATOM**: `[[api_standard_envelope]]`, `[[api_request_id]]`
**Secondary ATOMs**: `[[req_logging_traceability]]`
**File**: `edge_api_missing_request_id.js`
**Scenario Type**: Single Agent (requires direct HTTP client, may need unit test approach)

#### User Journey
1. Player makes API call without X-Request-ID header
2. System either generates one automatically or rejects with error
3. Player makes API call with X-Request-ID header
4. Request succeeds

#### Validation Points
- ✅ Missing request ID handled gracefully (auto-gen or error)
- ✅ With request ID, call succeeds
- ✅ Request ID is UUIDv7 format if manually provided
- ✅ Request ID appears in response envelope

#### Failure Conditions
- ❌ Missing request ID causes crash
- ❌ Request ID not properly tracked

---

### EC-38: Invalid UUID Format in Request
**Objective**: Verify that invalid UUIDs in parameters are rejected
**Primary ATOM**: `[[api_standard_envelope]]`
**Secondary ATOMs**: `[[entity_character]]`
**File**: `edge_api_invalid_uuid.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player attempts API call with invalid UUID format (e.g., "not-a-uuid")
2. System rejects request with 400 Bad Request
3. Player makes API call with valid UUID
4. Request succeeds

#### Validation Points
- ✅ Invalid UUID returns 400 Bad Request
- ✅ Clear error message about UUID format
- ✅ Valid UUID accepted
- ✅ UUID validation enforced for all UUID parameters

#### Failure Conditions
- ❌ Invalid UUID accepted (potential injection/security issue)
- ❌ Invalid UUID causes crash

---

### EC-39: Malformed JSON in Request Body
**Objective**: Verify that malformed JSON requests are rejected
**Primary ATOM**: `[[api_standard_envelope]]`
**Secondary ATOMs**: `[[api_laravel_gateway]]`
**File**: `edge_api_malformed_json.js`
**Scenario Type**: Unit Test (requires direct HTTP client)

#### User Journey
1. Client sends malformed JSON to API endpoint
2. System rejects request with 400 Bad Request
3. Client sends valid JSON
4. Request succeeds

#### Validation Points
- ✅ Malformed JSON returns 400 Bad Request
- ✅ Clear error message about JSON format
- ✅ Valid JSON accepted
- ✅ No crash on malformed input

#### Failure Conditions
- ❌ Malformed JSON causes crash
- ❌ Invalid JSON accepted

---

### EC-40: 5xx Server Error Handling
**Objective**: Verify that 5xx errors are properly caught and reported
**Primary ATOM**: `[[mechanic_mech_frontend_auth_bridge]]` (error propagation)
**Secondary ATOMs**: `[[api_standard_envelope]]`
**File**: `edge_api_5xx_error_handling.js`
**Scenario Type**: Unit Test (requires simulated 5xx error)

#### User Journey
1. API returns 500 Internal Server Error
2. Client catches error properly
3. Error details extracted from envelope
4. Appropriate user feedback displayed

#### Validation Points
- ✅ 5xx error caught and not causing crash
- ✅ Error message displayed to user
- ✅ Request ID available for debugging
- ✅ No retry logic for 5xx (unless implemented)

#### Failure Conditions
- ❌ 5xx error causes crash
- ❌ No error feedback to user
- ❌ Infinite retry loop

---

## Category 8: Leaderboard Edge Cases

### EC-41: Leaderboard Invalid Game Mode
**Objective**: Verify that leaderboard requests with invalid mode are rejected
**Primary ATOM**: `[[api_leaderboard]]`
**Secondary ATOMs**: `[[us_leaderboard_view_sort_leaderboard]]`
**File**: `edge_leaderboard_invalid_mode.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player requests leaderboard with invalid game mode
2. System rejects request with 400 Bad Request
3. Player requests leaderboard with valid game mode
4. Request succeeds

#### Validation Points
- ✅ Invalid mode returns 400 Bad Request
- ✅ Clear error message about invalid mode
- ✅ Valid modes accepted (1v1_PVP, 2v2_PVP, 1v1_PVE, 2v2_PVE)
- ✅ Leaderboard data returned correctly

#### Failure Conditions
- ❌ Invalid mode accepted
- ❌ Error causes crash

---

### EC-42: Leaderboard Pagination Beyond Results
**Objective**: Verify graceful handling of pagination beyond available results
**Primary ATOM**: `[[api_leaderboard]]`
**Secondary ATOMs**: `[[us_leaderboard_view_sort_leaderboard]]`
**File**: `edge_leaderboard_over_pagination.js`
**Scenario Type**: Single Agent

#### User Journey
1. Player requests leaderboard with page 9999
2. System returns empty results or 404 Not Found
3. Player requests leaderboard with page 1
4. Request succeeds with results

#### Validation Points
- ✅ Over-pagination handled gracefully
- ✅ Empty results returned or appropriate error
- ✅ No crash on over-pagination
- ✅ Valid pagination works correctly

#### Failure Conditions
- ❌ Over-pagination causes crash
- ❌ Invalid data returned

---

## Category 9: Admin Edge Cases

### EC-43: Admin View Private Data
**Objective**: Verify that admins cannot view private user data
**Primary ATOM**: `[[rule_admin_access_restriction]]`
**Secondary ATOMs**: `[[uc_admin_user_management]]`
**File**: `edge_admin_private_data_access.js`
**Scenario Type**: Single Agent (admin)

#### User Journey
1. Admin logs in
2. Admin requests user list
3. Verify private fields (full_address, birth_date) are null/omitted
4. Admin requests specific user profile
5. Verify private fields are still null/omitted

#### Validation Points
- ✅ User list excludes private fields
- ✅ User profile excludes private fields
- ✅ Admin can view non-private data
- ✅ Private data never exposed through admin APIs

#### Failure Conditions
- ❌ Admin can view private data (CRITICAL - GDPR violation)
- ❌ Private fields exposed in API responses

---

### EC-44: Admin Anonymize Non-Existent User
**Objective**: Verify graceful handling of anonymization for non-existent user
**Primary ATOM**: `[[uc_admin_user_management]]`
**Secondary ATOMs**: `[[rule_gdpr_compliance]]`
**File**: `edge_admin_anonymize_nonexistent.js`
**Scenario Type**: Single Agent (admin)

#### User Journey
1. Admin logs in
2. Admin attempts to anonymize non-existent user
3. System rejects with 404 Not Found
4. Admin anonymizes existing user successfully

#### Validation Points
- ✅ Non-existent user returns 404 Not Found
- ✅ Clear error message about user not found
- ✅ Existing user anonymization succeeds
- ✅ Private data overwritten with "ANONYMIZED"

#### Failure Conditions
- ❌ Anonymization of non-existent user causes crash
- ❌ Unclear error message

---

### EC-45: Admin Soft Delete Non-Existent User
**Objective**: Verify graceful handling of soft delete for non-existent user
**Primary ATOM**: `[[uc_admin_user_management]]`
**Secondary ATOMs**: `[[rule_gdpr_compliance]]`
**File**: `edge_admin_delete_nonexistent.js`
**Scenario Type**: Single Agent (admin)

#### User Journey
1. Admin logs in
2. Admin attempts to soft delete non-existent user
3. System rejects with 404 Not Found
4. Admin soft deletes existing user successfully

#### Validation Points
- ✅ Non-existent user returns 404 Not Found
- ✅ Clear error message about user not found
- ✅ Existing user soft delete succeeds
- ✅ deleted_at timestamp set correctly

#### Failure Conditions
- ❌ Soft delete of non-existent user causes crash
- ❌ Unclear error message

---

## Category 10: WebSocket Edge Cases

### EC-46: WebSocket Connection Without Token
**Objective**: Verify that private channel subscription requires authentication
**Primary ATOM**: `[[api_websocket]]`, `[[api_websocket_user_notifications]]`
**Secondary ATOMs**: `[[req_security_authorization]]`
**File**: `edge_ws_connection_no_token.js`
**Scenario Type**: Unit Test (requires WebSocket client)

#### User Journey
1. Client connects to WebSocket
2. Client attempts to subscribe to private channel without auth
3. Server rejects subscription with error
4. Client authenticates and subscribes successfully

#### Validation Points
- ✅ Private channel without auth rejected
- ✅ Auth signature required
- ✅ Public channels work without auth
- ✅ Authenticated subscription succeeds

#### Failure Conditions
- ❌ Private channel accessible without auth (CRITICAL - security issue)
- ❌ Authentication bypassed

---

### EC-47: WebSocket Subscription to Wrong Channel
**Objective**: Verify that subscription to non-existent or wrong channel is handled
**Primary ATOM**: `[[api_websocket]]`
**Secondary ATOMs**: `[[api_websocket_arena_updates]]`
**File**: `edge_ws_wrong_channel.js`
**Scenario Type**: Unit Test (requires WebSocket client)

#### User Journey
1. Client connects to WebSocket
2. Client attempts to subscribe to non-existent channel
3. Server rejects subscription or handles gracefully
4. Client subscribes to valid channel

#### Validation Points
- ✅ Invalid channel handled gracefully
- ✅ Clear error message or silent rejection
- ✅ Valid channel subscription succeeds
- ✅ No crash on invalid channel

#### Failure Conditions
- ❌ Invalid channel causes crash
- ❌ Subscription to any channel succeeds

---

### EC-48: WebSocket Ping/Pong Timeout
**Objective**: Verify that connection is closed after ping timeout
**Primary ATOM**: `[[api_websocket]]`
**Secondary ATOMs**: `[[req_logging_traceability]]`
**File**: `edge_ws_ping_timeout.js`
**Scenario Type**: Unit Test (requires WebSocket client)

#### User Journey
1. Client connects to WebSocket
2. Client stops sending pings
3. Server closes connection after timeout
4. Client receives close event

#### Validation Points
- ✅ Connection closed after timeout
- ✅ Close event received by client
- ✅ Timeout duration appropriate (typically 60s)
- ✅ Client can reconnect after timeout

#### Failure Conditions
- ❌ Connection never closes (resource leak)
- ❌ No close event sent

---

## Implementation Priority Matrix

| Priority | Category | Test Count | Business Impact | Complexity | Dependencies |
|---|---|---|---|---|---|
| **P0** | Movement Validation | 9 | Critical - prevents game-breaking bugs | Medium | None |
| **P0** | Attack Validation | 10 | Critical - prevents cheating/griefing | Medium | Movement tests |
| **P0** | Authentication | 5 | Critical - security & GDPR | Low | None |
| **P1** | Character/Progression | 6 | High - prevents stat exploitation | Medium | Authentication |
| **P1** | Matchmaking | 4 | High - prevents queue corruption | Medium | Authentication |
| **P2** | Match Resolution | 2 | Medium - prevents stale match states | Medium | Matchmaking |
| **P2** | API/Communication | 4 | Medium - prevents crashes | Low | None |
| **P2** | Leaderboard | 2 | Low - data integrity | Low | None |
| **P3** | Admin | 3 | High - GDPR compliance | Medium | Authentication |
| **P3** | WebSocket | 3 | Medium - connection stability | High | API |

**Total Tests**: 48 edge cases

---

## Unit Test Coverage (Backend Go)

While CLI e2e tests are excellent for API validation, unit tests should cover:

1. **Movement Validation** (`rules_move_test.go`):
   - ✅ Already covered: out of turn, wrong controller, occupied, obstacle, path length, not adjacent
   - Add: jump limitations, grid boundaries

2. **Attack Validation** (`rules_attack_test.go`):
   - ✅ Already covered: out of turn, wrong controller, target not found, target not in range, target not ground
   - Add: friendly fire, skill cooldown, entity targeting rules

3. **Action Economy** (`action_cost_rules`):
   - Add: attack ends movement, action state flags

4. **Progression** (new test file needed):
   - Add: attribute cap, movement gate, negative values, win tracking

5. **Authentication** (new test file needed):
   - Add: password validation, token expiration, admin access

---

## Next Steps

1. **Phase 1 (P0 - Movement & Attack)**: Implement EC-01 through EC-19
   - Create CLI test scripts for movement validation edge cases
   - Create CLI test scripts for attack validation edge cases
   - Update CI report generator to include new tests

2. **Phase 2 (P0 - Authentication)**: Implement EC-20 through EC-24
   - Extend existing password policy test for full coverage
   - Create tests for token expiration and missing token
   - Create test for admin access control

3. **Phase 3 (P1 - Character & Matchmaking)**: Implement EC-25 through EC-34
   - Extend existing reroll and progression tests
   - Create matchmaking edge case tests
   - Add unit tests for progression rules

4. **Phase 4 (P2/P3 - Integration)**: Implement remaining tests
   - Match resolution edge cases
   - API error handling tests
   - WebSocket and admin tests (mostly unit tests)

---

## Notes on Test Implementation

1. **Error Throwing**: Ensure CLI properly throws structured JSON for 4xx/5xx errors
2. **Try-Catch Pattern**: All edge case tests should use try-catch for expected failures
3. **Assertion Pattern**: Use `upsilon.assert(false, "ERROR: ...")` in catch blocks to verify error was thrown
4. **Cleanup**: Use `upsilon.onTeardown()` for account deletion and match cleanup
5. **Multi-Agent Coordination**: Use `syncGroup()` and `setShared()`/`getShared()` for coordination
6. **ATD Links**: Each test file should include `@spec-link` tags for relevant atoms
7. **Naming Convention**: Use `edge_` prefix for edge case tests to distinguish from e2e scenarios

---

## Success Metrics

- **Edge Case Coverage**: 95%+ of identified edge cases implemented
- **Pass Rate**: >90% across all edge case tests
- **Execution Time**: <30 seconds for single-agent tests, <2 minutes for multi-agent
- **Flakiness**: <3% intermittent failures
- **Error Code Accuracy**: 100% of tests verify correct error codes

This comprehensive edge case testing battery ensures the Upsilon Battle API enforces all business rules, handles errors gracefully, and provides clear feedback to clients.
