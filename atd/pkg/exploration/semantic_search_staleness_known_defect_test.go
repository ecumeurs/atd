package exploration

// KNOWN DEFECT pin for failures/20260901_atd_search_stale_index_returns_deleted_code.md:
// SemanticSearch (search.go) returns indexed chunks with no check against
// the current state of the file they came from. A chunk indexed for code
// that has since been deleted (or edited) on disk comes back looking
// identical to a live, fresh hit — same FilePath, same ChunkText, same
// similarity score, no error and no staleness signal of any kind.
//
// The gap is sharper than "nobody re-stats the file": store.IndexEntry
// already carries an Mtime column (see pkg/store/index.go) that gets
// persisted on every PutEmbedding call — but SemanticSearch's ListAll loop
// discards it before building the SearchResult (search.go: `SearchResult{
// entry.AtomPath, entry.Content, sim, projectName}` — no Mtime field at
// all). The staleness data already exists in storage; it is simply never
// compared against the live file or surfaced to the caller.
//
// If this test ever starts failing — the stale chunk stops coming back
// unchanged, or an error/marker appears — treat that as a fix landing:
// relax or delete this pin and mark the report resolved.
import (
	"os"
	"path/filepath"
	"testing"

	"atd-tools/pkg/store"
	"atd-tools/pkg/testutil/fakeprovider"
)

func TestSemanticSearch_KnownDefect_StaleIndexReturnsDeletedCode(t *testing.T) {
	fake := fakeprovider.InstallOllama(t)
	fake.SetEmbedTable(map[string][]float32{
		"zzfix stale query": {1, 0, 0},
	})

	codeDir := t.TempDir()
	codeFile := filepath.Join(codeDir, "zzfix_bridge_start.go")
	liveContent := "func BridgeStart() { PropertyToString(effectiveKey) }"
	if err := os.WriteFile(codeFile, []byte(liveContent), 0644); err != nil {
		t.Fatalf("writing fixture source file: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "index.db")
	seedIndex(t, dbPath, []store.IndexEntry{
		{ChunkID: "c1", Mtime: 12345, Content: liveContent, Embedding: []float32{1, 0, 0}, AtomPath: codeFile},
	})

	before, err := SemanticSearch("zzfix stale query", dbPath, 5, "all", "")
	if err != nil {
		t.Fatalf("SemanticSearch (pre-delete): %v", err)
	}
	if len(before) != 1 || before[0].ChunkText != liveContent {
		t.Fatalf("expected the live chunk to be returned before deletion, got %+v", before)
	}

	// The indexed expression is now deleted from disk — WITHOUT rebuilding
	// the index — exactly as in the report (an uncommitted refactor round
	// deleted the code; 'atd index' was never rerun before the next query).
	if err := os.Remove(codeFile); err != nil {
		t.Fatalf("removing fixture source file: %v", err)
	}

	after, err := SemanticSearch("zzfix stale query", dbPath, 5, "all", "")
	if err != nil {
		t.Fatalf("KNOWN DEFECT expectation changed: SemanticSearch now errors when a result's backing file is missing from disk (%v) — that would be a fix; if intentional, relax this pin", err)
	}
	if len(after) != 1 {
		t.Fatalf("KNOWN DEFECT expectation changed: expected the stale chunk to still be returned (got %d results) — filtering out vanished files would be a fix; if intentional, relax this pin", len(after))
	}
	if after[0].ChunkText != liveContent {
		t.Fatalf("KNOWN DEFECT expectation changed: stale chunk text differs from the pre-deletion result (%q) — investigate before relaxing this pin", after[0].ChunkText)
	}
	if after[0].FilePath != codeFile {
		t.Fatalf("KNOWN DEFECT expectation changed: FilePath changed after deletion (%q) — investigate before relaxing this pin", after[0].FilePath)
	}
	if after[0].Similarity != before[0].Similarity {
		t.Fatalf("KNOWN DEFECT expectation changed: similarity score changed after the backing file vanished (%.4f -> %.4f) — investigate before relaxing this pin", before[0].Similarity, after[0].Similarity)
	}

	if _, err := os.Stat(codeFile); !os.IsNotExist(err) {
		t.Fatalf("test setup invariant broken: fixture file should be gone, stat err = %v", err)
	}
}
