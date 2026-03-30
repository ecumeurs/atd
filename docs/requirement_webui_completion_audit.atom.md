---
id: requirement_webui_completion_audit
status: DRAFT
human_name: WebUI Completion Audit
layer: CUSTOMER
version: 1.0
priority: 3
parents: [[requirement_webui_platform]]
dependents: []
type: REQUIREMENT
tags: [webui, completion, audit]
---

# WebUI Completion Audit

## INTENT
Enable product owners to verify that every customer requirement is correctly decomposed into architecture and implementation.

## THE RULE / LOGIC
A "Requirement" is considered complete only if it has a path to at least one STABLE implementation atom with test coverage.
- **Ancestry Check**: Traverse `parents` from Implementation to Architecture to Customer.
- **Coverage Check**: Verify existence of `@spec-link` and `@test-link` in source code.
- **Status Check**: Status must be STABLE for full completion.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag**: `@spec-link [[requirement_webui_completion_audit]]`
- **Related Issue**: N/A
- **Test Names**: `TestCompletionAudit`

## EXPECTATION (For Testing)
An implementation status check must show a clear path from requirement to test for any STABLE requirement.
