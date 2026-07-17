---
id: mechanic_webui_breadcrumb_navigation
human_name: "Breadcrumb Navigation Logic"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 3
tags: [logic, navigation, breadcrumb]
parents:
  - [[ui_webui_tab_system]]
dependents: []
---

# Breadcrumb Navigation Logic

## INTENT
Provide a simple, clear mechanism for navigating back to the global overview from a zoomed health bucket.

## THE RULE / LOGIC
- **Path Structure:** `All > [Bucket Name]`.
- **Global Breadcrumb:** 'All' is always a clickable root that resets the view to the four health blobs.
- **Bucket Breadcrumb:** Displays the name of the currently zoomed-in health category.
- **Visibility:** Only visible when the `currentViewScope` is not 'all'.

## TECHNICAL INTERFACE (The Bridge)
- **HTML Element:** `.breadcrumb-container`
- **JS Function:** `renderBreadcrumb()`
- **Code Tag:** `@spec-link [[mechanic_webui_breadcrumb_navigation]]`

## EXPECTATION (For Testing)
When the user is in the 'Done' zoomed view, the breadcrumb must read `All > Done`, and clicking 'All' must reset the `currentViewScope` to 'all'.
