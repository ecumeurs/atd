package store

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) (*Store, string) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}

	return store, dbPath
}

func TestNewStore(t *testing.T) {
	store, dbPath := setupTestDB(t)
	defer store.Close()

	if store == nil {
		t.Fatal("store is nil")
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("database file was not created")
	}
}

func TestPutAndGetEmbedding(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	embedding := []float32{0.1, 0.2, 0.3}

	err := store.PutEmbedding(IndexEntry{
		ChunkID:   "chunk1",
		Embedding: embedding,
		Mtime:     12345,
		Content:   "test content",
		AtomID:    "atom1",
		AtomPath:  "path/to/atom.md",
	})
	if err != nil {
		t.Fatalf("failed to put embedding: %v", err)
	}

	retrieved, err := store.GetEmbedding("chunk1")
	if err != nil {
		t.Fatalf("failed to get embedding: %v", err)
	}

	if retrieved == nil {
		t.Fatal("retrieved embedding is nil")
	}

	if len(retrieved.Embedding) != len(embedding) {
		t.Fatalf("embedding length mismatch: got %d, want %d", len(retrieved.Embedding), len(embedding))
	}

	for i := range embedding {
		if retrieved.Embedding[i] != embedding[i] {
			t.Errorf("embedding value mismatch at index %d: got %f, want %f", i, retrieved.Embedding[i], embedding[i])
		}
	}
}

func TestPutAndGetAuditCache(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	embedding := []float32{0.4, 0.5, 0.6}

	err := store.PutAuditCache("atom1", embedding, "test intent", "test logic", 12345)
	if err != nil {
		t.Fatalf("failed to put audit cache: %v", err)
	}

	retrieved, err := store.GetAuditCache("atom1")
	if err != nil {
		t.Fatalf("failed to get audit cache: %v", err)
	}

	if retrieved == nil {
		t.Fatal("retrieved audit cache is nil")
	}

	if retrieved.AtomID != "atom1" {
		t.Errorf("atom ID mismatch: got %s, want atom1", retrieved.AtomID)
	}

	if retrieved.Intent != "test intent" {
		t.Errorf("intent mismatch: got %s, want test intent", retrieved.Intent)
	}

	if retrieved.Logic != "test logic" {
		t.Errorf("logic mismatch: got %s, want test logic", retrieved.Logic)
	}
}

func TestPutAndGetCollisionCache(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	collisionData := `{"collision": "detected"}`
	
	err := store.PutCollisionCache("atom1", collisionData, 12345)
	if err != nil {
		t.Fatalf("failed to put collision cache: %v", err)
	}

	retrieved, err := store.GetCollisionCache("atom1")
	if err != nil {
		t.Fatalf("failed to get collision cache: %v", err)
	}

	if retrieved == nil {
		t.Fatal("retrieved collision cache is nil")
	}

	if retrieved.AtomID != "atom1" {
		t.Errorf("atom ID mismatch: got %s, want atom1", retrieved.AtomID)
	}

	if retrieved.CollisionData != collisionData {
		t.Errorf("collision data mismatch: got %s, want %s", retrieved.CollisionData, collisionData)
	}
}

func TestDeleteAll(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	embedding := []float32{0.1, 0.2, 0.3}
	store.PutEmbedding(IndexEntry{ChunkID: "chunk1", Embedding: embedding, Mtime: 12345, Content: "content1", AtomID: "atom1", AtomPath: "path1"})
	store.PutEmbedding(IndexEntry{ChunkID: "chunk2", Embedding: embedding, Mtime: 12346, Content: "content2", AtomID: "atom2", AtomPath: "path2"})

	err := store.DeleteAll()
	if err != nil {
		t.Fatalf("failed to delete all: %v", err)
	}

	entries, err := store.ListAll()
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries after delete, got %d", len(entries))
	}
}

func TestDeleteAuditCache(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	embedding := []float32{0.1, 0.2, 0.3}
	store.PutAuditCache("atom1", embedding, "intent", "logic", 12345)

	err := store.DeleteAuditCache("atom1")
	if err != nil {
		t.Fatalf("failed to delete audit cache: %v", err)
	}

	retrieved, err := store.GetAuditCache("atom1")
	if err != nil {
		t.Fatalf("failed to get audit cache: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil after delete, got entry")
	}
}

func TestDeleteCollisionCache(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	store.PutCollisionCache("atom1", "collision data", 12345)

	err := store.DeleteCollisionCache("atom1")
	if err != nil {
		t.Fatalf("failed to delete collision cache: %v", err)
	}

	retrieved, err := store.GetCollisionCache("atom1")
	if err != nil {
		t.Fatalf("failed to get collision cache: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil after delete, got entry")
	}
}

func TestListAll(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	embedding := []float32{0.1, 0.2, 0.3}
	store.PutEmbedding(IndexEntry{ChunkID: "chunk1", Embedding: embedding, Mtime: 12345, Content: "content1", AtomID: "atom1", AtomPath: "path1"})
	store.PutEmbedding(IndexEntry{ChunkID: "chunk2", Embedding: embedding, Mtime: 12346, Content: "content2", AtomID: "atom2", AtomPath: "path2"})

	entries, err := store.ListAll()
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestGetEmbeddingNotFound(t *testing.T) {
	store, _ := setupTestDB(t)
	defer store.Close()

	embedding, err := store.GetEmbedding("nonexistent")
	if err != nil {
		t.Fatalf("failed to get embedding: %v", err)
	}

	if embedding != nil {
		t.Error("expected nil for nonexistent embedding")
	}
}