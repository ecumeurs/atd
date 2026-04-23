# Customer Requirements CI Scenarios

**Purpose**: Comprehensive E2E test scenarios for validating customer requirements via upsiloncli bot scripts  
**Based on**: ATD investigation and customer layer analysis (2026-04-17)  
**Integration**: Designed for `.github/workflows/e2e-battles.yml`  

## Scenario Design Principles

1. **Customer-First**: Each scenario validates a specific customer-facing requirement
2. **Atomic Validation**: Each scenario tests one primary requirement atom  
3. **BRD Alignment**: Scenarios map directly to BRD sections
4. **Multi-Agent Support**: Designed for both single and multi-agent testing
5. **Deterministic Outcomes**: Clear pass/fail assertions

---

## Priority 1: Critical Customer Journeys

### CR-01: Complete New Player Onboarding
**BRD Section**: 2.1 User Onboarding & Identity  
**Primary ATOM**: [[uc_player_registration]]  
**Secondary ATOMs**: [[us_new_player_onboard]], [[entity_player]], [[rule_password_policy]]  
**Scenario Type**: Single Agent  
**File**: `e2e_customer_onboarding.js`

#### User Journey
1. Player registers with valid credentials (name, email, password, address, birth date)
2. System creates account with initial 3-character roster
3. Player receives JWT token and can access dashboard
4. Player views initial character stats

#### Validation Points
- ✅ Registration succeeds with compliant password (15+ chars, uppercase, number, symbol)
- ✅ Account created with exactly 3 characters
- ✅ Each character has base stats (3 HP, 1 Move, 1 Attack, 1 Def) + 4 random points
- ✅ JWT token issued and valid for authentication
- ✅ Dashboard accessible with token
- ✅ Initial reroll count = 0

#### Failure Conditions
- ❌ Registration succeeds with non-compliant password
- ❌ Account created with ≠ 3 characters
- ❌ Character stats don't meet base requirements
- ❌ No JWT token issued
- ❌ Dashboard inaccessible

---

### CR-02: Player Login & Session Management
**BRD Section**: 2.1 User Onboarding & Identity  
**Primary ATOM**: [[uc_player_login]]  
**Secondary ATOMs**: [[api_auth_login]], [[requirement_req_ui_session_timeout]], [[api_standard_envelope]]  
**Scenario Type**: Single Agent  
**File**: `e2e_customer_login.js`

#### User Journey
1. Existing player attempts login with valid credentials
2. System validates credentials and issues new JWT token
3. Player accesses protected endpoints using token
4. Token expires after 15 minutes
5. System handles session timeout gracefully

#### Validation Points
- ✅ Login succeeds with correct credentials
- ✅ Login fails with incorrect credentials
- ✅ JWT token issued with 15-minute expiration
- ✅ Protected endpoints accessible with valid token
- ✅ Protected endpoints reject expired/invalid tokens
- ✅ Session timeout triggers appropriate UI feedback

#### Failure Conditions
- ❌ Login succeeds with incorrect credentials
- ❌ No JWT token issued on successful login
- ❌ Token doesn't expire after 15 minutes
- ❌ Protected endpoints accessible without token

---

### CR-03: Character Reroll Mechanics
**BRD Section**: 2.1 User Onboarding & Identity (frictionless entry)  
**Primary ATOM**: [[us_character_reroll]]  
**Secondary ATOMs**: [[uc_player_registration]], [[mech_character_reroll_limit]], [[rule_character_create_character]]  
**Scenario Type**: Single Agent  
**File**: `e2e_character_reroll.js`

#### User Journey
1. New player registers and receives initial character roster
2. Player chooses to reroll character stats (maximum 3 times)
3. System generates new random stat distribution
4. Reroll count increments appropriately
5. After 3 rerolls, further rerolls are blocked
6. After first match, rerolls are permanently blocked

#### Validation Points
- ✅ Initial reroll available (count = 0)
- ✅ Reroll generates new valid stat distribution
- ✅ Reroll count increments (1, 2, 3)
- ✅ Reroll blocked after 3 attempts
- ✅ Reroll blocked after first match participation
- ✅ Each character can be individually rerolled

