package prompt

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDissect(t *testing.T) {
	prompt := DissectBuild("001: hello")
	if !strings.Contains(prompt, "001: hello") {
		t.Error("Prompt doesn't contain numbered content")
	}

	schema := DissectFormat()
	_, err := json.Marshal(schema)
	if err != nil {
		t.Errorf("FormatSchema not serializable: %v", err)
	}
}

func TestAuditCode(t *testing.T) {
	prompt := AuditCodeBuild("rule", "code")
	if !strings.Contains(prompt, "rule") || !strings.Contains(prompt, "code") {
		t.Error("Prompt missing rule or code")
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
		DissectFormat(),
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
