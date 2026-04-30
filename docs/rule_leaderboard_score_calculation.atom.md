---
id: rule_leaderboard_score_calculation
status: STABLE
human_name: Score Calculation Rule
type: RULE
layer: ARCHITECTURE
parents:
  - [[us_leaderboard_view]]
dependents: []
priority: 5
version: 1.0
---

# New Atom

## INTENT
Define the scoring algorithm and eligibility for the leaderboard.

## THE RULE / LOGIC
- **Scoring Formula:** `Score = Wins / MAX(1, Losses)`
  - 10 Wins, 0 Losses => 10.0
  - 0 Wins, 0 Losses => 0.0
  - 10 Wins, 10 Losses => 1.0
- **Eligibility:** Only users with at least one match recorded within the current time range are included in the leaderboard.
- **Search Filtering:** Retrieval queries must exclude any user with 0 matches in the selected mode/period.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[rule_leaderboard_score_calculation]]`
- **Test Names:** `TestScoreCalculation`, `TestLeaderboardEligibility`

## EXPECTATION
