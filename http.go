package limiter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func NewRESTStorage(addr string, quota int, httpClient *http.Client) *RESTStorage {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &RESTStorage{
		addr:           addr,
		operationQuota: quota,
		client:         httpClient,
	}
}

type RESTStorage struct {
	addr           string
	operationQuota int
	client         *http.Client
}

var ErrCounterGet = errors.New("cannot get counter from storage")

func (r *RESTStorage) Increment(ctx context.Context) error {
	path := r.addr + "counter"
	counter, err := r.getCounter(ctx)
	if err != nil {
		return errors.Join(ErrCounterGet, err)
	}

	if counter == 0 {
		go func() {
			err = r.Refresh(ctx)
			if err != nil {
				fmt.Printf("counter refreshing failed: %v\n", err)
			}
		}()
	}

	_, err = r.doRequest(ctx, path, http.MethodPost)
	if err != nil {
		return fmt.Errorf("counter incrementing failed: %v", err)
	}

	return nil
}

func (r *RESTStorage) Allowed(ctx context.Context) (bool, error) {
	counter, err := r.getCounter(ctx)
	if err != nil {
		return false, errors.Join(ErrCounterGet, err)
	}

	if counter >= r.operationQuota {
		return false, nil
	}

	return true, nil
}

func (r *RESTStorage) Refresh(ctx context.Context) error {
	path := r.addr + "refresh"
	_, err := r.doRequest(ctx, path, http.MethodPost)
	if err != nil {
		return err
	}

	return nil
}

func (r *RESTStorage) getCounter(ctx context.Context) (int, error) {
	path := r.addr + "counter"

	b, err := r.doRequest(ctx, path, http.MethodGet)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(b))
}

func (r *RESTStorage) doRequest(ctx context.Context, path string, method string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, path, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("non 2xx status code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
