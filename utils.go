package db

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

func WithStmt(
	ctx context.Context,
	db StmtPreparor,
	query string,
	fn func(stmt *sql.Stmt) error,
) (err error) {
	var stmt *sql.Stmt
	stmt, err = db.PrepareContext(ctx, query)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	defer func() {
		e := stmt.Close()
		if e != nil && err == nil {
			err = errors.WithStack(e)
		}
	}()
	err = fn(stmt)
	if err != nil {
		return
	}
	return nil
}

func WithTx(
	ctx context.Context,
	db TxBeginor,
	txOpts *sql.TxOptions,
	fn func(tx *sql.Tx) error,
) (err error) {
	if txOpts == nil {
		txOpts = new(sql.TxOptions)
	}
	var tx *sql.Tx
	tx, err = db.BeginTx(ctx, txOpts)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		e := tx.Rollback()
		if e != nil && err == nil && !errors.Is(e, sql.ErrTxDone) {
			err = errors.WithStack(e)
		}
	}()
	err = fn(tx)
	if err != nil {
		return errors.WithStack(err)
	}
	err = errors.WithStack(tx.Commit())
	return
}

func WithTxStmt(
	ctx context.Context,
	db TxBeginor,
	txOpts *sql.TxOptions,
	query string,
	fn func(stmt *sql.Stmt) error,
) (err error) {
	return WithTx(ctx, db, txOpts, func(tx *sql.Tx) error {
		return WithStmt(ctx, tx, query, fn)
	})
}
