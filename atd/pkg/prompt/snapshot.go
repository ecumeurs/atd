package prompt

import (
	"encoding/json"
)

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

// AssembleBuild constructs the single-pass narrative prompt in JSON format.
func AssembleBuild(intent, length, content string) string {
	words := mapLengthToWords(length)
	msg := map[string]interface{}{
		"system_objective": "You are an ATD Document Assembler.",
		"intent":           intent,
		"target_length":    words + " words",
		"raw_content":      content,
		"instruction":      "Rewrite this fragmented logic into a cohesive, flowing document achieving the intent.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// LayerPassBuild constructs the intermediate summarization prompt for a specific layer in JSON format.
func LayerPassBuild(layer, length, content string) string {
	words := mapLengthToWords(length)
	msg := map[string]interface{}{
		"system_objective": "You are an ATD Layer Summarizer.",
		"layer":            layer,
		"target_length":    words + " words",
		"raw_fragments":    content,
		"instruction":      "Summarize their core intents and logic.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// FinalAssembleBuild constructs the final assessment prompt from structured layer summaries in JSON format.
func FinalAssembleBuild(intent, length, customer, arch, impl string) string {
	words := mapLengthToWords(length)
	msg := map[string]interface{}{
		"system_objective": "You are an ATD Analyst.",
		"intent":           intent,
		"target_length":    words + " words",
		"customer_layer":   customer,
		"arch_layer":       arch,
		"impl_layer":       impl,
		"instruction":      "Produce your final output ensuring it achieves the exact intent specified.",
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// AssembleFormat returns the JSON schema for assembly results.
func AssembleFormat() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"document": map[string]string{"type": "string"},
		},
		"required": []string{"document"},
	}
}
