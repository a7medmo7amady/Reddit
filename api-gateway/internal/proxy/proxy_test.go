package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api-gateway/internal/proxy"
)

// fakeBackend spins up a real HTTP server and returns its URL.
func fakeBackend(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestProxy_ForwardsRequest(t *testing.T) {
	backend := fakeBackend(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	p := proxy.NewSingle(backend.URL)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/123", nil)

	p.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if body := w.Body.String(); body != `{"ok":true}` {
		t.Errorf("body: want {\"ok\":true}, got %s", body)
	}
}

func TestProxy_ForwardsHeaders(t *testing.T) {
	var capturedUserID string
	backend := fakeBackend(t, func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = r.Header.Get("X-User-Id")
		w.WriteHeader(http.StatusOK)
	})

	p := proxy.NewSingle(backend.URL)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)
	req.Header.Set("X-User-Id", "user-99")

	p.ServeHTTP(w, req)

	if capturedUserID != "user-99" {
		t.Errorf("X-User-Id: want user-99, got %q", capturedUserID)
	}
}

func TestProxy_BackendDown_Returns502(t *testing.T) {
	// Point at a port nothing is listening on
	p := proxy.NewSingle("http://127.0.0.1:19999")
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/posts/1", nil)

	p.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("want 502, got %d", w.Code)
	}
}
