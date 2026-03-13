package prompt

import "fmt"

// ReconBuild constructs the recon engine prompt.
func ReconBuild(atomContent, candidateContent string) string {
	return fmt.Sprintf(`
<System Objective>
You are the ATD Recon Engine. Validate if the given candidate code logically acts as an implementation of the Target Atom even though it lacks the spec-link tag.
Output JSON: {"Confidence": int, "Mismatches": string}
</System Objective>

<Target Atom Intent>
%s
</Target Atom Intent>

<Candidate Search Hit>
%s
</Candidate Search Hit>
`, atomContent, candidateContent)
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
