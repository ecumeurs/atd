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
		"instruction":      "Does the code logically implement the atom? Output confidence and mismatches as JSON.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// ReconFormat returns the JSON schema for recon results.
func ReconFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"Confidence": map[string]string{"type": "integer"},
			"Mismatches": map[string]string{"type": "string"},
		},
		"required": []string{"Confidence", "Mismatches"},
	}
}
