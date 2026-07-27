package prompt

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAuditCode(t *testing.T) {
	atom := CuratedAuditAtom{
		ID:          "rule_password_policy",
		Type:        "RULE",
		Layer:       "BUSINESS",
		Intent:      "Passwords must be at least 12 characters",
		Logic:       "Reject any password shorter than 12 characters at signup",
		Expectation: "Signup returns a validation error for short passwords",
	}
	codeContent := "func ValidatePassword(pw string) bool { return len(pw) - 12 == 0 }"

	prompt := AuditCodeBuild(PersonaTechLead, atom, codeContent)
	if !strings.Contains(prompt, string(PersonaTechLead)) {
		t.Error("Prompt missing persona")
	}
	if !strings.Contains(prompt, atom.Intent) || !strings.Contains(prompt, atom.Logic) {
		t.Error("Prompt missing atom intent or logic")
	}
	if !strings.Contains(prompt, codeContent) {
		t.Error("Prompt missing code content")
	}

	schema := AuditCodeFormat()
	_, err := json.Marshal(schema)
	if err != nil {
		t.Errorf("FormatSchema not serializable: %v", err)
	}
}

func TestFixSplit(t *testing.T) {
	prompt := FixSplitBuild("atom content")
	if !strings.Contains(prompt, "atom content") {
		t.Error("Prompt missing atom content")
	}

	schema := FixSplitFormat()
	_, err := json.Marshal(schema)
	if err != nil {
		t.Errorf("FormatSchema not serializable: %v", err)
	}
}

func TestReconcile(t *testing.T) {
	prompt := ReconcileBuild("store", "inbound")
	if !strings.Contains(prompt, "store") || !strings.Contains(prompt, "inbound") {
		t.Error("Prompt missing store or inbound")
	}

	schema := ReconcileFormat()
	_, err := json.Marshal(schema)
	if err != nil {
		t.Errorf("FormatSchema not serializable: %v", err)
	}
}

func TestRecon(t *testing.T) {
	prompt := ReconBuild("atom", "code")
	if !strings.Contains(prompt, "atom") || !strings.Contains(prompt, "code") {
		t.Error("Prompt missing atom or code")
	}

	schema := ReconFormat()
	_, err := json.Marshal(schema)
	if err != nil {
		t.Errorf("FormatSchema not serializable: %v", err)
	}
}

func TestAllSchemas(t *testing.T) {
	// Simple validation that all Format functions return valid JSON (or nil)
	formats := []interface{}{
		AuditBloatFormat(),
		AuditCodeFormat(),
		CompareFormat(),
		FixSplitFormat(),
		ReconcileFormat(),
		CongruenceFormat(),
		ReconFormat(),
		AssembleFormat(),
		IntentExtractFormat(),
		DiscoverLinksFormat(),
	}

	for i, f := range formats {
		if f == nil {
			continue
		}
		_, err := json.Marshal(f)
		if err != nil {
			t.Errorf("Format index %d not serializable: %v", i, err)
		}
	}
}
