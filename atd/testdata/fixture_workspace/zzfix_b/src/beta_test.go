package src

import "testing"

// @test-link [[api_zzfix_ws_beta]]
func TestBeta(t *testing.T) {
	if Beta() != 2 {
		t.Fatalf("expected Beta() == 2, got %d", Beta())
	}
}
