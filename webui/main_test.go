package main

import (
	"testing"
)

func TestExtractIntent(t *testing.T) {
	content := `
# My Atom

## INTENT
To provide a secure proxy.

## THE RULE / LOGIC
Some logic here.
`
	intent := extractIntent(content)
	expected := "To provide a secure proxy."
	if intent != expected {
		t.Errorf("Expected intent '%s', got '%s'", expected, intent)
	}
}

func TestExtractIntentEmpty(t *testing.T) {
	content := `
# Dummy
`
	intent := extractIntent(content)
	if intent != "" {
		t.Errorf("Expected empty intent, got '%s'", intent)
	}
}

func TestExtractIntentMultiline(t *testing.T) {
	content := `
## INTENT
This is the intent
on multiple lines
## NEXT SECTION
`
	intent := extractIntent(content)
	expected := "This is the intent" // Current implementation stops at newline or next header
	if intent != expected {
		t.Errorf("Expected intent '%s', got '%s'", expected, intent)
	}
}
