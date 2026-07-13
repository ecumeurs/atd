package prompt

import (
	"encoding/json"
)

// ReconBuild constructs the recon engine prompt in JSON format.
func ReconBuild(atomContent, candidateContent string) string {
	msg := map[string]interface{}{
		"system_objective": "You are the ATD Recon Engine. Validate if the given candidate code logically acts as an implementation of the Target Atom even though it lacks the spec-link tag.",
		"atom_intent":      atomContent,
		"candidate_code":   candidateContent,
		"instruction": "Does the code logically implement the atom? You are only shown a single file: the " +
			"atom's logic may legitimately continue behind a service seam in a collaborating file you cannot " +
			"see here. Do not report Confidence: 0 with no rationale just because this one file is a partial " +
			"view — instead list, concretely, which specific aspects of the atom's logic you could not verify " +
			"from this file alone. Output a calibrated Confidence (0-100) reflecting how well the visible code " +
			"matches the atom, plus an itemized Mismatches array: one object per concrete discrepancy or " +
			"unverifiable aspect, each with 'aspect' (the piece of atom logic in question), 'expected' (what " +
			"the atom requires), and 'found' (what the code actually does, or 'not visible in this file' if it " +
			"may live elsewhere). If there are no mismatches, return an empty array — never a prose sentence.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// ReconFormat returns the JSON schema for recon results.
func ReconFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"Confidence": map[string]interface{}{
				"type":        "integer",
				"description": "Calibrated 0-100 confidence that the candidate code implements the atom.",
			},
			"Mismatches": map[string]interface{}{
				"type":        "array",
				"description": "Itemized concrete mismatches or unverifiable aspects. Empty array if none.",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"aspect":   map[string]string{"type": "string"},
						"expected": map[string]string{"type": "string"},
						"found":    map[string]string{"type": "string"},
					},
					"required": []string{"aspect", "expected", "found"},
				},
			},
		},
		"required": []string{"Confidence", "Mismatches"},
	}
}
