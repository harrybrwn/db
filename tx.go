package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

type Tx interface {
	DB
	Commit() error
	Rollback() error
}

// Begin will begin a transaction.
func Begin(ctx context.Context, opts *sql.TxOptions, database any) (Tx, error) {
	switch db := database.(type) {
	case Tx:
		return nil, errors.New("cannot start a transaction from a transaction")
	case DB:
		return db.BeginTx(ctx, opts)
	case *sql.DB:
		t, err := db.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &tx{Tx: t}, nil
	case TxBeginor:
		t, err := db.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &tx{Tx: t}, nil
	}
	return nil, fmt.Errorf("cannot start a transaction using %T", database)
}

func TxDo(ctx context.Context, tx Tx, fn func(tx Tx) error) (err error) {
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

// NewTx creates a wrapper around the standard library [sql.Tx] and returns a
// wrapper type that implements [DB].
func NewTx(tr *sql.Tx) *tx { return &tx{Tx: tr} }

type tx struct{ *sql.Tx }

func (tx *tx) QueryContext(ctx context.Context, query string, v ...any) (Rows, error) {
	return tx.Tx.QueryContext(ctx, query, v...)
}

// BeginTx is a noop because this is already a transaction. Should be used with caution.
func (tx *tx) BeginTx(context.Context, *sql.TxOptions) (Tx, error) {
	return tx, nil
}

var ErrCannotCloseTx = errors.New("cannot close a transaction. Use Commit or Rollback")

// Close does nothing because transactions cannot be closed
func (tx *tx) Close() error { return ErrCannotCloseTx }
