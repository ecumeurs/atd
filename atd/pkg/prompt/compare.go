package prompt

import (
	"encoding/json"
)

// AtomData is a subset of the atom metadata for prompt construction.
type AtomData struct {
	HumanName string
	Type      string
	Intent    string
	Logic     string
}

// CompareBuild constructs the atom comparison prompt in JSON format.
func CompareBuild(a, b AtomData) string {
	msg := map[string]interface{}{
		"system_objective": "You are an ATD architect reviewing two semantically overlapping atoms.",
		"atom_a":          a,
		"atom_b":          b,
		"instruction":      "Diagnose the relationship (MISSING_COMMON_PARENT, MERGE, REFACTOR, ACCEPTABLE_SIBLING) and propose a resolution.",
	}
	bj, _ := json.Marshal(msg)
	return string(bj)
}

// CompareFormat returns the JSON schema for atom comparison results.
func CompareFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"category":   map[string]string{"type": "string"},
			"diagnosis":  map[string]string{"type": "string"},
			"resolution": map[string]string{"type": "string"},
		},
		"required": []string{"category", "diagnosis", "resolution"},
	}
}
