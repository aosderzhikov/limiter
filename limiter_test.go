package limiter

import (
	"testing"
	"time"
)

func double(d int) int {
	return d * 2
}

func TestDoLimitExceed(t *testing.T) {
	inmem := NewInmemoryStorage(3, 3*time.Second)
	l := NewLimiter(inmem)

	var err error
	for i := range 4 {
		err = l.Do(nil, func() {
			_ = double(i)
		})
	}
	if err == nil {
		t.Error("got nil rate limit error, but wanted")
	}
}

func TestDoNotLimitExceed(t *testing.T) {
	inmem := NewInmemoryStorage(3, 1*time.Second)
	l := NewLimiter(inmem)

	var err error
	for i := range 4 {
		err = l.Do(nil, func() {
			_ = double(i)
		})
		time.Sleep(300 * time.Millisecond)
	}
	if err != nil {
		t.Error("got rate limit error, but didnt want")
	}
}
