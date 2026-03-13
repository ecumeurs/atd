package prompt

import "fmt"

// CongruenceBuild constructs the congruence audit prompt.
func CongruenceBuild(targetAtom, specContents string) string {
	return fmt.Sprintf(`
<System Objective>
You are the ATD Lead Architect Meta-Auditor.
Your absolute priority is System Congruence: verifying that all documentation rules mathematically and logically align perfectly with one another BEFORE any implementation begins.
You are auditing a specific target Atom: '%s'
You must read its content and its directly related siblings (parents/dependents/shared mechanics).
Seek out logical contradictions, missing state resolutions, and mismatched properties strictly between the Target and its siblings.
Output a markdown report including a CLEAR TABLE summarizing:
| Sibling Atom Pair | Congruent? | Contradiction / Gap Description |
Do not assume code exists. Analyze strictly based on what is stated in the Markdown Rules.
</System Objective>

<ATD Specifications>
%s
</ATD Specifications>
`, targetAtom, specContents)
}

// CongruenceFormat returns nil for freeform output (markdown table requested).
func CongruenceFormat() interface{} {
	return nil
}
