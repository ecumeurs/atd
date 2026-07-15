package store

import (
	"fmt"
)

func (s *Store) runMigrations() error {
	migrations := []string{
		s.createAtomIndexTable(),
		s.createAuditCacheTable(),
		s.createCollisionCacheTable(),
	}

	for _, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

func (s *Store) createAtomIndexTable() string {
	return `CREATE TABLE IF NOT EXISTS atom_index (
		chunk_id TEXT PRIMARY KEY,
		mtime INTEGER,
		content TEXT,
		embedding BLOB,
		atom_id TEXT,
		atom_path TEXT
	)`
}

func (s *Store) createAuditCacheTable() string {
	return `CREATE TABLE IF NOT EXISTS audit_cache (
		atom_id TEXT PRIMARY KEY,
		embedding BLOB,
		intent TEXT,
		logic TEXT,
		mtime INTEGER
	)`
}

func (s *Store) createCollisionCacheTable() string {
	return `CREATE TABLE IF NOT EXISTS collision_cache (
		atom_id TEXT PRIMARY KEY,
		collision_data TEXT,
		mtime INTEGER
	)`
}