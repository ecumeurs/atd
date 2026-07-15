package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type AuditCacheEntry struct {
	AtomID    string
	Embedding []float32
	Intent    string
	Logic     string
	Mtime     int64
}

type CollisionCacheEntry struct {
	AtomID        string
	CollisionData string
	Mtime         int64
}

func (s *Store) GetAuditCache(atomID string) (*AuditCacheEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var entry AuditCacheEntry
	var embJSON []byte
	err := s.db.QueryRow("SELECT atom_id, embedding, intent, logic, mtime FROM audit_cache WHERE atom_id = ?", atomID).Scan(
		&entry.AtomID, &embJSON, &entry.Intent, &entry.Logic, &entry.Mtime)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get audit cache: %w", err)
	}

	if embJSON != nil {
		if err := json.Unmarshal(embJSON, &entry.Embedding); err != nil {
			return nil, fmt.Errorf("failed to unmarshal embedding: %w", err)
		}
	}

	return &entry, nil
}

func (s *Store) PutAuditCache(atomID string, embedding []float32, intent, logic string, mtime int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var embBytes []byte
	var err error
	if embedding != nil {
		embBytes, err = json.Marshal(embedding)
		if err != nil {
			return fmt.Errorf("failed to marshal embedding: %w", err)
		}
	}

	_, err = s.db.Exec(`
		INSERT INTO audit_cache (atom_id, embedding, intent, logic, mtime) 
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(atom_id) DO UPDATE SET 
			embedding=excluded.embedding, 
			intent=excluded.intent, 
			logic=excluded.logic,
			mtime=excluded.mtime
	`, atomID, embBytes, intent, logic, mtime)

	if err != nil {
		return fmt.Errorf("failed to put audit cache: %w", err)
	}

	return nil
}

func (s *Store) GetCollisionCache(atomID string) (*CollisionCacheEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var entry CollisionCacheEntry
	err := s.db.QueryRow("SELECT atom_id, collision_data, mtime FROM collision_cache WHERE atom_id = ?", atomID).Scan(
		&entry.AtomID, &entry.CollisionData, &entry.Mtime)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get collision cache: %w", err)
	}

	return &entry, nil
}

func (s *Store) PutCollisionCache(atomID string, collisionData string, mtime int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(`
		INSERT INTO collision_cache (atom_id, collision_data, mtime) 
		VALUES (?, ?, ?)
		ON CONFLICT(atom_id) DO UPDATE SET 
			collision_data=excluded.collision_data, 
			mtime=excluded.mtime
	`, atomID, collisionData, mtime)

	if err != nil {
		return fmt.Errorf("failed to put collision cache: %w", err)
	}

	return nil
}

func (s *Store) DeleteAuditCache(atomID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM audit_cache WHERE atom_id = ?", atomID)
	if err != nil {
		return fmt.Errorf("failed to delete audit cache: %w", err)
	}

	return nil
}

func (s *Store) DeleteCollisionCache(atomID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM collision_cache WHERE atom_id = ?", atomID)
	if err != nil {
		return fmt.Errorf("failed to delete collision cache: %w", err)
	}

	return nil
}