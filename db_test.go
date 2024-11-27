package db

import (
	"database/sql"
	"io"
	"log/slog"
	"testing"

	"github.com/matryer/is"
	"github.com/pkg/errors"
	"go.uber.org/mock/gomock"

	"github.com/harrybrwn/db/mockrows"
)

func TestScanOne(t *testing.T) {
	var errTestError = errors.New("test error")
	run := func(name string, fn func(t *testing.T, r *mockrows.MockRows)) {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			r := mockrows.NewMockRows(ctrl)
			fn(t, r)
		})
	}

	run("happy path", func(t *testing.T, r *mockrows.MockRows) {
		is := is.New(t)
		r.EXPECT().Next().Return(true)
		r.EXPECT().Scan().Return(nil)
		r.EXPECT().Close().Return(errTestError)
		err := ScanOne(r)
		is.True(errors.Is(err, errTestError))
	})

	run("scan error", func(t *testing.T, r *mockrows.MockRows) {
		is := is.New(t)
		r.EXPECT().Next().Return(true)
		r.EXPECT().Scan().Return(errTestError)
		r.EXPECT().Close().Return(nil)
		err := ScanOne(r)
		is.True(errors.Is(err, errTestError))
	})

	run("scan error close error", func(t *testing.T, r *mockrows.MockRows) {
		is := is.New(t)
		r.EXPECT().Next().Return(true)
		r.EXPECT().Scan().Return(errTestError)
		r.EXPECT().Close().Return(errTestError)
		err := ScanOne(r)
		is.True(errors.Is(err, errTestError))
	})

	run("no next no rows", func(t *testing.T, r *mockrows.MockRows) {
		is := is.New(t)
		r.EXPECT().Next().Return(false)
		r.EXPECT().Err().Return(nil)
		r.EXPECT().Close().Return(nil)
		err := ScanOne(r)
		is.True(errors.Is(err, sql.ErrNoRows))
	})

	run("no next no rows error", func(t *testing.T, r *mockrows.MockRows) {
		is := is.New(t)
		r.EXPECT().Next().Return(false)
		r.EXPECT().Err().Return(nil)
		r.EXPECT().Close().Return(errTestError)
		err := ScanOne(r)
		is.True(errors.Is(err, sql.ErrNoRows))
	})

	run("no next with Err", func(t *testing.T, r *mockrows.MockRows) {
		is := is.New(t)
		r.EXPECT().Next().Return(false)
		r.EXPECT().Err().Return(errTestError)
		r.EXPECT().Close().Return(nil)
		err := ScanOne(r)
		is.True(errors.Is(err, errTestError))
	})

	run("no next with both Err", func(t *testing.T, r *mockrows.MockRows) {
		is := is.New(t)
		r.EXPECT().Next().Return(false)
		r.EXPECT().Err().Return(errTestError)
		r.EXPECT().Close().Return(errTestError)
		err := ScanOne(r)
		is.True(errors.Is(err, errTestError))
	})
}

func TestNew(t *testing.T) {
	is := is.New(t)
	l := slog.New(slog.NewTextHandler(io.Discard, nil))
	db := New(nil, WithLogger(l))
	is.True(db != nil)
	is.Equal(db.logger, l)
}

// TODO Uncomment when WaitFor is added.

/*
// swap out the function that gets the current time
func withNow(tm time.Time) func() {
	now = func() time.Time {
		return tm
	}
	return func() { now = time.Now }
}

func TestWaitFor(t *testing.T) {
	ctx := context.Background()
	TimeNow := time.Unix(1731461240, 0)
	defer withNow(TimeNow)()

	run := func(name string, fn func(t *testing.T, ping *mockrows.MockPingable)) {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			ping := mockrows.NewMockPingable(ctrl)
			fn(t, ping)
		})
	}

	run("successful first ping", func(t *testing.T, ping *mockrows.MockPingable) {
		is := is.New(t)
		ping.EXPECT().
			PingContext(gomock.All(gomock.AssignableToTypeOf(ctx), gomock.Not(gomock.Nil()))).
			Return(nil)
		err := WaitFor(ctx, ping)
		is.NoErr(err)
	})

	run("successful first ping timeout", func(t *testing.T, ping *mockrows.MockPingable) {
		is := is.New(t)
		timeout := time.Second
		timeoutCtx, cancel := context.WithDeadline(ctx, now().Add(timeout))
		defer cancel() // just to cleanup potential mem leaks
		ping.EXPECT().
			PingContext(gomock.All(
				gomock.AssignableToTypeOf(timeoutCtx),
				gomock.Not(gomock.Nil()),
				gomock.Eq(timeoutCtx),
			)).
			Return(nil)
		err := WaitFor(ctx, ping, WithTimeout(timeout))
		is.NoErr(err)
	})

	run("failed first n ping", func(t *testing.T, ping *mockrows.MockPingable) {
		is := is.New(t)
		ctxMatcher := gomock.All(
			gomock.AssignableToTypeOf(ctx),
			gomock.Not(gomock.Nil()),
		)
		ping.EXPECT().
			PingContext(ctxMatcher).
			Return(errors.New("throw away error 1"))
		ping.EXPECT().Ping().Return(errors.New("throw away error 2"))
		ping.EXPECT().Ping().Return(errors.New("throw away error 3"))
		ping.EXPECT().Ping().Return(nil)
		inter := time.Millisecond * 10
		start := time.Now()
		err := WaitFor(ctx, ping, WithInterval(inter), WithWaitLogger(log.SilentLogger()))
		is.NoErr(err)
		isWithinMargin(t, time.Since(start), inter*3, time.Millisecond*2)
	})
}

func isWithinMargin(t *testing.T, val, span, margin time.Duration) {
	t.Helper()
	between := val > span-margin && val < span+margin
	if !between {
		t.Errorf(
			"%v is not between %v and %v with a %v margin of error",
			val, span-margin, span+margin, margin,
		)
	}
}

func TestWaitFor_Functional(t *testing.T) {
	t.Skip()
	os.Unsetenv("PGSERVICEFILE")
	db, err := sql.Open("postgres", "host=localhost user=root password=testlab dbname=idp sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	ctx = log.StashInContext(ctx, log.GetLogger())
	err = WaitFor(
		ctx,
		db,
		WithTimeout(time.Millisecond*500),
		WithInterval(time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
}
*/
