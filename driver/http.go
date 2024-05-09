package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ConfigHTTPStorage struct {
	Addr       string
	Quota      int
	Interval   time.Duration
	HTTPClient *http.Client
}

func NewHTTPStorage(cfg ConfigHTTPStorage) *HTTPStorage {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{}
	}

	if !strings.HasSuffix(cfg.Addr, "/") {
		cfg.Addr += "/"
	}

	return &HTTPStorage{
		addr:           cfg.Addr,
		operationQuota: cfg.Quota,
		client:         cfg.HTTPClient,
	}
}

type HTTPStorage struct {
	addr           string
	operationQuota int
	interval       time.Duration
	client         *http.Client
}

var ErrCounterGet = errors.New("cannot get counter from storage")

const (
	counterEndpoint = "counter"
	refreshEndpoint = "refresh"
)

func (h *HTTPStorage) Increment(ctx context.Context) error {
	path := h.addr + counterEndpoint
	counter, err := h.getCounter(ctx)
	if err != nil {
		return errors.Join(ErrCounterGet, err)
	}

	if counter == 0 {
		go func() {
			err = h.Refresh(ctx)
			if err != nil {
				fmt.Printf("counter refreshing failed: %v\n", err)
			}
		}()
	}

	_, err = h.doRequest(ctx, path, http.MethodPost)
	if err != nil {
		return fmt.Errorf("counter incrementing failed: %v", err)
	}

	return nil
}

func (h *HTTPStorage) Allowed(ctx context.Context) (bool, error) {
	counter, err := h.getCounter(ctx)
	if err != nil {
		return false, errors.Join(ErrCounterGet, err)
	}

	if counter >= h.operationQuota {
		return false, nil
	}

	return true, nil
}

func (h *HTTPStorage) Refresh(ctx context.Context) error {
	time.Sleep(h.interval)
	path := h.addr + refreshEndpoint
	_, err := h.doRequest(ctx, path, http.MethodPost)
	if err != nil {
		return err
	}

	return nil
}

func (h *HTTPStorage) getCounter(ctx context.Context) (int, error) {
	path := h.addr + counterEndpoint

	b, err := h.doRequest(ctx, path, http.MethodGet)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(b))
}

func (h *HTTPStorage) doRequest(ctx context.Context, path string, method string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, path, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("non 2xx status code %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
