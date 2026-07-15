package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type IndexEntry struct {
	ChunkID   string
	Mtime     int64
	Content   string
	Embedding []float32
	AtomID    string
	AtomPath  string
}

func (s *Store) PutEmbedding(entry IndexEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	embBytes, err := json.Marshal(entry.Embedding)
	if err != nil {
		return fmt.Errorf("failed to marshal embedding: %w", err)
	}

	_, err = s.db.Exec(`
		INSERT INTO atom_index (chunk_id, mtime, content, embedding, atom_id, atom_path) 
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(chunk_id) DO UPDATE SET 
			mtime=excluded.mtime, 
			content=excluded.content, 
			embedding=excluded.embedding,
			atom_id=excluded.atom_id,
			atom_path=excluded.atom_path
	`, entry.ChunkID, entry.Mtime, entry.Content, embBytes, entry.AtomID, entry.AtomPath)

	if err != nil {
		return fmt.Errorf("failed to put embedding: %w", err)
	}

	return nil
}

func (s *Store) GetEmbedding(chunkID string) (*IndexEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	row := s.db.QueryRow(`
		SELECT chunk_id, mtime, content, embedding, atom_id, atom_path
		FROM atom_index WHERE chunk_id = ?
	`, chunkID)

	var entry IndexEntry
	var embJSON []byte
	err := row.Scan(&entry.ChunkID, &entry.Mtime, &entry.Content, &embJSON, &entry.AtomID, &entry.AtomPath)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(embJSON, &entry.Embedding); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (s *Store) ListAll() ([]IndexEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query("SELECT chunk_id, mtime, content, embedding, atom_id, atom_path FROM atom_index")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []IndexEntry
	for rows.Next() {
		var entry IndexEntry
		var embJSON []byte
		if err := rows.Scan(&entry.ChunkID, &entry.Mtime, &entry.Content, &embJSON, &entry.AtomID, &entry.AtomPath); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(embJSON, &entry.Embedding); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *Store) DeleteAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM atom_index")
	if err != nil {
		return fmt.Errorf("failed to delete all entries: %w", err)
	}

	return nil
}

func (s *Store) DeleteByAtomPath(atomPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM atom_index WHERE atom_path = ?", atomPath)
	if err != nil {
		return fmt.Errorf("failed to delete entries for atom path: %w", err)
	}

	return nil
}