package service

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ─────────────────── transientDBError classification ───────────────────

func TestTransientDBError_NilIsNotTransient(t *testing.T) {
	assert.False(t, transientDBError(nil))
}

func TestTransientDBError_DriverErrBadConn(t *testing.T) {
	assert.True(t, transientDBError(driver.ErrBadConn))
	// Wrapped form must also be detected.
	wrapped := fmt.Errorf("upsert: %w", driver.ErrBadConn)
	assert.True(t, transientDBError(wrapped))
}

func TestTransientDBError_KnownStringMarkers(t *testing.T) {
	cases := []string{
		"driver: bad connection",
		"read: connection reset by peer",
		"dial tcp: connect: connection refused",
		"write tcp: broken pipe",
		"unexpected EOF",
		"net/http: i/o timeout",
		"server closed the connection unexpectedly",
	}
	for _, msg := range cases {
		assert.True(t, transientDBError(errors.New(msg)), "should be transient: %q", msg)
	}
}

func TestTransientDBError_DeadlineAndCancelAreNotTransient(t *testing.T) {
	// Hard deadlines should NOT trigger an automatic retry — they're a signal
	// the operation is taking too long, not a network blip.
	assert.False(t, transientDBError(context.DeadlineExceeded))
	assert.False(t, transientDBError(context.Canceled))
}

func TestTransientDBError_OtherErrorsArePermanent(t *testing.T) {
	cases := []error{
		errors.New("ERROR: column does not exist"),
		errors.New("syntax error at or near"),
		errors.New("duplicate key value violates unique constraint"),
		errors.New("permission denied"),
	}
	for _, err := range cases {
		assert.False(t, transientDBError(err), "should NOT be transient: %v", err)
	}
}

// ─────────────────── retryTransient behaviour ───────────────────

func TestRetryTransient_SucceedsFirstTry(t *testing.T) {
	calls := 0
	err := retryTransient(context.Background(), "test", 3, func() error {
		calls++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, calls)
}

func TestRetryTransient_RetriesOnTransientErrorThenSucceeds(t *testing.T) {
	calls := 0
	err := retryTransient(context.Background(), "test", 4, func() error {
		calls++
		if calls < 3 {
			return driver.ErrBadConn
		}
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 3, calls, "should have retried twice before succeeding on the 3rd attempt")
}

func TestRetryTransient_GivesUpAfterMaxAttempts(t *testing.T) {
	calls := 0
	err := retryTransient(context.Background(), "test", 3, func() error {
		calls++
		return errors.New("bad connection")
	})
	assert.Error(t, err)
	assert.Equal(t, 3, calls)
	assert.Contains(t, err.Error(), "bad connection", "final error from last attempt is returned")
}

func TestRetryTransient_PermanentErrorDoesNotRetry(t *testing.T) {
	calls := 0
	permanent := errors.New("ERROR: column does not exist")
	err := retryTransient(context.Background(), "test", 5, func() error {
		calls++
		return permanent
	})
	assert.ErrorIs(t, err, permanent)
	assert.Equal(t, 1, calls, "permanent errors must NOT trigger retry")
}

func TestRetryTransient_HonoursContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	go func() {
		// Cancel before the first backoff completes.
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	err := retryTransient(ctx, "test", 5, func() error {
		calls++
		return driver.ErrBadConn
	})
	assert.ErrorIs(t, err, context.Canceled)
	assert.LessOrEqual(t, calls, 2, "should stop retrying once ctx is cancelled")
}
