package proxy

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"api-gateway/pkg/logger"

	"golang.org/x/net/http2"
)

func NewGRPC(target string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		logger.Fatalf("invalid gRPC target %q: %v", target, err)
	}
	transport := &http2.Transport{
		AllowHTTP: true,
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, network, addr)
		},
	}

	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			r.URL.Host = u.Host
		},
		Transport:     transport,
		FlushInterval: -1,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			logger.Errorf("gRPC proxy error → %s: %v", target, err)

			w.Header().Set("Content-Type", "application/grpc")
			w.Header().Set("Grpc-Status", "14") 
			w.Header().Set("Grpc-Message", "service unavailable")
			w.WriteHeader(http.StatusOK)
		},
	}
}
