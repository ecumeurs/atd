package prompt

import "fmt"

// DiscoverLinksBuild constructs the link discovery prompt.
func DiscoverLinksBuild(registryStr, codeContent string) string {
	return fmt.Sprintf(`
<System Objective>
You are an ATD Link Discoverer. Read the Target Source Code carefully and deduce which Atoms from the Known Atom Registry define this logic. Output a list of recommended IDs.
Only recommend IDs from the provided registry list if they truly map to the codebase logic.
</System Objective>

<Known Atom Registry (Top Semantic Matches)>
%s
</Known Atom Registry (Top Semantic Matches)>

<Target Source Code>
%s
</Target Source Code>
`, registryStr, codeContent)
}

// DiscoverLinksFormat returns nil for freeform output.
func DiscoverLinksFormat() interface{} {
	return nil
}
