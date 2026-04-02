package prompt

import "fmt"

func mapLengthToWords(length string) string {
	switch length {
	case "short":
		return "100"
	case "extended":
		return "500"
	case "long":
		return "1000"
	case "default", "":
		return "300"
	default:
		return "300"
	}
}

// AssembleBuild constructs the single-pass narrative prompt.
func AssembleBuild(intent, length, content string) string {
	words := mapLengthToWords(length)
	return fmt.Sprintf(`
<System Objective>
You are an ATD Document Assembler. Your primary intent is: %s.
Ensure the output is approximately %s words in length.
Rewrite this fragmented logic into a cohesive, flowing document achieving the intent.
</System Objective>

<Raw Mechanical Output>
%s
</Raw Mechanical Output>
`, intent, words, content)
}

// LayerPassBuild constructs the intermediate summarization prompt for a specific layer.
func LayerPassBuild(layer, length, content string) string {
	words := mapLengthToWords(length)
	return fmt.Sprintf(`
<System Objective>
You are an ATD Layer Summarizer.
Analyze the following ATD fragments from the %s layer.
Summarize their core intents and logic. Ensure the output is approximately %s words in length.
</System Objective>

<Raw Fragments>
%s
</Raw Fragments>
`, layer, words, content)
}

// FinalAssembleBuild constructs the final assessment prompt from structured layer summaries.
func FinalAssembleBuild(intent, length, customer, arch, impl string) string {
	words := mapLengthToWords(length)

	// Build the context string dynamically to omit empty summaries
	var context string
	if customer != "" && customer != "null" {
		context += fmt.Sprintf("\n<Customer Layer Summary>\n%s\n</Customer Layer Summary>\n", customer)
	}
	if arch != "" && arch != "null" {
		context += fmt.Sprintf("\n<Architecture Layer Summary>\n%s\n</Architecture Layer Summary>\n", arch)
	}
	if impl != "" && impl != "null" {
		context += fmt.Sprintf("\n<Implementation Layer Summary>\n%s\n</Implementation Layer Summary>\n", impl)
	}

	return fmt.Sprintf(`
<System Objective>
You are an ATD Analyst. Your primary intent is: %s.
You are provided with layer-by-layer summaries of the targeted scope.
Produce your final output ensuring it achieves the exact intent specified.
Ensure the output is approximately %s words in length.
</System Objective>
%s
`, intent, words, context)
}

// AssembleFormat returns nil for freeform output.
func AssembleFormat() interface{} {
	return nil
}
