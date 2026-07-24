package migration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Manager struct {
	DB        *sql.DB
	Directory string
}

func NewManager(db *sql.DB, directory string) *Manager { return &Manager{db, directory} }
func (m *Manager) Migrate(ctx context.Context) error {
	if _, err := m.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(version bigint primary key, applied_at timestamptz not null default now())`); err != nil {
		return err
	}
	entries, err := os.ReadDir(m.Directory)
	if err != nil {
		return err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		v, err := strconv.ParseInt(strings.Split(e.Name(), "_")[0], 10, 64)
		if err != nil {
			return err
		}
		var exists bool
		if err = m.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)", v).Scan(&exists); err != nil {
			return err
		}
		if exists {
			continue
		}
		sqlText, err := os.ReadFile(filepath.Join(m.Directory, e.Name()))
		if err != nil {
			return err
		}
		tx, err := m.DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, string(sqlText)); err == nil {
			_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES($1)", v)
		}
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", e.Name(), err)
		}
		if err = tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
