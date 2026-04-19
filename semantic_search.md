# Tour of Nomic & Semantic Search in ATD

Atomic Traceable Documentation (ATD) leverages vector embeddings and semantic search to bridge the gap between human requirements and source code. This tour explains how these technologies are integrated into the core engine.

---

## 🏗️ Core Stack

ATD uses a localized, offline-first approach to semantic search:

1.  **Embedding Model:** [Nomic Embed Text](https://huggingface.co/nomic-ai/nomic-embed-text-v1.5) (via Ollama). This model is specifically optimized for long-context retrieval and code-to-text matching.
2.  **Vector Store:** SQLite (using the `.atd_index.db` file). Instead of a heavy vector database, we store embeddings as JSON blobs in a standard SQLite table for portability and simplicity.
3.  **Comparison Engine:** A custom Go implementation of **Cosine Similarity** (`pkg/cosine/similarity.go`) used to rank chunks by relevance to a query.

```go 

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

---

## 📥 The Indexing Pipeline (`atd index`)

Before semantic search can be used, the codebase must be indexed. The `atd index` command performs the following steps:

### 1. File Discovery & Filtering
The engine crawls the project directory, respecting `.gitignore` and filtering files based on `SupportedExtensions` (defined in `.atd` config) and the `.atom.md` suffix.

### 2. Intelligent Chunking
To fit within the embedding model's context window and maintain semantic focus, files are split:
-   **Documentation (`.atom.md`):** Split by `## ` H2 sections (Intent, Logic, Interface, Expectation). This ensures that each section is indexed as a distinct semantic unit.
-   **Source Code:** Split by double newlines (`\n\n`), effectively grouping logical blocks (classes, functions, or cohesive logic units).

### 3. Embedding Generation
Each chunk is sent to the configured Ollama provider. 
> [!IMPORTANT]
> Unlike text generation tasks, **embedding generation has no IDE fallback**. It requires a local or remote Ollama server with the `nomic-embed-text` model installed.

### 4. Incremental Updates
The indexer tracks file modification times (`mtime`). Only files that have changed since the last indexing run are re-processed, significantly speeding up subsequent runs.

---

## 🔍 Semantic Search Workflow (`atd search`)

When you run `atd search --query "..."`, the following occurs:

1.  **Query Embedding:** Your search string is converted into a vector using the same Nomic model.
2.  **Similarity Scan:** The engine performs a linear scan of the SQLite database, calculating the cosine similarity between the query vector and every stored chunk vector.
3.  **Ranking & Filtering:** Results are sorted from highest to lowest similarity. You can scope searches to `code`, `docs`, or `all` via flags.

---

## ⚙️ Configuration

The search behavior is controlled via the `.atd` configuration file:

```json
{
  "llm": {
    "providers": [
      {
        "name": "local_ollama",
        "base_url": "http://localhost:11434",
        "timeout_ms": 10000
      }
    ],
    "models": {
      "nomic-embed-text": {
        "tasks": ["embed"],
        "priority": 10
      }
    }
  }
}
```

- **Tasks:** A model must be tagged with the `"embed"` task to be used by the indexer.
- **Priority:** ATD will try the highest priority provider that is currently online.

---

## 💡 Best Practices

-   **Re-index after refactors:** If you move large blocks of code or change documentation structures, run `atd index` to ensure your semantic links remain accurate.
-   **Use broad queries:** Semantic search excels at finding *meaning* rather than exact matches. Instead of searching for `func processUserData`, try `where do we handle user profile logic?`.
-   **Combine with `atd trace`:** Use search to find a starting atom, then use `atd trace [[id]]` to see its full relationship graph.
