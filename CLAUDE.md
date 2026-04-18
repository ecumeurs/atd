# UpsilonBattle Development Guide

## What is ATD?

**ATD (Atomic Traceable Documentation)** is a development governance system that ensures your code, documentation, and tests stay perfectly synchronized. It's the single source of truth for "what should this system do?" and "does this implementation match the requirements?"

Think of ATD as your **project's constitution**—it defines the rules that everyone (humans and AI agents) must follow when working on UpsilonBattle.

### Core Philosophy

| Principle | Description |
|---|---|
| **Documentation First** | Every feature starts with a `.atom.md` file, not code |
| **Atomicity** | Each atom describes exactly ONE rule or concept—no "and" or "also" allowed |
| **Bidirectional Links** | Documentation links to code via `@spec-link [[atom_id]]`; code links back via `@test-link [[atom_id]]` |
| **Co-evolution** | Documentation and code evolve together; neither is subordinate |

### The Power Model

```
Customer Need
     ↓
Design Solution (Architecture)
     ↓  
Code Implementation
     ↓
Test Verification
```

Every layer validates the one below it:
- **Customer**: "Does this meet the user need?"
- **Architecture**: "Is this the right technical approach?"  
- **Implementation**: "Did we build it correctly?"

---

## ATD Structure for UpsilonBattle

### Project Configuration
- **Docs Path**: `docs/`
- **Code Paths**: `upsilonapi/`, `upsilonbattle/`, `battleui/`, `upsiloncli/`
- **ATD Tools**: Located at `/home/bastien/work/skill/` (accessed via MCP)

### Type System (Simplified for UpsilonBattle)

| Type | Purpose | Examples |
|---|---|---|
| **REQUIREMENT** | Customer business requirements | `req_matchmaking`, `req_security` |
| **RULE** | Single business rules | `rule_password_policy`, `rule_progression` |
| **MECHANIC** | Implementation algorithms | `mech_initiative`, `mech_action_economy` |
| **API** | Interface contracts | `api_auth_login`, `api_matchmaking` |
| **ENTITY** | Data models | `entity_player`, `entity_character` |
| **MODULE** | Architectural grouping | `module_frontend`, `module_backend` |

### Layer System (3 Tiers)

| Layer | Responsibility | Examples | Link Expectations |
|---|---|---|---|
| **CUSTOMER** | Business requirements | User stories, rules | **No code links** (children link down) |
| **ARCHITECTURE** | System design & APIs | API contracts, UI components | **Links both ways** (to code + from customers) |
| **IMPLEMENTATION** | Algorithms & logic | Mechanics, validation | **Only code links** (links up to architecture) |

---

## Development Workflow at UpsilonBattle

### Phase 1: Discovery & Planning 📋
**Question**: "What do we need to build?"

1. **Explore Existing Documentation**
   ```bash
   # Find atoms related to your feature idea
   mcp__atd__atd_search(query="user authentication", scope="all")
   ```

2. **Review Dependencies**
   ```bash
   # See what atoms depend on what
   mcp__atd__atd_trace(atom="uc_player_login")
   ```

3. **Check Implementation Status**
   ```bash
   # Verify if already implemented
   mcp__atd__atd_stats()
   ```

**Decision Point**: Create new DRAFT atoms or proceed with existing architecture?

### Phase 2: Specification 📝
**Question**: "How should this work?"

1. **Create/Update Atoms**
   ```bash
   # Create new atom with proper structure
   mcp__atd__atd_update(
     file="docs/new_feature.atom.md",
     set=["id=new_auth_flow", "type=REQUIREMENT", "layer=CUSTOMER", "status=DRAFT"],
     intent="To provide secure session management",
     logic="The system must handle JWT tokens with 15-minute expiration"
   )
   ```

2. **Establish Dependencies**
   ```bash
   # Link to parent atoms
   mcp__atd__atd_weave()
   ```

3. **Define Technical Interface**
   - **API Endpoints**: What endpoints will this expose?
   - **Data Models**: What entities are involved?
   - **UI Components**: What screens need updating?

**Decision Point**: Architecture design complete and ready for implementation?

### Phase 3: Implementation 💻
**Question**: "Did we build it correctly?"

1. **Write Code with Traceability**
   ```go
   // Add @spec-link tags directly to implementation
   // @spec-link [[mech_action_economy_action_cost_rules]]
   func (gs *GameState) Attack(msg *message.Message, req rulermethods.ControllerAttack) {
       // implementation
   }
   ```

2. **Link Code to Architecture**
   - Place `@spec-link [[api_auth_login]]` above controller methods
   - Place `@spec-link [[ui_login_form]]` above Vue components
   - **Use specific atoms**: Not generic ones

3. **Follow Existing Patterns**
   - Copy structure from similar implementations
   - Use established error handling patterns
   - Follow existing authentication/authorization flows

**Decision Point**: Implementation complete and tests passing?

### Phase 4: Verification & Testing ✅
**Question**: "Does this actually work?"

1. **Manual Testing**
   - Test the feature manually via `upsiloncli`
   - Verify all user flows work end-to-end

