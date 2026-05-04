package middleware

import (
	"net/http"
	"strings"

	jwtpkg "api-gateway/pkg/jwt"
)


func GRPCAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			grpcDeny(w, "missing or malformed token")
			return
		}

		claims, err := jwtpkg.Verify(strings.TrimPrefix(header, "Bearer "), jwtSecret)
		if err != nil {
			grpcDeny(w, "invalid or expired token")
			return
		}

		r.Header.Set("X-User-Id", claims.UserID)
		r.Header.Set("X-Role", claims.Role)
		next.ServeHTTP(w, r)
	})
}

func grpcDeny(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/grpc")
	w.Header().Set("Grpc-Status", "16") // UNAUTHENTICATED
	w.Header().Set("Grpc-Message", msg)
	w.WriteHeader(http.StatusOK) // gRPC always uses HTTP 200
}
