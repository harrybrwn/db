package db

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

// TODO move to https://github.com/uber-go/mock
//  $ go install go.uber.org/mock/mockgen@latest

//go:generate mockgen -package=mockdb   -destination ./mockdb/db.go     . DB
//go:generate mockgen -package=mocktx   -destination ./mocktx/tx.go     . TxBeginor,StmtPreparor
//go:generate mockgen -package=mockrows -destination ./mockrows/rows.go . Rows,Pingable

var (
	ErrDBTimeout = errors.New("database ping timeout")
)

// DB is an abstract sql database type.
type DB interface {
	io.Closer
	QueryContext(context.Context, string, ...any) (Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

// Pingable is an abstract type that has Ping methods.
type Pingable interface {
	Ping() error
	PingContext(context.Context) error
}

// Scanner is an object that can be scanned such that the pointers passed will
// be populated with the correspoding row value of a database query.
type Scanner interface {
	Scan(...any) error
}

// TxBeginor is an abstract type that should be able to begin database
// transactions.
type TxBeginor interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// StmtPreparor is an abstract type that should be able to prepare database
// statements.
type StmtPreparor interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// Rows is an abstract iterator type that is returned by database queries.
type Rows interface {
	Scanner
	io.Closer
	Next() bool
	Err() error
}

// Scanable is an abstract type for objects that can scan themselves given an
// database scanner.
type Scanable interface {
	Scan(scanner Scanner) error
}

// StdDB is an interface that is here just in case you need a DB interface that
// is more easily compatible with the standard library's [database/sql.DB].
type StdDB interface {
	io.Closer
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// ScanOne will scan one row from a query and then close the Rows object.
func ScanOne(r Rows, dest ...any) (err error) {
	if !r.Next() {
		if err = r.Err(); err != nil {
			r.Close()
			return err
		}
		r.Close()
		return sql.ErrNoRows
	}
	if err = r.Scan(dest...); err != nil {
		r.Close()
		return err
	}
	return r.Close()
}

type dbOptions struct {
	logger *slog.Logger
}

type Option func(*dbOptions)

// WithLogger sets the logger to use with an resource that takes an [Option].
func WithLogger(l *slog.Logger) Option { return func(d *dbOptions) { d.logger = l } }

// New will wrap an [sql.DB] and return a type that implements [DB]. Use this
// function if you want fancy features like configuration and logging but if you
// don't need those features then use [Simple].
func New(pool *sql.DB, opts ...Option) *database {
	options := dbOptions{}
	for _, o := range opts {
		o(&options)
	}
	if options.logger == nil {
		// TODO Create a silent log handler.
		options.logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	d := &database{
		DB:     pool,
		logger: options.logger,
	}
	return d
}

type database struct {
	*sql.DB
	logger *slog.Logger
}

func (db *database) QueryContext(ctx context.Context, query string, v ...any) (Rows, error) {
	rows, err := db.DB.QueryContext(ctx, query, v...)
	if err != nil {
		db.logger.Debug(query, slog.Any("error", err))
	}
	return rows, err
}

func (db *database) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	t, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &tx{Tx: t}, nil
}

// Simple creates a bare bones simple wrapper around a [sql.DB] that implements
// [DB]. If you want fancy features like logging then use [New] with its
// options.
func Simple(db *sql.DB) *simple { return &simple{db} }

type simple struct{ *sql.DB }

func (db *simple) QueryContext(ctx context.Context, query string, v ...any) (Rows, error) {
	return db.DB.QueryContext(ctx, query, v...)
}

func (db *simple) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	t, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &tx{Tx: t}, nil
}

type waitOpts struct {
	interval time.Duration
	timeout  time.Duration
	logger   *slog.Logger
}

// WaitOpt respresents an option type for the [WaitFor] function.
type WaitOpt func(*waitOpts)

// WithInterval sets the interval to wait after each time [db.Ping] is called when using [WaitFor].
func WithInterval(d time.Duration) WaitOpt {
	return func(wo *waitOpts) { wo.interval = d }
}

// WithTimeout sets the timeout to use when calling [WaitFor].
func WithTimeout(d time.Duration) WaitOpt {
	return func(wo *waitOpts) { wo.timeout = d }
}

// WithWaitLogger sets the logger to use when calling [WaitFor].
func WithWaitLogger(l *slog.Logger) WaitOpt { return func(wo *waitOpts) { wo.logger = l } }

var now = time.Now

// WaitFor will block until the database is up and can be connected to.
func WaitFor(ctx context.Context, database Pingable, opts ...WaitOpt) (err error) {
	wo := waitOpts{
		interval: time.Second * 2,
		logger:   slog.Default(),
	}
	for _, o := range opts {
		o(&wo)
	}

	var cancel context.CancelFunc = func() {}
	if wo.timeout > 0 {
		ctx, cancel = context.WithDeadline(ctx, now().Add(wo.timeout))
	}
	defer cancel()

	// Don't wait to send the first ping.
	if err = database.PingContext(ctx); err == nil {
		return nil
	}

	ticker := time.NewTicker(wo.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			err = database.Ping()
			if err == nil {
				wo.logger.Info("database connected")
				return nil
			}
			wo.logger.Warn("failed to ping database, retrying...", slog.Any("error", err))
		case <-ctx.Done():
			return errors.Wrap(ErrDBTimeout, "could not reach database")
		}
	}
}

type noopLogHandler struct{}

func (nh *noopLogHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (nh *noopLogHandler) Handle(context.Context, slog.Record) error { return nil }
func (nh *noopLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler  { return nh }
func (nh *noopLogHandler) WithGroup(name string) slog.Handler        { return nh }
