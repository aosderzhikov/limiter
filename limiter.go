package limiter

import (
	"context"
	"errors"
	"time"

	"github.com/aosderzhikov/limiter/driver"
)

func DefaultCounterStorage(quota int, interval time.Duration) CounterStorage {
	return driver.NewInmemoryStorage(quota, interval)
}

func NewLimiter(storage CounterStorage) *Limiter {
	return &Limiter{
		storage: storage,
	}
}

type Limiter struct {
	storage CounterStorage
}

type CounterStorage interface {
	Increment(ctx context.Context) error
	Allowed(ctx context.Context) (bool, error)
	Refresh(ctx context.Context) error
}

var ErrLimitExceed error = errors.New("operation limit exceeded")

func (l *Limiter) Do(ctx context.Context, f func() error) error {
	ok, err := l.storage.Allowed(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLimitExceed
	}

	err = f()
	if err != nil {
		return err
	}

	err = l.storage.Increment(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (l *Limiter) Allowed(ctx context.Context) (bool, error) {
	return l.storage.Allowed(ctx)
}
