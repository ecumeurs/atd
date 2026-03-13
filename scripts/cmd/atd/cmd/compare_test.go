package cmd

import (
	"testing"
)

func TestSharedKeywords(t *testing.T) {
	a := "To implement a secure authentication system."
	b := "The system handles user authentication and session management."
	
	shared := sharedKeywords(a, b)
	
	found := false
	for _, k := range shared {
		if k == "authentication" {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Expected 'authentication' in shared keywords, got %v", shared)
	}
}
