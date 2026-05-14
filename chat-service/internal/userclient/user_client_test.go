package userclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestUserExistsEscapesAndTrimsID(t *testing.T) {
	var gotPath string

	client := New("http://user-service/")
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotPath = r.URL.EscapedPath()
			return response(http.StatusOK, `{"exists":true}`), nil
		}),
	}

	exists, err := client.UserExists(context.Background(), " 12/34 ")
	if err != nil {
		t.Fatalf("UserExists returned error: %v", err)
	}
	if !exists {
		t.Fatal("UserExists = false, want true")
	}
	if gotPath != "/internal/users/12%2F34/exists" {
		t.Fatalf("path = %q, want escaped user id path", gotPath)
	}
}

func TestUserExistsNotFound(t *testing.T) {
	client := New("http://user-service")
	client.httpClient = &http.Client{
		Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
			return response(http.StatusNotFound, ""), nil
		}),
	}

	exists, err := client.UserExists(context.Background(), "42")
	if err != nil {
		t.Fatalf("UserExists returned error: %v", err)
	}
	if exists {
		t.Fatal("UserExists = true, want false")
	}
}

func response(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