2. **Create Tests with Traceability**
   ```go
   // Add @test-link tags
   // @test-link [[uc_player_login]]
   func TestPlayerLogin(t *testing.T) {
       // test implementation
   }
   ```

3. **Verify Documentation Coverage**
   ```bash
   # Check if implementation has proper @spec-link coverage
   mcp__atd__atd_trace(atom="mech_action_economy_action_cost_rules")
   ```

4. **Update Atom Status**
   ```bash
   # Mark atom as stable once implemented
   mcp__atd__atd_update(
     file="docs/mech_action_economy_action_cost_rules.atom.md",
     set=["status=STABLE"]
   )
   ```

**Decision Point**: Ready for production?

---

## ATD Tool Usage Guide

### For Feature Development

#### Finding Documentation
```bash
# Find atoms by type
mcp__atd__atd_query(field="type", search="MECHANIC")

# Find atoms by layer
mcp__atd__atd_query(field="layer", search="ARCHITECTURE")

# Semantic search
mcp__atd__atd_search(query="turn timer implementation", scope="all")
```

#### Creating New Atoms
```bash
# Create new atom (agent: Architect mode)
mcp__atd__atd_update(
  file="docs/feature_name.atom.md",
  set=["id=feature_name", "type=MECHANIC", "layer=IMPLEMENTATION", "status=DRAFT"],
  intent="To implement X functionality",
  logic="The algorithm must process Y when Z happens"
)

# Set dependencies (agent: Architect mode)
mcp__atd__atd_weave()
```

#### Checking Coverage
```bash
# Check what's implemented
mcp__atd__atd_stats()

# Check specific atom coverage
mcp__atd__atd_trace(atom="your_feature_atom")
```

### For Bug Fixing

#### Finding Related Atoms
```bash
# Search for atoms related to bug area
mcp__atd__atd_search(query="authentication error handling", scope="all")

# Trace dependencies
mcp__atd__atd_trace(atom="req_security_authorization")
```

#### Understanding Impact
```bash
# Check blast radius before changing
mcp__atd__atd_crawl()

# Find all code using the atom
grep -r "@spec-link \[\[atom_name\]\]" --include="*.go"
```

#### Updating Code & Docs
```bash
# Fix implementation
# (make code changes)

# Update atom if behavior changed
mcp__atd__atd_update(
  file="docs/affected_atom.atom.md",
  set=["logic=New behavior after fix", "status=STABLE"]
)
```

### For Code Review

#### Checking Compliance
```bash
# Verify changes match documentation
mcp__atd__atd_verify()

# Check if new code violates existing atoms
mcp__atd__atd_audit()
```

#### Finding Orphans
```bash
# Find atoms with no implementation
mcp__atd__atd_crawl(gaps=true)

# Check if removed code has orphaned atoms
mcp__atd__atd_lint()
```

---

## Common Patterns & Best Practices

### DO ✅

**1. Always start with ATD**
- Never implement without first creating/updating documentation
- Use `atd_query` to find related atoms before starting

**2. Be specific with @spec-link placement**
```go
// GOOD: Specific function
// @spec-link [[mech_combat_standard_attack_computation]]
func computeDamage(attacker, defender) int {
    return max(1, attacker.attack - defender.defense)
}

// BAD: File-level tag
// @spec-link [[mech_combat_standard_attack_computation]]
package combat
```

**3. Use appropriate atom types**
- **Rules**: Single business constraints (`rule_password_policy`)
- **Mechanics**: Complex algorithms (`mech_initiative`)
- **APIs**: Interface contracts (`api_auth_login`)

**4. Update atom status progressively**
- DRAFT → REVIEW → STABLE
- Update status after implementation, after testing, and after review

**5. Weave dependencies after creating atoms**
```bash
# Always run after creating new atoms
mcp__atd__atd_weave()
```

### DON'T ❌

**1. Don't skip documentation for "simple" features**
- Even "add a button" needs ATD coverage
- Future you will thank yourself for the traceability

**2. Don't use file-level @spec-link tags**
- Place tags directly above the specific function/class being implemented
- File-level tags make it impossible to trace what code implements what

**3. Don't create overly broad atoms**
- "And" in intent = split into multiple atoms
- One atom = one state-changing rule

**4. Don't ignore atom status**
- DRAFT means "not ready yet"
- STABLE means "production-ready"
- Don't implement DRAFT atoms without reviewing

**5. Don't break existing @spec-link chains**
- When refactoring, maintain existing links
- If you need to change an atom, update it rather than creating new ones

---

## Project-Specific Conventions

### Naming Patterns
- **Authentication**: `uc_auth_*` for use cases, `api_auth_*` for endpoints
- **Matchmaking**: `uc_matchmaking`, `api_matchmaking`
- **Combat**: `uc_combat_turn`, `mech_*`
- **Progression**: `uc_progression_*`, `rule_progression`
- **Characters**: `uc_player_registration`, `us_character_reroll`
- **UI**: `ui_*` (components, screens, flows)

