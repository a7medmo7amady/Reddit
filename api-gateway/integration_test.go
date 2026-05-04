package main_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-gateway/internal/middleware"
	"api-gateway/internal/proxy"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	jwtpkg "api-gateway/pkg/jwt"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const integrationSecret = "integration-secret"

func buildGateway(userServiceURL, feedServiceURL string) *httptest.Server {
	middleware.SetJWTSecret(integrationSecret)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Logging())

	r.Any("/auth/*path", gin.WrapH(proxy.NewSingle(userServiceURL)))

	protected := r.Group("/")
	protected.Use(middleware.Auth())
	{
		protected.Any("/users/*path", gin.WrapH(proxy.NewSingle(userServiceURL)))
		protected.Any("/posts/*path", gin.WrapH(proxy.NewSingle(feedServiceURL)))
	}

	// Use a real HTTP server so ReverseProxy's CloseNotifier cast works.
	return httptest.NewServer(r)
}

func validToken() string {
	claims := jwtpkg.Claims{
		UserID: "user-1",
		Role:   "member",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(integrationSecret))
	return token
}

func fakeService(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func get(t *testing.T, url, token string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestIntegration_PublicRoute_NoToken(t *testing.T) {
	userSvc := fakeService(t, http.StatusOK, `{"token":"jwt"}`)
	gw := buildGateway(userSvc.URL, "")
	t.Cleanup(gw.Close)

	resp := get(t, gw.URL+"/auth/login", "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("public /auth route: want 200, got %d", resp.StatusCode)
	}
}

func TestIntegration_ProtectedRoute_NoToken(t *testing.T) {
	gw := buildGateway("", "")
	t.Cleanup(gw.Close)

	resp := get(t, gw.URL+"/posts/feed", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("protected route without token: want 401, got %d", resp.StatusCode)
	}
}

func TestIntegration_ProtectedRoute_ValidToken(t *testing.T) {
	feedSvc := fakeService(t, http.StatusOK, `{"posts":[]}`)
	gw := buildGateway("", feedSvc.URL)
	t.Cleanup(gw.Close)

	resp := get(t, gw.URL+"/posts/feed", validToken())
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("protected route with valid token: want 200, got %d — body: %s", resp.StatusCode, body)
	}
}

func TestIntegration_HeadersInjected(t *testing.T) {
	var gotUserID, gotRole string
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = r.Header.Get("X-User-Id")
		gotRole = r.Header.Get("X-Role")
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)

	gw := buildGateway("", backend.URL)
	t.Cleanup(gw.Close)

	get(t, gw.URL+"/posts/1", validToken())

	if gotUserID != "user-1" {
		t.Errorf("X-User-Id: want user-1, got %q", gotUserID)
	}
	if gotRole != "member" {
		t.Errorf("X-Role: want member, got %q", gotRole)
	}
}

func TestIntegration_CORS_Preflight(t *testing.T) {
	gw := buildGateway("", "")
	t.Cleanup(gw.Close)

	req, _ := http.NewRequest(http.MethodOptions, gw.URL+"/posts/feed", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("CORS preflight: want 204, got %d", resp.StatusCode)
	}
}
