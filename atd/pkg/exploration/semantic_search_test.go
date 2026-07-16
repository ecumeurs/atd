package exploration

// WP-6 (test_atd_07_26.md §3.5, §6): deterministic tests for semantic
// search's embedding mode, using pkg/testutil/fakeprovider's fake embedder
// (fixed vectors, real cosine math — pkg/cosine is already fully covered)
// instead of a live Ollama. Before this, embedding-mode search had zero
// deterministic coverage: the only search tests exercised the grep
// fallback and workspace plumbing.

import (
	"os"
	"path/filepath"
	"testing"

	"atd-tools/pkg/store"
	"atd-tools/pkg/testutil/fakeprovider"
)

// seedIndex creates a real sqlite index db at dbPath with one embedded
// chunk per entry. Returns nothing; fails the test on any error.
func seedIndex(t *testing.T, dbPath string, entries []store.IndexEntry) {
	t.Helper()
	st, err := store.NewStore(dbPath)
	if err != nil {
		t.Fatalf("store.NewStore: %v", err)
	}
	defer st.Close()
	for _, e := range entries {
		if err := st.PutEmbedding(e); err != nil {
			t.Fatalf("PutEmbedding(%s): %v", e.ChunkID, err)
		}
	}
}

// TestSemanticSearch_RankingWithFixedVectors pins the core embedding-mode
// contract: results come back ranked by cosine similarity to the query
// embedding, highest first, truncated to limit. Vectors are chosen so the
// ranking is unambiguous: query = (1,0,0); docs at decreasing similarity.
func TestSemanticSearch_RankingWithFixedVectors(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetEmbedTable(map[string][]float32{
		"zzfix query": {1, 0, 0},
	})

	dbPath := filepath.Join(t.TempDir(), "index.db")
	seedIndex(t, dbPath, []store.IndexEntry{
		{ChunkID: "c1", Content: "exact match", Embedding: []float32{1, 0, 0}, AtomPath: "docs/zzfix_exact.atom.md"},
		{ChunkID: "c2", Content: "close match", Embedding: []float32{0.9, 0.1, 0}, AtomPath: "docs/zzfix_close.atom.md"},
		{ChunkID: "c3", Content: "orthogonal", Embedding: []float32{0, 1, 0}, AtomPath: "docs/zzfix_far.atom.md"},
	})

	results, err := SemanticSearch("zzfix query", dbPath, 2, "all", "")
	if err != nil {
		t.Fatalf("SemanticSearch: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected limit=2 results, got %d", len(results))
	}
	if results[0].FilePath != "docs/zzfix_exact.atom.md" {
		t.Errorf("expected the identical vector to rank first, got %s (sim %.4f)", results[0].FilePath, results[0].Similarity)
	}
	if results[1].FilePath != "docs/zzfix_close.atom.md" {
		t.Errorf("expected the near vector second, got %s (sim %.4f)", results[1].FilePath, results[1].Similarity)
	}
	if !(results[0].Similarity > results[1].Similarity) {
		t.Errorf("similarities must be strictly ordered: %.4f then %.4f", results[0].Similarity, results[1].Similarity)
	}
	if results[0].Similarity < 0.999 {
		t.Errorf("identical vectors must score ~1.0, got %.4f", results[0].Similarity)
	}
}

// TestSemanticSearch_ScopeFilter pins the docs/code scope split: entries
// whose AtomPath ends in .atom.md are "docs", everything else is "code".
func TestSemanticSearch_ScopeFilter(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetEmbedTable(map[string][]float32{
		"zzfix query": {1, 0, 0},
	})

	dbPath := filepath.Join(t.TempDir(), "index.db")
	seedIndex(t, dbPath, []store.IndexEntry{
		{ChunkID: "d1", Content: "a doc", Embedding: []float32{1, 0, 0}, AtomPath: "docs/zzfix_doc.atom.md"},
		{ChunkID: "s1", Content: "some code", Embedding: []float32{1, 0, 0}, AtomPath: "src/zzfix_code.go"},
	})

	docs, err := SemanticSearch("zzfix query", dbPath, 10, "docs", "")
	if err != nil {
		t.Fatalf("SemanticSearch(docs): %v", err)
	}
	if len(docs) != 1 || docs[0].FilePath != "docs/zzfix_doc.atom.md" {
		t.Errorf("scope=docs must return only the atom entry, got %+v", docs)
	}

	code, err := SemanticSearch("zzfix query", dbPath, 10, "code", "")
	if err != nil {
		t.Fatalf("SemanticSearch(code): %v", err)
	}
	if len(code) != 1 || code[0].FilePath != "src/zzfix_code.go" {
		t.Errorf("scope=code must return only the code entry, got %+v", code)
	}
}

// TestSemanticSearch_MissingIndexIsLoud pins S10's invariant at the package
// level: a missing index db must yield ErrIndexMissing, never a silent
// empty result set.
func TestSemanticSearch_MissingIndexIsLoud(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetEmbedTable(map[string][]float32{"anything": {1}})

	_, err := SemanticSearch("anything", filepath.Join(t.TempDir(), "nonexistent.db"), 5, "all", "")
	if err == nil {
		t.Fatal("expected ErrIndexMissing for an absent index db, got nil")
	}
	if err != ErrIndexMissing {
		t.Errorf("expected ErrIndexMissing, got: %v", err)
	}
}

// TestSemanticSearch_EmptyIndexIsLoud: a present but empty index (schema
// exists, 0 rows) must also be ErrIndexMissing — not an empty success.
func TestSemanticSearch_EmptyIndexIsLoud(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetEmbedTable(map[string][]float32{"anything": {1}})

	dbPath := filepath.Join(t.TempDir(), "index.db")
	seedIndex(t, dbPath, nil) // creates schema, inserts nothing

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected db file to exist: %v", err)
	}

	_, err := SemanticSearch("anything", dbPath, 5, "all", "")
	if err != ErrIndexMissing {
		t.Errorf("expected ErrIndexMissing for an empty index, got: %v", err)
	}
}

// TestSemanticSearch_EmbedFailureIsLoud: if embedding the query itself
// fails (provider down mid-session), the search must error, not return an
// empty result.
func TestSemanticSearch_EmbedFailureIsLoud(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	// Deliberately no vector for the query text: the fake embedder errors.
	fake.SetEmbedTable(map[string][]float32{})

	dbPath := filepath.Join(t.TempDir(), "index.db")
	seedIndex(t, dbPath, []store.IndexEntry{
		{ChunkID: "c1", Content: "x", Embedding: []float32{1}, AtomPath: "docs/zzfix.atom.md"},
	})

	_, err := SemanticSearch("unseeded query", dbPath, 5, "all", "")
	if err == nil {
		t.Fatal("expected an error when query embedding fails, got nil")
	}
}
