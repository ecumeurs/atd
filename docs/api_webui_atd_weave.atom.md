---
id: api_webui_atd_weave
status: DRAFT
type: API
layer: ARCHITECTURE
priority: 3
dependents:
  - [[mechanic_webui_atd_weave_handler]]
human_name: WebUI Weave API
version: 1.0
parents:
  - [[ui_webui_details_weave_button]]
---

# WebUI Weave API

## INTENT
Expose a REST-like endpoint for the WebUI to trigger the ATD weaving operation.

## THE RULE / LOGIC
Endpoint: `POST /api/atd/weave`
Input: None
Output: `{ "message": "Link Weaving Complete. Updated X files.", "count": X }`
Standard 200/500 status codes.

## TECHNICAL INTERFACE
@spec-link [[api_webui_atd_weave]]

## EXPECTATION
Successful response returns a message indicating the number of files updated.
