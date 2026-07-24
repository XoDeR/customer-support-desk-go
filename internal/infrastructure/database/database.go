package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewPostgresConnection(c *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.Name, c.Database.SSLMode)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if c.Database.MaxOpenConns > 0 {
		db.SetMaxOpenConns(c.Database.MaxOpenConns)
	}
	db.SetConnMaxLifetime(time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

type txKey struct{}
type TransactionManager interface {
	WithTransaction(context.Context, func(context.Context) error) error
}
type transactionManager struct{ db *sql.DB }

func NewTransactionManager(db *sql.DB) TransactionManager { return transactionManager{db} }
func (m transactionManager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	if err = fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
func Tx(ctx context.Context) (*sql.Tx, bool) { tx, ok := ctx.Value(txKey{}).(*sql.Tx); return tx, ok }
