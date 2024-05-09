package limiter

import (
	"errors"
	"testing"
	"time"
)

func double(d int) int {
	return d * 2
}

var (
	errExpectErr    error = errors.New("got nil rate limit error, but wanted")
	errNotExpectErr error = errors.New("got rate limit error, but didnt want")
)

func TestDoLimitExceed(t *testing.T) {
	t.Parallel()

	defDriver := DefaultCounterStorage(3, 1*time.Second)
	l := NewLimiter(defDriver)

	var err error
	for i := range 4 {
		err = l.Do(nil, func() error {
			_ = double(i)
			return nil
		})
	}
	if err == nil {
		t.Error(errExpectErr)
	}
}

func TestDoNoLimitExceed(t *testing.T) {
	t.Parallel()

	defDriver := DefaultCounterStorage(3, 1*time.Second)
	l := NewLimiter(defDriver)

	var err error
	for i := range 4 {
		err = l.Do(nil, func() error {
			_ = double(i)
			return nil
		})
		time.Sleep(350 * time.Millisecond)
	}
	if err != nil {
		t.Error(errNotExpectErr)
	}
}

func TestComplexDo(t *testing.T) {
	t.Parallel()

	defDriver := DefaultCounterStorage(2, 100*time.Millisecond)
	l := NewLimiter(defDriver)

	var err error
	for i := range 3 {
		err = l.Do(nil, func() error {
			_ = double(i)
			return nil
		})
	}
	if err == nil {
		t.Error(errExpectErr)
	}

	time.Sleep(100 * time.Millisecond)

	err = l.Do(nil, func() error {
		_ = double(1)
		return nil
	})

	if err != nil {
		t.Error(errNotExpectErr)
	}
}
