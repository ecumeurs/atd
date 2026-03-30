---
id: ui_webui_global_theme
human_name: "Global UI Look and Feel"
type: UI
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [ui, theme, glassmorphism]
parents:
  - [[module_webui]]
dependents: [[[rule_webui_color_palette]]]
---

# Global UI Look and Feel

## INTENT
Maintain a premium, high-tech, and cohesive user experience across the entire Web UI.

## THE RULE / LOGIC
- **Color Mode:** Dark mode defaults; deep charcoal background (#0f1115).
- **Typography:** Primary font is 'Inter' (sans-serif); weights: 400 (regular), 600 (semibold).
- **Visual Effects:**
    - **Glassmorphism:** Use semi-transparent backgrounds with `backdrop-filter: blur(12px)`.
    - **Glows:** High-priority or status-coded elements must feature subtle outer-glows (box-shadow).
- **Iconography:** Minimalist, thin-line icons (matching the visual weight of the Inter font).

## TECHNICAL INTERFACE (The Bridge)
- **CSS Variables:** `--bg-dark`, `--bg-panel`, `--text-main`, `--accent`
- **Font-Family:** `'Inter', sans-serif`
- **Code Tag:** `@spec-link [[ui_webui_global_theme]]`

## EXPECTATION (For Testing)
When the 'Glassmorphism' effect is active on a panel, the background content must be blurred (at least 10px) behind it.
