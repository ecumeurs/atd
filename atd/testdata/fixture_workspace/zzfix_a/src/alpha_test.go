package src

import "testing"

// @test-link [[req_zzfix_ws_alpha]]
func TestAlpha(t *testing.T) {
	if Alpha() != 1 {
		t.Fatalf("expected Alpha() == 1, got %d", Alpha())
	}
}
