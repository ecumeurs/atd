---
id: ui_webui_document_viewer
status: DRAFT
priority: 3
dependents: []
human_name: Document Generation UI
type: UI
layer: CUSTOMER
parents: [[[module_webui]]]
tags: [webui, modal, ui, documentation]
version: 1.0
---

# New Atom

## INTENT
To allow users to rapidly synthesize broad documentation narratives on the fly via a contextual overlay picker and viewer modal.

## THE RULE / LOGIC
1. Using Ctrl+K search, users trigger the Create Document flow.
2. A setup modal allows checkbox selection of 5 top search ATDs and adding additional ones via a nested picker.
3. Users can reformulate the base search query.
4. Clicking 'Generate' invokes the backend logic, triggering a loading modal.
5. On succcess, a comprehensive Viewer Modal renders the markdown with a 'Download .md' utility.
6. A 'Recent Docs' dropdown retains up to 5 previously generated sessions for fast access.

## TECHNICAL INTERFACE

## EXPECTATION
The document modal correctly renders markdown and provides a valid download file. The setup picker properly populates from `/search-document-context` limit 5 array.
