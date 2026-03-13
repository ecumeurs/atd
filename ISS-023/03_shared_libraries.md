# Task 03 — Shared Libraries

**Depends on:** Task 01 (module structure exists)  
**Produces:** `internal/atom/`, `internal/cosine/`, `internal/ollama/client.go`

## Context

These structs and functions are duplicated across multiple tools today:
- `EmbeddingRequest`/`EmbeddingResponse` — in `atd-ollama-indexer`, `atd-ollama-search`, `atd-audit`, `atd-discover-links`
- `GenerateRequest`/`GenerateResponse` — in `atd-ollama-audit`, `atd-ollama-generate`, `atd-compare`, `atd-audit-fixer`, `atd-audit`, `atd-discover-links`
- `cosineSimilarity()` — in `atd-ollama-search`, `atd-audit`, `atd-discover-links`
- Atom parsers — in `atd-audit` (`parseAtomFile`), `atd-compare` (`parseAtom`), `atd-audit-fixer` (`readAtomMeta`)

## Steps

### 3.1 Create `scripts/internal/atom/parse.go`

Package: `atom`

Extract and unify all atom parsing into one file. Must handle:

```go
package atom

// AtomData holds parsed atom metadata and content sections.
type AtomData struct {
    ID        string
    HumanName string
    Type      string
    Status    string
    Priority  string
    Tags      []string
    Parents   []string
    Dependents []string
    Intent    string  // Content of ## INTENT section
    Logic     string  // Content of ## THE RULE / LOGIC section
    Interface string  // Content of ## TECHNICAL INTERFACE section
    FilePath  string  // Original file path
}

// Parse reads a full .atom.md file and returns all metadata + content sections.
// Reference: atd-compare/main.go parseAtom() (lines 62-149)
//            atd-audit/main.go parseAtomFile() (lines 106-190)
func Parse(path string) (AtomData, error) { ... }

// ParseMeta reads only the YAML frontmatter (fast, no body parsing).
// Reference: atd-audit-fixer/main.go readAtomMeta() (lines 153-174)
func ParseMeta(path string) (id, humanName, atomType string, err error) { ... }

// BuildContent generates a complete .atom.md file from an AtomData struct.
// Reference: atd-audit-fixer/main.go buildAtomContent() (lines 71-96)
func BuildContent(a AtomData) string { ... }

// BuildParentContent generates a MODULE parent .atom.md.
// Reference: atd-audit-fixer/main.go buildParentContent() (lines 98-122)
func BuildParentContent(id, humanName, intent, logic string) string { ... }
```

**Implementation note:** The parser must handle both inline parents (`parents: [[a]], [[b]]`) and multi-line list format (`parents:\n  - [[a]]\n  - [[b]]`). See `atd-audit/main.go` lines 130-170 for the complete logic.

### 3.2 Create `scripts/internal/cosine/similarity.go`

```go
package cosine

import "math"

// Similarity computes the cosine similarity between two float64 vectors.
// Returns 0.0 if either vector is empty or lengths don't match.
func Similarity(a, b []float64) float64 {
    if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
        return 0.0
    }
    var dot, magA, magB float64
    for i := range a {
        dot += a[i] * b[i]
        magA += a[i] * a[i]
        magB += b[i] * b[i]
    }
    if magA*magB == 0 {
        return 0.0
    }
    return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}
```

### 3.3 Create `scripts/internal/ollama/client.go`

Package: `ollama`

Unified HTTP client replacing all hardcoded calls:

```go
package ollama

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

// GenerateRequest is the Ollama /api/generate request body.
type GenerateRequest struct {
    Model   string      `json:"model"`
    Prompt  string      `json:"prompt"`
    Stream  bool        `json:"stream"`
    Format  interface{} `json:"format,omitempty"`  // string "json" or JSON schema object
    Options *Options    `json:"options,omitempty"`
}

type Options struct {
    Temperature float64 `json:"temperature,omitempty"`
    NumCtx      int     `json:"num_ctx,omitempty"`
}

// GenerateResponse is the Ollama /api/generate response.
type GenerateResponse struct {
    Response        string `json:"response"`
    PromptEvalCount int    `json:"prompt_eval_count"`
    EvalCount       int    `json:"eval_count"`
}

// EmbeddingRequest is the Ollama /api/embeddings request body.
type EmbeddingRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
}

// EmbeddingResponse is the Ollama /api/embeddings response.
type EmbeddingResponse struct {
    Embedding []float64 `json:"embedding"`
}

// TagsResponse is the Ollama /api/tags response.
type TagsResponse struct {
    Models []struct {
        Name string `json:"name"`
    } `json:"models"`
}

// Generate sends a generation request to an Ollama endpoint.
func Generate(baseURL, model, prompt string, format interface{}, opts *Options) (*GenerateResponse, error) { ... }

// Embed sends an embedding request to an Ollama endpoint.
func Embed(baseURL, model, text string) ([]float64, error) { ... }

// ListModels queries /api/tags to get available models on an endpoint.
func ListModels(baseURL string, timeoutMs int) ([]string, error) { ... }
```

**Key design decision:** `Format` is `interface{}` to support both `"json"` (simple mode) and full JSON schema objects (structured output). See Task 05 for format schemas.

### 3.4 Write tests

Create `scripts/internal/atom/parse_test.go`:
- Test `Parse()` against a sample `.atom.md` (inline in test as string)
- Test `ParseMeta()` returns correct id/humanName/type
- Test multi-line vs inline parents parsing

Create `scripts/internal/cosine/similarity_test.go`:
- Test identical vectors → 1.0
- Test orthogonal vectors → 0.0
- Test empty vectors → 0.0

## Acceptance Criteria

- [ ] `go test ./internal/atom/` passes
- [ ] `go test ./internal/cosine/` passes
- [ ] `internal/ollama/client.go` compiles (HTTP tests require live Ollama, skip for now)
- [ ] No duplicate struct definitions remain in new code