#### Failure Conditions
- ❌ More than 3 rerolls allowed
- ❌ Rerolls allowed after match participation
- ❌ Invalid stat distributions generated
- ❌ Reroll count not tracked correctly

---

## Priority 2: Core Gameplay Features

### CR-04: Matchmaking Flow (PvE Instant)
**BRD Section**: 2.3 Matchmaking Ecosystem  
**Primary ATOM**: [[uc_matchmaking]]  
**Secondary ATOMs**: [[us_queue_selection]], [[req_matchmaking]], [[api_matchmaking]]  
**Scenario Type**: Single Agent  
**File**: `e2e_matchmaking_pve_instant.js`

#### User Journey
1. Player selects PvE mode from queue selection
2. Player views current win/loss record
3. Player joins queue
4. System instantly starts match against AI
5. Player enters battle arena with full state

#### Validation Points
- ✅ Queue selection shows available modes
- ✅ Win/loss record displayed in queue UI
- ✅ PvE queue provides instant match start
- ✅ Match created with correct game mode (1v1_PVE)
- ✅ Player enters battle with proper character roster
- ✅ AI opponents properly configured

#### Failure Conditions
- ❌ Queue selection doesn't show available modes
- ❌ Win/loss record not displayed
- ❌ PvE queue doesn't provide instant start
- ❌ Match created with wrong game mode

---

### CR-05: Matchmaking Flow (PvP Queue)
**BRD Section**: 2.3 Matchmaking Ecosystem  
**Primary ATOM**: [[uc_matchmaking]]  
**Secondary ATOMs**: [[us_queue_selection]], [[req_matchmaking_pve_pvp_transition]], [[rule_matchmaking_single_queue]]  
**Scenario Type**: Multi-Agent (2-4 agents)  
**File**: `e2e_matchmaking_pvp_queue.js`

#### User Journey
1. Multiple players select PvP mode
2. Players view current win/loss records
3. Players join queue simultaneously
4. System matches players based on game mode
5. All players enter same battle arena
6. Players can leave queue before matching

#### Validation Points
- ✅ Multiple players can join same queue
- ✅ System matches players correctly (1v1 or 2v2)
- ✅ All matched players enter same arena
- ✅ Win/loss records displayed before matching
- ✅ Players can leave queue before match
- ✅ Queue properly handles player disconnects

#### Failure Conditions
- ❌ Players end up in different matches
- ❌ Wrong number of players matched
- ❌ Players can't leave queue
- ❌ Queue doesn't handle disconnects

---

### CR-06: Combat Turn Management
**BRD Section**: 2.4 Combat Engine & Turn Management  
**Primary ATOM**: [[uc_combat_turn]]  
**Secondary ATOMs**: [[us_take_combat_turn]], [[mech_initiative]], [[mech_action_economy]], [[rule_turn_clock]]  
**Scenario Type**: Multi-Agent (2-4 agents)  
**File**: `e2e_combat_turn_management.js`

#### User Journey
1. Match starts with initiative-based turn order
2. Current player has 30-second turn timer
3. Player executes valid action (move, attack, or pass)
4. Turn passes to next entity in initiative order
5. If player times out, auto-pass with penalty
6. Match continues until win condition met

#### Validation Points
- ✅ Initiative order calculated and displayed
- ✅ 30-second timer starts on each turn
- ✅ Valid actions (move/attack/pass) executed correctly
- ✅ Turn passes to correct next entity
- ✅ Timeout triggers auto-pass with +400 delay penalty
- ✅ Match ends when win condition met

#### Failure Conditions
- ❌ Initiative order incorrect
- ❌ Timer doesn't start or doesn't timeout
- ❌ Invalid actions allowed
- ❌ Turn doesn't pass correctly
- ❌ Timeout doesn't trigger penalty
- ❌ Match doesn't end on win condition

---

