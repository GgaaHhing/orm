package orm

import (
	"context"
	"database/sql"
	"errors"
)

var (
	_ Session = &Tx{}
	_ Session = &DB{}
)

// Session DB和Tx的共同抽象
type Session interface {
	getCore() core
	// 查询
	queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	// 执行
	execContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx struct {
	tx *sql.Tx
	db *DB
}

func (t *Tx) getCore() core {
	return t.db.core
}

func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

func (t *Tx) RollbackIfNotCommit() error {
	err := t.tx.Rollback()
	// ErrTxDone 对已提交或回滚的事务执行的任何操作都会返回 ErrTxDone
	if errors.Is(err, sql.ErrTxDone) {
		return nil
	}
	return err
}
