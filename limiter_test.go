package limiter

import (
	"testing"
	"time"
)

func double(d int) int {
	return d * 2
}

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
		t.Error("got nil rate limit error, but wanted")
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
		t.Error("got rate limit error, but didnt want")
	}
}
