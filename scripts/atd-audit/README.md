# atd-audit

## Intent
Analyzes ATD files to ensure architectural integrity by checking for documentation bloat and detecting semantic overlaps or missing abstractions.

## Phases
1. **The Bloat Metric (Syntactic Validator)**: Validates that `.atom.md` intents and rules strictly follow the "Minimum Atomic Scale" rule without describing multiple distinct features.
2. **The Collision Map (Semantic Overlap Detection)**: Computes vector embeddings of ATD files and flags refactoring opportunities if atoms share significant logic without a shared parent (Missing Abstraction).