### CR-07: Friendly Fire Prevention
**BRD Section**: 2.4 Combat Engine & Turn Management  
**Primary ATOM**: [[rule_friendly_fire]]  
**Secondary ATOMs**: [[uc_combat_turn]], [[rule_friendly_fire_team_validation]], [[rule_friendly_fire_match_type]]  
**Scenario Type**: Multi-Agent (2v2)  
**File**: `e2e_friendly_fire_prevention.js`

#### User Journey
1. 2v2 match starts with players on same team
2. Player attempts to attack teammate
3. System blocks friendly fire attack
4. Player attempts to move to teammate's position
5. System prevents collision with teammate
6. Match continues with only valid attacks

#### Validation Points
- ✅ Attacks on teammates are blocked
- ✅ Movement to teammate position is blocked
- ✅ Friendly fire prevention applies to all team configurations
- ✅ Error messages clearly indicate friendly fire violation
- ✅ Team identification is accurate
- ✅ Match continues normally after blocked actions

#### Failure Conditions
- ❌ Attacks on teammates succeed
- ❌ Movement to teammate position allowed
- ❌ Team identification incorrect
- ❌ No error messages for violations

---

### CR-08: Match Resolution (Standard)
**BRD Section**: 2.4 Combat Engine & Turn Management  
**Primary ATOM**: [[uc_match_resolution]]  
**Secondary ATOMs**: [[rule_forfeit_battle]], [[us_win_progression]], [[mech_combat_standard_attack_computation]]  
**Scenario Type**: Multi-Agent (2-4 agents)  
**File**: `e2e_match_resolution_standard.js`

#### User Journey
1. Match progresses through normal combat
2. One team eliminates all opposing characters
3. System detects win condition
4. Match ends with proper winner declaration
5. Winning players receive progression rewards
6. Match results recorded in history

#### Validation Points
- ✅ Win condition detected when opposing team eliminated
- ✅ Match ends immediately on win condition
- ✅ Winner declared correctly (team ID)
- ✅ Winning players receive +1 attribute point
- ✅ Match results recorded with correct metadata
- ✅ Losers receive no progression rewards

#### Failure Conditions
- ❌ Win condition not detected
- ❌ Match doesn't end on win condition
- ❌ Wrong winner declared
- ❌ Progression rewards not awarded
- ❌ Match results not recorded

---

### CR-09: Match Resolution (Forfeit)
**BRD Section**: 2.4 Combat Engine & Turn Management  
**Primary ATOM**: [[uc_match_resolution]]  
**Secondary ATOMs**: [[rule_forfeit_battle]], [[api_go_battle_action]]  
**Scenario Type**: Multi-Agent (2-4 agents)  
**File**: `e2e_match_resolution_forfeit.js`

#### User Journey
1. Match in progress
2. Player chooses to forfeit during their turn
3. System processes forfeit immediately
4. Match ends with opposing team as winner
5. Forfeiting player receives loss
6. Match results recorded appropriately

#### Validation Points
- ✅ Forfeit only allowed during player's turn
- ✅ Forfeit ends match immediately
- ✅ Opposing team declared winner
- ✅ Forfeiting player receives loss
- ✅ Match results recorded with forfeit status
- ✅ No progression rewards for forfeiting player

#### Failure Conditions
- ❌ Forfeit allowed outside player's turn
- ❌ Forfeit doesn't end match
- ❌ Wrong winner declared
- ❌ Match status not recorded correctly

---

## Priority 3: Progression & Rewards

### CR-10: Character Progression (Post-Win)
**BRD Section**: 2.5 Character Progression  
**Primary ATOM**: [[uc_progression_stat_allocation]]  
**Secondary ATOMs**: [[rule_progression]], [[us_win_progression]], [[us_win_progression_win_alloc_point]]  
**Scenario Type**: Multi-Agent (2 agents)  
**File**: `e2e_progression_post_win.js`

#### User Journey
1. Two players compete in match
2. Winner receives +1 attribute point
3. Winner allocates point to character stat
4. System validates allocation against progression rules
5. Loser receives no points
6. Both players can view updated stats

