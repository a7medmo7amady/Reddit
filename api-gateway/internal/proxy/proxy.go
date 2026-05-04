package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"api-gateway/pkg/logger"
)

func NewSingle(target string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		logger.Fatalf("invalid proxy target %q: %v", target, err)
	}
	p := httputil.NewSingleHostReverseProxy(u)
	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Errorf("proxy error for %s %s → %s: %v", r.Method, r.URL.Path, target, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"service unavailable"}`))
	}
	return p
}
