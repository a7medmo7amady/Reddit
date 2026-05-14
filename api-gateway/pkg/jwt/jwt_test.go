package jwt_test

import (
	"testing"
	"time"

	jwtpkg "api-gateway/pkg/jwt"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "test-secret"

func makeToken(userID, role string, expiry time.Duration) string {
	claims := jwtpkg.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token
}

func TestVerify_ValidToken(t *testing.T) {
	token := makeToken("user-123", "member", time.Hour)
	claims, err := jwtpkg.Verify(token, secret)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if claims.UserID != "user-123" {
		t.Errorf("UserID: want user-123, got %s", claims.UserID)
	}
	if claims.Role != "member" {
		t.Errorf("Role: want member, got %s", claims.Role)
	}
}

func TestVerify_UsesSubjectAsUserID(t *testing.T) {
	claims := jwt.MapClaims{
		"sub":  "42",
		"role": "USER",
		"exp":  time.Now().Add(time.Hour).Unix(),
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))

	verified, err := jwtpkg.Verify(token, secret)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if verified.UserID != "42" {
		t.Errorf("UserID: want 42, got %s", verified.UserID)
	}
}

func TestVerify_ExpiredToken(t *testing.T) {
	token := makeToken("user-123", "member", -time.Minute)
	_, err := jwtpkg.Verify(token, secret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestVerify_WrongSecret(t *testing.T) {
	token := makeToken("user-123", "member", time.Hour)
	_, err := jwtpkg.Verify(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestVerify_MalformedToken(t *testing.T) {
	_, err := jwtpkg.Verify("not.a.token", secret)
	if err == nil {
		t.Fatal("expected error for malformed token, got nil")
	}
}

func TestVerify_WrongSigningMethod(t *testing.T) {
	// Craft a token whose header claims RS256 but with an HS256 signature —
	// the gateway must reject any non-HMAC algorithm regardless of signature.
	// header: {"alg":"RS256","typ":"JWT"}, payload: {"sub":"x"}
	fakeRS256 := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9" +
		".eyJzdWIiOiJ4In0" +
		".invalidsignature"
	_, err := jwtpkg.Verify(fakeRS256, secret)
	if err == nil {
		t.Fatal("expected error for RS256-alg token, got nil")
	}
}