#### Validation Points
- ✅ Winner receives exactly +1 attribute point
- ✅ Allocation allowed: total stats ≤ 10 + total_wins
- ✅ Movement upgrade only allowed every 5 wins
- ✅ Loser receives no attribute points
- ✅ Character stats updated correctly
- ✅ Progression rules enforced consistently

#### Failure Conditions
- ❌ Winner receives wrong number of points
- ❌ Allocation exceeds cap (10 + wins)
- ❌ Movement upgrade allowed outside 5-win milestone
- ❌ Loser receives points
- ❌ Stats not updated correctly

---

### CR-11: Progression Constraints
**BRD Section**: 2.5 Character Progression  
**Primary ATOM**: [[rule_progression]]  
**Secondary ATOMs**: [[uc_progression_stat_allocation]], [[us_win_progression_stat_reflection]]  
**Scenario Type**: Single Agent  
**File**: `e2e_progression_constraints.js`

#### User Journey
1. Player with multiple wins attempts various upgrades
2. Player tries to exceed stat cap (10 + wins)
3. Player tries to upgrade movement outside 5-win milestone
4. Player tries to upgrade without winning matches
5. System enforces all progression constraints
6. Player can only make valid upgrades

#### Validation Points
- ✅ Stat cap enforced: total ≤ 10 + total_wins
- ✅ Movement upgrade locked to every 5 wins
- ✅ No upgrades allowed without wins
- ✅ Clear error messages for constraint violations
- ✅ Valid upgrades processed successfully
- ✅ Progression reflected in character stats

#### Failure Conditions
- ❌ Stat cap not enforced
- ❌ Movement upgrade allowed outside milestone
- ❌ Upgrades allowed without wins
- ❌ No error messages for violations

---

### CR-12: Leaderboard Viewing
**BRD Section**: 4. UI & Dashboard  
**Primary ATOM**: [[us_leaderboard_view]]  
**Secondary ATOMs**: [[ui_leaderboard]], [[api_leaderboard]], [[rule_leaderboard_score_calculation]]  
**Scenario Type**: Single Agent  
**File**: `e2e_leaderboard_viewing.js`

#### User Journey
1. Player accesses leaderboard from dashboard
2. Player selects game mode (1v1_PVP, 2v2_PVP, etc.)
3. System displays ranked players
4. Player can see their own ranking
5. Leaderboard shows wins, losses, and scores
6. Leaderboard updates with recent match results

#### Validation Points
- ✅ Leaderboard accessible from dashboard
- ✅ Game mode selection works correctly
- ✅ Players ranked by score (descending)
- ✅ Current player's ranking highlighted
- ✅ Wins, losses, and scores displayed correctly
- ✅ Leaderboard reflects recent match results

#### Failure Conditions
- ❌ Leaderboard not accessible
- ❌ Game mode selection doesn't work
- ❌ Players not ranked correctly
- ❌ Current player not highlighted
- ❌ Data not displayed correctly

---

## Priority 4: Security & Privacy

### CR-13: Password Policy Enforcement
**BRD Section**: 3.1 Security & Access Control  
**Primary ATOM**: [[rule_password_policy]]  
**Secondary ATOMs**: [[req_security]], [[uc_player_registration]]  
**Scenario Type**: Single Agent  
**File**: `e2e_password_policy.js`

#### User Journey
1. Player attempts registration with various password strengths
2. System validates password against policy requirements
3. Registration rejected for non-compliant passwords
4. Registration accepted for compliant passwords
5. Clear error messages guide password requirements

#### Validation Points
- ✅ Password < 15 characters rejected
- ✅ Password without uppercase rejected
- ✅ Password without number rejected
- ✅ Password without special symbol rejected
- ✅ Compliant passwords (15+ chars, uppercase, number, symbol) accepted
- ✅ Clear error messages for each violation

#### Failure Conditions
- ❌ Weak passwords accepted
- ❌ Strong passwords rejected
- ❌ No error messages for violations
- ❌ Error messages unclear or misleading

---

