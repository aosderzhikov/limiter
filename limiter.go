package limiter

import (
	"context"
	"errors"
	"sync"
	"time"
)

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

func NewInmemoryStorage(quota int, interval time.Duration) *InmemoryStorage {
	return &InmemoryStorage{
		mu:             sync.Mutex{},
		operationQuota: quota,
		operationDone:  0,
		interval:       interval,
	}
}

type InmemoryStorage struct {
	operationQuota int
	interval       time.Duration

	mu            sync.Mutex
	operationDone int
}

var ErrLimitExceed error = errors.New("operation limit exceeded")

func (l *Limiter) Do(ctx context.Context, f func()) error {
	ok, err := l.storage.Allowed(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return ErrLimitExceed
	}

	err = l.storage.Increment(ctx)
	if err != nil {
		return err
	}

	f()
	return nil
}

func (i *InmemoryStorage) Increment(_ context.Context) error {
	i.mu.Lock()
	if i.operationDone == 0 {
		go i.startInterval()
	}

	i.operationDone++
	i.mu.Unlock()
	return nil
}

func (i *InmemoryStorage) Refresh(_ context.Context) error {
	i.mu.Lock()
	i.operationDone = 0
	i.mu.Unlock()
	return nil
}

func (i *InmemoryStorage) Allowed(_ context.Context) (bool, error) {
	i.mu.Lock()
	allowed := i.operationDone < i.operationQuota
	i.mu.Unlock()
	return allowed, nil
}

func (i *InmemoryStorage) startInterval() {
	time.Sleep(i.interval)
	_ = i.Refresh(context.TODO())
}
