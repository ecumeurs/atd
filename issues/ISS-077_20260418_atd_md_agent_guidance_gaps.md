# Issue: ATD.md Missing Agent-Specific Integration Guidance

**ID:** `20260418_atd_md_agent_guidance_gaps`
**Ref:** `ISS-077`
**Date:** 2026-04-18
**Severity:** Medium
**Status:** Open
**Component:** `ATD.md`, documentation system
**Affects:** Agent Decision Making, Error Handling, Development Workflow Efficiency

---

## Summary

ATD.md provides excellent tool documentation but lacks critical agent-specific integration guidance: no Claude Code context, error handling patterns, tool decision frameworks, or performance optimization guidance for agent workflows.

---

## Technical Description

### Background
ATD.md should serve as the definitive reference for IDE agents, developers, and CI tooling, providing comprehensive guidance for ATD usage within development environments.

### The Problem Scenario
1. **Missing Agent Context**: No guidance for Claude Code specifically, which is a primary consumer
2. **Unclear Error Handling**: What to do when tools fail or return unexpected results?
3. **No Decision Framework**: When to use which tool for specific tasks?
4. **Missing Recovery Patterns**: How to handle broken states or tool failures?
5. **No Performance Guidance**: Token economy not explained practically for agent workflows

### Where This Pattern Exists Today
- **ATD.md Structure**: Comprehensive tool reference but lacks agent workflow integration
- **Gap Areas**: Error handling, decision frameworks, performance optimization, recovery patterns
- **Agent Impact**: Agents may use tools inefficiently or fail to recover from errors

### Evidence from Investigation
- **Current State**: 70% sufficient for basic agent usage, 95% with enhancements
- **Key Missing Element**: No Claude Code-specific integration patterns
- **Impact**: Agents struggle with ATD tool selection, error recovery, and performance optimization

---

## Risk Assessment

| Factor | Value |
|---|---|
| Likelihood | High (Confirmed via analysis of agent usage patterns and ATD.md gaps) |
| Impact if triggered | Medium (Inefficient agent workflows, increased failure rates) |
| Detectability | Medium (Requires observation of agent behavior and error patterns) |
| Current mitigant | Manual agent guidance in CLAUDE.md and trial-and-error learning |

---

## Recommended Fix

**Short term**: Add Agent-Specific Guidance section to ATD.md with Claude Code integration patterns, tool decision framework flowchart, and common error handling patterns.

**Medium term**: Expand with detailed recovery patterns for common failures: atd_index failures, atd_weave circular dependencies, atd_trace coverage issues, LLM tool timeouts. Add performance optimization guidance for agent workflows.

**Long term**: Create comprehensive agent workflow examples for common development scenarios: feature development, bug fixing, refactoring, code review. Integrate with CLAUDE.md for project-specific guidance.

---

## References

- [ATD System Analysis](file:///home/bastien/work/skill/upsilon-hub/atd_investigation/atd_system_analysis.md)
- [Current ATD.md](file:///home/bastien/work/skill/ATD.md)
- [CLAUDE.md](file:///home/bastien/work/skill/CLAUDE.md)