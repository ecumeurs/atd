package exploration

// Regression test for failures/20260901_atd_search_stale_index_returns_deleted_code.md
// (fixed): SemanticSearch used to return indexed chunks with no check
// against the current state of the file they came from. A chunk indexed
// for code that had since been deleted (or edited) on disk came back
// looking identical to a live, fresh hit — same FilePath, same ChunkText,
// same similarity score, no error and no staleness signal of any kind.
//
// The fix compares each entry's stored store.IndexEntry.Mtime (persisted on
// every PutEmbedding call, see pkg/store/index.go) against the live file's
// current mtime, stat'ing the file at query time. A missing file, or an
// mtime that no longer matches, now sets SearchResult.Stale = true instead
// of silently presenting the result as fresh. The result is still
// returned — not dropped — so a caller can still see "here's what used to
// be here" while knowing to treat it as historical rather than live.
import (
	"os"
	"path/filepath"
	"testing"

	"atd-tools/pkg/store"
	"atd-tools/pkg/testutil/fakeprovider"
)

func TestSemanticSearch_StaleIndexEntryIsMarkedStale(t *testing.T) {
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

	info, err := os.Stat(codeFile)
	if err != nil {
		t.Fatalf("stat fixture source file: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "index.db")
	seedIndex(t, dbPath, []store.IndexEntry{
		{ChunkID: "c1", Mtime: info.ModTime().Unix(), Content: liveContent, Embedding: []float32{1, 0, 0}, AtomPath: codeFile},
	})

	before, err := SemanticSearch("zzfix stale query", dbPath, 5, "all", "")
	if err != nil {
		t.Fatalf("SemanticSearch (pre-delete): %v", err)
	}
	if len(before) != 1 || before[0].ChunkText != liveContent {
		t.Fatalf("expected the live chunk to be returned before deletion, got %+v", before)
	}
	if before[0].Stale {
		t.Fatalf("expected the unmodified, still-present chunk to be reported fresh (Stale=false), got Stale=true")
	}

	// The indexed expression is now deleted from disk — WITHOUT rebuilding
	// the index — exactly as in the report (an uncommitted refactor round
	// deleted the code; 'atd index' was never rerun before the next query).
	if err := os.Remove(codeFile); err != nil {
		t.Fatalf("removing fixture source file: %v", err)
	}

	after, err := SemanticSearch("zzfix stale query", dbPath, 5, "all", "")
	if err != nil {
		t.Fatalf("SemanticSearch (post-delete): %v", err)
	}
	// The vanished chunk is still returned — dropping it would lose
	// information a caller might still want ("here's what used to be
	// here") — but it must now be flagged as stale.
	if len(after) != 1 {
		t.Fatalf("expected the stale chunk to still be returned (got %d results)", len(after))
	}
	if after[0].ChunkText != liveContent {
		t.Fatalf("stale chunk text should be unchanged from the pre-deletion result, got %q", after[0].ChunkText)
	}
	if after[0].FilePath != codeFile {
		t.Fatalf("FilePath should be unchanged after deletion, got %q", after[0].FilePath)
	}
	if after[0].Similarity != before[0].Similarity {
		t.Fatalf("similarity score should be unchanged after the backing file vanished (%.4f -> %.4f)", before[0].Similarity, after[0].Similarity)
	}
	if !after[0].Stale {
		t.Fatalf("expected the result for a deleted backing file to be marked Stale=true, got Stale=false")
	}

	if _, err := os.Stat(codeFile); !os.IsNotExist(err) {
		t.Fatalf("test setup invariant broken: fixture file should be gone, stat err = %v", err)
	}
}
