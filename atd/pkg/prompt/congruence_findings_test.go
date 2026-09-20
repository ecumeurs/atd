package prompt

// Pins the FIX for Problem 1 of
// failures/20260917_atd_congruence_empty_verdict_and_no_workspace_resolution.md
// ("bare/unusable verdict, no validation"). See also
// cmd/atd/cmd/congruence_known_defects_test.go, which pins the CURRENT
// (broken) behavior as a negative pin, and cmd/atd/cmd/congruence_findings_test.go,
// which pins RunE's validation of a response built to this shape.
//
// Desired CongruenceFormat() contract (once fixed) — the schema must declare
// a structured `findings` array alongside the existing `is_congruent` bool
// and `audit_report` free-string summary, so a congruent-false verdict can
// carry which related atom contradicts the target and in which section:
//
//	{
//	  "is_congruent": bool,
//	  "audit_report": string,
//	  "findings": [
//	    {
//	      "atom_id":       string,  // the related atom that contradicts the target
//	      "section":       string,  // one of "INTENT", "THE RULE", "LOGIC"
//	      "contradiction": string   // what the contradiction actually is
//	    },
//	    ...
//	  ]
//	}
//
// RunE is expected to reject (non-nil error) any is_congruent:false response
// whose findings array is missing or empty — a title-only audit_report must
// no longer be accepted as a complete result. This file only pins the
// *schema* shape; cmd/atd/cmd/congruence_findings_test.go pins the RunE-level
// validation behavior against that shape.

import (
	"testing"
)

// asMap is a small helper for navigating CongruenceFormat()'s
// map[string]interface{} schema without a JSON round-trip, failing loudly
// (not panicking) when a key is absent or of the wrong shape.
func asMap(t *testing.T, v interface{}, path string) map[string]interface{} {
	t.Helper()
	m, ok := v.(map[string]interface{})
	if !ok {
		t.Fatalf("CongruenceFormat(): expected %s to be a map[string]interface{}, got %T (%v)", path, v, v)
	}
	return m
}

func TestCongruenceFormat_RequiresStructuredFindings(t *testing.T) {
	schema := asMap(t, CongruenceFormat(), "schema root")

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("CongruenceFormat(): expected schema[\"properties\"] to be a map[string]interface{}, got %T", schema["properties"])
	}

	findingsRaw, ok := properties["findings"]
	if !ok {
		t.Fatal("CongruenceFormat(): schema has no \"findings\" property — a structured findings array (atom_id/section/contradiction per entry) is required so a congruent-false verdict is actionable, not just a free-string audit_report")
	}
	findings := asMap(t, findingsRaw, "properties.findings")

	if findings["type"] != "array" {
		t.Fatalf("CongruenceFormat(): properties.findings.type = %v, want \"array\"", findings["type"])
	}

	itemsRaw, ok := findings["items"]
	if !ok {
		t.Fatal("CongruenceFormat(): properties.findings has no \"items\" schema describing each finding entry's shape")
	}
	items := asMap(t, itemsRaw, "properties.findings.items")

	itemProps, ok := items["properties"].(map[string]interface{})
	if !ok {
		t.Fatalf("CongruenceFormat(): properties.findings.items.properties is not a map[string]interface{}, got %T", items["properties"])
	}

	for _, field := range []string{"atom_id", "section", "contradiction"} {
		fieldSchemaRaw, ok := itemProps[field]
		if !ok {
			t.Errorf("CongruenceFormat(): findings item schema missing required field %q (contract: atom_id/section/contradiction per finding)", field)
			continue
		}
		fieldSchema := asMap(t, fieldSchemaRaw, "properties.findings.items.properties."+field)
		if fieldSchema["type"] != "string" {
			t.Errorf("CongruenceFormat(): findings item field %q has type %v, want \"string\"", field, fieldSchema["type"])
		}
	}

	itemRequired, ok := items["required"].([]string)
	if !ok {
		// Schema helpers in this package build "required" as []string
		// elsewhere (see CongruenceFormat's own top-level "required" field);
		// accept the same shape here rather than []interface{}.
		t.Fatalf("CongruenceFormat(): properties.findings.items.required is not a []string, got %T (%v)", items["required"], items["required"])
	}
	wantRequired := map[string]bool{"atom_id": true, "section": true, "contradiction": true}
	for _, r := range itemRequired {
		delete(wantRequired, r)
	}
	if len(wantRequired) > 0 {
		t.Errorf("CongruenceFormat(): findings item schema's \"required\" list is missing: %v (got required=%v)", wantRequired, itemRequired)
	}
}
