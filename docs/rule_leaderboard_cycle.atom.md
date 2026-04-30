---
id: rule_leaderboard_cycle
status: STABLE
human_name: Leaderboard Reset Cycle
type: RULE
layer: ARCHITECTURE
parents:
  - [[us_leaderboard_view]]
version: 1.0
dependents: []
priority: 5
---

# New Atom

## INTENT
Define the weekly filtering window for leaderboard rankings (Resets conceptually every Sunday).

## THE RULE / LOGIC
- The leaderboard displays a "Weekly Cycle" of performance.
- Results are dynamically filtered to only include matches with a `concluded_at` timestamp greater than or equal to the most recent Sunday 00:01 UTC.
- All match data remains in the database for historical auditing; the "reset" is a temporal filter applied at query time.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[rule_leaderboard_cycle]]`
- **Cron Expression:** `1 0 * * 0` (Sunday 00:01)

## EXPECTATION
