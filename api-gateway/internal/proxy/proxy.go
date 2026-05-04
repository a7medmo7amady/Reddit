package proxy

import (
	"net"
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

	defaultDirector := p.Director
	p.Director = func(r *http.Request) {
		defaultDirector(r)
		r.Host = u.Host
		if clientIP := r.RemoteAddr; clientIP != "" {
			if ip, _, err := net.SplitHostPort(clientIP); err == nil {
				prior := r.Header.Get("X-Forwarded-For")
				if prior == "" {
					r.Header.Set("X-Forwarded-For", ip)
				} else {
					r.Header.Set("X-Forwarded-For", prior+", "+ip)
				}
			}
		}
	}

	p.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Errorf("proxy error for %s %s → %s: %v", r.Method, r.URL.Path, target, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"service unavailable"}`))
	}
	return p
}
