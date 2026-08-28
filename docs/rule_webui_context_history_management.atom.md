---
id: rule_webui_context_history_management
status: DRAFT
human_name: Context History Management Logic
type: RULE
layer: ARCHITECTURE
priority: 3
parents:
  - [[mechanic_webui_gemini_proxy]]
version: 1.0
dependents: []
---

# Context History Management Logic

## INTENT
Manage the selection and forwarding of conversation history to the Gemini API based on user settings.

## THE RULE / LOGIC
1. If omit_history is false: Include the entire messages[] array in the GenerateContent call. 2. If omit_history is true: Only include the most recent user message in the GenerateContent call.

## TECHNICAL INTERFACE

## EXPECTATION
Verification of request payload shows history omission when Chat Mode is disabled.