### CR-14: GDPR Data Portability
**BRD Section**: 3.2 GDPR & Data Privacy  
**Primary ATOM**: [[api_profile_export]]  
**Secondary ATOMs**: [[rule_gdpr_compliance]], [[requirement_customer_player_profile]]  
**Scenario Type**: Single Agent  
**File**: `e2e_gdpr_portability.js`

#### User Journey
1. Authenticated player requests data export
2. System compiles complete player data
3. System returns machine-readable JSON export
4. Export includes all personal data and match history
5. Player can download and verify export completeness

#### Validation Points
- ✅ Export endpoint requires authentication
- ✅ Export includes all personal data (name, email, address, birth date)
- ✅ Export includes character roster and stats
- ✅ Export includes complete match history
- ✅ Export format is valid JSON
- ✅ Export data matches current database state

#### Failure Conditions
- ❌ Export accessible without authentication
- ❌ Export missing personal data
- ❌ Export missing character data
- ❌ Export missing match history
- ❌ Export format invalid

---

### CR-15: Admin User Management
**BRD Section**: 2.2 System Administration & Management  
**Primary ATOM**: [[uc_admin_user_management]]  
**Secondary ATOMs**: [[uc_admin_login]], [[rule_admin_access_restriction]], [[rule_gdpr_compliance]]  
**Scenario Type**: Single Agent (admin role)  
**File**: `e2e_admin_user_management.js`

#### User Journey
1. Admin authenticates with admin credentials
2. Admin accesses user management dashboard
3. Admin views list of all users (excluding private data)
4. Admin performs soft delete on user account
5. Admin anonymizes user's private data (GDPR)
6. Admin cannot view private user data

#### Validation Points
- ✅ Admin login requires admin role
- ✅ User list excludes private fields (address, birth date)
- ✅ Soft delete sets deleted_at timestamp
- ✅ Anonymization overwrites private data with "ANONYMIZED"
- ✅ Admin cannot view private data through any endpoint
- ✅ Audit trail maintained for admin actions

#### Failure Conditions
- ❌ Non-admin can access admin functions
- ❌ Private data exposed in user list
- ❌ Soft delete doesn't work correctly
- ❌ Anonymization doesn't overwrite data
- ❌ Admin can view private data

---

## Priority 5: Advanced Features

### CR-16: Session Timeout Handling
**BRD Section**: 3.4 API-First & Developer Experience  
**Primary ATOM**: [[requirement_req_ui_session_timeout]]  
**Secondary ATOMs**: [[api_auth_login]], [[mech_sanctum_token_renewal]]  
**Scenario Type**: Single Agent  
**File**: `e2e_session_timeout.js`

#### User Journey
1. Player logs in and receives JWT token
2. Player performs actions with valid token
3. Token expires after 15 minutes
4. Player attempts action with expired token
5. System handles timeout gracefully with modal
6. Player can re-authenticate and continue

#### Validation Points
- ✅ Token expires after exactly 15 minutes
- ✅ Actions with expired token are rejected
- ✅ Clear timeout message displayed to user
- ✅ Session timeout modal appears appropriately
- ✅ Re-authentication flow works smoothly
- ✅ User can continue after re-authentication

#### Failure Conditions
- ❌ Token doesn't expire
- ❌ Expired tokens still accepted
- ❌ No timeout message displayed
- ❌ Re-authentication doesn't work

---

### CR-17: API Self-Discovery
**BRD Section**: 3.4 API-First & Developer Experience  
**Primary ATOM**: [[requirement_customer_api_first]]  
**Secondary ATOMs**: [[api_help_endpoint]], [[api_laravel_gateway]]  
**Scenario Type**: Single Agent  
**File**: `e2e_api_discovery.js`

#### User Journey
1. Developer/access client requests API help endpoint
2. System returns comprehensive API registry
3. Registry includes all available endpoints
4. Each endpoint includes HTTP method, URI, and description
5. Registry is machine-readable for automation
6. Registry stays current with API changes

#### Validation Points
- ✅ Help endpoint accessible without authentication
- ✅ Registry includes all game-critical endpoints
- ✅ Each endpoint has method, URI, and description
- ✅ Registry format is consistent (JSON/structured)
- ✅ Registry reflects current API state
- ✅ Registry can be parsed programmatically

