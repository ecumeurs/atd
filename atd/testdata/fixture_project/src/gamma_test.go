package src

import "testing"

// @test-link [[mech_zzfix_gamma]]
func TestGamma(t *testing.T) {
	if Gamma() != 7 {
		t.Fatalf("expected Gamma() == 7, got %d", Gamma())
	}
}
