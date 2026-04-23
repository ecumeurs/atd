# Orphan Atom Categorization

## Summary
- **Total True Orphans**: 144 atoms (no @spec-link tags found in code)
- **Analysis Date**: 2026-04-17
- **Method**: Manual analysis of atom content, status, and relationships

## Category 1: Planned Features (DRAFT Status) - 8 atoms

**Characteristics**: Status = DRAFT, clearly documented as future work

### Customer Requirements (DRAFT)
- `requirement_customer_action_reporting` - Rich action feedback for UI animations
- `requirement_customer_api_first` - API-first approach requirement

### Rules (DRAFT)  
- `rule_pvp_stalemate_draw` - PvP stalemate detection (ISS-029)

### Other DRAFT Atoms
- `rule_friendly_fire_match_type` - Friendly fire by match type
- `script_farm` - Script farming functionality

**Action Items**: These are intentionally unimplemented - keep as DRAFT, prioritize based on product roadmap.

---

## Category 2: Parent/Grouping Atoms (MODULE Type) - 15 atoms

**Characteristics**: Type = MODULE, exist to aggregate child atoms, don't need direct code links

### Architecture Modules
- `module_backend` - Backend module aggregation
- `module_frontend` - Frontend module aggregation  
- `module_game` - Game module aggregation
- `module_ui_tactical_layout` - UI layout module
- `module_backend_action_economy` - Action economy backend
- `module_backend_board_generation` - Board generation backend
- `module_backend_combat_math` - Combat math backend
- `module_backend_initiative_evaluation` - Initiative evaluation backend
- `module_frontend_board_ui_rendering` - Board UI rendering
- `module_frontend_character_entity_creation` - Character entity creation
- `module_frontend_integration_constraint` - Frontend integration constraints
- `module_frontend_matchmaking_orchestration` - Matchmaking orchestration
- `module_frontend_session_management` - Session management

### Requirement Modules  
- `req_matchmaking` - Matchmaking requirements aggregation
- `req_security` - Security requirements aggregation

**Action Items**: These are **NOT true orphans** - parent MODULE atoms should NOT have direct code links. Update ATD system to exclude MODULE types from orphan detection.

---

## Category 3: Implemented But Not Tagged - ~40 atoms

**Characteristics**: Status = STABLE, describe core functionality that likely exists but lacks @spec-link tags

### Mechanics Likely Implemented
- `mech_action_economy_time_constraint_rules` - Turn timeout (30s)
- `mech_ai_name_generation` - AI name generation  
- `mech_board_generation*` - Board generation rules
- `mech_character_reroll_*` - Character reroll mechanics
- `mech_combat_shielding` - Combat shielding
- `mech_entity_properties*` - Entity property systems
- `mech_initiative_*` - Initiative systems
- `mech_move_validation_*` - Movement validation (9 atoms)
- `mech_skill_validation_*` - Skill validation (7 atoms)

### UI/UX Likely Implemented  
- `ui_leaderboard_*` - Leaderboard UI components (4 atoms)
- `ui_dashboard_*` - Dashboard components (5 atoms)
- `ui_registration_*` - Registration flows (3 atoms)
- `ui_board`, `ui_theme`, `ui_holo_obstacle` - Various UI elements

### Requirements Likely Satisfied
- `req_admin_experience` - Admin experience
- `req_logging_traceability` - Logging requirements
- `req_player_experience` - Player experience  
- `req_security_*` - Security requirements (4 atoms)

**Action Items**: Add @spec-link tags to existing code implementations. These are **missing documentation links**, not missing features.

---

## Category 4: Domain/Architecture Specifications - 15 atoms

**Characteristics**: High-level architecture and domain definitions

### Domain Models
- `domain_ruler_state*` - Ruler state domains (5 atoms)
- `domain_skill_system` - Skill system domain
- `domain_upsilon_engine*` - Upsilon engine domains (6 atoms)

### Entity Definitions  
- `entity_character_distribute_remaining_points` - Character point distribution
- `entity_player_*` - Player entity rules (4 atoms)

### Infrastructure
- `infra_mvp_docker` - Docker infrastructure
- `battleui_upsilon_api_service` - API service integration

**Action Items**: These are **architectural specifications** that may not need direct code links. Verify if child atoms have proper implementation coverage.

---

## Category 5: User Stories & Use Cases - 20 atoms

**Characteristics**: High-level user-facing descriptions

### User Stories
- `us_auth_logout` - Logout flow
- `us_character_reroll_*` - Character reroll stories (5 atoms)
- `us_leaderboard_view_*` - Leaderboard viewing (3 atoms)
- `us_queue_selection_*` - Queue selection (5 atoms)
- `us_win_progression_*` - Win progression (5 atoms)
- `us_new_player_onboard` - New player onboarding

### Use Cases
- `uc_admin_history_management` - Admin history
- `uc_auth_logout` - Authentication logout
- `uc_combat_turn` - Combat turn flow
- `uc_matchmaking` - Matchmaking flow
- `uc_player_login` - Player login
- `uc_player_registration` - Player registration  
- `uc_progression_stat_allocation` - Stat allocation

**Action Items**: These are **requirement-level atoms**. Verify that child atoms have proper implementation coverage.

---

## Category 6: API Specifications - 5 atoms

**Characteristics**: API contracts that may be implemented differently

- `api_go_action_feedback` - Action feedback endpoint
- `api_profile_export` - Profile export
- `api_websocket*` - WebSocket endpoints (3 atoms)

**Action Items**: Verify if these APIs exist with different implementations or naming.

---

## Key Insights

### 1. ATD System Issues
- **False Orphans**: ~100 atoms are incorrectly classified as orphans
- **MODULE Type Problem**: Parent grouping atoms should never require direct code links
- **Detection Failure**: Many STABLE atoms are implemented but lack @spec-link tags

### 2. Documentation Health
- **Excellent Coverage**: Most features ARE implemented and documented
- **Linking Gap**: Primary issue is missing @spec-link tags, not missing features
- **Good Hierarchy**: Customer → Architecture → Implementation layers are well-structured

### 3. True Gaps
- **Planned Features**: ~8 DRAFT atoms represent intentional future work
- **Tagging Needed**: ~40 atoms need @spec-link tags added to existing code
- **Architecture Docs**: ~20 domain/spec atoms may need verification

## Recommendations

1. **Fix ATD Orphan Detection**: Exclude MODULE type atoms and REQUIREMENT-level atoms
2. **Add Missing Tags**: Prioritize adding @spec-link tags to STABLE mechanic atoms  
3. **Verify Architecture**: Check domain/spec atoms for child implementation coverage
4. **Plan DRAFT Items**: Prioritize the 8 DRAFT atoms based on product needs
