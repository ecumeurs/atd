package prompt

import (
	"encoding/json"
)

// CongruenceBuild constructs the congruence audit prompt in JSON format.
//
// The instruction explicitly ties `findings` to `is_congruent`: a
// congruent-false verdict is useless to callers without at least one
// specific, named contradiction -- see
// failures/20260917_atd_congruence_empty_verdict_and_no_workspace_resolution.md
// and pkg/prompt/congruence_findings_test.go / cmd/atd/cmd/congruence_findings_test.go
// for the pinned contract this enforces at the RunE layer.
func CongruenceBuild(targetAtom, specContents string) string {
	msg := map[string]interface{}{
		"system_objective": "You are the ATD Lead Architect Meta-Auditor. Your absolute priority is System Congruence: verifying that all documentation rules mathematically and logically align perfectly with one another BEFORE any implementation begins.",
		"target_atom":      targetAtom,
		"spec_contents":    specContents,
		"instruction": "Seek out logical contradictions, missing state resolutions, and mismatched properties strictly between the Target and its siblings. " +
			"If is_congruent is false, the 'findings' array MUST be non-empty: include one entry per contradiction found, each naming the specific related atom id that contradicts the target ('atom_id'), the section the contradiction lives in ('section': one of \"INTENT\", \"THE RULE\", \"LOGIC\"), and precisely what the contradiction is ('contradiction'). " +
			"A bare title or one-line summary in 'audit_report' with no matching 'findings' entries is not an acceptable result when is_congruent is false. " +
			"If is_congruent is true, 'findings' may be an empty array.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// CongruenceFormat returns the JSON schema for congruence results.
//
// The schema declares a structured `findings` array (one entry per
// contradiction: atom_id/section/contradiction) alongside the free-string
// `audit_report`, so a congruent-false verdict is actionable instead of a
// bare title -- see cmd/atd/cmd/congruence.go's RunE, which now rejects an
// is_congruent:false response with no findings.
//
// Unlike pkg/prompt/discover_links.go's discoverLinksSchema, this schema is
// intentionally built as nested map[string]interface{} literals rather than
// a literal JSON string: pkg/prompt/congruence_findings_test.go's asMap
// helper navigates the returned schema via direct map[string]interface{}
// type assertions (no JSON round-trip), which a json.RawMessage-typed
// return value would fail outright. That means encoding/json's alphabetical
// map-key sort applies when this schema is marshaled for the wire (unlike
// discoverLinksSchema, "findings" cannot be guaranteed to precede
// "audit_report" in the actual request sent to the model) -- the
// completeness requirement ("findings must be non-empty when is_congruent
// is false") is carried by CongruenceBuild's instruction text instead of by
// schema field order.
func CongruenceFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"is_congruent": map[string]interface{}{"type": "boolean"},
			"findings": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"atom_id":       map[string]interface{}{"type": "string"},
						"section":       map[string]interface{}{"type": "string"},
						"contradiction": map[string]interface{}{"type": "string"},
					},
					"required": []string{"atom_id", "section", "contradiction"},
				},
			},
			"audit_report": map[string]interface{}{"type": "string"},
		},
		"required": []string{"is_congruent", "audit_report"},
	}
}
