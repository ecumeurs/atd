---
id: req_logging_traceability
human_name: Logging Request Traceability
type: REQUIREMENT
layer: BUSINESS
version: 1.0
status: STABLE
priority: 5
tags: [logging, traceability, requirements, observability]
parents:
  - [[upsilonapi:api_request_id]]
dependents:
  - [[rule_tracing_logging]]
---
# Logging Request Traceability

## INTENT
To guarantee that every log entry produced by the system can be definitively mapped back to the specific user request or asynchronous flow that triggered it, ensuring zero-gap observability.

## THE RULE / LOGIC
1.  **Mandatory Tagging:** Every log entry (stdout, file, or external sink) MUST contain a reference ID derived from the transaction's unique identifier.
2.  **ID Sources:**
    - **Header Source:** For HTTP interactions, use the value from the `X-Request-ID` header (see [[api_request_id]]).
    - **Payload Source:** If the header is missing or unavailable (e.g., internal processing), use the `request_id` field from the [[api_standard_envelope]].
3.  **Consistency:** The same ID must persist across the entire execution stack of a single flow, including cross-service calls (Laravel <-> Go).
4.  **Fallback:** If no ID is provided, the logging middleware/layer MUST generate a temporary "orphan" ID to at least group local logs, while flagging the missing traceability as a warning.

## TECHNICAL INTERFACE (The Bridge)
- **Header:** `X-Request-ID`
- **JSON Field:** `request_id`
- **Code Tag:** `@spec-link [[req_logging_traceability]]`
- **Related Issue:** `ISS-023`

## EXPECTATION (For Testing)
- Every line of logs emitted during a supervised request contains the matching `request_id` or its `ref_id` prefix.
- Logs emitted before an ID is established (extreme early boot) are exempt but should be minimal.
