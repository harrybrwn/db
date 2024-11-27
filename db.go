package db

import (
	"context"
	"database/sql"
	"io"
	"log/slog"

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

type DB interface {
	io.Closer
	QueryContext(context.Context, string, ...any) (Rows, error)
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type Pingable interface {
	Ping() error
	PingContext(context.Context) error
}

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

type Rows interface {
	Scanner
	io.Closer
	Next() bool
	Err() error
}

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

func WithLogger(l *slog.Logger) Option { return func(d *dbOptions) { d.logger = l } }

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

// TODO Add standard logging that will be compatible with either slog or logrus
//      before adding WaitFor.

/*
type waitOpts struct {
	interval time.Duration
	timeout  time.Duration
	logger   log.FieldLogger
}

type WaitOpt func(*waitOpts)

func WithInterval(d time.Duration) WaitOpt {
	return func(wo *waitOpts) { wo.interval = d }
}

func WithTimeout(d time.Duration) WaitOpt {
	return func(wo *waitOpts) { wo.timeout = d }
}

func WithWaitLogger(l log.FieldLogger) WaitOpt { return func(wo *waitOpts) { wo.logger = l } }

var now = time.Now

// WaitFor will block until the database is up and can be connected to.
func WaitFor(ctx context.Context, database Pingable, opts ...WaitOpt) (err error) {
	wo := waitOpts{
		interval: time.Second * 2,
		logger:   log.FromContext(ctx),
	}
	for _, o := range opts {
		o(&wo)
	}
	if wo.logger == nil {
		wo.logger = log.SilentLogger()
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
			wo.logger.WithError(err).Warn("failed to ping database, retrying...")
		case <-ctx.Done():
			return errors.Wrap(ErrDBTimeout, "could not reach database")
		}
	}
}
*/
