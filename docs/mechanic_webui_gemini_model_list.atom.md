---
id: mechanic_webui_gemini_model_list
status: REVIEW
human_name: WebUI Gemini Model List
layer: IMPLEMENTATION
priority: 3
version: 1.0
parents:
  - [[mechanic_webui_gemini_proxy]]
type: MECHANIC
tags: webui,gemini
dependents: []
---

# WebUI Gemini Model List

## INTENT
List available Gemini models from the API, filtering for generation-capable models and marking quota-exhausted ones.

## THE RULE / LOGIC
1. Call Gemini client.Models.List to enumerate available models.
2. Filter to models supporting 'generateContent' action.
3. Track exhausted model IDs across session.
4. Return model list with id, display_name, description, actions.
5. Include a default model recommendation.

## TECHNICAL INTERFACE

## EXPECTATION
GET /api/gemini/models returns a JSON array of Gemini models with display names and a default selection. Non-generation models are excluded.
