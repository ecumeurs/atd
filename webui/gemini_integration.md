# Gemini Integration: Spec Builder Tool (Refined)

This document outlines the detailed strategy for the Gemini-powered "Spec Builder" in the ATD WebUI.

## Core Features

1.  **Conversational ATD Evolution**: Real-time discussion with an LLM that understands ATD rules.
2.  **Batch Proposals**: Ability to suggest multiple creation/update/deletion actions in a single response.
3.  **Session Management**: Support for model switching and context window resets (new sessions).
4.  **ATD Manifesto Guided**: Every session starts with a strict set of rules (the Manifesto) to ensure model compliance.

## Technical Architecture

### 1. The ATD Manifesto (System Prompt)
The system prompt will incorporate the core rules of Atomic Traceable Documentation:
- **Atom Blueprint**: Strict YAML frontmatter and H2 sections. Especially important: ATD type and layer (mandatory), and the `parents` and `dependents` fields (best effort). It will be the role of the webui tool to use appropriate atd commands to resolve the parents and dependents fields based on informations provided by the LLM.
- **Minimum Atomic Scale**: Exactly ONE state-changing rule per atom.
- **Hierarchy**: Customer -> Architecture -> Implementation layers.

### 2. Structured JSON Object (JSON Schema)
The response from Gemini will follow this refined schema:
```json
{
  "type": "object",
  "properties": {
    "message": { 
      "type": "string",
      "description": "The human-readable conversational part of the response."
    },
    "proposals": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "action": { "enum": ["CREATE", "UPDATE", "DELETE"] },
          "atom_id": { "type": "string" },
          "content": {
            "type": "object",
            "properties": {
              "human_name": { "type": "string" },
              "type": { "type": "string" },
              "status": { "type": "string" },
              "priority": { "type": "string" },
              "tags": { "type": "array", "items": { "type": "string" } },
              "intent": { "type": "string" },
              "logic": { "type": "string" },
              "technical_interface": { "type": "string" },
              "expectation": { "type": "string" }
            }
          },
          "impact_summary": { "type": "string" }
        },
        "required": ["action", "atom_id"]
      }
    }
  },
  "required": ["message"]
}
```

### 3. Session & Context Management
- **Model Switching**: Users can select between available models.
- **Restart Session**: A UI button to clear the conversation history and start fresh with the Manifesto.
- **Context Injection**: 
  - On every user message, the system searches existing atoms.
  - New atoms are added to the conversation context if relevant.
  - Avoid redundant injection if an atom is already in the chat history.

## Integration Workflow

1.  **Discuss**: User chats with Gemini about a feature.
2.  **Propose**: Gemini returns a `proposals` array.
3.  **Reconcile**: WebUI calls `atd reconcile` for each proposal to show deltas.
4.  **Evaluate**: WebUI calls `atd trace` to alert on implementation/test impact.
5.  **Apply**: User clicks "Apply" to execute `atd update`
