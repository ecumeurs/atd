# Issue: CLAUDE.md Project Context Mismatch

**ID:** `20260418_claude_md_project_context_mismatch`
**Ref:** `ISS-078`
**Date:** 2026-04-18
**Severity:** Medium
**Status:** Open
**Component:** `CLAUDE.md`, documentation system
**Affects:** Agent Understanding, Project Context Accuracy, Development Workflow

---

## Summary

Current CLAUDE.md was copied from upsilon-hub and describes UpsilonBattle development using ATD tools, but this project IS the ATD project itself. The documentation needs complete overhaul to reflect the correct project context.

---

## Technical Description

### Background
CLAUDE.md should provide accurate project context for agents working on this codebase, describing the project structure, conventions, and development workflow.

### The Problem Scenario
1. **Project Identity Mismatch**: CLAUDE.md describes UpsilonBattle using ATD, but this IS the ATD project
2. **Incorrect Structure References**: References to upsilonapi/, upsilonbattle/, battleui/ don't exist in this project
3. **Wrong Development Focus**: Describes game development instead of tool development
4. **Missing ATD Development Context**: No guidance for working on ATD tools themselves

### Where This Pattern Exists Today
- **Current CLAUDE.md**: Copy-pasted from upsilon-hub with game development context
- **Project Structure**: This project contains atd/, extension/, docs/ not mentioned in CLAUDE.md
- **Tool Development**: No guidance for working on ATD CLI tools, MCP server, or VS Code extension

### Evidence from Investigation
- **True Project**: ATD (Atomic Traceable Documentation) framework and tooling
- **Current Content**: Describes UpsilonBattle game development
- **Impact**: Agents have wrong context about project purpose and structure

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Current CLAUDE.md content is clearly wrong for this project) |
| Impact if triggered | Medium (Agents confused about project purpose, inefficient workflows) |
| Detectability | High (Obvious mismatch between CLAUDE.md and actual project structure) |
| Current mitigant | Manual project context correction during agent interactions |

---

## Recommended Fix

**Short term**: Complete CLAUDE.md rewrite to accurately describe this as the ATD project. Include correct project structure: atd/ (core tools), extension/ (VS Code extension), docs/ (ATD documentation), .agent/ (agent rules).

**Medium term**: Add ATD development workflow guidance: tool development, MCP integration, extension development, documentation updates. Include testing, build processes, and deployment instructions.

**Long term**: Integrate with ATD.md for tool-specific guidance. Create project-specific conventions and patterns for ATD development itself. Establish CI/CD and testing standards.

---

## References

- [Current CLAUDE.md](file:///home/bastien/work/skill/CLAUDE.md)
- [Project Structure](file:///home/bastien/work/skill/)
- [ATD.md](file:///home/bastien/work/skill/ATD.md)