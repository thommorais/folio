package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogin(t *testing.T) {
	t.Run("exchanges credentials for a token", func(t *testing.T) {
		var gotPath, gotMethod string
		var gotBody map[string]string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			_, _ = w.Write([]byte(`{"token":"tok123","record":{"id":"u1","email":"a@b.c","name":"Demo"}}`))
		}))
		defer server.Close()

		// No token yet: login is the one call that must work without one.
		session, err := New(server.URL, "").Login("a@b.c", "secret")
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}

		if gotMethod != http.MethodPost {
			t.Errorf("method = %q, want POST", gotMethod)
		}
		if gotPath != "/api/collections/users/auth-with-password" {
			t.Errorf("path = %q", gotPath)
		}
		if gotBody["identity"] != "a@b.c" || gotBody["password"] != "secret" {
			t.Errorf("body = %v", gotBody)
		}
		if session.Token != "tok123" {
			t.Errorf("token = %q, want tok123", session.Token)
		}
		if session.Email != "a@b.c" {
			t.Errorf("email = %q, want a@b.c", session.Email)
		}
	})

	t.Run("reports bad credentials", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"Failed to authenticate."}`))
		}))
		defer server.Close()

		_, err := New(server.URL, "").Login("a@b.c", "wrong")
		if err == nil {
			t.Fatal("Login() error = nil, want an error")
		}
		if got := err.Error(); got != "400: Failed to authenticate." {
			t.Errorf("err = %q", got)
		}
	})
}
