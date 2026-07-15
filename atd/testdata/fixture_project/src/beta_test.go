package src

import "testing"

// @test-link [[api_zzfix_beta]]
func TestBeta(t *testing.T) {
	if Beta() != 42 {
		t.Fatalf("expected Beta() == 42, got %d", Beta())
	}
	if BetaHelper() != 1 {
		t.Fatalf("expected BetaHelper() == 1, got %d", BetaHelper())
	}
}
