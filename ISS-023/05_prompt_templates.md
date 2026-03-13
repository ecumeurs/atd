# Task 05 — Prompt Templates

**Depends on:** Task 03 (ollama client format types)  
**Produces:** `internal/prompt/*.go` — one file per LLM task type

## Context

Prompt text is currently embedded inline in each tool's `main.go`. This makes prompt evolution risky — changing one prompt can break the tool's logic. Isolating prompts into dedicated files lets us evolve them independently and attach JSON format schemas for structured output.

## Steps

### 5.1 Create one file per task type

Each file exports two functions:
- `Build(args...) string` — constructs the prompt from inputs
- `FormatSchema() interface{}` — returns the Ollama JSON format schema (or `nil` for freeform)

**Files to create:**

| File | Task Type | Source Tool | Format |
|------|-----------|-------------|--------|
| `internal/prompt/dissect.go` | `dissect` | `atd-dissect/main.go` L50-65 | JSON schema: `{atoms: [{id, responsibility, line_range}]}` |
| `internal/prompt/audit_bloat.go` | `audit_bloat` | `atd-audit/main.go` L192-218 | Freeform: `YES` or `NO` |
| `internal/prompt/audit_code.go` | `audit_code` | `atd-ollama-audit/main.go` L92-106 | JSON schema: `{passed: bool, resolutionMessage: str}` |
| `internal/prompt/compare.go` | `compare` | `atd-compare/main.go` L190-216 | Freeform: plaintext diagnosis |
| `internal/prompt/fix_split.go` | `fix_split` | `atd-audit-fixer/main.go` L244-264 | JSON schema: `{parent_logic, splits: [{id_suffix, human_name, intent, logic}]}` |
| `internal/prompt/reconcile.go` | `reconcile` | `atd-reconcile/main.go` L48-60 | JSON schema: `[{proposed_id, relationship, change_context}]` |
| `internal/prompt/congruence.go` | `congruence` | `atd-congruence/main.go` L120-130 | Freeform: markdown table |
| `internal/prompt/recon.go` | `recon` | `atd-recon/main.go` L48-61 | JSON schema: `{Confidence: int, Mismatches: str}` |
| `internal/prompt/snapshot.go` | `snapshot` | `atd-generate-snapshot/main.go` L46-54 | Freeform: markdown prose |
| `internal/prompt/intent_extract.go` | `intent_extract` | `atd-discover-links/main.go` L136-137 | Freeform: 2 sentences |
| `internal/prompt/discover_links.go` | — (uses intent_extract) | `atd-discover-links/main.go` L211-224 | Freeform: ID list |

### 5.2 Example structure (dissect.go)

```go
package prompt

import "fmt"

// DissectBuild constructs the dissect prompt for a numbered document.
func DissectBuild(numberedContent string) string {
    return fmt.Sprintf(`
<System_Context>
You are an ATD Architect. The target document below has line numbers prepended (e.g., 001:). 
Identify atomic boundaries where a single architectural responsibility starts and ends.
</System_Context>

<Instruction>
1. Map each Atom to its exact line_range [start, end].
2. Identify the 'responsibility' as a deterministic skill definition.
3. Response must strictly follow the JSON schema.
</Instruction>

<Document>
%s
</Document>
`, numberedContent)
}

// DissectFormat returns the Ollama JSON format schema for dissect output.
func DissectFormat() interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "atoms": map[string]interface{}{
                "type": "array",
                "items": map[string]interface{}{
                    "type": "object",
                    "properties": map[string]interface{}{
                        "id":             map[string]string{"type": "string"},
                        "responsibility": map[string]string{"type": "string"},
                        "line_range": map[string]interface{}{
                            "type":     "array",
                            "items":    map[string]string{"type": "integer"},
                            "minItems": 2,
                            "maxItems": 2,
                        },
                    },
                    "required": []string{"id", "responsibility", "line_range"},
                },
            },
        },
    }
}
```

### 5.3 Prompt design rules

- Every prompt must include `<System_Context>` and `<Instruction>` XML tags
- JSON format schemas must be provided for any task expecting structured output
- Temperature should default to 0.0 for deterministic tasks
- `num_ctx` should be set to 16384 for tasks processing code files

### 5.4 Write tests

Create `scripts/internal/prompt/dissect_test.go`:
- Test that `DissectBuild("001: hello")` contains the content
- Test that `DissectFormat()` returns a valid map with "atoms" key
- Test all `*Format()` functions return valid JSON-serializable objects

## Acceptance Criteria

- [ ] 11 prompt files created in `internal/prompt/`
- [ ] Each file compiles cleanly
- [ ] `go test ./internal/prompt/` passes
- [ ] Every prompt with structured output has a corresponding `Format()` function
