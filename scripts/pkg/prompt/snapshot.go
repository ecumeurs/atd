package prompt

import "fmt"

// SnapshotBuild constructs the narrative snapshot prompt.
func SnapshotBuild(theme, content string) string {
	return fmt.Sprintf(`
<System Objective>
You are an ATD Narrative Generator. Rewrite this fragmented logic into a cohesive, flowing document styled strictly as an: %s. Emphasize clarity and narrative over raw mathematical values.
</System Objective>

<Raw Mechanical Output>
%s
</Raw Mechanical Output>
`, theme, content)
}

// SnapshotFormat returns nil for freeform output.
func SnapshotFormat() interface{} {
	return nil
}