#### Failure Conditions
- ❌ Help endpoint not accessible
- ❌ Registry missing endpoints
- ❌ Endpoint information incomplete
- ❌ Registry format inconsistent
- ❌ Registry out of sync with API

---

## Implementation Priority Matrix

| Priority | Scenario | Complexity | Business Impact | Dependencies |
|---|---|---|---|---|
| **P0** | CR-01: New Player Onboarding | Medium | Critical | None |
| **P0** | CR-02: Player Login & Session | Low | Critical | CR-01 |
| **P0** | CR-03: Character Reroll | Low | High | CR-01 |
| **P1** | CR-04: Matchmaking PvE | Medium | High | CR-02 |
| **P1** | CR-05: Matchmaking PvP | High | High | CR-02 |
| **P1** | CR-06: Combat Turn Management | High | Critical | CR-04, CR-05 |
| **P1** | CR-07: Friendly Fire Prevention | Medium | High | CR-06 |
| **P1** | CR-08: Match Resolution Standard | Medium | Critical | CR-06 |
| **P1** | CR-09: Match Resolution Forfeit | Low | High | CR-06 |
| **P2** | CR-10: Character Progression | Medium | High | CR-08 |
| **P2** | CR-11: Progression Constraints | Medium | High | CR-10 |
| **P2** | CR-12: Leaderboard Viewing | Low | Medium | CR-08 |
| **P2** | CR-13: Password Policy | Low | Critical | CR-01 |
| **P3** | CR-14: GDPR Portability | Low | Medium | CR-01 |
| **P3** | CR-15: Admin User Management | Medium | Medium | Admin setup |
| **P3** | CR-16: Session Timeout | Low | Medium | CR-02 |
| **P3** | CR-17: API Self-Discovery | Low | Low | None |

---

## CI Integration Notes

### Workflow Integration
Each scenario should be added to `.github/workflows/e2e-battles.yml` with:
- Descriptive name referencing BRD section
- Timeout appropriate for scenario complexity
- Clear pass/fail criteria
- Proper agent coordination for multi-agent scenarios

### Script Naming Convention
- Prefix: `e2e_` (end-to-end)
- Customer-focused: descriptive business language
- Suffix: `.js` for upsiloncli bot scripts

### Assertion Patterns
```javascript
// Success assertion
upsilon.assert(condition, "Failure message");

// Value assertion  
upsilon.assertEquals(actual, expected, "Values don't match");

// State assertion
upsilon.assertState(expectedState, "Current state incorrect");
```

### Multi-Agent Coordination
```javascript
// Shared state for coordination
const sharedKey = "scenario_sync_key";
upsilon.setShared(sharedKey, value);
const value = upsilon.getShared(sharedKey);

// Role-based behavior
const role = upsilon.getContext("agent_index");
if (role === "0") {
    // Primary agent behavior
} else {
    // Secondary agent behavior
}
```

---

## Success Metrics

### Coverage Goals
- **P0 Scenarios**: 100% implemented (critical customer journeys)
- **P1 Scenarios**: 80% implemented (core gameplay features)
- **P2 Scenarios**: 60% implemented (progression & rewards)
- **P3 Scenarios**: 40% implemented (advanced features)

### Quality Metrics
- **Scenario Pass Rate**: >95% across all scenarios
- **Execution Time**: <30 seconds for single-agent, <2 minutes for multi-agent
- **Flakiness**: <5% intermittent failures
- **Maintenance**: <2 hours per month for scenario updates

---

## Next Steps

1. **Immediate**: Implement P0 scenarios (CR-01, CR-02, CR-03)
2. **Short-term**: Implement P1 scenarios for core gameplay validation
3. **Medium-term**: Implement P2 scenarios for progression testing
4. **Long-term**: Implement P3 scenarios for advanced feature coverage

This comprehensive scenario library provides end-to-end validation of all customer requirements identified in the ATD investigation, ensuring the system delivers on its business promises.