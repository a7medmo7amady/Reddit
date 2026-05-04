package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"api-gateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

func corsRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.CORS())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func TestCORS_HeadersPresent(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	corsRouter().ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("ACAO: want *, got %q", got)
	}
}

func TestCORS_Preflight(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	corsRouter().ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204 for preflight, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Access-Control-Allow-Methods header missing on preflight")
	}
}
