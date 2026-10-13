package sql

import (
	"context"
	"database/sql"
)

func (db *Sqlite) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if db.db == nil {
		return nil, ErrNilDB
	}
	return db.db.ExecContext(ctx, query, args...)
}

// Transaction commits related control writes together. The caller serializes
// access with its Manager lock; no cached statement is shared with the Tx.
func (db *Sqlite) Transaction(f func(*sql.Tx) error) error {
	return db.TransactionContext(context.Background(), f)
}

// TransactionContext bounds admin snapshot/restore work without sharing statements.
func (db *Sqlite) TransactionContext(ctx context.Context, f func(*sql.Tx) error) error {
	if db.db == nil {
		return ErrNilDB
	}
	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err = f(tx); err != nil {
		return err
	}
	return tx.Commit()
}
