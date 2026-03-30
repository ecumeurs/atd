---
id: rule_webui_color_palette
human_name: "Explorer Color Scheme"
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [ui, colors, explorer]
parents:
  - [[ui_webui_global_theme]]
dependents: []
---

# Explorer Color Scheme

## INTENT
Define the precise hex codes and meanings for health-based status categorization in the Explorer view.

## THE RULE / LOGIC
The following Palette Tokens will be used for categorization:
- **Done:** Emerald Green (#2e7d32, #4caf50 glow).
- **Almost Done:** Amber/Gold (#fbc02d, #fdd835 glow).
- **WIP:** Cyan/Electric Blue (#0288d1, #03a9f4 glow).
- **Doc Jungle:** Charcoal/Deep Red mix (#546e7a with #d32f2f accents).

## TECHNICAL INTERFACE (The Bridge)
- **CSS Variables:** `--color-done`, `--color-almost-done`, `--color-wip`, `--color-jungle`
- **Code Tag:** `@spec-link [[rule_webui_color_palette]]`

## EXPECTATION (For Testing)
When an atom is categorized as 'Almost Done', it must be clearly rendered with the Amber palette (CSS: `--color-almost-done`).
