package prompt

import (
	"encoding/json"
)

// DiscoverLinksBuild constructs the link discovery prompt in JSON format.
func DiscoverLinksBuild(registryStr, codeContent string) string {
	msg := map[string]interface{}{
		"system_objective": "You are an ATD Link Discoverer. Read the Target Source Code carefully and deduce which Atoms from the Known Atom Registry define this logic. Output a list of recommended IDs.",
		"atom_registry":    registryStr,
		"target_code":      codeContent,
		"instruction":      "Only recommend IDs from the provided registry list if they truly map to the codebase logic.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// DiscoverLinksFormat returns the JSON schema for link discovery results.
func DiscoverLinksFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"recommendations": map[string]interface{}{
				"type":  "array",
				"items": map[string]string{"type": "string"},
			},
			"rationale": map[string]string{"type": "string"},
		},
		"required": []string{"recommendations", "rationale"},
	}
}
