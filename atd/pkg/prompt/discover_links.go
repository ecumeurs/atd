package prompt

import (
	"encoding/json"
)

// DiscoverLinksBuild constructs the link discovery prompt in JSON format.
//
// The instruction block is deliberately explicit about what a recommendation
// is (a verbatim id copied out of atom_registry) and what it is not (prose,
// advice, URLs, refusals). Three field reports — failures/20260901_atd_map_
// hallucinated_code_values.md and the 20260917 recurrence — show models
// happily filling `recommendations` with code-review sentences and fabricated
// URLs when the prompt only says "recommend IDs". The Go side now drops those
// (see cmd/atd/cmd/map_link_validation.go); this tightens the request so the
// model produces fewer of them in the first place, and caps the rationale so
// it cannot consume the whole output budget.
func DiscoverLinksBuild(registryStr, codeContent string) string {
	msg := map[string]interface{}{
		"system_objective": "You are an ATD Link Discoverer. Read the Target Source Code carefully and deduce which Atoms from the Known Atom Registry define this logic. Output a list of recommended IDs.",
		"atom_registry":    registryStr,
		"target_code":      codeContent,
		"instruction": "Answer with the 'recommendations' array FIRST, then 'rationale'. " +
			"Every entry of 'recommendations' must be an atom id copied verbatim from atom_registry " +
			"(lowercase snake_case, e.g. 'rule_password_policy' or 'project:api_auth_login') and nothing else: " +
			"no sentences, no advice, no URLs, no explanations, no refusals. " +
			"Only recommend ids from the provided registry list if they truly map to the codebase logic. " +
			"If none of them apply, return an empty 'recommendations' array — never an apology or a sentence inside the array. " +
			"Keep 'rationale' under 400 characters (two short sentences at most).",
		"response_shape": `{"recommendations": ["atom_id", ...], "rationale": "at most two short sentences"}`,
	}
	b, _ := json.Marshal(msg)
	return string(b)
}

// discoverLinksSchema is the JSON schema handed to the provider as the
// structured-output format.
//
// Field order is load-bearing, which is why this is a literal rather than a
// map[string]interface{}. encoding/json sorts map keys alphabetically, so the
// previous map-built schema serialized "rationale" BEFORE "recommendations";
// grammar-constrained decoders (Ollama/llama.cpp) emit object keys in schema
// order, so every response began with an unbounded free-form rationale and the
// structured result was whatever survived the token budget. When the budget ran
// out mid-rationale the entire response was unparseable and `atd map` failed
// hard with "unexpected end of JSON input" — the crash in
// failures/20260901_atd_map_malformed_json_from_link_recommender.md and its two
// recurrences, reproduced on both a remote deepseek-r1:7b and a local
// llama3.2:latest.
//
// Putting "recommendations" first means truncation now costs at most the
// rationale, and the partial-JSON salvage in cmd/atd/cmd/map_link_validation.go
// can still recover the ids. "rationale" is also dropped from "required" so a
// model is free to stop after the useful part.
//
// Keywords are restricted to the widely-supported core (type/properties/items/
// required) on purpose: a "pattern"/"maxLength" constraint would be the natural
// way to pin the id shape, but schema-to-grammar support for those is uneven
// across providers and an unsupported keyword risks a hard 400 instead of a
// merely sloppy answer. Shape enforcement therefore lives in Go, where it is
// deterministic and testable.
const discoverLinksSchema = `{
  "type": "object",
  "properties": {
    "recommendations": {
      "type": "array",
      "items": {"type": "string"}
    },
    "rationale": {"type": "string"}
  },
  "required": ["recommendations"]
}`

// DiscoverLinksFormat returns the JSON schema for link discovery results.
func DiscoverLinksFormat() interface{} {
	return json.RawMessage(discoverLinksSchema)
}
