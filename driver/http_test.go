package driver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHandler struct {
	counter int
}

func (m *mockHandler) incrementCounertHandle(w http.ResponseWriter, r *http.Request) {
	m.counter++
}

func (m *mockHandler) getCouneterHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(fmt.Sprintf("%d", m.counter)))
}

func (m *mockHandler) refreshCouneterHandle(w http.ResponseWriter, r *http.Request) {
	m.counter = 0
}

func registerHandler(mux *http.ServeMux, handler *mockHandler) {
	mux.HandleFunc("GET /counter", handler.getCouneterHandle)
	mux.HandleFunc("POST /counter", handler.incrementCounertHandle)
	mux.HandleFunc("POST /refresh", handler.refreshCouneterHandle)
}

func TestIncrement(t *testing.T) {
	handler := &mockHandler{0}
	mux := http.NewServeMux()
	registerHandler(mux, handler)

	srv := httptest.NewServer(mux)

	driver := NewHTTPStorage(ConfigHTTPStorage{
		Addr:       srv.URL,
		HTTPClient: &http.Client{},
	})

	ctx := context.Background()
	err := driver.Increment(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if handler.counter != 1 {
		t.Error("counter didnt increase")
	}
}

func TestAllowed(t *testing.T) {
	handler := &mockHandler{4}
	mux := http.NewServeMux()
	registerHandler(mux, handler)

	srv := httptest.NewServer(mux)

	driver := NewHTTPStorage(ConfigHTTPStorage{
		Addr:       srv.URL,
		Quota:      3,
		HTTPClient: &http.Client{},
	})

	ctx := context.Background()
	allowed, err := driver.Allowed(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("allowed but shoudn't")
	}
}

func TestRefresh(t *testing.T) {
	handler := &mockHandler{4}
	mux := http.NewServeMux()
	registerHandler(mux, handler)

	srv := httptest.NewServer(mux)

	driver := NewHTTPStorage(ConfigHTTPStorage{
		Addr:       srv.URL,
		HTTPClient: &http.Client{},
	})

	ctx := context.Background()
	err := driver.Refresh(ctx)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if handler.counter != 0 {
		t.Error("didnt refresh counter")
	}
}
