package prompt

import (
	"fmt"
)

// AtomData is a subset of the atom metadata for prompt construction.
type AtomData struct {
	HumanName string
	Type      string
	Intent    string
	Logic     string
}

// CompareBuild constructs the atom comparison prompt.
func CompareBuild(a, b AtomData) string {
	return fmt.Sprintf(`You are an ATD (Atomic Traceable Documentation) architect reviewing two semantically overlapping atoms.

Atom A — %s (%s):
Intent: %s
Logic: %s

Atom B — %s (%s):
Intent: %s
Logic: %s

These two atoms have high semantic similarity. Diagnose the relationship and propose a concrete resolution in plain text (3-5 sentences).
Choose the most appropriate category: MISSING_COMMON_PARENT, MERGE, REFACTOR, or ACCEPTABLE_SIBLING.
If MISSING_COMMON_PARENT, suggest a parent id and one-line intent for it.
If MERGE, specify which atom should absorb the other and why.
If REFACTOR, specify what specific content should move where.
If ACCEPTABLE_SIBLING, explain why the similarity is harmless.`,
		a.HumanName, a.Type, a.Intent, a.Logic,
		b.HumanName, b.Type, b.Intent, b.Logic,
	)
}

// CompareFormat returns nil for freeform output.
func CompareFormat() interface{} {
	return nil
}
