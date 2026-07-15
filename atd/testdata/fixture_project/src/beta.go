package src

// @spec-link [[api_zzfix_beta]]
func Beta() int {
	return 42
}

// Beta is deliberately spec-linked twice in this one file (Beta and
// BetaHelper both implement api_zzfix_beta) so scenario S2
// (test_atd_07_26.md §4) has a real "check --file" dedup case: the report
// must list api_zzfix_beta exactly once for this file, not twice.

// @spec-link [[api_zzfix_beta]]
func BetaHelper() int {
	return Beta() - 41
}