### Atom Status Workflow
```
1. Developer creates DRAFT atom
2. Developer implements feature
3. Developer writes tests with @test-link
4. Developer requests review
5. Architect changes status to REVIEW
6. Developer approves and changes to STABLE
```

### File Organization
```
docs/
├── customer/           # Business requirements, user stories
├── architecture/         # System design, APIs, data models  
└── implementation/      # Mechanics, algorithms, validation rules
```

---

## Getting Started with ATD

### First Time Setup
```bash
# 1. Initialize ATD in your project
cd /home/bastien/work/upsilon/projbackend
atd init

# 2. Verify configuration
cat .atd
```

### Daily Development Workflow
```bash
# 1. Check what you're working on
git status

# 2. Find related documentation
mcp__atd__atd_search(query="current task")

# 3. Plan your approach
# (Create atoms? Update existing? Link to code?)

# 4. Execute the workflow
# (Follow the phases: Discovery → Specification → Implementation → Verification)
```

---

## Troubleshooting Common Issues

### "Atom not found" errors
**Cause**: Typo in @spec-link tag or atom not created yet  
**Solution**: 
```bash
# Verify atom exists
mcp__atd__atd_query(field="id", search="your_atom_id")

# Check spelling in existing atoms
mcp__atd__atd_query(field="id", search="partial_match")
```

### "No code links found" in atd_trace
**Cause**: @spec-link tags not placed in code or ATD index out of date  
**Solution**:
```bash
# Rebuild ATD index
atd_index --force

# Verify links exist
grep -r "@spec-link \[\[atom_name\]\]" --include="*.go"
```

### Circular dependency errors
**Cause**: Atom A depends on B, B depends on A  
**Solution**:
```bash
# Check for circular dependencies
mcp__atd__atd_lint()

# Fix by removing one dependency
mcp__atd__atd_update(
  file="docs/atom_a.atom.md",
  set=["dependents=[]"]  # Remove circular reference
)
```

---

## Integration with CI/CD

### Pre-Commit Checks
```bash
# 1. Check coverage of atoms you're changing
mcp__atd__atd_trace(atom="atom_youre_modifying")

# 2. Verify tests exist
grep -r "@test-link \[\[atom_youre_modifying\]\]" tests/

# 3. Check for broken links
mcp__atd__atd_lint()
```

### Post-Commit Updates
```bash
# 1. If implementation added, mark atom STABLE
mcp__atd__atd_update(
  file="docs/implemented_feature.atom.md",
  set=["status=STABLE"]
)

# 2. If tests added, verify coverage
mcp__atd__atd_test_links(atom="implemented_feature")
```

---

## Success Metrics

### What "Good ATD Usage" Looks Like

- **Documentation Coverage**: 100% of features have atoms
- **Implementation Coverage**: 100% of code has @spec-link tags
- **Test Coverage**: 100% of critical features have tests
- **Traceability**: Perfect chain from customer requirement to code to test

### Current Project Status (2026-04-17)

- **Documentation Coverage**: 243 atoms created
- **Implementation Coverage**: 421 @spec-link tags (82% accurate)
- **Test Coverage**: 40+ test files with @test-link tags
- **True Orphans**: Only 8 atoms (3% intentionally unimplemented)

### Goals for Next Quarter

- **Reduce Implementation Atoms**: Move simple mechanics to DESIGN layer (-40% target)
- **Improve Agent Guidance**: Add Claude Code specific patterns to ATD.md
- **Enhance CI Integration**: Automated BRD compliance testing via customer scenarios
- **Fix ATD Tooling**: Resolve indexing and orphan detection issues

---

## Quick Reference

### Essential Commands
```bash
# Find documentation
mcp__atd__atd_query(field="type", search="RULE")

# Create new atom
mcp__atd__atd_update(file="docs/new.atom.md", set=["id=new", "type=RULE", "layer=CUSTOMER"])

# Check implementation coverage
mcp__atd__atd_trace(atom="your_atom_id")

# Find orphaned atoms
mcp__atd__atd_crawl(gaps=true)

# Rebuild index
atd_index --force

# Run health check
mcp__atd__atd_stats()
```

### Key ATD Concepts
- **@spec-link [[atom_id]]**: Links code to documentation
- **@test-link [[atom_id]]**: Links tests to documentation  
- **Atomicity**: One rule per atom
- **Bidirectional**: Documentation ↔ Code ↔ Tests

---

## Conclusion

ATD is your **project's foundation**—not just documentation, but a living system that governs how UpsilonBattle evolves. When used correctly, it ensures:

✅ **Clear requirements** through customer stories and rules  
✅ **Solid architecture** through well-designed APIs and data models  
✅ **Correct implementation** through traceable code with @spec-link tags  
✅ **Verified quality** through tests with @test-link coverage  
✅ **Easy maintenance** through clear dependency graphs and status tracking

When ATD and code are in sync, development is **predictable, verifiable, and maintainable**. When they're out of sync, you have **immediate visibility** into what's broken and why.

**Start every feature with ATD, link every implementation with @spec-link, and you'll have a project that scales gracefully with confidence.**