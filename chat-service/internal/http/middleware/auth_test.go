package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuth_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Auth())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_EmptyHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Auth())
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-Id", "")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ValidHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var gotUserID, gotRole string

	r := gin.New()
	r.Use(Auth())
	r.GET("/test", func(c *gin.Context) {
		gotUserID = c.GetString("userID")
		gotRole = c.GetString("userRole")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-User-Id", "user123")
	req.Header.Set("X-Role", "MODERATOR")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if gotUserID != "user123" {
		t.Errorf("userID = %q, want %q", gotUserID, "user123")
	}
	if gotRole != "MODERATOR" {
		t.Errorf("userRole = %q, want %q", gotRole, "MODERATOR")
	}
}

func TestAuth_AbortsPipeline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false

	r := gin.New()
	r.Use(Auth())
	r.GET("/test", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if called {
		t.Error("handler was called despite missing auth")
	}
}
